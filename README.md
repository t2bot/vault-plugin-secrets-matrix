# vault-plugin-secrets-matrix
Vault secrets engine for Matrix

# How it works

When you request an access token for a user, .well-known auto discovery is
performed for that homeserver. There must be a result returned by the discovery
and it is assumed to be pointing to a valid homeserver. The homeserver is
also assumed to be utilizing the `io.t2bot.vault` login type, such as via
the [synapse-vault-auth-provider](https://github.com/t2bot/synapse-vault-auth-provider).

