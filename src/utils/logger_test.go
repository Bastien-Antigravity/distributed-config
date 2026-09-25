package utils
 
 /*
 ESSENTIAL PROCESS:
 Unit tests for EnsureSafeLogger in distributed-config.
 Verifies safe fallback and strict mode enforcement via STRICT_LOGGER.
 
 DATA FLOW:
 Injected Logger -> EnsureSafeLogger() -> Validated Logger instance
 
 KEY PARAMETERS:
 - STRICT_LOGGER: Environment variable toggling fail-fast panic.
 */
 
 import (
 	"os"
 	"testing"
 
 	"github.com/stretchr/testify/assert"
 )
 
 // -----------------------------------------------------------------------------
 
 func TestEnsureSafeLogger_NormalFallback(t *testing.T) {
 	_ = os.Unsetenv("STRICT_LOGGER")
 	l := EnsureSafeLogger(nil)
 	assert.NotNil(t, l)
 	// Must not panic on method calls
 	l.Info("test info")
 	l.Debug("test debug")
 	l.Warning("test warning")
 	l.Error("test error")
 	l.Critical("test critical")
 }
 
 // -----------------------------------------------------------------------------
 
 func TestEnsureSafeLogger_StrictModePanics(t *testing.T) {
 	_ = os.Setenv("STRICT_LOGGER", "true")
 	defer func() { _ = os.Unsetenv("STRICT_LOGGER") }()
 
 	assert.Panics(t, func() {
 		EnsureSafeLogger(nil)
 	}, "EnsureSafeLogger must panic on nil logger when STRICT_LOGGER=true")
 }
