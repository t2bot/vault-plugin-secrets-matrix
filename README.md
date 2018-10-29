# vault-plugin-secrets-matrix

Vault secrets engine for Matrix. 

# Install / Upgrade

You'll need to set the `plugin_directory` for Vault so that the plugin can be
loaded. The example here assumes you have the plugin directory set to `/etc/vault/plugins`. 

```bash
# Clone and build the repository
git clone https://github.com/t2bot/vault-plugin-secrets-matrix.git
cd vault-plugin-secrets-matrix
go build -o /etc/vault/plugins/matrix

# Register the plugin in the catalog
sha256=$(shasum -a 256 "/etc/vault/plugins/matrix" | cut -d " " -f1)
vault write sys/plugins/catalog/matrix sha_256=$sha256 command="matrix"

# Enable the plugin and mount it to matrix/
vault secrets enable -path=matrix -plugin-name=matrix plugin
```

# Configuration

You'll need to configure the homeservers you want the plugin to interact with. Each homeserver
needs a client/server API URL to talk to. This is typically the same as what you'd set for clients.

```bash
vault write matrix/config/homeserver/example.org cs_url=https://example.org
```

You'll also need to configure which users should be accessible through the plugin. Each uses a 
login secret that will be given to the `io.t2bot.vault` login type during authentication.

```bash
# for @alice:example.org
vault write matrix/config/user/alice/example.org login_secret=SomeSecretString
```

# Usage with Synapse

1. Install the [synapse-vault-auth-provider](https://github.com/t2bot/synapse-vault-auth-provider)
2. Register any homeservers and users you want to log in with (as per the "Configuration" section above)
3. Write the same `login_secret` used for the user to the kv store:
   ```bash
   vault kv put secret/matrix/users/@alice:example.org login_secret=SomeSecretString
   ```
4. Read an access token from Vault:
   ```bash
   # for @alice:example.org
   vault read matrix/user/alice/example.org
   ```
