# Determine OS and Library Extension
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	LIB_EXT = dylib
	RPATH = -Wl,-rpath,@loader_path/../../libdistconf
else ifeq ($(UNAME_S),Linux)
	LIB_EXT = so
	RPATH = -Wl,-rpath,'$$ORIGIN/../../libdistconf'
else
	LIB_EXT = dll
	RPATH = 
endif

LIB_PATH = ../libdistconf/libdistconf.$(LIB_EXT)

CXX = g++
CXXFLAGS = -std=c++17 -Wall -Wextra -I. -I../libdistconf
LDFLAGS = -L../libdistconf -ldistconf $(RPATH)

# If DISTCONF_DIR is set, use it for headers and lib
ifneq ($(DISTCONF_DIR),)
    CXXFLAGS += -I$(DISTCONF_DIR)
    LDFLAGS = -L$(DISTCONF_DIR) -ldistconf -Wl,-rpath,$(DISTCONF_DIR)
endif

TEST_SRC = tests/test_distconf.cpp
TEST_BIN = bin/test_distconf

.PHONY: all clean test

all: $(TEST_BIN)

# Dependency on the library file
$(LIB_PATH):
	@if [ -f ../../Makefile ]; then \
		echo "Shared library missing or outdated, attempting to build..."; \
		$(MAKE) -C ../.. build-lib; \
	else \
		echo "Warning: Root Makefile not found, cannot build library automatically."; \
	fi

$(TEST_BIN): $(TEST_SRC) $(LIB_PATH)
	mkdir -p bin
	$(CXX) $(CXXFLAGS) $(TEST_SRC) -o $(TEST_BIN) $(LDFLAGS)

test: $(TEST_BIN)
	./$(TEST_BIN)

clean:
	rm -rf bin/
