package config

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
