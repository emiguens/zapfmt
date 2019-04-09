package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the interface that wraps methods needed for a valid logger implementation.
type Logger interface {
	// Check returns a CheckedEntry if logging a message at the specified level
	// is enabled. It's a completely optional optimization; in high-performance
	// applications, Check can help avoid allocating a slice to hold fields.
	Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry

	// Named adds a new path segment to the logger's name. Segments are joined by
	// periods. By default, Loggers are unnamed.
	Named(s string) Logger

	// Sugar wraps the logger to provide a more ergonomic, but slightly slower,
	// API. Sugaring a logger is quite inexpensive, so it's reasonable for a
	// single application to use both Loggers and SugaredLoggers, converting
	// between them on the boundaries of performance-sensitive code.
	Sugar() *zap.SugaredLogger

	// With creates a child logger and adds structured context to it. Fields added
	// to the child don't affect the parent, and vice versa.
	With(fields ...zap.Field) Logger

	// WithLevel created a child logger that logs on the given level.
	// Child logger contains all fields from the parent.
	WithLevel(lvl zapcore.Level) Logger

	// DPanic logs a message at DPanicLevel. The message includes any fields
	// passed at the log site, as well as any fields accumulated on the logger.
	DPanic(msg string, fields ...zap.Field)

	// Debug logs a message at DebugLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Debug(msg string, fields ...zap.Field)

	// Error logs a message at ErrorLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Error(msg string, fields ...zap.Field)

	// Fatal logs a message at FatalLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	//
	// The logger then calls os.Exit(1), even if logging at FatalLevel is
	// disabled.
	Fatal(msg string, fields ...zap.Field)

	// Info logs a message at InfoLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Info(msg string, fields ...zap.Field)

	// Panic logs a message at PanicLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	//
	// The logger then panics, even if logging at PanicLevel is disabled.
	Panic(msg string, fields ...zap.Field)

	// Warn logs a message at WarnLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Warn(msg string, fields ...zap.Field)
}
