package report

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"

	"github.com/mfojtik/goodmod/pkg/config"
	"github.com/mfojtik/goodmod/pkg/golang"
	"github.com/mfojtik/goodmod/pkg/resolve/branch"
)

type Options struct {
	ConfigPath string
	GoModPath  string

	GithubClient *http.Client
}

func (opts *Options) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&opts.ConfigPath, "config", "goodmod.yaml", "Specify file to read the replace rules from")
	flags.StringVar(&opts.GoModPath, "gomod-file-path", "go.mod", "Specify the path to go.mod file")
}

type module struct {
	path           string
	replacePath    string
	currentVersion string
	trackingType   string
	desiredVersion string
}

func (m module) CommitsMissing(client *http.Client) string {
	lister := branch.NewGithubBranchCommitsLister(client)
	commits, err := lister.List(context.TODO(), m.replacePath, m.currentVersion, m.desiredVersion)
	if err != nil {
		return err.Error()
	}
	if commits == 0 {
		return "up to date"
	}
	return fmt.Sprintf("%d commits", commits)
}

func formatModuleVersion(v string) string {
	// v0.0.0-20191016115129-c07a134afb42 => c07a134afb42
	parts := strings.Split(v, "-")
	if len(parts) == 3 {
		return strings.TrimSuffix(parts[2], "+incompatible")
	}
	return strings.TrimSuffix(v, "+incompatible")
}

// parseModules will parse the existing go.mod file and filter out only modules matching the name prefixes specified with this command
func (opts *Options) parseModules(rules []config.Rule) ([]module, error) {
	modBytes, err := ioutil.ReadFile(opts.GoModPath)
	if err != nil {
		return nil, err
	}
	s, err := golang.ParseModFile("go.mod", modBytes, nil)
	if err != nil {
		return nil, err
	}
	modules := []module{}
	for _, r := range s.Replace {
		newModule := module{
			path:           r.Old.Path,
			replacePath:    r.New.Path,
			currentVersion: formatModuleVersion(r.New.Version),
		}
		if rule := config.RuleForPath(rules, r.Old.Path); rule != nil {
			trackingType, version := formatRuleSource(*rule)
			newModule.desiredVersion = version
			newModule.trackingType = trackingType
		} else {
			newModule.desiredVersion = formatModuleVersion(r.New.Version)
			newModule.trackingType = "manual"
		}
		modules = append(modules, newModule)
	}
	for _, r := range s.Require {
		foundReplace := false
		for _, m := range modules {
			if m.path == r.Mod.Path {
				foundReplace = true
				break
			}
		}
		if foundReplace {
			continue
		}
		modules = append(modules, module{
			path:           r.Mod.Path,
			currentVersion: formatModuleVersion(r.Mod.Version),
			desiredVersion: " ",
			trackingType:   "required",
		})
	}
	return modules, nil
}

func (opts *Options) run(cmd *cobra.Command, args []string) {
	if ghToken := os.Getenv("GITHUB_TOKEN"); len(ghToken) > 0 {
		opts.GithubClient = oauth2.NewClient(context.TODO(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghToken}))
	}
	c, err := config.ReadConfig(opts.ConfigPath)
	if err != nil {
		reportFatal(err)
	}

	modules, err := opts.parseModules(c.Rules)
	if err != nil {
		reportFatal(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Path", "Current Version", "Tracking Type", "Desired Version", "Updates"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	tableData := [][]string{}
	for _, m := range modules {
		row := []string{
			m.path,
			m.currentVersion,
			m.trackingType,
			m.desiredVersion,
		}
		if m.trackingType == "branch" {
			row = append(row, m.CommitsMissing(opts.GithubClient))
		} else {
			row = append(row, "")
		}
		tableData = append(tableData, row)
	}
	sort.Slice(tableData, func(i, j int) bool {
		return tableData[i][0] >= tableData[j][0]
	})
	table.AppendBulk(tableData)
	table.Render()
}

func formatRuleSource(rule config.Rule) (string, string) {
	switch {
	case len(rule.Commit) > 0:
		return "commit", rule.Commit[0:12]
	case len(rule.TagName) > 0:
		return "tag", rule.TagName
	case len(rule.BranchName) > 0:
		return "branch", rule.BranchName
	default:
		return "<unknown>", "<unknown>"
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

func NewReportCommand() *cobra.Command {
	reportOptions := &Options{}

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Report the current levels of dependencies",
		Long:  "Report the current levels of dependencies with branches and possible updates",
		Run:   reportOptions.run,
	}

	reportOptions.AddFlags(cmd.Flags())

	return cmd
}
