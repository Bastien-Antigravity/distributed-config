.PHONY: all build test build-lib clean

all: build

build: build-lib

# Attempt to build the library if it's missing
build-lib:
	@OS=$$(go env GOOS 2>/dev/null || uname -s | tr '[:upper:]' '[:lower:]'); \
	case $$OS in \
		darwin)  EXT="dylib" ;; \
		windows*) EXT="dll" ;; \
		*)       EXT="so" ;; \
	esac; \
	if [ ! -f ../libdistconf/libdistconf.$$EXT ] && [ -f ../../Makefile ]; then \
		echo "Shared library missing ($$EXT), attempting to build..."; \
		$(MAKE) -C ../.. build-lib; \
	fi

test: build-lib
	python3 tests/test_distconf.py


clean:
	rm -rf __pycache__ tests/__pycache__
