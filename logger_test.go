package log_test

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	log "github.com/emiguens/zapfmt"
	"github.com/kami-zh/go-capturer"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// logRegex returns the level as the first group, discards the timestamp, logger as the
	// second group, caller is discarded and everything after that as the fourth group.
	//
	// Examples of matched lines:
	//   [ts:2019-04-01T15:39:09.142773Z][level:debug][caller:log/logger_test.go:21][msg:before contextualicing]
	//   [ts:2019-04-01T17:19:16.290081Z][level:warn][logger:first_level.second_level.third_level][caller:log/logger_test.go:97][msg:my Warn message]
	logRegex = regexp.MustCompile(`\[ts:(?:[0-9-T:\.]+Z)\]\[level:([a-z]+)\](\[logger:(?:.*?)\])?\[caller:(.*?)\](.*)`)

	// stacktraceRegex finds the stacktrace segment within a log line.
	stacktraceRegex = regexp.MustCompile(`(\[stacktrace:(?:.*?)\])`)
)

type LogLine struct {
	Level      string
	LoggerName string
	Message    string
}

func TestKeyValueLogger(t *testing.T) {
	parseLogLine := func(t *testing.T, line string) LogLine {
		matches := logRegex.FindAllStringSubmatch(line, -1)

		if len(matches[0]) != 5 {
			t.Fatalf("expected regex to have 5 matches, %d found", len(matches[0]))
		}

		lvl, name, msg := matches[0][1], matches[0][2], matches[0][4]

		return LogLine{
			Level:      lvl,
			LoggerName: name,
			Message:    msg,
		}
	}

	assertLine := func(t *testing.T, line, level, content, name string) {
		l := parseLogLine(t, line)

		if l.Level != level {
			t.Fatalf("expected log level to be %s, got: %s", level, l.Level)
		}

		if l.Message != content {
			t.Fatalf("expected content to be %s, got: %s", content, l.Message)
		}

		if l.LoggerName != name {
			t.Fatalf("expected logger name to be %s, got: %s", name, l.LoggerName)
		}
	}

	assertAndRemoveStacktrace := func(t *testing.T, line string) string {
		if !stacktraceRegex.MatchString(line) {
			t.Fatalf("expected line to have stacktrace, none found")
		}
		return stacktraceRegex.ReplaceAllString(line, "")
	}

	tt := []struct {
		Name       string
		Level      zapcore.Level
		SetupFunc  func(t *testing.T, l log.Logger)
		AssertFunc func(t *testing.T, lines []string)
	}{
		{
			Name:  "Log Using Raw Logger",
			Level: zap.DebugLevel,
			SetupFunc: func(t *testing.T, l log.Logger) {
				l.Debug("my Debug message")
				l.Info("my Info message")
				l.Warn("my Warn message")
				l.Error("my Error message")
			},
			AssertFunc: func(t *testing.T, lines []string) {
				assertLine(t, lines[0], "debug", "[msg:my Debug message]", "")
				assertLine(t, lines[1], "info", "[msg:my Info message]", "")
				assertLine(t, lines[2], "warn", "[msg:my Warn message]", "")
				assertLine(t, assertAndRemoveStacktrace(t, lines[3]), "error", `[msg:my Error message]`, "")
			},
		},
		{
			Name:  "Log Using Context",
			Level: zap.DebugLevel,
			SetupFunc: func(t *testing.T, l log.Logger) {
				ctx := log.Context(context.Background(), l)

				log.Debug(ctx, "my Debug message")
				log.Info(ctx, "my Info message")
				log.Warn(ctx, "my Warn message")
				log.Error(ctx, "my Error message")
			},
			AssertFunc: func(t *testing.T, lines []string) {
				assertLine(t, lines[0], "debug", "[msg:my Debug message]", "")
				assertLine(t, lines[1], "info", "[msg:my Info message]", "")
				assertLine(t, lines[2], "warn", "[msg:my Warn message]", "")
				assertLine(t, assertAndRemoveStacktrace(t, lines[3]), "error", `[msg:my Error message]`, "")
			},
		},
		{
			Name:  "Named Logger",
			Level: zap.DebugLevel,
			SetupFunc: func(t *testing.T, l log.Logger) {
				ctx := log.Context(context.Background(), l)

				ctx = log.Named(ctx, "first_level")
				log.Debug(ctx, "my Debug message")

				ctx = log.Named(ctx, "second_level")
				log.Info(ctx, "my Info message")

				ctx = log.Named(ctx, "third_level")
				log.Warn(ctx, "my Warn message")
			},
			AssertFunc: func(t *testing.T, lines []string) {
				assertLine(t, lines[0], "debug", "[msg:my Debug message]", "[logger:first_level]")
				assertLine(t, lines[1], "info", "[msg:my Info message]", "[logger:first_level.second_level]")
				assertLine(t, lines[2], "warn", "[msg:my Warn message]", "[logger:first_level.second_level.third_level]")
			},
		},
		{
			Name:  "Check Works (Should log)",
			Level: zap.DebugLevel,
			SetupFunc: func(t *testing.T, l log.Logger) {
				ctx := log.Context(context.Background(), l)
				if ce := log.Check(ctx, zap.DebugLevel, "my Debug message"); ce != nil {
					ce.Write(zap.String("foo", "bar"))
				}
			},
			AssertFunc: func(t *testing.T, lines []string) {
				assertLine(t, lines[0], "debug", "[msg:my Debug message][foo:bar]", "")
			},
		},
		{
			Name:  "Check Works (Should not log)",
			Level: zap.InfoLevel,
			SetupFunc: func(t *testing.T, l log.Logger) {
				ctx := log.Context(context.Background(), l)
				if ce := log.Check(ctx, zap.DebugLevel, "my Debug message"); ce != nil {
					ce.Write(zap.String("foo", "bar"))
				}
			},
			AssertFunc: func(t *testing.T, lines []string) {
				if len(lines) > 0 {
					t.Fatalf("expected 0 lines in output buffer, found: %d", len(lines))
				}
			},
		},
		{
			Name:  "Log Message With Fields",
			Level: zap.DebugLevel,
			SetupFunc: func(t *testing.T, l log.Logger) {
				ctx := log.Context(context.Background(), l)

				log.Debug(ctx, "my Debug message",
					zap.String("string_key", "value"),
					zap.Time("time_key", time.Unix(0, 0)),
					zap.Int64("int64_key", 1234),
					zap.Float64("float64_key", 1234.5678),
					zap.Error(fmt.Errorf("my error")),
				)
			},
			AssertFunc: func(t *testing.T, lines []string) {
				assertLine(t, lines[0], "debug", "[msg:my Debug message][string_key:value][time_key:1970-01-01T00:00:00.000000Z][int64_key:1234][float64_key:1234.5678][error:my error]", "")
			},
		},
		{
			Name:  "Log Message With Context Fields",
			Level: zap.DebugLevel,
			SetupFunc: func(t *testing.T, l log.Logger) {
				ctx := log.Context(context.Background(), l)

				ctx = log.With(ctx,
					zap.String("string_key", "value"),
					zap.Time("time_key", time.Unix(0, 0)),
					zap.Int64("int64_key", 1234),
					zap.Float64("float64_key", 1234.5678),
					zap.Error(fmt.Errorf("my error")),
					zap.Duration("duration_key", 374*time.Millisecond),
				)

				log.Debug(ctx, "my Debug message", zap.String("extra", "debug_extra"))
				log.Info(ctx, "my Info message", zap.String("extra", "info_extra"))
			},
			AssertFunc: func(t *testing.T, lines []string) {
				assertLine(t, lines[0], "debug", "[msg:my Debug message][string_key:value][time_key:1970-01-01T00:00:00.000000Z][int64_key:1234][float64_key:1234.5678][error:my error][duration_key:0.374][extra:debug_extra]", "")
				assertLine(t, lines[1], "info", "[msg:my Info message][string_key:value][time_key:1970-01-01T00:00:00.000000Z][int64_key:1234][float64_key:1234.5678][error:my error][duration_key:0.374][extra:info_extra]", "")
			},
		},
		{
			Name:  "Test Panic Levels",
			Level: zap.DebugLevel,
			SetupFunc: func(t *testing.T, l log.Logger) {
				defer func() {
					if r := recover(); r == nil {
						t.Fatal("expected panic to happen")
					}
				}()

				ctx := log.Context(context.Background(), l)
				log.Panic(ctx, "my Panic message")
			},
			AssertFunc: func(t *testing.T, lines []string) {
				line := assertAndRemoveStacktrace(t, lines[0])
				assertLine(t, line, "panic", "[msg:my Panic message]", "")
			},
		},
		{
			Name:  "Test DPanic Levels",
			Level: zap.DebugLevel,
			SetupFunc: func(t *testing.T, l log.Logger) {
				ctx := log.Context(context.Background(), l)
				log.DPanic(ctx, "my DPanic message")
			},
			AssertFunc: func(t *testing.T, lines []string) {
				line := assertAndRemoveStacktrace(t, lines[0])
				assertLine(t, line, "dpanic", "[msg:my DPanic message]", "")
			},
		},
		{
			Name:  "Test Sugar Logger",
			Level: zap.DebugLevel,
			SetupFunc: func(t *testing.T, l log.Logger) {
				ctx := log.Context(context.Background(), l)

				logger := log.Sugar(ctx)
				logger.Debugw("my Debug message", "string_key", "value", "int64_key", 123456)
			},
			AssertFunc: func(t *testing.T, lines []string) {
				assertLine(t, lines[0], "debug", "[msg:my Debug message][string_key:value][int64_key:123456]", "")
			},
		},
		{
			Name:  "Test Change Levels",
			Level: zap.ErrorLevel,
			SetupFunc: func(t *testing.T, l log.Logger) {
				ctx1 := log.Context(context.Background(), l)
				log.Debug(ctx1, "should not appear", zap.String("log_level", "error"))
				log.Info(ctx1, "should not appear", zap.String("log_level", "error"))

				ctx2 := log.WithLevel(ctx1, zap.InfoLevel)
				// Previous contexts should remain at their own level.
				log.Debug(ctx1, "should not appear", zap.String("log_level", "error"))
				log.Info(ctx1, "should not appear", zap.String("log_level", "error"))

				// New context should accept new level.
				log.Debug(ctx2, "should not appear", zap.String("log_level", "info"))
				log.Info(ctx2, "should appear", zap.String("log_level", "info"))

				ctx3 := log.WithLevel(ctx2, zap.DebugLevel)
				// Previous contexts should remain at their own level.
				log.Debug(ctx1, "should not appear", zap.String("log_level", "error"))
				log.Info(ctx1, "should not appear", zap.String("log_level", "error"))
				log.Debug(ctx2, "should not appear", zap.String("log_level", "info"))

				// New context should accept new level.
				log.Debug(ctx3, "should appear", zap.String("log_level", "debug"))
				log.Info(ctx3, "should appear", zap.String("log_level", "debug"))
			},
			AssertFunc: func(t *testing.T, lines []string) {
				assertLine(t, lines[0], "info", "[msg:should appear][log_level:info]", "")
				assertLine(t, lines[1], "debug", "[msg:should appear][log_level:debug]", "")
				assertLine(t, lines[2], "info", "[msg:should appear][log_level:debug]", "")
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			out := capturer.CaptureStderr(func() {
				lvl := zap.NewAtomicLevelAt(tc.Level)
				l := log.NewProductionLogger(&lvl)
				tc.SetupFunc(t, l)
			})

			var lines []string

			s := bufio.NewScanner(strings.NewReader(out))
			for s.Scan() {
				lines = append(lines, s.Text())
			}

			if err := s.Err(); err != nil {
				t.Fatalf("error reading stdErr output buffer: %v", err)
			}

			tc.AssertFunc(t, lines)
		})
	}

}
