Deep technical review — distributed-config
Repository: Bastien-Antigravity/distributed-config · Branch: develop · Focus: configuration discovery, precedence, merging, profiles, live synchronization, concurrency, encryption and cross-language integration

This repository is arguably even more important than config-server.

Why? Because config-server distributes configuration, whereas distributed-config decides what configuration each microservice actually uses.

It is effectively the configuration runtime shared by the entire ecosystem.

I see a fairly ambitious design:
- Multiple configuration sources.
- Profile-based loading.
- Environment-variable expansion.
- Remote configuration synchronization.
- Atomic pointer swaps.
- Dynamic callbacks.
- RSA-based secret handling.
- A C ABI for Python, Rust, C/C++ and other consumers.
The biggest architectural question is whether all these sources have a clearly defined and deterministic precedence, and whether the same precedence is respected by the Go implementation, the remote server and the polyglot interfaces.

That is where I would concentrate the review.
1. Understanding the architecture
The library has two distinct configuration domains:
- Static configuration: Common, Capabilities, profile settings, YAML and environment-derived values.
- Live configuration: dynamic key/value overrides received from the configuration server or changed at runtime.
This distinction is important because the two domains are merged at different points.
Configuration sources

Defaults · YAML · Environment · Remote server · Runtime updates


Static Config

CommonConfig + Capabilities


LiveConfig

Atomic pointer to nested maps




Config access and merge layer

Get · GetCapability · GetAddress · GetGRPCAddress


Go services


C ABI / FFI

Python · Rust · C++ · VBA




The separation is reasonable. It also makes lock-free reads possible without repeatedly parsing YAML or contacting the configuration server.

The main implementation files I examined include:
- core/config.go
- strategies/cloud.go
- strategies/test.go
- network/client.go
- network/proto_handler.go

2. Configuration precedence: the first thing to clarify
This is one of the most important aspects of the library.

The project documentation describes a layered configuration system, but the precedence is not expressed consistently everywhere.

For example:

The project DNA describes:
CLI flags
    >
Local YAML
    >
Config Server
    >
Environment variables


However, the CloudStrategy implementation does something more nuanced.
A. Actual production loading sequence
Looking at CloudStrategy.Load():
1. Load local YAML

   Initial static configuration and capabilities
2. Apply environment overrides

   Common settings such as name and public IP
3. Connect to config-server

   Retrieve dynamic LiveConfig
4. Merge server capabilities into static capabilities

   Populate capability baseline
5. Reload local YAML

   Local file wins for static configuration
6. Synchronize capabilities into LiveConfig

   Create the final live capability representation
7. Validate mandatory services

   Check log_server and config_server

Source: cloud.go.
GitHub+2

B. What does this mean in practice?
Suppose we have the following values:
Source	database.host
Default	127.0.0.1
YAML	database.local
Config server	database.remote
Environment interpolation	database.env

The final value depends on which configuration domain is being read.

For static capabilities, the local YAML takes precedence over the remote capability baseline.

For live configuration, remote values are stored in LiveConfig, and GetCapability() merges them over static capability values.

Environment interpolation occurs while loading YAML, whereas LoadCommonFromEnv() separately overrides selected common fields.

Therefore, there isn't one universal precedence chain.

My recommendation: define precedence independently for:
1. Common configuration.
2. Static capabilities.
3. Dynamic LiveConfig.
4. Secret resolution.
Then expose one documented precedence table.

This will prevent different parts of the ecosystem from implementing their own interpretations.
3. The merging mechanism inside Config.GetCapability()
This is where the library becomes interesting.

The method:
func (c *Config) GetCapability(
    key string,
    target interface{},
) error


does not simply return a capability from YAML.

It constructs a merged representation.

Imagine the static YAML:
capabilities:
  database:
    ip: 192.168.1.10
    port: "5432"
    max_connections: "100"


And the remote server provides:
database:
  ip: 10.0.0.20
  max_connections: "200"


The method first copies the static capability map.

Then it applies the live overrides.
Static capability
ip: 192.168.1.10
port: "5432"
max_connections: "100"



Live override
ip: 10.0.0.20
max_connections: "200"



Merged result
ip: 10.0.0.20
port: "5432"
max_connections: "200"


Live values win

There is another feature: the method also supports dotted capability keys in the capabilities LiveConfig section.

