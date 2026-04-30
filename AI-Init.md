# ⚡ AI Initialization: distributed-config

> [!IMPORTANT] MANDATORY INITIALIZATION
> Copy and paste this prompt when starting a new session in this repository:
> 
> *"Read the ecosystem map in **[[00-Master-MOC]]** and restore session state from **[[distributed-config/AI-Session-State]]**. Follow the standardized loop in **[[00-Daily-AI-Playbook]]**."*

## 🛡️ Architectural Guardrails (v1.9.1+)
- **Networking**: This library relies on `safe-socket v1.8.2+` for Infinite Wait support. Always ensure `go.mod` is synchronized during propagation.
- **Secrets Management**: The `secret` package is critical. Any changes to configuration loading MUST preserve the `ProcessConfigSecrets` callback to ensure RSA-encrypted fields are decrypted at boot.
- **CLI Tools**: Maintain `cmd/config-keygen` and `cmd/config-encrypt` as the primary bootstrap utilities.
