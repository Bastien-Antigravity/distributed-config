package utils

import "fmt"

// Logger defines the interface for structured logging.
// This interface is structurally compatible with universal-logger and microservice-toolbox.
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
func EnsureSafeLogger(log Logger) Logger {
	if log == nil {
		return &noOpLogger{}
	}
	return log
}

type noOpLogger struct{}

func (n *noOpLogger) Debug(string, ...any)    {}
func (n *noOpLogger) Info(format string, args ...any) { fmt.Printf(format+"\n", args...) } // Keep basic console output if no logger provided for core config
func (n *noOpLogger) Warning(string, ...any)  {}
func (n *noOpLogger) Error(string, ...any)    {}
func (n *noOpLogger) Critical(string, ...any) {}
func (n *noOpLogger) Logon(string, ...any)    {}
func (n *noOpLogger) Logout(string, ...any)   {}
func (n *noOpLogger) Trade(string, ...any)    {}
func (n *noOpLogger) Schedule(string, ...any) {}
func (n *noOpLogger) Report(string, ...any)   {}
func (n *noOpLogger) Stream(string, ...any)   {}
