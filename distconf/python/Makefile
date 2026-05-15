.PHONY: all test build-lib clean

all: test

# Attempt to build the library if it's missing
build-lib:
	@if [ ! -f ../libdistconf/libdistconf.so ] && [ ! -f ../libdistconf/libdistconf.dylib ] && [ -f ../../Makefile ]; then \
		echo "Shared library missing, attempting to build..."; \
		$(MAKE) -C ../.. build-lib; \
	fi

test: build-lib
	python3 tests/test_distconf.py

clean:
	rm -rf __pycache__ tests/__pycache__
