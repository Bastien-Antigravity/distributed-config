GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

# Determine OS and Library Extension
OS=$(shell go env GOOS)
ifeq ($(OS),windows)
	LIB_EXT=dll
else ifeq ($(OS),darwin)
	LIB_EXT=dylib
else
	LIB_EXT=so
endif

.PHONY: all build clean test build-lib build-dll build-all test-sdk

all: build

build: build-lib
	mkdir -p bin
	$(GOBUILD) -o bin/ ./cmd/...

build-lib:
	mkdir -p distconf/libdistconf
	# Use dynamic extension based on host OS
	$(GOBUILD) -buildmode=c-shared -o distconf/libdistconf/libdistconf.$(LIB_EXT) ./cmd/libdistconf
	# For macOS, set the install_name to @rpath to allow relative loading via LC_RPATH
	@if [ "$(LIB_EXT)" = "dylib" ]; then \
		install_name_tool -id @rpath/libdistconf.dylib distconf/libdistconf/libdistconf.dylib; \
		cp distconf/libdistconf/libdistconf.dylib distconf/libdistconf/libdistconf.so || true; \
	fi

build-dll:
	mkdir -p distconf/libdistconf
	# Requires mingw-w64 if cross-compiling from macOS/Linux
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) -buildmode=c-shared -o distconf/libdistconf/libdistconf.dll ./cmd/libdistconf

build-all: build build-lib build-dll

clean:
	$(GOCLEAN)
	rm -rf bin/
	rm -rf release/
	rm -f distconf/libdistconf/libdistconf.so
	rm -f distconf/libdistconf/libdistconf.dylib
	rm -f distconf/libdistconf/libdistconf.h
	rm -f distconf/libdistconf/libdistconf.dll
	# Purge transient test artifacts
	rm -f distconf-*.yaml
	rm -f Python.yaml
	rm -f *.yaml.bak

test:
	$(GOTEST) -v ./...

test-sdk: build-lib
	@echo "--- Running Polyglot SDK Tests ---"
	# Python
	@echo "Testing Python..."
	cd distconf/python && python3 ffi_validation.py && python3 -m unittest discover tests
	# Rust
	@echo "Testing Rust..."
	cd distconf/rust && cargo run --example ffi_validation && cargo test
	# C++
	@echo "Testing C++..."
	cd distconf/cpp && g++ -std=c++11 examples/ffi_validation.cpp -ldl -o ffi_val && ./ffi_val && rm ffi_val