For example:
capabilities:
    database.ip = 10.0.0.20


The method extracts the database. prefix and merges the remaining keys into the database capability.

This is useful for backward compatibility and flexible updates.

GitHub+2
A. A potential weakness: duplicated merge logic
The library currently has several places where capability values are merged:
- GetCapability()
- getCapabilityField()
- syncCapabilitiesToLiveConfig()
- CloudStrategy.Load()
- TestStrategy.Load()
This creates a risk of precedence drift.

For example, GetCapability() and getCapabilityField() both prioritize live values, but they have different implementations for extracting them.

I would introduce a single internal function:
func (c *Config) ResolveCapability(
    name string,
) (map[string]interface{}, error)


Its responsibilities:
1. Read static capability.
2. Read direct LiveConfig overrides.
3. Read dotted LiveConfig overrides.
4. Apply the documented precedence.
5. Return a detached merged representation.
Then GetCapability(), GetAddress() and other accessors can reuse it.

This would remove duplicated merging logic and make future changes much safer.
B. The missing-capability problem
There is a particularly interesting issue in getAddr().

The current implementation checks:
if c.Capabilities != nil {
    if _, ok := c.Capabilities[capability]; !ok {
        return "", fmt.Errorf(
            "capability %s not found",
            capability,
        )
    }
}


This occurs before attempting to resolve the live override.

Imagine:
# Static configuration
capabilities:
  database:
    ip: localhost
    port: "5432"


Then the remote server introduces:
analytics.ip = 10.0.0.20
analytics.port = 9000


The live configuration contains a perfectly valid analytics capability.

But because Capabilities is non-nil and does not contain analytics, getAddr() rejects it.
Confirmed code-level issue
The TODO file explicitly recognizes this limitation.

Recommendation: resolve the merged capability first, then check whether its host and port are available. Do not require the capability to exist in the static baseline if dynamic capabilities are intended to be supported.

GitHub+1
4. The most serious issue: concurrent writers can overwrite each other
This is the first thing I would fix.

The library uses an atomic pointer for reads, which is good.

However, atomic pointer replacement does not automatically make a read-modify-write operation atomic.

Look at Config.Set():
func (c *Config) Set(
    updates map[string]map[string]string,
) {
    if updates == nil {
        return
    }

    newMap := c.PreviewSet(updates)
    c.LiveConfig.Store(newMap)
}


PreviewSet() performs:
1. Load current snapshot.
2. Deep-copy current map.
3. Apply updates.
4. Return new map.
Then Set() stores it.

The problem is that two writers can perform those steps simultaneously.
A. Concrete example
Initial state:
database:
  host: localhost


Two goroutines execute:
// Goroutine A
cfg.Set(map[string]map[string]string{
    "database": {"host": "db.internal"},
})

// Goroutine B
cfg.Set(map[string]map[string]string{
    "redis": {"host": "redis.internal"},
})


Possible execution:
Lost update scenario
Initial snapshot
database.host = localhost


Goroutine A

Reads snapshot 100.

Creates snapshot 101:
database.host = db.internal



Goroutine B

Reads snapshot 100.

Creates snapshot 102:
database.host = localhost
redis.host = redis.internal





Final atomic pointer swap

Goroutine A stores 101.

Goroutine B stores 102.

The database update is lost.

The atomic operation guarantees that readers never see a partially written map. It does not guarantee that concurrent modifications are merged.

The same fundamental problem exists in Delete() and in the PreviewSet() → Apply() pattern.

The repository's TODO actually identifies this exact issue as a deferred architecture improvement.
GitHub+2

B. My preferred correction
Use a writer mutex while retaining lock-free reads.
type Config struct {
    // Existing configuration fields

    writeMu sync.Mutex
}


Then:
func (c *Config) Set(
    updates map[string]map[string]string,
) {
    if updates == nil {
        return
    }

    c.writeMu.Lock()
    defer c.writeMu.Unlock()

    next := c.PreviewSet(updates)
    c.LiveConfig.Store(next)
}


This ensures that each writer reads the snapshot produced by the previous writer.

The result:
Initial
   |
   v
Update A
   |
   v
Update B
   |
   v
Final snapshot containing A + B


Reads remain lock-free.

One important detail: the public PreviewSet() and Apply() methods should not be treated as a transaction by themselves. The strategy-level operations must also serialize their read/preview/remote-commit/local-apply lifecycle, or use revision-based compare-and-swap.

