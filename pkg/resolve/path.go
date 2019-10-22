package resolve

import (
	"fmt"
	"strings"
)

// RepositoryModulePath try to do the best to resolve the Go module path into Git(hub) repository.
// NOTE: Only kubernetes repositories are special cased here, there are many more (gopkg.in, etc..)
func RepositoryModulePath(path string) string {
	// snowflake kubernetes
	if strings.HasPrefix(path, "k8s.io/") {
		return fmt.Sprintf("https://github.com/kubernetes/%s", strings.TrimPrefix(path, "k8s.io/"))
	}
	// TODO: probably snowflake others here
	return fmt.Sprintf("https://%s", path)
}

// GetGithubOwnerAndRepo splits the repository into owner and repository name.
func GetGithubOwnerAndRepo(r string) (string, string) {
	r = strings.TrimPrefix(r, "https://github.com/")
	parts := strings.Split(r, "/")
	return parts[0], parts[1]
}
