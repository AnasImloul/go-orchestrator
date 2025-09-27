package logger

// Logger represents a logging interface
type Logger interface {
	// Debug logs a debug message
	Debug(msg string, args ...interface{})

	// Info logs an info message
	Info(msg string, args ...interface{})

	// Warn logs a warning message
	Warn(msg string, args ...interface{})

	// Error logs an error message
	Error(msg string, args ...interface{})

	// WithComponent returns a new logger with the component field added
	WithComponent(component string) Logger
}
