package replace

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/mfojtik/gomod-helpers/pkg/golang"
)

type moduleReplace struct {
	oldPath        string
	newPath        string
	newPathVersion string
}

type options struct {
	Branch string
	Commit string
	Tag    string

	Paths     []string
	Exclude   []string
	GoModPath string

	replaces []moduleReplace
}

func (opts *options) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&opts.Branch, "branch", "", "Specify branch to use for this bump")
	flags.StringVar(&opts.Tag, "tag", "", "Specify tag to use for this bump")
	flags.StringVar(&opts.Commit, "commit", "", "Specify commit to use for this bump")
	flags.StringVar(&opts.GoModPath, "gomod-file-path", "go.mod", "Specify the path to go.mod file")
	flags.StringArrayVar(&opts.Paths, "paths", []string{}, "Specify dependency path prefixes to update separated by comma (eg. 'github.com/openshift/api' or 'k8s.io/')")
	flags.StringArrayVar(&opts.Exclude, "exclude", []string{}, "Specify dependency path prefixes to exclude (eg. 'github.com/openshift/api' or 'k8s.io/')")
}

func (opts *options) matchRepository(r string) bool {
	for _, item := range opts.Paths {
		exclude := false
		for _, excludes := range opts.Exclude {
			if strings.HasPrefix(r, excludes) {
				exclude = true
				break
			}
		}
		if exclude {
			continue
		}
		if strings.HasPrefix(r, item) {
			return true
		}
	}
	return false
}

// parseModules will parse the existing go.mod file and filter out only modules matching the name prefixes specified with this command
func (opts *options) parseModules() error {
	modBytes, err := ioutil.ReadFile(opts.GoModPath)
	if err != nil {
		return err
	}
	s, err := golang.ParseModFile("go.mod", modBytes, nil)
	if err != nil {
		return err
	}
	for _, r := range s.Replace {
		if opts.matchRepository(r.New.Path) {
			opts.replaces = append(opts.replaces, moduleReplace{newPath: r.New.Path, oldPath: r.Old.Path})
		}
	}
	return nil
}

func pathToRepository(path string) string {
	if strings.HasPrefix(path, "k8s.io/") {
		return fmt.Sprintf("https://github.com/kubernetes/%s", strings.TrimPrefix(path, "k8s.io/"))
	}
	return fmt.Sprintf("https://%s", path)
}

func reportError(repo string, inErr error) {
	if _, err := fmt.Fprintf(os.Stderr, "# ERROR: %s: %v\n", repo, inErr); err != nil {
		panic(err)
	}
}

func (opts *options) Complete() error {
	if err := opts.parseModules(); err != nil {
		return err
	}

	if len(opts.replaces) == 0 {
		return fmt.Errorf("no modules found with given path prefixes: %#v", opts.Paths)
	}

	errChan := make(chan error)
	defer close(errChan)

	var wg sync.WaitGroup
	wg.Add(len(opts.replaces))

	for i := range opts.replaces {
		go func(index int) {
			replace := opts.replaces[index]

			defer wg.Done()
			repository, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{URL: pathToRepository(replace.newPath)})
			if err != nil {
				reportError(replace.newPath, fmt.Errorf("failed to clone git repository %s: %v", pathToRepository(replace.newPath), err))
				return
			}
			var (
				commit *object.Commit
			)

			if len(opts.Branch) > 0 {
				ref, err := repository.Storer.Reference(plumbing.NewBranchReferenceName(opts.Branch))
				if err != nil {
					reportError(replace.newPath, fmt.Errorf("unable to find branch %s: %v", opts.Branch, err))
					return
				}
				commit, err = repository.CommitObject(ref.Hash())
				if err != nil {
					reportError(replace.newPath, err)
					return
				}
			}

			if len(opts.Tag) > 0 {
				ref, err := repository.Storer.Reference(plumbing.NewTagReferenceName(opts.Tag))
				if err != nil {
					reportError(replace.newPath, fmt.Errorf("unable to find tag %s: %v", opts.Tag, err))
					return
				}
				obj, err := repository.TagObject(ref.Hash())
				if err != nil {
					reportError(replace.newPath, fmt.Errorf("object not found: %v", err))
				}
				// fake this, we only need hash and timestamp
				commit = &object.Commit{Hash: obj.Hash, Committer: obj.Tagger}
			}

			if len(opts.Commit) > 0 {
				// This seems to be unnecessary, but it is ok to check if the target commit really exists in the repo
				var err error
				commit, err = repository.CommitObject(plumbing.NewHash(opts.Commit))
				if err != nil {
					reportError(replace.newPath, fmt.Errorf("unable to find commit %s: %v", opts.Commit, err))
					return
				}
			}

			opts.replaces[index].newPathVersion = commitToGoModString(commit)
			if _, err := fmt.Fprintf(os.Stdout, "# Using %q for import path %q ...\n", opts.replaces[index].newPathVersion, replace.newPath); err != nil {
				reportError(replace.newPath, err)
				return
			}
		}(i)
	}

	wg.Wait()
	return nil
}

// commitToGoModString convert the Git commit to go.mod compatible version string that includes timestamp and the first 12 characters from commit hash.
func commitToGoModString(c *object.Commit) string {
	t := c.Committer.When
	timestamp := fmt.Sprintf("%d%d%d%d%d%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	return fmt.Sprintf("v0.0.0-%s-%s", timestamp, c.Hash.String()[0:12])
}

func (opts *options) Run() error {
	for _, replace := range opts.replaces {
		if _, err := fmt.Fprintf(os.Stdout, `go mod edit -replace %s=%s@"%s"`+"\n", replace.oldPath, replace.newPath, replace.newPathVersion); err != nil {
			return err
		}
	}
	return nil
}

func (opts *options) Validate() error {
	if len(opts.Branch) == 0 && len(opts.Commit) == 0 && len(opts.Tag) == 0 {
		return fmt.Errorf("either branch, commit or tag must be specified")
	}
	if len(opts.Paths) == 0 {
		return fmt.Errorf("dependency name must be specified")
	}
	return nil
}

func NewReplaceCommand() *cobra.Command {
	replaceOptions := &options{}
	cmd := &cobra.Command{
		Use:   "replace",
		Short: "Replace help to bulk update modules versions",
		Run: func(cmd *cobra.Command, args []string) {
			if err := replaceOptions.Validate(); err != nil {
				log.Fatal(err)
			}
			if err := replaceOptions.Complete(); err != nil {
				log.Fatal(err)
			}
			if err := replaceOptions.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	replaceOptions.AddFlags(cmd.Flags())

	return cmd
}
