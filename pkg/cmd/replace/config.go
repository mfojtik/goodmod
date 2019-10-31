package replace

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/mfojtik/goodmod/pkg/config"
)

func ReadConfigToOptions(configPath string, originalOptions Options) ([]*Options, bool, error) {
	configBytes, err := ioutil.ReadFile(configPath)
	if os.IsNotExist(err) {
		return nil, true, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("error reading %q: %v", configPath, err)
	}
	c := config.Config{}
	if err := yaml.Unmarshal(configBytes, &c); err != nil {
		return nil, false, fmt.Errorf("error parsing %q: %v", configPath, err)
	}
	options := make([]*Options, len(c.Rules))
	for i, rule := range c.Rules {
		options[i] = &Options{
			Branch:       rule.BranchName,
			Commit:       rule.Commit,
			Tag:          rule.TagName,
			Paths:        rule.Paths,
			Excludes:     rule.Excludes,
			GoModPath:    originalOptions.GoModPath,
			GithubClient: originalOptions.GithubClient,
			ApplyReplace: originalOptions.ApplyReplace,
		}
	}
	return options, false, nil
}
