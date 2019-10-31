package tag

import (
	"context"
	"fmt"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/mfojtik/goodmod/pkg/resolve"
	"github.com/mfojtik/goodmod/pkg/resolve/types"
)

type GitTagResolver struct {
}

func (g *GitTagResolver) Resolve(ctx context.Context, modulePath string, name string) (*types.Commit, error) {
	repository, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{URL: resolve.RepositoryModulePath(modulePath)})
	if err != nil {
		return nil, fmt.Errorf("failed to clone githelper repository %s: %v", resolve.RepositoryModulePath(modulePath), err)
	}

	ref, err := repository.Storer.Reference(plumbing.NewTagReferenceName(name))
	if err != nil {
		return nil, fmt.Errorf("unable to find tag %s: %v", name, err)
	}
	obj, err := repository.TagObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("object not found: %v", err)
	}
	commit, err := repository.CommitObject(obj.Target)
	if err != nil {
		return nil, fmt.Errorf("unable to find commit %s: %v", name, err)
	}

	return &types.Commit{
		SHA:       commit.Hash.String(),
		Timestamp: commit.Author.When,
	}, nil
}

func NewGitTagResolver() resolve.ModulerResolver {
	return &GitTagResolver{}
}
