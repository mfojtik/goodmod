package bump

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v28/github"

	"github.com/mfojtik/goodmod/pkg/resolve"
)

func sanitizeCommitMessage(message string) string {
	firstLine := strings.Split(message, "\n")[0]
	return strings.TrimSpace(firstLine)
}

func ListCommits(modulePath string, fromCommit, toCommit string, oauthClient *http.Client) ([]string, error) {
	client := github.NewClient(oauthClient)
	owner, repo := resolve.GetGithubOwnerAndRepo(resolve.RepositoryModulePath(modulePath))
	commits, _, err := client.Repositories.ListCommits(context.TODO(), owner, repo, &github.CommitsListOptions{
		SHA: fromCommit,
	})
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, c := range commits {
		if strings.HasPrefix(c.GetSHA(), toCommit) {
			break
		}
		if strings.HasPrefix(c.GetCommit().GetMessage(), "Merge pull request") {
			continue
		}
		result = append(result, fmt.Sprintf("%s: %s", c.GetSHA()[0:8], sanitizeCommitMessage(c.GetCommit().GetMessage())))
	}
	return result, nil
}
