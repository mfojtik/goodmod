package commit

import (
	"context"

	"github.com/google/go-github/v28/github"

	"github.com/mfojtik/gomod-helpers/pkg/resolve"
	"github.com/mfojtik/gomod-helpers/pkg/resolve/types"
)

type GithubCommitResolver struct{}

func (g *GithubCommitResolver) Resolve(ctx context.Context, modulePath string, name string) (*types.Commit, error) {
	client := github.NewClient(nil)
	owner, repo := resolve.GetGithubOwnerAndRepo(resolve.RepositoryModulePath(modulePath))
	commit, _, err := client.Git.GetCommit(ctx, repo, owner, name)
	if err != nil {
		return nil, err
	}
	return &types.Commit{
		SHA:       commit.GetSHA(),
		Timestamp: commit.GetCommitter().GetDate(),
	}, nil
}
