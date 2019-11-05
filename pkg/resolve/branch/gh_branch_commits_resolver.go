package branch

import (
	"context"
	"net/http"

	"github.com/google/go-github/v28/github"

	"github.com/mfojtik/goodmod/pkg/resolve"
)

type GithubBranchCommitsLister struct {
	oauthClient *http.Client
}

func NewGithubBranchCommitsLister(oauthClient *http.Client) *GithubBranchCommitsLister {
	return &GithubBranchCommitsLister{oauthClient: oauthClient}
}

func (g *GithubBranchCommitsLister) List(ctx context.Context, modulePath string, startingCommit string, branchName string) (int, error) {
	client := github.NewClient(g.oauthClient)
	owner, repo := resolve.GetGithubOwnerAndRepo(resolve.RepositoryModulePath(modulePath))
	comparison, _, err := client.Repositories.CompareCommits(ctx, owner, repo, startingCommit, branchName)
	if err != nil {
		return 0, err
	}
	return comparison.GetBehindBy(), nil
}
