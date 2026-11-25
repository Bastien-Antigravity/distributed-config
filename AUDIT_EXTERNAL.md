
Initial high-priority findings
One security issue stands out immediately: the RSA key-generation command does not protect the generated private key file.

In cmd/config-tool/main.go, handleKeygen() creates private.pem with os.Create(), which uses permissions of 0666 filtered by the process umask. On a typical Unix system with umask 022, that produces a file readable by other local users (0644). The command labels the key “KEEP THIS SECRET”, but does not enforce that protection.

This is inconsistent with the newer tlsgen implementation, which writes private keys with 0600 permissions and has tests verifying those permissions.

There is also a concrete networking defect: GetAddress() builds addresses using fmt.Sprintf("%s:%s", host, port). For an IPv6 wildcard address, it returns :::7000; the existing test explicitly expects that malformed address instead of a valid bracketed address such as [::]:7000. IPv6 service endpoints therefore need correction, not just additional documentation.

Sources: config-tool/main.go · tlsgen.go · config.go · config_test.go

A further important finding concerns the new certificate tooling: tlsgen can generate per-service certificates, and safe-socket supports TLS and mTLS configuration fields. However, the inspected distributed-config client still connects using the tcp-hello profile and supplies no certificate, key or CA configuration.

So certificate generation is not the same as securing the configuration channel. I cannot establish mTLS protection on this client path from the current implementation. The TLS tooling needs to be connected to the actual client/server transport configuration, with peer verification and service identity checks.

Sources: network/client.go · safe-socket/socket_config.go · tlsgen.go
distributed-config — repository audit
Repository: Bastien-Antigravity/distributed-config · Branch: develop · Reviewed commit: 695bfba · 24 November 2026

Strong architecture
 Live-sync reliability and security gaps

The overall design is ambitious and generally well structured. This is more than a YAML loader: it is the configuration foundation for the wider microservice fleet, with dynamic updates, optimistic concurrency, service discovery, secret handling and a cross-language API.

The recent work on atomic configuration snapshots, serialized writers, stale-revision protection and acknowledged updates is particularly valuable.

My assessment is that the architecture is stronger than the current operational guarantees. The most important concerns are the disconnected-client behavior, missing acknowledgement correlation, and the fact that the newly generated certificates are not yet wired into the configuration client's transport.

1. Detailed findings
   P1 — Disconnected writes can be reported as successful
   This is one of the most important defects in the current implementation.

In src/network/client.go, sendWithUpdateAck() contains a branch that returns nil when the socket is absent:
if sock == nil {
    c.Handler.parentConfig.Logger.Info("Mock: sendWithUpdateAck simulated")
    return nil
}

The same principle appears in requestSync().

The problem is that this behavior is not isolated to a test implementation. A real client can lose its socket while its Client object still exists. The Watch() goroutine sets c.sock to nil while it tries to reconnect.

During that period:

1. The application calls Set().
2. CloudStrategy sends the update through the existing client.
3. The client finds no socket and returns success.
4. CloudStrategy applies the update locally.
5. The caller receives no indication that the remote server never acknowledged the change.
   When the connection recovers, the next full synchronization may overwrite the local-only update.

There is a second problem: if the initial NewClient() connection fails, CloudStrategy.Load() does not start a Watch() loop. The application remains in local-only mode, with no automatic client reconstruction.

Recommendation: remove the mock-success behavior from production network code. A disconnected live client must return an explicit error or an explicit offline status. If offline writes are allowed, queue them with durable status and reconcile them after reconnecting. Distinguish LOCAL_ONLY, PENDING, ACKNOWLEDGED and REJECTED.

Source: client.go · cloud.go
P1 — Server ACKs have no request correlation ID
The current protocol uses:
message ConfigMsg {
    Cmd Command = 1;
    bytes Payload = 2;
    int32 Version = 3;
}

There is no request ID.

The client serializes outstanding synchronization and update operations using syncMu, which is useful. However, consider this sequence:

1. Update A is sent.
2. It times out after five seconds.
3. The callback for A is removed.
4. Update B is sent, installing a new callback.
5. A delayed ACK for A arrives.
6. The handler invokes the currently installed callback, which belongs to B.
   The client can therefore interpret an acknowledgement for A as confirmation of B.

That is a serious consistency problem for configuration changes involving service endpoints, credentials or operational policies.

