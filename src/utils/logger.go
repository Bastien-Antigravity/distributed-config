package utils

// =============================================================================
// ESSENTIAL PROCESS:
// Standardized logging abstraction compatible with universal-logger and
// microservice-toolbox, providing fallback no-op and console sinks.
//
// DATA FLOW:
// 1. Input: Log messages, formats, and arbitrary arguments across severity levels.
// 2. Logic: Routes log calls to active Logger implementation or default stderr sink.
// 3. Output: Formatted diagnostic messages to console or external log streams.
//
// KEY PARAMETERS:
// - Logger: Interface defining standard fleet log methods.
// =============================================================================

import (
	"os"
)

// -----------------------------------------------------------------------------
type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warning(format string, args ...any)
	Error(format string, args ...any)
	Critical(format string, args ...any)
	Logon(format string, args ...any)
	Logout(format string, args ...any)
	Trade(format string, args ...any)
	Schedule(format string, args ...any)
	Report(format string, args ...any)
	Stream(format string, args ...any)
}

// EnsureSafeLogger provides a no-op implementation fallback.
// In strict mode (STRICT_LOGGER=true), it panics immediately.
func EnsureSafeLogger(log Logger) Logger {
	if log == nil {
		if os.Getenv("STRICT_LOGGER") == "true" {
			panic("CRITICAL ARCHITECTURAL VIOLATION: Nil logger passed to EnsureSafeLogger in strict mode.")
		}
		return &noOpLogger{}
	}
	return log
}

type noOpLogger struct{}

func (n *noOpLogger) Debug(string, ...any)    {}
func (n *noOpLogger) Info(string, ...any)     {}
func (n *noOpLogger) Warning(string, ...any)  {}
func (n *noOpLogger) Error(string, ...any)            {}
func (n *noOpLogger) Critical(string, ...any)         {}
func (n *noOpLogger) Logon(string, ...any)            {}
func (n *noOpLogger) Logout(string, ...any)           {}
func (n *noOpLogger) Trade(string, ...any)            {}
func (n *noOpLogger) Schedule(string, ...any)         {}
func (n *noOpLogger) Report(string, ...any)           {}
func (n *noOpLogger) Stream(string, ...any)           {}
