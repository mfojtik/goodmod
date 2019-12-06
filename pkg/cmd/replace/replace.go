package replace

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"

	"github.com/mfojtik/goodmod/pkg/config"
	"github.com/mfojtik/goodmod/pkg/golang"
	"github.com/mfojtik/goodmod/pkg/resolve"
	"github.com/mfojtik/goodmod/pkg/resolve/branch"
	"github.com/mfojtik/goodmod/pkg/resolve/commit"
	"github.com/mfojtik/goodmod/pkg/resolve/tag"
	"github.com/mfojtik/goodmod/pkg/resolve/types"
)

type moduleReplace struct {
	oldPath        string
	oldPathVersion string
	newPath        string
	newPathVersion string
}

type Options struct {
	Branch string
	Commit string
	Tag    string

	Paths      []string
	Excludes   []string
	GoModPath  string
	ConfigPath string
	SingleRule string

	ApplyReplace bool

	GithubClient *http.Client
	Verbose      bool

	replaces []moduleReplace
}

func (opts *Options) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&opts.ConfigPath, "config", "goodmod.yaml", "Specify file to read the replace rules from")
	flags.StringVar(&opts.Branch, "branch", "", "Specify branch to use for this bump")
	flags.StringVar(&opts.Tag, "tag", "", "Specify tag to use for this bump")
	flags.StringVar(&opts.Commit, "commit", "", "Specify commit to use for this bump")
	flags.StringVar(&opts.GoModPath, "gomod-file-path", "go.mod", "Specify the path to go.mod file")
	flags.BoolVar(&opts.ApplyReplace, "apply", false, "Apply the replace rules (execute 'go mod edit -replace' directly)")
	flags.BoolVar(&opts.Verbose, "verbose", false, "Print more information about progress")
	flags.StringSliceVar(&opts.Paths, "paths", []string{}, "Specify dependency path prefixes to update separated by comma (eg. 'github.com/openshift/api' or 'k8s.io/*')")
	flags.StringSliceVar(&opts.Excludes, "excludes", []string{}, "Specify dependency path prefixes to exclude (eg. 'github.com/openshift/api' or 'k8s.io/')")
}

func (opts *Options) GetVersionsForPath(path string) (string, string) {
	for _, p := range opts.replaces {
		if p.oldPath == path {
			return p.oldPathVersion, p.newPathVersion
		}
	}
	return "", ""
}

func (opts *Options) hasReplacePath(path string) bool {
	for _, p := range opts.replaces {
		if p.oldPath == path {
			return true
		}
	}
	return false
}

// parseModules will parse the existing go.mod file and filter out only modules matching the name prefixes specified with this command
func (opts *Options) parseModules() error {
	modBytes, err := ioutil.ReadFile(opts.GoModPath)
	if err != nil {
		return err
	}
	s, err := golang.ParseModFile("go.mod", modBytes, nil)
	if err != nil {
		return err
	}
	opts.replaces = []moduleReplace{}
	for _, r := range s.Replace {
		if config.MatchPath(opts.Paths, opts.Excludes, r.Old.Path) {
			opts.replaces = append(opts.replaces, moduleReplace{newPath: r.New.Path, oldPath: r.Old.Path, oldPathVersion: r.New.Version})
		}
	}
	for _, r := range s.Require {
		if !opts.hasReplacePath(r.Mod.Path) && config.MatchPath(opts.Paths, opts.Excludes, r.Mod.Path) {
			opts.replaces = append(opts.replaces, moduleReplace{newPath: r.Mod.Path, oldPath: r.Mod.Path, oldPathVersion: r.Mod.Version})
		}
	}
	return nil
}

// reportErrorForPath will report errors to standard error output, but prefix all messages with comment, so it can still
// be passed via pipe to command.
func reportErrorForPath(path string, reportedError error) {
	d := color.New(color.FgHiRed, color.Bold)
	if _, err := fmt.Fprintf(os.Stderr, "### "+d.Sprintf("ERROR: %s: %v\n", path, reportedError)); err != nil {
		panic(err)
	}
}

func reportVerbose(message string, objects ...interface{}) {
	if _, err := fmt.Fprintf(os.Stderr, "# "+message+"\n", objects...); err != nil {
		panic(err)
	}
}

func (opts *Options) resolveByTag(modulePath string) *types.Commit {
	resolvers := []resolve.ModulerResolver{
		tag.NewGithubTagResolver(opts.GithubClient),
		tag.NewGitTagResolver(),
	}
	if opts.Verbose {
		reportVerbose("Resolving module path %q using tag %q ...", modulePath, opts.Tag)
	}
	for _, r := range resolvers {
		c, err := r.Resolve(context.TODO(), modulePath, opts.Tag)
		if err != nil {
			reportErrorForPath(modulePath, fmt.Errorf("failed to resolve tag using %T: %v", r, err))
			continue
		}
		if opts.Verbose {
			reportVerbose("Module path %q resolved to %q ...", modulePath, c.String())
		}
		return c
	}
	return nil
}

