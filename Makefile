GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

.PHONY: all build clean test

all: build

build:
	mkdir -p bin
	$(GOBUILD) -o bin/ ./cmd/...

clean:
	$(GOCLEAN)
	rm -rf bin/

test:
	$(GOTEST) -v ./...