Recommendation: add a unique request_id to the protocol and echo it in ACK and ERROR messages. Maintain a pending-request map keyed by ID. A late response must never satisfy a newer operation.

Also, consider changing Version from int32 to uint64: the public API uses uint64, but outgoing messages cast it to int32, which creates an eventual overflow/truncation problem.

Sources: config.proto · proto_handler.go · client.go
P1 — TLS certificate generation is not integrated into this client
The new tlsgen command is a good addition. It generates a fleet CA and per-service certificates. The tests verify certificate chains, distinct keys and private-key permissions.

However, the actual distributed-config client still creates its socket using:
profile := fmt.Sprintf("tcp-hello:%s", identity)

And the SocketConfig passed to safe-socket contains only PublicIP and Deadline. It does not provide CertFile, KeyFile, CAFile or ServerName.

The underlying safe-socket library supports TLS profiles and those certificate fields, but the inspected configuration client does not use them.

Consequently, the presence of tlsgen does not establish that the config-server connection is protected by mTLS.

This is particularly important because the configuration channel can distribute service addresses and sensitive configuration values. Protecting secrets at rest is not sufficient if the channel distributing configuration is not authenticated and encrypted.

Recommendation: make transport security an explicit configuration profile. For production, require a TLS transport, server certificate verification, client certificate authentication and peer identity validation. Do not silently fall back to plain TCP.

Sources: network/client.go · tlsgen.go · safe-socket SocketConfig
P1 — Configuration discovery can silently select a different file
ResolveConfigPath() checks CONFIG_PATH, but if the specified path does not exist or is a directory, it continues searching.

That means a deployment explicitly requesting:
CONFIG_PATH=/etc/my-service/production.yaml

could unexpectedly load another discovered configuration file if the requested file is missing.

This is dangerous because configuration selection is part of the security boundary. An operator expects the explicit path to be authoritative.

Recommendation: if CONFIG_PATH is set, treat it as authoritative. Return a clear error if it cannot be resolved or loaded. Apply the same principle to explicitly configured shared paths.

Source: path_resolver.go
P1 — YAML parsing changes quoted strings into booleans
In ProcessNode(), all scalar values are retyped according to their text content.

The code checks whether the lowercased value is true or false, then assigns the YAML boolean tag.

It does not first check whether the original YAML node was explicitly quoted as a string.

For example:
capabilities:
  service:
    token: "false"
    mode: "true"

These values can be converted into booleans instead of strings.

That can produce unexpected type errors when unmarshalling into typed capability structs. It is especially undesirable for tokens, passwords, identifiers or other values whose content happens to be true or false.

Recommendation: preserve explicit YAML string tags and only coerce unquoted plain scalars when that behavior is intended. Better still, avoid a global forced retyping policy and apply field-aware conversion at the capability boundary.

Source: loader.go
P2 — Apply() does not preserve the RCU immutability contract
The implementation correctly uses atomic.Pointer and writer serialization for most configuration updates.

However, Apply() stores the supplied pointer directly:
c.LiveConfig.Store(newMap)

Unlike ReplaceLive(), it does not deep-copy the map.

A caller can retain the original map and mutate it after calling Apply(). Readers access the map without locks, so this breaks the immutability assumption of the RCU model and can introduce data races.

Recommendation: either clone the entire nested map inside Apply() or explicitly transfer ownership and make that contract enforceable. I would favor cloning for a public API.

Source: config.go
P2 — IPv6 address construction is incorrect
As noted above, the current implementation concatenates host and port using a string format rather than net.JoinHostPort().

For IPv6, it produces malformed addresses such as:
:::7000

The current test explicitly expects this result.

Use net.JoinHostPort(host, port) and update the tests to expect [::]:7000.

There is another useful improvement here: validate ports and address formats when loading configuration, rather than allowing malformed endpoints to survive until the networking layer attempts to use them.

Source: config.go · config_test.go
P2 — TLS generator needs stronger certificate validation and rotation rules
The new tlsgen implementation is a good start, but its idempotency rules are too trusting.

When both certificate and key files already exist and force=false, generation is skipped. The code does not establish that:

- The certificate is still valid and unexpired.
- The private key actually matches the certificate.
- The certificate was signed by the current CA.
- The CA is still a valid CA certificate.
- The certificate has the intended service identity.
  writePEMFile() also writes using O_TRUNC. If writing the certificate succeeds but writing the associated key fails, the pair can be left in an inconsistent state.

