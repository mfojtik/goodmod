package branch

import (
	"context"
	"fmt"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/mfojtik/goodmod/pkg/resolve"
	"github.com/mfojtik/goodmod/pkg/resolve/types"
)

type GitBranchResolver struct {
}

func (g *GitBranchResolver) Resolve(ctx context.Context, modulePath string, name string) (*types.Commit, error) {
	repository, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{URL: resolve.RepositoryModulePath(modulePath)})
	if err != nil {
		return nil, fmt.Errorf("failed to clone githelper repository %s: %v", resolve.RepositoryModulePath(modulePath), err)
	}
	ref, err := repository.Storer.Reference(plumbing.NewBranchReferenceName(name))
	if err != nil {
		return nil, fmt.Errorf("unable to find branch %s: %v", name, err)
	}
	commit, err := repository.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	return &types.Commit{
		SHA:       commit.Hash.String(),
		Timestamp: commit.Author.When,
	}, nil
}

func NewGitBranchResolver() resolve.ModulerResolver {
	return &GitBranchResolver{}
}
