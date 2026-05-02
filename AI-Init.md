# ⚡ AI Initialization: distributed-config

> [!IMPORTANT] MANDATORY INITIALIZATION
> Copy and paste this prompt when starting a new session in this repository:
> 
> *"1. Read the ecosystem map in **[[00-Master-MOC]]**."*
> *"2. Load project constraints from **[[AI-Project-DNA]]**."*
> *"3. Restore session state from **[[AI-Session-State]]**."*
> *"4. **Audit**: Run `git branch --show-current` and `cat VERSION.txt` to verify state matches the Session State."*
> *"5. **Spec Gate**: Before implementing any feature, you MUST read the behavioral spec in `business-bdd-brain`."*

## 🛡️ Architectural Guardrails (v1.9.1+)
- **Networking**: This library relies on `safe-socket v1.8.2+` for Infinite Wait support. Always ensure `go.mod` is synchronized during propagation.
- **Secrets Management**: The `secret` package is critical. Any changes to configuration loading MUST preserve the `ProcessConfigSecrets` callback to ensure RSA-encrypted fields are decrypted at boot.
- **CLI Tools**: Maintain `cmd/config-keygen` and `cmd/config-encrypt` as the primary bootstrap utilities.