Finally, the generator puts localhost and loopback SANs into every service certificate. That is convenient for development, but it weakens the distinction between service identities if consumers rely on those shared names rather than validating the intended service identity.

Recommendation: validate existing certificate/key pairs before treating them as reusable. Use temporary files and atomic rename for generation. Separate development SAN defaults from production SAN policy, and keep production certificates service-specific.

Sources: tlsgen.go · tlsgen_test.go
P2 — Production IP validation does not cover every endpoint
CheckProductionIPs() checks whether the static capability map contains an ip equal to 127.0.0.2.

It does not check grpc_ip or rest_ip, and it does not comprehensively validate alternate capability representations.

Consequently, a production service could use a test address in an alternate endpoint field without this check detecting it.

The separate CheckTestIPs() function is not currently integrated into the test strategy. The TODO explains that it was disabled because Docker and host-based testing use different IP ranges.

Recommendation: validate all endpoint types through a common capability schema, with profile-specific policies. Avoid relying on one hardcoded test IP as the only environment integrity check.

Source: validator.go · TODO.md
P2 — CGO error handling has an acknowledged OS-thread issue
The FFI bridge is a major architectural feature, and the handle/lifecycle code has received useful concurrency attention.

However, the error handling uses C _Thread_local variables:
_Thread_local char* last_error;
_Thread_local int last_error_code;

Go goroutines can migrate between OS threads between separate CGO calls. Consequently, an error written in one C call may not be available through a subsequent GetLastError() call if it runs on a different OS thread.

There is also a memory-lifetime concern: the C code allocates error strings using strdup(), and frees them only when another error is set on that same thread.

This is not just a hypothetical refactoring concern: the repository's own TODO still records the need to address thread affinity and strdup leaks.

Recommendation: use an error retrieval contract that does not depend on OS-thread affinity, such as returning an allocated error value tied to the call or a per-handle error store with explicit ownership and release.

Sources: errors.c · helpers.h · TODO.md
Other improvements worth including
There are some lower-priority issues that should be covered in a broader hardening pass.
Area	Observation	Recommendation
Network protocol	Watch() ignores errors returned by HandleIncoming()	Log and count malformed or rejected messages; trigger recovery where appropriate
Capability lookup	GetCapability() does not overlay LiveConfig when the static value is map[string]string	Normalize capability representations before merging
Secret handling	config-tool encrypt --token accepts plaintext as a command-line argument	Support stdin or a hidden interactive prompt to avoid shell history and process argument exposure
Secret processing	ProcessConfigSecrets() substitutes plaintext into raw YAML before parsing	Prefer parsing first and decrypting individual scalar values
Versioning	Protocol version field is int32, while application revisions are uint64	Use a consistent unsigned 64-bit revision field
CI	Shared CI workflow is referenced using mutable @develop	Pin the reusable workflow to a commit SHA
Release checks	Release workflow tests selected packages, not the complete Go test suite	Run the full test and race suites before publishing

2. What is particularly good
   There are several design choices I would preserve.

The RCU architecture is appropriate. LiveConfig uses atomic pointer swaps for read-heavy configuration access. The writer mutex added in the recent work addresses concurrent local mutation and lost-update problems. This is a good fit for services that read configuration frequently but update it relatively infrequently.

The OCC direction is correct. The ability to submit a versioned mutation and receive a conflict response is much safer than blindly overwriting shared configuration. The recent changes to submit deltas and apply them locally after an ACK are also a positive development.

The code is modular. Loading, path discovery, profile strategies, network synchronization, secrets, core state and FFI are separated into distinct packages. That gives the project room to evolve without rewriting the whole system.

The loader has a useful explicit environment-expansion mechanism. The ${VAR:default} syntax is practical and makes configuration portable across local, staging and production deployments.

Secret handling has a good foundation. RSA-OAEP with SHA-256, encrypted values remaining encrypted at load time, on-demand decryption and private-key caching are reasonable design decisions. The problem is more about key handling and end-to-end integration than the choice of encryption primitive.

The polyglot strategy is ambitious but coherent. Using a central Go implementation exposed through a C ABI is a reasonable way to avoid implementing separate configuration engines in every language. The repository has meaningful handle-lifecycle and concurrency tests, which is important for this kind of boundary.