Otherwise, the race simply moves into CloudStrategy.Set().

I would use a mutex first rather than a CAS loop. Configuration updates are relatively infrequent, and correctness is more important than optimizing write contention.

5. CloudStrategy.Set(): a problem with remote synchronization
This is another important finding.

Consider the current method:
func (s *CloudStrategy) Set(
    cfg *core.Config,
    updates map[string]map[string]string,
) error {

    nextState := cfg.PreviewSet(updates)

    if err := s.Client.UpdateConfigMap(nextState); err != nil {
        return err
    }

    cfg.Apply(nextState)
    return nil
}


At first glance, this looks like a transactional operation:
1. Calculate the next state.
2. Push it to the authoritative server.
3. Apply locally if successful.
But the network method being called is not actually waiting for server acceptance.
A. UpdateConfigMap only sends the message
Inside network/client.go:
func (c *Client) UpdateConfigMap(
    m *map[string]map[string]string,
) error {

    data, err := c.Handler.HandleOutgoing(
        pb.ConfigMsg_PUT_SYNC,
        m,
    )

    if err != nil {
        return err
    }

    if sock != nil {
        return sock.Send(data)
    }

    return nil
}


It returns success when the socket accepts the outgoing data.

It does not wait for the server's ACK or ERROR response.

The client has a separate sendWithUpdateAck() method that does wait for the response.

This means CloudStrategy.Set() can apply a local configuration that the server has subsequently rejected.
Important consistency issue
Source: cloud.go and client.go.
GitHub+2

B. How I would correct it
Use an acknowledged update operation:
if err := s.Client.UpdateConfigMapWithOCC(
    updates,
    expectedVersion,
); err != nil {
    return err
}


Then commit locally only after acceptance.

However, I would go further.

The ideal operation should return the committed server revision:
type UpdateResult struct {
    Revision uint64
}


Then the local store can associate its snapshot with the exact server revision that accepted it.

This also makes reconnection and recovery much easier.
C. Another subtle issue: sending the entire snapshot
CloudStrategy.Set() calculates the complete next state and sends it to the server, rather than sending only the delta.

This is not inherently wrong, but it introduces unnecessary coupling.

Imagine:
Current configuration: 1000 keys
Actual update: 1 key


The client serializes and transmits all 1000 keys.

That means:
- Larger network messages.
- More JSON serialization.
- More server-side processing.
- More opportunities for concurrent updates to overwrite one another.
I would send the delta to the server and let the server perform the authoritative merge.

Then the server returns the new revision.

This makes the server the single authority over shared configuration.

6. Ordering and revision handling: an issue in the protobuf handler
This is particularly important because config-server already has a revision mechanism.

The client tracks:
lastSeenVersion uint64


And checks incoming broadcasts:
if uint64(msg.Version) > h.lastSeenVersion+1 {
    // Sequence gap detected
}


When a gap is detected, it triggers a FULL_REFRESH.

This is a good concept.

But I found two important problems.
A. Older messages can move the revision backwards
The current implementation does:
if received > lastSeenVersion+1 {
    // Gap
} else {
    lastSeenVersion = received
}


There is no explicit handling of:
received < lastSeenVersion


For example:
Last seen: 105
Incoming:  104


The code enters the else branch and sets:
lastSeenVersion = 104


The client has now moved backwards.

It can subsequently accept stale updates that should have been rejected.

Correction:
if received <= lastSeenVersion {
    // Ignore duplicate or stale update
    return nil
}

if received > lastSeenVersion+1 {
    // Trigger full refresh
}


Then advance the revision only after the relevant update has been safely applied.
B. A full refresh can overwrite a newer update
This is more subtle.

Imagine:
Client at revision 100

Receives revision 103
    |
    v
Detects gap
    |
    v
Starts FULL_REFRESH request

Receives revision 104
    |
    v
Applies delta 104

FULL_REFRESH response arrives with revision 103
    |
    v
updateLiveConfig() replaces entire snapshot


The client can end up with revision 103 after having already applied revision 104.

The current updateLiveConfig() performs an unconditional atomic pointer swap:
h.parentConfig.LiveConfig.Store(&sections)


It does not check whether the snapshot is older than the currently applied state.
Potential state rollback

