package matrix

import (
	"context"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathHomeserverConfig(b *backend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: "config/homeserver/?",
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ListOperation: b.pathHomeserverConfigList,
			},
		},
		{
			Pattern:         "config/homeserver/" + framework.GenericNameRegex("homeserver"),
			HelpSynopsis:    "Configures a homeserver's access.",
			HelpDescription: "Configures a homeserver's access.",
			Fields: map[string]*framework.FieldSchema{
				"homeserver": {
					Type:        framework.TypeString,
					Description: "The hostname of a Matrix homeserver.",
				},
				"cs_url": {
					Type:        framework.TypeString,
					Description: "The Client-Server URL for the homeserver. Eg: https://matrix.org",
				},
			},
			ExistenceCheck: b.pathExistenceCheck,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ReadOperation:   b.pathHomeserverConfigRead,
				logical.CreateOperation: b.pathHomeserverConfigCreateUpdate,
				logical.UpdateOperation: b.pathHomeserverConfigCreateUpdate,
				logical.DeleteOperation: b.pathHomeserverConfigDelete,
			},
		},
	}
}

func (b *backend) getHomeserverCsUrl(ctx context.Context, storage logical.Storage, domain string) (string, error) {
	reqPath := "config/homeserver/" + domain

	entry, err := storage.Get(ctx, reqPath)
	if err != nil {
		return "", err
	}

	if entry == nil {
		return "", nil
	}

	return string(entry.Value), nil
}

func (b *backend) pathHomeserverConfigList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	vals, err := req.Storage.List(ctx, "config/homeserver/")
	if err != nil {
		return nil, err
	}
	return logical.ListResponse(vals), nil
}

func (b *backend) pathHomeserverConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	entry, err := req.Storage.Get(ctx, req.Path)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	value := string(entry.Value)

	b.Logger().Info("reading homeserver value", "key", req.Path, "value", value)
	resp := &logical.Response{
		Data: map[string]interface{}{
			"cs_url": value,
		},
	}
	return resp, nil
}

func (b *backend) pathHomeserverConfigCreateUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	csUrl := data.Get("cs_url").(string)

	b.Logger().Info("storing homeserver value", "key", req.Path, "value", csUrl)
	entry := &logical.StorageEntry{
		Key:   req.Path,
		Value: []byte(csUrl),
	}

	s := req.Storage
	err := s.Put(ctx, entry)
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"cs_url": csUrl,
		},
	}, nil
}

func (b *backend) pathHomeserverConfigDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if err := req.Storage.Delete(ctx, req.Path); err != nil {
		return nil, err
	}

	return nil, nil
}
