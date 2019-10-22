package resolve

import (
	"context"

	"github.com/mfojtik/gomod-helpers/pkg/resolve/types"
)

type ModulerResolver interface {
	Resolve(ctx context.Context, modulePath string, name string) (*types.Commit, error)
}
