package replace

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/mfojtik/goodmod/pkg/config"
)

var NoConfigError = errors.New("no config file")

func readConfig(configPath string) (*config.Config, error) {
	configBytes, err := ioutil.ReadFile(configPath)
	if os.IsNotExist(err) {
		return nil, NoConfigError
	}
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %v", configPath, err)
	}
	c := config.Config{}
	if err := yaml.Unmarshal(configBytes, &c); err != nil {
		return nil, fmt.Errorf("error parsing %q: %v", configPath, err)
	}
	return &c, nil
}

func ReadConfigToOptions(configPath string, singleRule string, originalOptions Options) ([]*Options, bool, error) {
	c, err := readConfig(configPath)
	if err == NoConfigError {
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
