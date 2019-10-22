package tag

import (
	"context"

	"github.com/google/go-github/v28/github"

	"github.com/mfojtik/gomod-helpers/pkg/resolve"
	"github.com/mfojtik/gomod-helpers/pkg/resolve/types"
)

type GithubTagResolver struct{}

func (r *GithubTagResolver) Resolve(ctx context.Context, path string, tagName string) (*types.Commit, error) {
	client := github.NewClient(nil)
	owner, repo := resolve.GetGithubOwnerAndRepo(resolve.RepositoryModulePath(path))
	tree, _, err := client.Git.GetTree(ctx, owner, repo, tagName, false)
	if err != nil {
		return nil, err
	}
	commit, _, err := client.Git.GetCommit(ctx, owner, repo, tree.GetSHA())
	if err != nil {
		return nil, err
	}
	return &types.Commit{
		SHA:       commit.GetSHA(),
		Timestamp: commit.GetCommitter().GetDate(),
	}, err
}
