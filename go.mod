module github.com/Bastien-Antigravity/distributed-config

go 1.25.4

require (
	github.com/Bastien-Antigravity/safe-socket v0.0.1
	github.com/spf13/pflag v1.0.10
	google.golang.org/protobuf v1.36.12
	gopkg.in/yaml.v3 v3.0.1
)

require (
	capnproto.org/go/capnp/v3 v3.1.0-alpha.2 // indirect
	github.com/colega/zeropool v0.0.0-20230505084239-6fb4a4f75381 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/edsrzf/mmap-go v1.2.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace (
	github.com/Bastien-Antigravity/flexible-logger => ../flexible-logger
	github.com/Bastien-Antigravity/microservice-toolbox => ../microservice-toolbox
	github.com/Bastien-Antigravity/safe-socket => ../safe-socket
	github.com/Bastien-Antigravity/universal-logger => ../universal-logger
)
