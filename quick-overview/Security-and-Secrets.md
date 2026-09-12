---
tags:
- '#ai/ignore'
- '#zone/3-fleet'
- '#service/distributed-config'
- '#type/overview'
- '#state/active'
microservice: distributed-config
type: overview
status: active
---
# Security & Secrets

\`distributed-config\` provides native support for managing sensitive information (API keys, passwords) using RSA-based encryption and environment variable expansion.

## 1. Environment Variable Expansion

The simplest way to handle secrets is through environment variables. In your YAML:

\`\`\`yaml
capabilities:
  timescale_db:
    password: "\${DB_PASSWORD}"
\`\`\`

The library will automatically replace \`\${DB_PASSWORD}\` with the value of the environment variable at runtime.

---

## 2. RSA Encryption (ENC)

For higher security, sensitive fields can be encrypted. These values are prefixed with `ENC(...)`.

### How it Works
1.  The developer encrypts a secret using the **Public Key**.
2.  The encrypted blob is stored in the YAML file as `ENC(base64_blob)`.
3.  At boot, configuration loaders preserve `ENC(...)` values untouched in memory.
4.  At runtime, authorized consumer microservices use the **Private Key** to explicitly decrypt the secret on-demand via `AppConfig.DecryptSecret(val)` (or `secret.Decrypt(val)`).

### Key Distribution & Per-Service Keys
- **Public Key (`public.pem`)**: Used for encryption. Can be shared with developers or embedded in service public specs.
- **Private Key (`private.pem`)**: **Critical Secret**. Must be stored securely (e.g., `/etc/bastien/private.pem` or injected via `BASTIEN_PRIVATE_KEY` / `BASTIEN_PRIVATE_KEY_PATH`) and never committed to version control.
- **Per-Service Key Pairs**: The architecture supports generating distinct RSA key pairs per service (e.g., `tele-remote`, `notif-server`). Secrets encrypted with Service A's public key cannot be decrypted by Service B even if Service B holds its own private key.
- **Zero-Knowledge Distribution**: Neither `config-server` nor `web-interface` holds private keys or performs decryption. Centralized distributors serve ciphertext blobs directly.

---

## 3. The `config-tool`

A unified CLI tool is provided for key management and encryption.

### Generate Key Pair
```bash
go run ./cmd/config-tool keygen --dir .
```

### Encrypt a Secret
```bash
go run ./cmd/config-tool encrypt --key public.pem --token "my-secret-token"
```

---

## 4. Key Discovery Logic

The library searches for the private key in the following order:
1.  Inline PEM string in `BASTIEN_PRIVATE_KEY` environment variable.
2.  Path specified in `BASTIEN_PRIVATE_KEY_PATH` environment variable or `--key` CLI argument.
3.  `/etc/bastien/private.pem` (Production default).
4.  `./private.pem` (Local development fallback).

> [!WARNING]
> If the private key is missing, decryption requests for `ENC(...)` values will fail, returning an explicit error. Plaintext secrets are never silently guessed or exposed.

---

## 5. Decryption Safety & Zero-Knowledge Architecture

Decryption does **NOT** occur during the initial configuration loading phase. Encrypted values (`ENC(...)`) remain untouched in the internal memory-mapped AST and are accessible via standard `Get()` or `GetCapability()` methods as encrypted strings.

Decryption occurs **strictly on-demand** inside the authorized consumer microservice:
```go
token := appConfig.GetCapability("telegram.bot_token") // Returns "ENC(...)"
plainToken, err := appConfig.DecryptSecret(token)      // Decrypts in-memory using local private key
```

Plaintext secrets exist only in volatile runtime memory for the shortest time necessary and must never be written to disk, synced to `config-server`, or logged.
