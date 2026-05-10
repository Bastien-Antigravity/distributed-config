CXX = g++
CXXFLAGS = -std=c++17 -Wall -Wextra -I. -I../libdistconf
LDFLAGS = -L../libdistconf -ldistconf -Wl,-rpath,../libdistconf

TEST_SRC = tests/test_distconf.cpp
TEST_BIN = bin/test_distconf

.PHONY: all clean test

all: $(TEST_BIN)

$(TEST_BIN): $(TEST_SRC)
	mkdir -p bin
	$(CXX) $(CXXFLAGS) $(TEST_SRC) -o $(TEST_BIN) $(LDFLAGS)

test: $(TEST_BIN)
	./$(TEST_BIN)

clean:
	rm -rf bin/