GitHub+1
C. My recommended revision protocol
I would make every update carry a revision and use one common application function:
func (c *Config) ApplySnapshot(
    revision uint64,
    data map[string]map[string]string,
) bool


Its responsibilities:
1. Serialize against other writers.
2. Reject revisions older than the current snapshot.
3. Ignore duplicate revisions.
4. Apply the complete snapshot.
5. Publish the new revision and map consistently.
A revision and map should ideally be published together in one immutable snapshot:
type LiveSnapshot struct {
    Revision uint64
    Data     map[string]map[string]string
}


This is cleaner than maintaining a separate atomic pointer and revision variable.

One atomic pointer swap then represents one complete version of the configuration.

That is the model I would strongly recommend.

7. Configuration discovery: path resolution is more complicated than necessary
I like the intention behind ResolveConfigPath().

It attempts to support:
- CONFIG_PATH.
- SHARED_CONFIG_PATH.
- Executable directory.
- Current working directory.
- Source-code directory when using go run.
- Project-root discovery using .git, go.mod, Cargo.toml and pyproject.toml.
- Profile-specific YAML files.
- standalone.yaml fallback.
This is quite flexible.

However, I think the implementation has accumulated too many discovery rules.
A. The SHARED_CONFIG_PATH ambiguity
The README describes SHARED_CONFIG_PATH as accepting a configuration file or root directory.

But the implementation does:
if customPath := os.Getenv("SHARED_CONFIG_PATH"); customPath != "" {
    if _, err := os.Stat(customPath); err == nil {
        return customPath
    }
}


It returns the path directly.

If that path is a directory, LoadYAML() subsequently attempts to read it as a file.

That will fail.

I would define the contract explicitly:
CONFIG_PATH
    = exact YAML file

SHARED_CONFIG_PATH
    = directory containing profile YAML files


Then resolve the profile filename inside the shared directory.
B. Invalid explicit paths should not silently fall back
Consider:
CONFIG_PATH=/etc/myapp/production.yaml


If the file does not exist, the resolver continues searching other directories.

That is potentially dangerous.

You might think you are loading the production configuration from /etc/myapp/production.yaml, while the application actually loads another file.

For an explicitly configured path, I would fail immediately with a clear error.

Automatic discovery should only happen when no explicit path was supplied.
C. Caller-stack discovery is expensive conceptually
The getCallerDir() approach uses runtime.Caller() to identify the original application's source directory.

It is clever, especially for go run.

But it is also quite difficult to reason about and test across:
- Different build systems.
- Plugins.
- Generated binaries.
- Test execution.
- Symlinked executables.
- Different deployment layouts.
I would simplify the resolution order:
1. Explicit configuration file.
2. Explicit shared configuration directory.
3. Executable directory.
4. Working directory.
5. Clearly defined fallback.
I would retain caller discovery only if a concrete deployment requirement cannot be solved otherwise.

GitHub+3
8. YAML merging and environment expansion
There are two interesting mechanisms here.
A. Deep merging of capabilities
LoadYAML() preserves existing capabilities and deep-merges new YAML values.

This is useful because the configuration server can hydrate capabilities before the local YAML is reloaded.

For example:

Initial server-provided state:
capabilities:
  database:
    ip: db.internal
    port: "5432"
    pool: "20"


Local YAML:
capabilities:
  database:
    pool: "50"


Expected merged result:
capabilities:
  database:
    ip: db.internal
    port: "5432"
    pool: "50"


This is a good pattern.

However, it reinforces the importance of defining exactly when a configuration is being merged versus replaced.
B. The environment expansion engine
The current implementation uses:
var envRegex = regexp.MustCompile(
    `\$\{([^}]+)\}`,
)


And supports:
database:
  password: "${DB_PASSWORD}"
  host: "${DB_HOST:localhost}"


This is a lightweight solution.

There are a few details I would improve:

1. Missing variables silently become empty strings.

For example:
password: "${DB_PASSWORD}"


If the variable is absent, the result becomes an empty string.

For mandatory configuration values, that can create confusing failures much later.

I would support an explicit strict mode:
password: "${DB_PASSWORD:?required}"


Or make mandatory variables part of validation.

2. Boolean handling is special-cased.

The implementation preserves true and false as booleans and converts other scalar values to strings.

That is understandable given the string-based configuration model, but it should be documented because numeric YAML values will not retain their native numeric types.