The recent testing work is substantial. The network tests cover stale broadcasts, sequence gaps, version conflicts, full-refresh protection and deletion. The CGO stress test exercises concurrent handle creation, access and closure. The certificate generator also has dedicated certificate-chain and file-permission tests.

Sources: config.go · network tests · CGO stress tests · TLS generator tests
3. Score
Overall engineering quality
7.0/10

Qualitative source-level assessment. Not a benchmark, runtime certification or penetration-test result.

Dimension	Score
Architecture	8.7
Code organization	8.4
RCU and concurrency design	7.8
Configuration discovery	7.0
Live synchronization	7.3
Network reliability	6.3
Security and secrets	6.2
Validation	6.6
Error handling	6.6
CGO / FFI safety	6.5
Testing strategy	7.6
CI and build reproducibility	7.6
Performance potential	8.5
Production readiness	6.5

The architecture and performance potential are strong. The lower scores reflect the difference between having the right mechanisms and ensuring that they behave correctly during disconnection, delayed responses, invalid input and security-sensitive operation.
4. Priority remediation plan
I would organize the next work into four groups.
Priority	Work	Desired result
P0	Protect generated RSA private keys with 0600 permissions	Secure default key generation
P0	Remove false-success behavior when the live socket is disconnected	No silent local-only writes reported as acknowledged
P0	Wire mTLS into the real config client/server path	Authenticated and encrypted fleet configuration
P1	Add request IDs to ACK/ERROR protocol messages	Prevent delayed ACKs from confirming the wrong operation
P1	Implement reconnection after initial connection failure	Production clients recover without process restart
P1	Fail explicitly on invalid CONFIG_PATH	No unexpected configuration selection
P1	Preserve explicitly quoted YAML strings	Correct interpretation of tokens and string values
P1	Validate CA and certificate/key pairs	Reliable certificate lifecycle
P2	Fix IPv6 address formatting and expand endpoint validation	Correct and consistent network addressing
P2	Enforce RCU ownership in Apply()	Immutable snapshots remain safe
P2	Correct CGO error ownership and thread-affinity design	Reliable cross-language error retrieval
P2	Add integration tests for disconnect, timeout, late ACK and recovery	Demonstrate real synchronization guarantees

The most important new tests should exercise these failure sequences:

- Config server becomes unavailable immediately after startup.
- Config server disconnects during an update.
- Update A times out, then its ACK arrives after update B is sent.
- Client reconnects while local changes have accumulated.
- mTLS is enabled and an untrusted client attempts to connect.
- A certificate/key pair is corrupt, expired or signed by an old CA.
- Several writers update and delete different keys while full snapshots arrive.
  These are more valuable than adding a large number of additional happy-path unit tests.

5. Comparison with the other audited repositories
   Using the same qualitative scoring approach:
   Repository	Overall score	Main strength	Main weakness
   tele-remote	7.8	Operational control architecture	Authorization and execution guarantees
   flexible-logger	7.2	Modular sinks and memory ownership	Audit durability and failure visibility
   distributed-config	7.0	RCU, configuration architecture and OCC direction	Disconnected writes and transport security

The three projects are reasonably consistent in architecture and code organization. The recurring weakness is at the operational boundary: what happens when a remote service is unavailable, a request times out, a response arrives late, or an operation appears successful without being verified?

This is a common reliability theme across the fleet, rather than a reason to redesign any one repository.
6. Final verdict
Strong foundation — retain the architecture
I would classify distributed-config as a well-architected configuration platform with meaningful production-oriented engineering, but not yet fully hardened for critical fleet-wide configuration management.

For local configuration loading and high-frequency reads, the design is promising.

For production live configuration, endpoint management, credentials and coordinated fleet changes, I would address the P0 and P1 items before relying on it as an authoritative configuration control plane.

The biggest risk is not that the configuration engine is too complicated. It is that certain failure paths can make the application believe a change has been synchronized when it has not.

My recommendation is to keep the current architecture and concentrate engineering effort on three contracts:

1. Identity: every configuration connection must authenticate the intended service.
2. Delivery: every remote mutation must have an unambiguous, correlated acknowledgement.
3. Consistency: offline, concurrent and delayed operations must never silently overwrite or misrepresent authoritative state.
   One limitation of this review: I inspected the current source, tests, documentation and CI workflows, but did not execute the build, race detector, integration suite or security tests. The scores are therefore source-level assessments, not independently verified runtime results.
