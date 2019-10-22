package commit

import (
	"context"
	"fmt"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/mfojtik/gomod-helpers/pkg/resolve"
	"github.com/mfojtik/gomod-helpers/pkg/resolve/types"
)

type GitCommitResolver struct {
}

func (g *GitCommitResolver) Resolve(ctx context.Context, modulePath string, name string) (*types.Commit, error) {
	repository, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{URL: resolve.RepositoryModulePath(modulePath)})
	if err != nil {
		return nil, fmt.Errorf("failed to clone githelper repository %s: %v", resolve.RepositoryModulePath(modulePath), err)
	}
	commit, err := repository.CommitObject(plumbing.NewHash(name))
	if err != nil {
		return nil, err
	}
	return &types.Commit{
		SHA:       commit.Hash.String(),
		Timestamp: commit.Author.When,
	}, nil
}

func NewGitCommitResolver() resolve.ModulerResolver {
	return &GitCommitResolver{}
}