3. The expansion parser is simplistic.

It does not provide a proper escaping mechanism for literal ${...} expressions, and nested/default expressions are not supported.

I would not introduce a large templating engine unnecessarily, but I would make the supported syntax explicit and test it thoroughly.

GitHub
9. Secret management: good foundation, but I would tighten the failure behavior
The repository uses RSA-OAEP with SHA-256.

The implementation:
- Supports PEM public and private keys.
- Uses RSA encryption.
- Encodes ciphertext as Base64.
- Wraps encrypted values in ENC(...).
- Provides decryption helpers.
This is a reasonable approach for distributing encrypted configuration values.

However, there are some important details.
A. Missing private key silently bypasses processing
In ProcessConfigSecrets():
_, err := getPrivateKey()

if err != nil {
    return content, nil
}


If a private key is missing, the function returns the original content without reporting an error.

Likewise, when individual decryption fails, it writes a warning to stderr and retains the original ENC(...) value.

This may be acceptable for an optional decryption helper, but it is dangerous if a caller assumes that all secrets have been decrypted successfully.

I would introduce explicit policies:
Mode	Missing key / decryption failure
Optional	Keep encrypted value, report status
Strict	Return error and reject configuration
Audit	Report all unresolved encrypted values

For production configuration, I would use strict mode.
B. Private key lookup
The private key is searched in:
1. BASTIEN_PRIVATE_KEY.
2. BASTIEN_PRIVATE_KEY_PATH.
3. /etc/bastien/private.pem.
4. ./private.pem fallback.
I would remove the working-directory fallback in production.

The current working directory is not a reliable security boundary. A process launched from an unexpected directory could pick up an unintended private key.
C. Repeated private-key parsing
Every decryption calls getPrivateKey(), which reads and parses the PEM file again.

This is unnecessary for a long-running service.

I would load and validate the key once during initialization, retain the parsed key in a dedicated secret manager and provide an explicit reload operation if key rotation is required.

The TODO already identifies this optimization.

One further architectural point: RSA-OAEP is suitable for relatively short secrets, not arbitrary large configuration documents. If the system later needs to encrypt larger payloads, use envelope encryption with a symmetric authenticated cipher such as AES-GCM.

GitHub+2

10. Network client: reconnection and synchronization
The networking implementation is more sophisticated than a basic TCP client.

It includes:
- Safe-Socket framing.
- Handshake identity.
- Background watching.
- Exponential backoff.
- Reconnection.
- Initial synchronization.
- Full refresh.
- A single-reader synchronization mechanism.
- Acknowledged OCC operations.
I particularly like the use of syncMu to serialize request/response operations while the watcher is responsible for reading incoming frames.

This avoids multiple goroutines competing to read the same socket.
A. One weakness: non-OCC operations don't wait for acknowledgement
We already discussed this with CloudStrategy.Set().

It also affects:
DeleteKeys()


which sends a request without waiting for the server's response.

I would make the behavior consistent:
Set()
    → server acknowledgement
    → local commit

Delete()
    → server acknowledgement
    → local commit


No successful return should imply authoritative acceptance if only a local socket write occurred.
B. Reconnection and lifecycle
The Watch() loop reconnects when the socket becomes unavailable.

It uses exponential backoff and resynchronizes after reconnection.

That is good.

However, there is a possible lifecycle race:
1. The watcher detects a disconnected socket.
2. It starts reconnecting.
3. Another goroutine calls Close().
4. The reconnect operation finishes and assigns a newly opened socket.
The watcher may then exit because quit is closed, leaving the newly opened socket without an owner.

I would serialize connection creation and shutdown, and check the closed state before publishing a newly connected socket.

Also, a reconnect should always be followed by a full snapshot synchronization, not merely an assumption that the last local state remains current.

The current code sends GET_SYNC after reconnection, which is the right direction.

GitHub+3
C. Callback execution
HandleIncoming() invokes registered callbacks directly from the incoming-message processing path.

For example:
if liveCb != nil {
    liveCb(parsed)
}


This is simple, but it means a slow or blocked application callback can delay processing of subsequent network messages.

For a high-frequency configuration service, I would either:
- Document that callbacks must be fast and non-blocking; or
- Dispatch callbacks through a controlled notification mechanism.
I would not introduce an unrestricted goroutine per callback, because that would create another ordering and lifecycle problem.

