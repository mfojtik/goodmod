package bump

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/mfojtik/goodmod/pkg/cmd/replace"
)

var example = `
# Update 'github.com/openshift/library-go' dependency and commit result
goodmod bump github.com/openshift/library-go
`

type Options struct {
	Path       string
	ConfigPath string
	SingleRule string

	oldVersion string
	newVersion string

	Verbose      bool
	GoModPath    string
	GithubClient *http.Client
}

func (opts *Options) AddFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&opts.Verbose, "verbose", false, "Print more information about progress")
	flags.StringVar(&opts.ConfigPath, "config", "goodmod.yaml", "Specify file to read the replace rules from")
	flags.StringVar(&opts.GoModPath, "gomod-file-path", "go.mod", "Specify the path to go.mod file")
}

func NewBumpCommand() *cobra.Command {
	o := &Options{}

	cmd := &cobra.Command{
		Use:     "bump [path]",
		Example: example,
		Short:   "Bump specified path to latest version",
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(args); err != nil {
				reportFatal("complete failed: %v", err)
			}
			if err := o.Validate(); err != nil {
				reportFatal("validate failed: %v", err)
			}
			if err := o.Run(cmd, args); err != nil {
				reportFatal("run failed: %v", err)
			}
		},
	}

	o.AddFlags(cmd.Flags())

	return cmd
}

func (opts *Options) Complete(args []string) error {
	if len(args) > 0 {
		opts.Path = args[0]
	}
	return nil
}

func (opts *Options) Validate() error {
	if len(opts.Path) == 0 {
		return fmt.Errorf("path argument must be specified")
	}
	return nil
}

func (opts *Options) runReplace(cmd *cobra.Command, args []string) error {
	replaceOpts := &replace.Options{
		ConfigPath:   opts.ConfigPath,
		GoModPath:    opts.GoModPath,
		Verbose:      opts.Verbose,
		ApplyReplace: true,
	}
	replaceOpts.RunCommand(cmd, args)

	// inherit data we gathered in replace command
	opts.oldVersion, opts.newVersion = replaceOpts.GetVersionsForPath(args[0])
	opts.GithubClient = replaceOpts.GithubClient
	return nil
}

func (opts *Options) Run(cmd *cobra.Command, args []string) error {
	if err := opts.runReplace(cmd, args); err != nil {
		return err
	}
	if len(opts.oldVersion) == 0 || len(opts.newVersion) == 0 {
		return fmt.Errorf("path %q old version (%q) or new version (%q) is empty", args[0], opts.oldVersion, opts.newVersion)
	}
	reportVerbose("Listing %q commits from %s to %s", args[0], versionToCommit(opts.oldVersion), versionToCommit(opts.newVersion))
	commits, err := ListCommits(args[0], versionToCommit(opts.oldVersion), versionToCommit(opts.newVersion), opts.GithubClient)
	if err != nil {
		return err
	}
	for _, c := range commits {
		reportVerbose("%s", c)
	}
	if err := commitGoMod(); err != nil {
		return err
	}
	if err := commitVendor(commits); err != nil {
		return err
	}
	return nil
}

func versionToCommit(version string) string {
	parts := strings.Split(version, "-")
	lastPart := parts[len(parts)-1]
	return lastPart
}

func reportVerbose(message string, objects ...interface{}) {
	if _, err := fmt.Fprintf(os.Stderr, "# "+message+"\n", objects...); err != nil {
		panic(err)
	}
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