func (opts *Options) resolveByBranch(modulePath string) *types.Commit {
	resolvers := []resolve.ModulerResolver{
		branch.NewGithubBranchResolver(opts.GithubClient),
		branch.NewGitBranchResolver(),
	}
	if opts.Verbose {
		reportVerbose("Resolving module path %q using branch %q ...", modulePath, opts.Branch)
	}
	for _, r := range resolvers {
		c, err := r.Resolve(context.TODO(), modulePath, opts.Branch)
		if err != nil {
			reportErrorForPath(modulePath, fmt.Errorf("failed to resolve branch using %T: %v", r, err))
			continue
		}
		if opts.Verbose {
			reportVerbose("Module path %q resolved to %q ...", modulePath, c.String())
		}
		return c
	}
	return nil
}

func (opts *Options) resolveByCommit(modulePath string) *types.Commit {
	resolvers := []resolve.ModulerResolver{
		commit.NewGithubCommitResolver(opts.GithubClient),
		commit.NewGitCommitResolver(),
	}
	if opts.Verbose {
		reportVerbose("Resolving module path %q using commit %q ...", modulePath, opts.Commit)
	}
	for _, r := range resolvers {
		c, err := r.Resolve(context.TODO(), modulePath, opts.Commit)
		if err != nil {
			reportErrorForPath(modulePath, fmt.Errorf("failed to resolve commit: %v", err))
			continue
		}
		if opts.Verbose {
			reportVerbose("Module path %q resolved to %q ...", modulePath, c.String())
		}
		return c
	}
	return nil
}

func (opts *Options) Complete() error {
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
				reportErrorForPath(replace.newPath, fmt.Errorf("unable to get commit"))
				return
			}
			opts.replaces[index].newPathVersion = foundCommit.String()
		}(i)
	}

	wg.Wait()
	return nil
}

func (opts *Options) applyReplace(replace moduleReplace) error {
	cmd := exec.Command("go", "mod", "edit", "-replace", fmt.Sprintf(`%s=%s@%s`, replace.oldPath, replace.newPath, replace.newPathVersion))
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %v\n%s", strings.Join(cmd.Args, " "), err, string(outBytes))
	}
	return nil
}

func (opts *Options) Run() error {
	for _, replace := range opts.replaces {
		if len(replace.newPathVersion) == 0 {
			continue
		}
		if _, err := fmt.Fprintf(os.Stdout, `go mod edit -replace %s=%s@"%s"`+"\n", replace.oldPath, replace.newPath, replace.newPathVersion); err != nil {
			return err
		}
		if opts.ApplyReplace {
			if err := opts.applyReplace(replace); err != nil {
				return err
			}
		}
	}
	return nil
}

func (opts *Options) Validate() error {
	if len(opts.Branch) == 0 && len(opts.Commit) == 0 && len(opts.Tag) == 0 {
		return fmt.Errorf("either branch, commit or tag must be specified")
	}
	if len(opts.Paths) == 0 {
		return fmt.Errorf("dependency name must be specified")
	}
	return nil
}

func reportFatal(message interface{}, objects ...interface{}) {
	formatMessage := ""
	switch v := message.(type) {
	case error:
		formatMessage = v.Error()
	case string:
		formatMessage = v
	}
	if _, err := fmt.Fprintf(os.Stderr, "ERROR: "+formatMessage+"\n", objects...); err != nil {
		panic(err)
	}
	os.Exit(1)
}

func (opts *Options) RunCommand(cmd *cobra.Command, args []string) {
	if ghToken := os.Getenv("GITHUB_TOKEN"); len(ghToken) > 0 {
		opts.GithubClient = oauth2.NewClient(context.TODO(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghToken}))
	}
	if len(args) == 1 {
		opts.SingleRule = strings.TrimSpace(args[0])
	}
	options, noConfig, err := ConfigToOptions(opts.ConfigPath, opts.SingleRule, *opts)
	if err != nil {
		reportFatal(err)
	}
	// we don't have config passed, RunCommand using flags
	if noConfig {
		opts.RunOnce(cmd, args)
		return
	}
	for _, o := range options {
		o.RunOnce(cmd, args)

		// TODO: This will only preserve last replace set
		opts.replaces = o.replaces
	}
}

func (opts *Options) RunOnce(cmd *cobra.Command, args []string) {
	if err := opts.Validate(); err != nil {
		reportFatal(err)
	}
	if err := opts.Complete(); err != nil {
		reportFatal(err)
	}
	if err := opts.Run(); err != nil {
		reportFatal(err)
	}
}

var example = `
# Update all k8s.io/* paths to 'kubernetes-1.16.2' tag
goodmod replace --paths=k8s.io/* --tag=kubernetes-1.16.2

# Update all github.com/openshift/* to HEAD commit in 'master' branch
goodmod replace --paths=github.com/openshift/* --branch=master

# Update all github.com/openshift/library-go to HEAD commit in 'master' branch
# The goodmod.yaml config file MUST specify at least one rule for this path.
goodmod replace github.com/openshift/library-go

# Update all modules specified in goodmod.yaml and apply changes to go.mod directly
goodmod replace --apply --verbose
`

func NewReplaceCommand() *cobra.Command {
	replaceOptions := &Options{}

	cmd := &cobra.Command{
		Use:     "replace [path]",
		Example: example,
		Short:   "Replace multiple modules at once",
		Long:    "Replace help to perform bulk operations on go.mod replace in case you want to track branch, tag or commit for single path",
		Run:     replaceOptions.RunCommand,
	}
	replaceOptions.AddFlags(cmd.Flags())

	return cmd
}
