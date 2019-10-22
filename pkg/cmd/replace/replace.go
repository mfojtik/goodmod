package replace

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/mfojtik/gomod-helpers/pkg/golang"
	"github.com/mfojtik/gomod-helpers/pkg/resolve"
	"github.com/mfojtik/gomod-helpers/pkg/resolve/branch"
	"github.com/mfojtik/gomod-helpers/pkg/resolve/commit"
	"github.com/mfojtik/gomod-helpers/pkg/resolve/tag"
	"github.com/mfojtik/gomod-helpers/pkg/resolve/types"
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
	Excludes  []string
	GoModPath string

	GithubClient *http.Client
	replaces     []moduleReplace
}

func (opts *options) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&opts.Branch, "branch", "", "Specify branch to use for this bump")
	flags.StringVar(&opts.Tag, "tag", "", "Specify tag to use for this bump")
	flags.StringVar(&opts.Commit, "commit", "", "Specify commit to use for this bump")
	flags.StringVar(&opts.GoModPath, "gomod-file-path", "go.mod", "Specify the path to go.mod file")
	flags.StringSliceVar(&opts.Paths, "paths", []string{}, "Specify dependency path prefixes to update separated by comma (eg. 'github.com/openshift/api' or 'k8s.io/')")
	flags.StringSliceVar(&opts.Excludes, "excludes", []string{}, "Specify dependency path prefixes to exclude (eg. 'github.com/openshift/api' or 'k8s.io/')")
}

func (opts *options) matchPath(r string) bool {
	for _, item := range opts.Paths {
		exclude := false
		for _, e := range opts.Excludes {
			if strings.HasPrefix(r, e) {
				exclude = true
			}
		}
		if !exclude && strings.HasPrefix(r, item) {
			return true
		}
	}
	return false
}

func (opts *options) hasReplacePath(path string) bool {
	for _, p := range opts.replaces {
		if p.newPath == path {
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
		if opts.matchPath(r.New.Path) {
			opts.replaces = append(opts.replaces, moduleReplace{newPath: r.New.Path, oldPath: r.Old.Path})
		}
	}
	for _, r := range s.Require {
		if !opts.hasReplacePath(r.Mod.Path) && opts.matchPath(r.Mod.Path) {
			opts.replaces = append(opts.replaces, moduleReplace{newPath: r.Mod.Path, oldPath: r.Mod.Path})
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

func (opts *options) resolveByTag(modulePath string) *types.Commit {
	resolvers := []resolve.ModulerResolver{
		tag.NewGithubTagResolver(nil),
		tag.NewGitTagResolver(),
	}
	for _, r := range resolvers {
		c, err := r.Resolve(context.TODO(), modulePath, opts.Tag)
		if err != nil {
			reportError(modulePath, fmt.Errorf("failed to resolve tag using %T: %v", r, err))
			continue
		}
		return c
	}
	return nil
}

func (opts *options) resolveByBranch(modulePath string) *types.Commit {
	resolvers := []resolve.ModulerResolver{
		branch.NewGithubBranchResolver(nil),
		branch.NewGitBranchResolver(),
	}
	for _, r := range resolvers {
		c, err := r.Resolve(context.TODO(), modulePath, opts.Branch)
		if err != nil {
			reportError(modulePath, fmt.Errorf("failed to resolve branch using %T: %v", r, err))
			continue
		}
		return c
	}
	return nil
}

func (opts *options) resolveByCommit(modulePath string) *types.Commit {
	resolvers := []resolve.ModulerResolver{
		commit.NewGithubCommitResolver(nil),
		commit.NewGitCommitResolver(),
	}
	for _, r := range resolvers {
		c, err := r.Resolve(context.TODO(), modulePath, opts.Commit)
		if err != nil {
			reportError(modulePath, fmt.Errorf("failed to resolve commit using %T: %v", r, err))
			continue
		}
		return c
	}
	return nil
}

func (opts *options) Complete() error {
	ctx := context.Background()

	if ghToken := os.Getenv("GITHUB_TOKEN"); len(ghToken) > 0 {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: ghToken},
		)
		opts.GithubClient = oauth2.NewClient(ctx, ts)
	} else {
		reportError("", fmt.Errorf("Using Github client without authentication, set GITHUB_TOKEN if you get rate limited"))
	}

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
			defer wg.Done()
			replace := opts.replaces[index]
			var (
				foundCommit *types.Commit
			)

			if len(opts.Branch) > 0 {
				foundCommit = opts.resolveByBranch(replace.newPath)
			}

			if len(opts.Tag) > 0 {
				foundCommit = opts.resolveByTag(replace.newPath)
			}

			if len(opts.Commit) > 0 {
				foundCommit = opts.resolveByCommit(replace.newPath)
			}

			if foundCommit == nil {
				reportError(replace.newPath, fmt.Errorf("unable to get commit"))
				return
			}
			opts.replaces[index].newPathVersion = foundCommit.String()
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
	t := c.Committer.When.UTC()
	timestamp := fmt.Sprintf("%d%.2d%.2d%.2d%.2d%.2d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	return fmt.Sprintf("v0.0.0-%s-%s", timestamp, c.Hash.String()[0:12])
}

func (opts *options) Run() error {
	for _, replace := range opts.replaces {
		if len(replace.newPathVersion) == 0 {
			continue
		}
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
