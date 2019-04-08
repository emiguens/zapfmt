package log

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const contextKeyLogger = contextKey("zap-Logger")

// Context returns a copy of the parent context in which the Logger associated
// with it is the one given.
//
// Usually you'll call Context with the Logger returned by NewProductionLogger.
// Once you have a context with a Logger, all additional logging should be
// made by using the static methods exported by this package.
func Context(ctx context.Context, log Loggerer) context.Context {
	l, ok := log.(*Logger)
	if ok {
		l.Logger.WithOptions(zap.AddCallerSkip(1))
	}
	return context.WithValue(ctx, contextKeyLogger, log)
}

// Sugar wraps the logger to provide a more ergonomic, but slightly slower,
// API. Sugaring a logger is quite inexpensive, so it's reasonable for a
// single application to use both Loggers and SugaredLoggers, converting
// between them on the boundaries of performance-sensitive code.
func Sugar(ctx context.Context) *zap.SugaredLogger {
	return getLogger(ctx).Sugar()
}

// Named adds a new path segment to the Logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func Named(ctx context.Context, s string) context.Context {
	logger := getLogger(ctx).Named(s)
	return context.WithValue(ctx, contextKeyLogger, logger)
}

// With creates a child Logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func With(ctx context.Context, fields ...zap.Field) context.Context {
	logger := getLogger(ctx).With(fields...)
	return context.WithValue(ctx, contextKeyLogger, logger)
}

// WithLevel created a child logger that logs on the given level.
// Child logger contains all fields from the parent.
func WithLevel(ctx context.Context, lvl zapcore.Level) context.Context {
	logger := getLogger(ctx).WithLevel(lvl)
	return context.WithValue(ctx, contextKeyLogger, logger)
}

// Check returns a CheckedEntry if logging a message at the specified level
// is enabled. It's a completely optional optimization; in high-performance
// applications, Check can help avoid allocating a slice to hold fields.
func Check(ctx context.Context, lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	return getLogger(ctx).Check(lvl, msg)
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the Logger.
//
// If the Logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func DPanic(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger(ctx).DPanic(msg, fields...)
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the Logger.
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger(ctx).Debug(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the Logger.
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger(ctx).Error(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the Logger.
//
// The Logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger(ctx).Fatal(msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the Logger.
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger(ctx).Info(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the Logger.
//
// The Logger then panics, even if logging at PanicLevel is disabled.
func Panic(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger(ctx).Panic(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the Logger.
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	getLogger(ctx).Warn(msg, fields...)
}

func getLogger(ctx context.Context) Loggerer {
	l, ok := ctx.Value(contextKeyLogger).(Loggerer)
	if ok {
		return l
	}
	return DefaultLogger
}
