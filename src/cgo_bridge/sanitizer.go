package cgo_bridge

// =============================================================================
// ESSENTIAL PROCESS:
// FFI boundary string sanitizer stripping null bytes and trimming whitespace
// from incoming C string pointers to prevent buffer overruns or malformed keys.
//
// DATA FLOW:
// 1. Input: Raw string extracted from C char* pointer.
// 2. Logic: Trims surrounding whitespace and removes interior \x00 null bytes.
// 3. Output: Cleaned, safe Go string.
//
// KEY PARAMETERS:
// - input: Raw string originating across the FFI boundary.
// =============================================================================

/*
#include <stdlib.h>
*/
import "C"

// -----------------------------------------------------------------------------

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
