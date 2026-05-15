CXX = g++
CXXFLAGS = -std=c++17 -Wall -Wextra -I. -I../libdistconf
LDFLAGS = -L../libdistconf -ldistconf -Wl,-rpath,../libdistconf

# If DISTCONF_DIR is set, use it for headers and lib
ifneq ($(DISTCONF_DIR),)
    CXXFLAGS += -I$(DISTCONF_DIR)
    LDFLAGS = -L$(DISTCONF_DIR) -ldistconf -Wl,-rpath,$(DISTCONF_DIR)
endif

TEST_SRC = tests/test_distconf.cpp
TEST_BIN = bin/test_distconf

.PHONY: all clean test build-lib

all: build-lib $(TEST_BIN)

# Attempt to build the library if it's missing and we are in a Go environment
build-lib:
	@if [ ! -f ../libdistconf/libdistconf.so ] && [ ! -f ../libdistconf/libdistconf.dylib ] && [ -f ../../Makefile ]; then \
		echo "Shared library missing, attempting to build..."; \
		cd ../.. && make build-lib; \
	fi

$(TEST_BIN): $(TEST_SRC)
	mkdir -p bin
	$(CXX) $(CXXFLAGS) $(TEST_SRC) -o $(TEST_BIN) $(LDFLAGS)

test: $(TEST_BIN)
	./$(TEST_BIN)

clean:
	rm -rf bin/
