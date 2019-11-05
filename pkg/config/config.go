package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gobwas/glob"
	"gopkg.in/yaml.v2"
)

var NotFoundError = errors.New("no config file found")

type Config struct {
	// Rules include rules for individual go mod paths
	Rules         []Rule `yaml:"rules,omitempty"`
	GoModFilePath string `yaml:"gomodPath,omitempty"`
}

type Rule struct {
	Paths      []string `yaml:"paths"`
	Excludes   []string `yaml:"excludes,omitempty"`
	BranchName string   `yaml:"branch,omitempty"`
	TagName    string   `yaml:"tag,omitempty"`
	Commit     string   `yaml:"commit,omitempty"`
}

func ReadConfig(configPath string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(configPath)
	if os.IsNotExist(err) {
		return nil, NotFoundError
	}
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %v", configPath, err)
	}

	c := Config{}
	if err := yaml.Unmarshal(configBytes, &c); err != nil {
		return nil, fmt.Errorf("error parsing %q: %v", configPath, err)
	}
	return &c, nil
}

func matchRulePath(s, rulePath string) bool {
	return glob.MustCompile(rulePath).Match(s)
}

func RuleForPath(rules []Rule, modulePath string) *Rule {
	for i := range rules {
		if MatchPath(rules[i].Paths, rules[i].Excludes, modulePath) {
			return &rules[i]
		}
	}
	return nil
}

func MatchPath(paths, excludes []string, modulePath string) bool {
	for _, item := range paths {
		exclude := false
		for _, e := range excludes {
			if matchRulePath(modulePath, e) {
				exclude = true
			}
		}
		if !exclude && matchRulePath(modulePath, item) {
			return true
		}
	}
	return false
}
