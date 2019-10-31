package resolve

import (
	"context"

	"github.com/mfojtik/goodmod/pkg/resolve/types"
)

type ModulerResolver interface {
	Resolve(ctx context.Context, modulePath string, name string) (*types.Commit, error)
}
