# External Key-Value Storage

Tyk Gateway supports storing configuration data in key-value (KV) stores such as AWS Secrets Manager, Consul, and Vault
and then referencing these values the gateway configuration or API definitions deployed on the gateway.

## Supported KV Stores

Tyk Gateway supports the following KV stores:

- [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/)
- [Consul](https://www.consul.io/)
- [Vault](https://www.vaultproject.io/)
- Tyk Config Secrets
- Environment Variables

## Accessing KV Store Data

You can configure Tyk Gateway to retrieve values from KV stores in the following places:

- Tyk Gateway configuration file (tyk.conf)
- API definitions
- Transform Middleware

### Gateway Configuration

You can retrieve values from KV stores for the following fields Tyk Gateway configuration fields:

- `secret`
- `node_secret`
- `storage.password`
- `cache_storage.password`
- `security.private_certificate_encoding_secret`
- `db_app_conf_options.connection_string`
- `policies.policy_connection_string`

Use the following notation to reference KV store values in the gateway configuration file:

| Store                       | Notation                                                    |
|-----------------------------|-------------------------------------------------------------|
| AWS Secrets Manager \[1\]   | `secretsmanager://{key}` or `secretsmanager://{path}.{key}` |
| Consul                      | `consul://{key}`                                            |
| Vault  \[1\]                | `vault://{path}.{key}`                                      |
| Tyk config secrets          | `secrets://{key}`                                           |
| Environment variables \[2\] | `env://{key}`                                               |

\[1\] Path key value must be a JSON document to support `{path}.{key}` references (e.g., `{"key": "value"}`).\
\[2\] Gateway configuration environment variable names must be prefixed with `TYK_SECRET_{KEY}` (e.g.,
`TYK_SECRET_MY_API_SECRET` is referenced as `env://MY_API_SECRET`). Environment variables names must be uppercase.

### API Definition

Use the following notation to reference KV store values in API definition fields:

| Store                       | Notation                 |
|-----------------------------|--------------------------|
| AWS Secrets Manager \[1\]   | `secretsmanager://{key}` |
| Consul \[1\]                | `consul://{key}`         |
| Vault \[1\]                 | `vault://{key}`          |
| Tyk config secrets          | `secrets://{key}`        |
| Environment variables \[2\] | `env://{key}`            |

\[1\] External API definition secrets must be stored in a JSON document in the KV store with the key `tyk-apis`.
Vault API definition secrets must be stored in `/secrets/data/tyk-apis`. In this case, `{key}` refers to the JSON 
document key (e.g., `{"key": "value"}`).\
\[2\] API definition environment variables can be any name (e.g. `MY_API_SECRET` is referenced as
`env://MY_API_SECRET`). API definition environment variables **do not** require the prefix `TYK_SECRET_{KEY}`.

### Transform Middleware

You can retrieve values from KV stores for the following transform middleware:

- Authentication Token (signature secret)
- Persist GraphQL Operation (request path)
- Rate Limiting (rate limit pattern)
- Request Body
- Request Headers
- Response Headers
- URL Rewrite

Use the following notation to reference KV store values in transform middleware:

| Store                       | Notation                                                                |
|-----------------------------|-------------------------------------------------------------------------|
| AWS Secrets Manager \[1\]   | `$secret_secretsmanager.{key}` or `$secret_secretsmanager.{path}.{key}` |
| Consul                      | `$secret_consul.{key}`                                                  |
| Vault  \[1\]                | `$secret_vault.{path}.{key}`                                            |
| Tyk config secrets          | `$secret_conf.{key}`                                                    |
| Environment variables \[2\] | `$secret_env.{key}`                                                     |

\[1\] Path key value must be a JSON document to support `{path}.{key}` references (e.g., `{"key": "value"}`).\
\[2\] Transform middleware environment variable names must be prefixed with `TYK_SECRET_{KEY}` (e.g.,
`TYK_SECRET_MY_API_SECRET` is referenced as `$secret_env.my_api_secret`).
