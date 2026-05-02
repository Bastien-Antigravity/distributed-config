package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"github.com/Bastien-Antigravity/distributed-config"
)

// -------------------------------------------------------------------------

// Decrypt is a Go-native wrapper for distributed_config.Decrypt.
func Decrypt(ciphertext string) (string, error) {
	return distributed_config.Decrypt(sanitizeString(ciphertext))
}
