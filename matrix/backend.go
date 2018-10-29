package matrix

import (
	"context"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/pkg/errors"
)

// Factory creates a new usable instance of this secrets engine.
func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	b := Backend(c)
	if err := b.Setup(ctx, c); err != nil {
		return nil, errors.Wrap(err, "failed to create factory")
	}
	return b, nil
}

// backend is the actual backend.
type backend struct {
	*framework.Backend
	storagePrefix string
}

func Backend(c *logical.BackendConfig) *backend {
	var b backend

	b.Backend = &framework.Backend{
		BackendType: logical.TypeLogical,
		Help:        "The Matrix secrets engine provides access tokens to Matrix users on a homeserver.",
		Secrets:     []*framework.Secret{},
		Invalidate:  b.invalidate,
		Paths: framework.PathAppend(
			pathHomeserverConfig(&b),
			pathUserConfig(&b),
			pathAccessToken(&b),
		),
	}

	return &b
}

func (b *backend) invalidate(ctx context.Context, key string) {
	// TODO: Do something here?
}
