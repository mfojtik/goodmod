package commit

import (
	"context"
	"net/http"

	"github.com/google/go-github/v28/github"

	"github.com/mfojtik/goodmod/pkg/resolve"
	"github.com/mfojtik/goodmod/pkg/resolve/types"
)

type GithubCommitResolver struct {
	oauthClient *http.Client
}

func NewGithubCommitResolver(oauthClient *http.Client) resolve.ModulerResolver {
	return &GithubCommitResolver{oauthClient: oauthClient}
}

func (g *GithubCommitResolver) Resolve(ctx context.Context, modulePath string, name string) (*types.Commit, error) {
	client := github.NewClient(g.oauthClient)
	owner, repo := resolve.GetGithubOwnerAndRepo(resolve.RepositoryModulePath(modulePath))
	commit, _, err := client.Git.GetCommit(ctx, owner, repo, name)
	if err != nil {
		return nil, err
	}
	return &types.Commit{
		SHA:       commit.GetSHA(),
		Timestamp: commit.GetCommitter().GetDate(),
	}, nil
}
