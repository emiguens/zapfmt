package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Loggerer ...
type Loggerer interface {
	// Check returns a CheckedEntry if logging a message at the specified level
	// is enabled. It's a completely optional optimization; in high-performance
	// applications, Check can help avoid allocating a slice to hold fields.
	Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry

	// Named adds a new path segment to the Logger's name. Segments are joined by
	// periods. By default, Loggers are unnamed.
	Named(s string) Loggerer

	// Sugar wraps the Logger to provide a more ergonomic, but slightly slower,
	// API. Sugaring a Logger is quite inexpensive, so it's reasonable for a
	// single application to use both Loggers and SugaredLoggers, converting
	// between them on the boundaries of performance-sensitive code.
	Sugar() *zap.SugaredLogger

	// With creates a child Logger and adds structured context to it. Fields added
	// to the child don't affect the parent, and vice versa.
	With(fields ...zap.Field) Loggerer

	// WithLevel created a child logger that logs on the given level.
	// Child logger contains all fields from the parent.
	WithLevel(lvl zapcore.Level) Loggerer

	// DPanic logs a message at DPanicLevel. The message includes any fields
	// passed at the log site, as well as any fields accumulated on the Logger.
	DPanic(msg string, fields ...zap.Field)

	// Debug logs a message at DebugLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the Logger.
	Debug(msg string, fields ...zap.Field)

	// Error logs a message at ErrorLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the Logger.
	Error(msg string, fields ...zap.Field)

	// Fatal logs a message at FatalLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the Logger.
	//
	// The Logger then calls os.Exit(1), even if logging at FatalLevel is
	// disabled.
	Fatal(msg string, fields ...zap.Field)

	// Info logs a message at InfoLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the Logger.
	Info(msg string, fields ...zap.Field)

	// Panic logs a message at PanicLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the Logger.
	//
	// The Logger then panics, even if logging at PanicLevel is disabled.
	Panic(msg string, fields ...zap.Field)

	// Warn logs a message at WarnLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the Logger.
	Warn(msg string, fields ...zap.Field)
}
