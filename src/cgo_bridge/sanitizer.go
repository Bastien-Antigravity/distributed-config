package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"strings"
)

// sanitizeString ensures that strings coming from C are clean.
// It trims whitespace and removes any potential null-bytes or control characters
// that might have leaked through the FFI boundary.
func sanitizeString(input string) string {
	// 1. Trim whitespace
	s := strings.TrimSpace(input)

	// 2. Remove null bytes (common in C-string buffers)
	s = strings.ReplaceAll(s, "\x00", "")

	return s
}
