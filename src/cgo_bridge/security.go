package cgo_bridge

// =============================================================================
// ESSENTIAL PROCESS:
// CGO secret decryption bridge allowing non-Go callers to decrypt inline ENC(...)
// configuration tokens via RSA-OAEP.
//
// DATA FLOW:
// 1. Input: Ciphertext string (ENC(...) format) originating across C ABI.
// 2. Logic: Sanitizes input string and delegates to distributed_config.Decrypt.
// 3. Output: Plaintext secret string or decryption error.
//
// KEY PARAMETERS:
// - ciphertext: Encrypted token string.
// =============================================================================

/*
#include <stdlib.h>
*/
import "C"

// -----------------------------------------------------------------------------

import (
	"github.com/Bastien-Antigravity/distributed-config"
)

// -------------------------------------------------------------------------

// Decrypt is a Go-native wrapper for distributed_config.Decrypt.
func Decrypt(ciphertext string) (string, error) {
	return distributed_config.Decrypt(sanitizeString(ciphertext))
}
