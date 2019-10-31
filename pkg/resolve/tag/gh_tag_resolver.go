package tag

import (
	"context"
	"net/http"

	"github.com/google/go-github/v28/github"

	"github.com/mfojtik/goodmod/pkg/resolve"
	"github.com/mfojtik/goodmod/pkg/resolve/types"
)

type GithubTagResolver struct {
	oauthClient *http.Client
}

func NewGithubTagResolver(oauthClient *http.Client) resolve.ModulerResolver {
	return &GithubTagResolver{oauthClient: oauthClient}
}

func (r *GithubTagResolver) Resolve(ctx context.Context, path string, tagName string) (*types.Commit, error) {
	client := github.NewClient(r.oauthClient)
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