11. A protocol-level issue: uint64 versus int32 revisions
There is a small but important inconsistency in proto_handler.go.

The public API uses:
expectedVersion uint64


And the client tracks:
lastSeenVersion uint64


But when building the protobuf message:
Version: int32(version),


This converts a 64-bit unsigned revision into a signed 32-bit integer.

The maximum positive value is:
\[
2^{31}-1 = 2,147,483,647
\]

Beyond that, the conversion can produce negative values.

The receiver then has code such as:
if msg.Version > 0 {
    // Update revision
}


A negative version will therefore no longer be processed correctly.

This is unlikely to happen quickly in a configuration system, but it is an avoidable inconsistency.

Recommendation: use uint64 consistently in the protobuf schema, regenerate the bindings, and preserve that type throughout the client and server.

GitHub+2
12. The polyglot architecture: ambitious, but needs strict ownership rules
This is one of the most interesting aspects of the project.

The library is not limited to Go.

It exposes a C ABI through libdistconf, allowing Python, Rust and C/C++ applications to use the same configuration engine.

I particularly like this direction because it avoids independently implementing configuration precedence and synchronization in every language.

However, a C ABI introduces responsibilities that ordinary Go applications don't have.
Ownership boundaries
Go runtime

Configuration snapshots, networking, callbacks


C ABI

Handles, pointers, strings, allocation



Python / Rust

Foreign-memory ownership and lifecycle


C / C++

ABI compatibility and thread safety




I would insist on four rules:
1. Every exported handle has an explicit lifecycle.
2. Every returned allocated string has a documented release function.
3. Closing a handle cannot race with an ongoing operation.
4. Callbacks must have explicit threading and reentrancy guarantees.
The TODO already identifies thread-local error reporting and strdup memory-leak concerns.

I would also maintain a small ABI compatibility test suite that verifies exported symbols and function signatures across releases.

This is especially important because the C ABI is intended to become a shared ecosystem-wide interface.

Bastien-Antigravity/distributed-config · GitHub+2

13. My proposed target architecture
I would keep the existing architecture but make configuration resolution and mutation more centralized.
Configuration sources

Defaults · YAML · Environment · Config Server


Config Loader

Path discovery · expansion · parsing · validation


Config Resolver

One authoritative merge and precedence implementation


Versioned Immutable Snapshot

Revision + static configuration + live configuration

Read API

Lock-free access


Write API

Serialized updates · OCC




Strategy adapters

Standalone · Test · Staging · Production

Go API


C ABI / FFI




What I would retain
- Go as the configuration engine.
- Lock-free atomic snapshot reads.
- Copy-On-Write.
- Profile-based strategies.
- YAML and environment expansion.
- Safe-Socket integration.
- Revision-aware synchronization.
- Optimistic concurrency control.
- RSA-OAEP secret support.
- C ABI for other languages.
What I would simplify
- Multiple independent capability merge implementations.
- Unclear path-discovery fallback rules.
- Separate revision and configuration state.
- Unprotected read-modify-write operations.
- Fire-and-forget remote mutations.
- Repeated full-configuration transmission.
- Implicit secret-decryption failure handling.

14. Implementation priorities
I would approach this repository in five phases.
Phase 1 — Data consistency
Priority 1

- [ ] Protect Set(), Delete() and Apply() with a coherent writer synchronization strategy.
- [ ] Introduce LiveSnapshot{Revision, Data}.
- [ ] Reject stale broadcasts.
- [ ] Prevent stale FULL_REFRESH responses from rolling back newer state.
- [ ] Serialize strategy-level preview/commit operations.


  Phase 2 — Remote synchronization
  Priority 2
- [ ] Make normal Set operations wait for acknowledgement.
- [ ] Make Delete operations wait for acknowledgement.
- [ ] Return the committed server revision.
- [ ] Send deltas instead of unnecessary full snapshots.
- [ ] Define offline update semantics.


  Phase 3 — Configuration resolution
  Priority 3
- [ ] Centralize capability merging.
- [ ] Fix dynamically introduced capabilities.
- [ ] Formalize precedence rules.
- [ ] Fix explicit path failure behavior.
- [ ] Define SHARED_CONFIG_PATH as a directory or file, not ambiguously both.


  Phase 4 — Security and lifecycle
  Priority 4
