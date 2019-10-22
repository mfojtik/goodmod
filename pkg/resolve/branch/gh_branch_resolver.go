package branch

import (
	"context"

	"github.com/google/go-github/v28/github"

	"github.com/mfojtik/gomod-helpers/pkg/resolve"
	"github.com/mfojtik/gomod-helpers/pkg/resolve/types"
)

type GithubBranchResolver struct{}

func (g *GithubBranchResolver) Resolve(ctx context.Context, modulePath string, name string) (*types.Commit, error) {
	client := github.NewClient(nil)
	owner, repo := resolve.GetGithubOwnerAndRepo(resolve.RepositoryModulePath(modulePath))
	branch, _, err := client.Repositories.GetBranch(ctx, owner, repo, name)
	if err != nil {
		return nil, err
	}
	commit, _, err := client.Git.GetCommit(ctx, repo, owner, branch.GetCommit().GetSHA())
	if err != nil {
		return nil, err
	}
	return &types.Commit{
		SHA:       commit.GetSHA(),
		Timestamp: commit.GetCommitter().GetDate(),
	}, nil
}
