package logger

import (
	"log/slog"
)

// SlogAdapter implements the Logger interface using slog.
type SlogAdapter struct {
	slog *slog.Logger
}

// NewSlogAdapter creates a new SlogAdapter.
func NewSlogAdapter(slogLogger *slog.Logger) *SlogAdapter {
	return &SlogAdapter{slog: slogLogger}
}

// Debug logs a debug message.
func (s *SlogAdapter) Debug(msg string, args ...interface{}) {
	s.slog.Debug(msg, args...)
}

// Info logs an info message.
func (s *SlogAdapter) Info(msg string, args ...interface{}) {
	s.slog.Info(msg, args...)
}

// Warn logs a warning message.
func (s *SlogAdapter) Warn(msg string, args ...interface{}) {
	s.slog.Warn(msg, args...)
}

// Error logs an error message.
func (s *SlogAdapter) Error(msg string, args ...interface{}) {
	s.slog.Error(msg, args...)
}

// WithComponent returns a new logger with the component field added.
func (s *SlogAdapter) WithComponent(component string) Logger {
	return &SlogAdapter{slog: s.slog.With("component", component)}
}
