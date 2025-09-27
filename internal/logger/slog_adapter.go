package logger

import (
	"log/slog"
)

// SlogAdapter adapts slog.Logger to the logger.Logger interface
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new slog adapter
func NewSlogAdapter(logger *slog.Logger) *SlogAdapter {
	return &SlogAdapter{logger: logger}
}

// Debug logs a debug message
func (a *SlogAdapter) Debug(msg string, args ...interface{}) {
	a.logger.Debug(msg, args...)
}

// Info logs an info message
func (a *SlogAdapter) Info(msg string, args ...interface{}) {
	a.logger.Info(msg, args...)
}

// Warn logs a warning message
func (a *SlogAdapter) Warn(msg string, args ...interface{}) {
	a.logger.Warn(msg, args...)
}

// Error logs an error message
func (a *SlogAdapter) Error(msg string, args ...interface{}) {
	a.logger.Error(msg, args...)
}

// WithComponent returns a new logger with a component context
func (a *SlogAdapter) WithComponent(component string) Logger {
	return &SlogAdapter{
		logger: a.logger.With("component", component),
	}
}
