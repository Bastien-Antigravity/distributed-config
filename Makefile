GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

.PHONY: all build clean test

all: build build-lib

build:
	mkdir -p bin
	$(GOBUILD) -o bin/ ./cmd/...

build-lib:
	mkdir -p distconf/libdistconf
	# For macOS we use .dylib, for others we use .so
	# go build generates the .h file automatically with -buildmode=c-shared
	$(GOBUILD) -buildmode=c-shared -o distconf/libdistconf/libdistconf.so ./cmd/libdistconf
	# On macOS, rename to .dylib for clarity if needed, but .so works for many loaders
	cp distconf/libdistconf/libdistconf.so distconf/libdistconf/libdistconf.dylib || true

build-dll:
	mkdir -p distconf/libdistconf
	# Requires mingw-w64 if cross-compiling from macOS/Linux
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) -buildmode=c-shared -o distconf/libdistconf/libdistconf.dll ./cmd/libdistconf

build-all: build build-lib build-dll

clean:
	$(GOCLEAN)
	rm -rf bin/
	rm -rf release/

test:
	$(GOTEST) -v ./...

