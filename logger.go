package log

import (
	"os"
	"time"

	"github.com/emiguens/zapfmt/encoders"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// DefaultLogger is the default logger and is used when given a context
// with no associated log instance.
//
// DefaultLogger by default discards all logs. You can change it's implementation
// by settings this variable to an instantiated logger of your own.
var DefaultLogger = &logger{
	Logger: zap.NewNop(),
}

// NewProductionLogger is a reasonable production logging configuration.
// Logging is enabled at given level and above. The level can be later
// adjusted dynamically in runtime by calling SetLevel method.
//
// It uses the custom Key Value encoder, writes to standard error, and enables sampling.
// Stacktraces are automatically included on logs of ErrorLevel and above.
func NewProductionLogger(lvl *zap.AtomicLevel) Logger {
	zapCore := newZapCoreAtLevel(zap.DebugLevel)
	l := zap.New(
		zapCore,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel),
		wrapCoreWithLevel(lvl),
	)

	return &logger{
		Logger: l,
	}
}

// logger provides a fast, leveled, structured logging. All methods are safe
// for concurrent use.
//
// The logger is designed for contexts in which every microsecond and every
// allocation matters, so its API intentionally favors performance and type
// safety over brevity. For most applications, the SugaredLogger strikes a
// better balance between performance and ergonomics.
type logger struct {
	*zap.Logger
}

var _ Logger = &logger{}

// WithLevel creates a child logger that logs on the given level.
// Child logger contains all fields from the parent.
func (l *logger) WithLevel(level zapcore.Level) Logger {
	lvl := zap.NewAtomicLevelAt(level)
	child := l.Logger.WithOptions(wrapCoreWithLevel(&lvl))
	return &logger{
		Logger: child,
	}
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func (l *logger) With(fields ...zapcore.Field) Logger {
	child := l.Logger.With(fields...)
	return &logger{
		Logger: child,
	}
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func (l *logger) Named(s string) Logger {
	child := l.Logger.Named(s)
	return &logger{
		Logger: child,
	}
}

func newZapCoreAtLevel(lvl zapcore.Level) zapcore.Core {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     rfc3399NanoTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := encoders.NewKeyValueEncoder(encoderConfig)
	writer := zapcore.Lock(zapcore.AddSync(os.Stderr))

	return zapcore.NewCore(encoder, writer, lvl)
}

// rfc3399NanoTimeEncoder serializes a time.Time to an RFC3399-formatted string
// with microsecond precision padded with zeroes to make it fixed width.
func rfc3399NanoTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	const RFC3339Micro = "2006-01-02T15:04:05.000000Z07:00"

	enc.AppendString(t.UTC().Format(RFC3339Micro))
}
