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

For higher security, sensitive fields can be encrypted. These values are prefixed with \`ENC(...)\`.

### How it Works
1.  The developer encrypts a secret using the **Public Key**.
2.  The encrypted blob is stored in the YAML file as \`ENC(base64_blob)\`.
3.  At runtime, the library uses the **Private Key** to automatically decrypt the value before it's used by the application.

### Key Distribution
- **Public Key (\`public.pem\`)**: Used for encryption. Can be shared with developers.
- **Private Key (\`private.pem\`)**: **Critical Secret**. Must be stored securely (e.g., \`/etc/bastien/private.pem\`) and never committed to version control.

---

## 3. The \`config-tool\`

A unified CLI tool is provided for key management and encryption.

### Generate Key Pair
\`\`\`bash
go run ./cmd/config-tool keygen --dir .
\`\`\`

### Encrypt a Secret
\`\`\`bash
go run ./cmd/config-tool encrypt --key public.pem --token "my-secret-token"
\`\`\`

---

## 4. Key Discovery Logic

The library searches for the private key in the following order:
1.  Path specified in \`BASTIEN_PRIVATE_KEY_PATH\` environment variable.
2.  \`/etc/bastien/private.pem\` (Production default).
3.  \`./private.pem\` (Local development fallback).

> [!WARNING]
> If the private key is missing and the configuration contains \`ENC(...)\` values, the library will log a warning and return the raw \`ENC(...)\` string.

---

## 5. Decryption Safety

Decryption happens during the "Loading" phase (Step 4 of the resolution hierarchy). Decrypted values are stored in the internal memory-mapped snapshot and are accessible via the standard \`Get()\` or \`GetCapability()\` methods. The raw encrypted strings are never exposed to the application logic.
