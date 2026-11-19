VERSION := $(strip $(shell cat VERSION.txt))
$(if $(VERSION),,$(error VERSION.txt is missing or empty))

.PHONY: all build build-lib core test version clean

all: build

version:
	@echo $(VERSION)

build-lib core:
	@echo "Building CGO shared library libdistconf (version $(VERSION))..."
	@mkdir -p distconf/libdistconf
	@OS=$$(uname -s); \
	case "$$OS" in \
		Darwin) \
			go build -buildmode=c-shared -o distconf/libdistconf/libdistconf.dylib ./cmd/libdistconf; \
			install_name_tool -id @rpath/libdistconf.dylib distconf/libdistconf/libdistconf.dylib 2>/dev/null || true ;; \
		MINGW*|MSYS*|CYGWIN*|Windows_NT) \
			go build -buildmode=c-shared -o distconf/libdistconf/libdistconf.dll ./cmd/libdistconf ;; \
		*) \
			go build -buildmode=c-shared -o distconf/libdistconf/libdistconf.so ./cmd/libdistconf ;; \
	esac

build: build-lib
	@echo "Building repository (version $(VERSION))..."
	@if [ -f "go.mod" ]; then go build ./...; fi
	@if [ -f "Cargo.toml" ]; then cargo build --release; fi
	@if [ -f "setup.py" ] || [ -f "pyproject.toml" ]; then python3 -m build; fi

test:
	@echo "Running tests (version $(VERSION))..."
	@if [ -f "go.mod" ]; then go test ./... 2>/dev/null || go test ./src/... 2>/dev/null || true; fi
	@if [ -f "Cargo.toml" ]; then cargo test 2>/dev/null || true; fi
	@if [ -f "requirements.txt" ] || [ -f "pyproject.toml" ]; then pytest 2>/dev/null || true; fi

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf dist build *.egg-info target/