- [ ] Make secret resolution fail explicitly in strict mode.
- [ ] Remove private-key working-directory fallback in production.
- [ ] Cache parsed private keys safely.
- [ ] Fix reconnect versus Close lifecycle races.
- [ ] Audit FFI allocation and handle ownership.


  Phase 5 — Performance and validation
  Priority 5
- [ ] Benchmark lock-free reads.
- [ ] Benchmark concurrent writes.
- [ ] Measure YAML and JSON allocation costs.
- [ ] Benchmark FFI overhead.
- [ ] Test all profiles and cross-language interfaces.

15. Tests I would specifically add
Test	Expected behavior
100 concurrent Set() operations	No lost updates
Concurrent Set() and Delete()	Consistent final state
Stale broadcast revision	Ignored
Delayed FULL_REFRESH response	Cannot roll back newer state
Concurrent OCC updates	Exactly one conflicting revision accepted
Server rejects update	Local configuration unchanged
Server disconnects during update	Explicit uncertain/failed result
New remote capability	Correctly resolvable
YAML plus server override	Documented precedence
Missing CONFIG_PATH	Explicit error
Missing mandatory environment variable	Validation failure
Missing private key in strict mode	Configuration rejected
Concurrent Close and reconnect	No socket leak
C ABI concurrent Close/Get	No use-after-free
Repeated FFI operations	No memory growth

For concurrency, I would make these mandatory:
go test -race ./...
go vet ./...


I would also add property-based tests for configuration merging.

For example, the important invariant is:
Apply(A)
Apply(B)


must preserve both independent updates regardless of their execution interleaving.

16. Overall assessment
8.2 / 10
Architectural potential

Qualitative code-level assessment, not a measured benchmark or completed test-suite result.
Area	Assessment
Overall architecture	Very good
Configuration discovery	Good, too complex
Merge/precedence model	Good, needs consolidation
Atomic reads	Very good
Concurrent writes	Needs correction
Remote synchronization	Needs correction
Reconnection	Good, lifecycle refinement needed
Secret management	Good foundation
FFI architecture	Ambitious, needs strict ownership
Maintainability	Good potential


What I particularly appreciate is that the developer is thinking beyond a simple configuration loader.

This is an attempt to build a genuine configuration infrastructure layer that can serve a heterogeneous microservice ecosystem.

The use of a common Go engine with a C ABI is a particularly interesting architectural decision.

The main problem is that the project is starting to accumulate several different mechanisms for solving similar problems:
- Multiple merge paths.
- Multiple ways to commit configuration.
- Multiple synchronization semantics.
- Multiple representations of configuration revisions.
The result is a system that is conceptually powerful but more complicated than it needs to be.

I would not rewrite it. I would consolidate its state management and make its consistency guarantees explicit.

17. The most important finding: the three repositories must agree
Now that we have examined log-server, config-server and distributed-config, I think there is a larger architectural issue worth addressing.

These components form a dependency chain:
distributed-config

Loads, merges and applies configuration


config-server

Owns and distributes shared configuration


log-server

Receives, orders and persists operational events


There are four cross-repository contracts I would standardize.
Contract	What must be consistent
Configuration revisions	Same revision semantics on client and server
Merge operations	PATCH, REPLACE and DELETE behave identically
Acknowledgements	Accepted versus persisted versus applied
Lifecycle	Disconnect, reconnect, shutdown and recovery

There is also a subtle issue with logging: if configuration changes affect the logging endpoint, the client must not lose its ability to report errors during the transition.

I would therefore keep a small independent bootstrap logging mechanism rather than make the entire logging system dependent on successful configuration synchronization.
My final recommendation
I would prioritize the repositories in this order:
1. distributed-config: fix concurrent state updates and revision handling.
2. config-server: make publication and persistence revision-aware.
3. log-server: simplify ordering and ensure one consistent persistence pipeline.
This order addresses the shared state contract first, then the services that depend on it.

One important qualification: I reviewed the current public develop source and documentation, but did not clone the repositories or execute their test suites. The findings above are code-level findings; runtime confirmation should be the next step.

My preferred next step would be a cross-repository consistency audit of these three projects, producing one shared technical specification for revisions, merging, acknowledgements and failure recovery. That would eliminate the risk of each repository implementing a slightly different interpretation of the same distributed-system contract.