package matrix

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathUserConfig(b *backend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: "config/user/?",
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ListOperation: b.pathUserConfigList,
			},
		},
		{
			Pattern:         "config/user/" + framework.GenericNameRegex("localpart") + "/" + framework.GenericNameRegex("domain"),
			HelpSynopsis:    "Configures a user's access.",
			HelpDescription: "Configures a user's access.",
			Fields: map[string]*framework.FieldSchema{
				"localpart": {
					Type:        framework.TypeString,
					Description: "The Matrix user ID's localpart to get login details for.",
				},
				"domain": {
					Type:        framework.TypeString,
					Description: "The Matrix user ID's domain to get login details for.",
				},
				"login_secret": {
					Type:        framework.TypeString,
					Description: "The secret used to authenticate the user.",
				},
			},
			ExistenceCheck: b.pathExistenceCheck,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ReadOperation:   b.pathUserConfigRead,
				logical.CreateOperation: b.pathUserConfigCreateUpdate,
				logical.UpdateOperation: b.pathUserConfigCreateUpdate,
				logical.DeleteOperation: b.pathUserConfigDelete,
			},
		},
	}
}

func (b *backend) getUserLoginSecret(ctx context.Context, storage logical.Storage, userId string) (string, error) {
	reqPath := "config/user/" + userId

	entry, err := storage.Get(ctx, reqPath)
	if err != nil {
		return "", err
	}

	if entry == nil {
		return "", nil
	}

	return string(entry.Value), nil
}

func (b *backend) pathUserConfigList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	vals, err := req.Storage.List(ctx, "config/user/")
	if err != nil {
		return nil, err
	}
	return logical.ListResponse(vals), nil
}

func (b *backend) pathUserConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	localpart := data.Get("localpart").(string)
	domain := data.Get("domain").(string)
	userId := fmt.Sprintf("@%s:%s", localpart, domain)
	reqPath := "config/user/" + userId

	entry, err := req.Storage.Get(ctx, reqPath)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	value := string(entry.Value)

	b.Logger().Info("reading user value", "key", reqPath, "value", value)
	resp := &logical.Response{
		Data: map[string]interface{}{
			"login_secret": value,
		},
	}
	return resp, nil
}

func (b *backend) pathUserConfigCreateUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	localpart := data.Get("localpart").(string)
	domain := data.Get("domain").(string)
	userId := fmt.Sprintf("@%s:%s", localpart, domain)
	reqPath := "config/user/" + userId

	loginSecret := data.Get("login_secret").(string)

	b.Logger().Info("storing user value", "key", reqPath, "value", loginSecret)
	entry := &logical.StorageEntry{
		Key:   reqPath,
		Value: []byte(loginSecret),
	}

	s := req.Storage
	err := s.Put(ctx, entry)
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"login_secret": loginSecret,
		},
	}, nil
}

func (b *backend) pathUserConfigDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	localpart := data.Get("localpart").(string)
	domain := data.Get("domain").(string)
	userId := fmt.Sprintf("@%s:%s", localpart, domain)
	reqPath := "config/user/" + userId

	if err := req.Storage.Delete(ctx, reqPath); err != nil {
		return nil, err
	}

	return nil, nil
}
