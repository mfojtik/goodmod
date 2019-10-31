package branch

import (
	"context"
	"net/http"

	"github.com/google/go-github/v28/github"

	"github.com/mfojtik/goodmod/pkg/resolve"
	"github.com/mfojtik/goodmod/pkg/resolve/types"
)

type GithubBranchResolver struct {
	oauthClient *http.Client
}

func NewGithubBranchResolver(oauthClient *http.Client) resolve.ModulerResolver {
	return &GithubBranchResolver{oauthClient: oauthClient}
}

func (g *GithubBranchResolver) Resolve(ctx context.Context, modulePath string, name string) (*types.Commit, error) {
	client := github.NewClient(g.oauthClient)
	owner, repo := resolve.GetGithubOwnerAndRepo(resolve.RepositoryModulePath(modulePath))
	branch, _, err := client.Repositories.GetBranch(ctx, owner, repo, name)
	if err != nil {
		return nil, err
	}
	commit, _, err := client.Git.GetCommit(ctx, owner, repo, branch.GetCommit().GetSHA())
	if err != nil {
		return nil, err
	}
	return &types.Commit{
		SHA:       commit.GetSHA(),
		Timestamp: commit.GetCommitter().GetDate(),
	}, nil
}
