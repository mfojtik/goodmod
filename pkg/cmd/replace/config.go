package replace

import (
	"fmt"
	"os"

	"github.com/mfojtik/goodmod/pkg/config"
)

func configToOptions(configPath string, singleRule string, originalOptions Options) ([]*Options, bool, error) {
	c, err := config.ReadConfig(configPath)
	if err == config.NotFoundError {
		return nil, true, nil
	}
	if err != nil {
		return nil, false, err
	}
	if originalOptions.Verbose {
		if _, err := fmt.Fprintf(os.Stdout, "# Loaded %d go.mod rules\n", len(c.Rules)); err != nil {
			return nil, false, err
		}
	}
	options := []*Options{}
	for _, rule := range c.Rules {
		if len(singleRule) > 0 {
			found := false
			for _, p := range rule.Paths {
				if singleRule == p {
					found = true
					rule.Paths = []string{p}
					break
				}
			}
			if !found {
				continue
			}
		}
		options = append(options, &Options{
			Branch:       rule.BranchName,
			Commit:       rule.Commit,
			Tag:          rule.TagName,
			Paths:        rule.Paths,
			Excludes:     rule.Excludes,
			GoModPath:    originalOptions.GoModPath,
			GithubClient: originalOptions.GithubClient,
			ApplyReplace: originalOptions.ApplyReplace,
			Verbose:      originalOptions.Verbose,
		})
	}
	if len(singleRule) > 0 && len(options) == 0 {
		return nil, false, fmt.Errorf("no rule matched %q", singleRule)
	}
	return options, false, nil
}
