package matrix

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/pkg/errors"
)

type userCredentials struct {
	AccessToken string
	DeviceId    string
}

func pathAccessToken(b *backend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern:         "user/" + framework.GenericNameRegex("localpart") + "/" + framework.GenericNameRegex("domain"),
			HelpSynopsis:    "Gets login information for a user.",
			HelpDescription: "Gets login information for a user.",
			Fields: map[string]*framework.FieldSchema{
				"localpart": {
					Type:        framework.TypeString,
					Description: "The Matrix user ID's localpart to get login details for.",
				},
				"domain": {
					Type:        framework.TypeString,
					Description: "The Matrix user ID's domain to get login details for.",
				},
				"logout_other_devices": {
					Type:        framework.TypeBool,
					Description: "If true, all devices will be logged out prior to generating a token.",
					Default:     false,
				},
			},
			// ExistenceCheck: b.pathExistenceCheck,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ReadOperation: b.pathAccessTokenRead,
			},
		},
	}
}

func (b *backend) generateUserCredentials(ctx context.Context, storage logical.Storage, userId string, homeserver string) (*userCredentials, error) {
	csUrl, err := b.getHomeserverCsUrl(ctx, storage, homeserver)
	if err != nil {
		return nil, err
	}
	if csUrl == "" {
		return nil, errors.New("homeserver client/server url not found")
	}

	secret, err := b.getUserLoginSecret(ctx, storage, userId)
	if err != nil {
		return nil, err
	}
	if secret == "" {
		return nil, errors.New("login secret not found")
	}

	// First make sure that the login flow is supported
	loginFlows, err := getJson(csUrl, "/_matrix/client/r0/login")
	if err != nil {
		return nil, err
	}

	hasVaultLoginType := false
	if flowsRaw, ok := loginFlows["flows"]; ok {
		flows, ok := flowsRaw.([]interface{})
		if !ok {
			return nil, errors.New("invalid response for flows: not a list")
		}

		for _, flowRaw := range flows {
			flow, ok := flowRaw.(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid response for flows: flow is not an object")
			}
			if loginTypeRaw, ok := flow["type"]; ok {
				loginType, ok := loginTypeRaw.(string)
				if !ok {
					return nil, errors.New("invalid response for flows: invalid login type: not a string")
				}

				if loginType == "io.t2bot.vault" {
					hasVaultLoginType = true
					break
				}
			} else {
				return nil, errors.New("invalid response for flows: invalid login type: not an object")
			}
		}
	} else {
		return nil, errors.New("failed to get login flows")
	}

	if !hasVaultLoginType {
		return nil, errors.New("vault login type not supported")
	}

	// Now generate the hash we'll need
	mac := hmac.New(sha256.New, []byte( secret))
	mac.Write([]byte(userId))
	sum := hex.EncodeToString(mac.Sum(nil))

	// Finally actually make the login request
	creds, err := postJson(csUrl, "/_matrix/client/r0/login", map[string]interface{}{
		"type":       "io.t2bot.vault",
		"token_hash": sum,
		"identifier": map[string]interface{}{
			"type": "m.id.user",
			"user": userId,
		},
	})
	if err != nil {
		return nil, err
	}

	result := &userCredentials{}
	if errcodeRaw, ok := creds["errcode"]; ok {
		errcode, ok := errcodeRaw.(string)
		if !ok {
			return nil, errors.New("there was an error logging in, however the errcode is illegible")
		}

		errmsg := ""
		if errmsgRaw, ok := creds["error"]; ok {
			errmsg, ok = errmsgRaw.(string)
			if !ok {
				return nil, errors.New("there was an error logging in, however the error message is illegible")
			}
		}

		return nil, errors.New("error logging in: " + errcode + " " + errmsg)
	}
	if accessTokenRaw, ok := creds["access_token"]; ok {
		accessToken, ok := accessTokenRaw.(string)
		if !ok {
			return nil, errors.New("invalid login response: access_token is not a string")
		}
		result.AccessToken = accessToken
	} else {
		return nil, errors.New("missing access_token in response")
	}
	if deviceIdRaw, ok := creds["device_id"]; ok {
		deviceId, ok := deviceIdRaw.(string)
		if !ok {
			return nil, errors.New("invalid login response: device_id is not a string")
		}
		result.DeviceId = deviceId
	} else {
		return nil, errors.New("missing device_id in response")
	}

	return result, nil
}

func (b *backend) pathAccessTokenRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	localpart := data.Get("localpart").(string)
	domain := data.Get("domain").(string)
	userId := fmt.Sprintf("@%s:%s", localpart, domain)

	logoutFirst := data.Get("logout_other_devices").(bool)

	b.Logger().Info("Generating credentials for " + userId)
	creds, err := b.generateUserCredentials(ctx, req.Storage, userId, domain)
	if err != nil {
		return nil, err
	}

	if logoutFirst {
		b.Logger().Info("Logging out of all devices before continuing")

		csUrl, err := b.getHomeserverCsUrl(ctx, req.Storage, domain)
		if err != nil {
			return nil, err
		}
		if csUrl == "" {
			return nil, errors.New("homeserver client/server url not found")
		}

		_, err = postJson(csUrl, "/_matrix/client/r0/logout/all", map[string]interface{}{})
		if err != nil {
			return nil, err
		}
		// No error means that the logout was successful

		b.Logger().Info("All devices logged out - generating new credentials")
		creds, err = b.generateUserCredentials(ctx, req.Storage, userId, domain)
		if err != nil {
			return nil, err
		}
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"access_token": creds.AccessToken,
			"device_id":    creds.DeviceId,
		},
	}, nil
}
