package encoders_test

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/emiguens/zapfmt/encoders"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestEncoderBasicFormat(t *testing.T) {
	logger, buffer := testLogger()

	logger.Debug("my debug message",
		zap.String("key", "my_value"),
		zap.Duration("duration", 673*time.Millisecond),
		zap.Error(fmt.Errorf("timeout")),
	)

	lvl, msg := deconstructLogLine(buffer.String())
	requireEqual(t, "debug", lvl)
	requireEqual(t, "[msg:my debug message][key:my_value][duration:0.673][error:timeout]", msg)

	buffer.Reset()

	logger.Warn("my warn message",
		zap.String("key", "my_value"),
		zap.Duration("duration", 673*time.Millisecond),
		zap.Error(fmt.Errorf("timeout")),
	)

	lvl, msg = deconstructLogLine(buffer.String())
	requireEqual(t, "warn", lvl)
	requireEqual(t, "[msg:my warn message][key:my_value][duration:0.673][error:timeout]", msg)

}

func TestEncoderTypedFunctions(t *testing.T) {
	logger, buffer := testLogger()
	logger = logger.With(
		zap.Binary("binary_key", []byte{0x56}),
		zap.Bool("bool_key", true),
		zap.ByteString("bytestring_key", []byte("æ")),
		zap.Complex128("complex128_key", complex(float64(1), float64(1))),
		zap.Complex64("complex64_key", complex(float32(1), float32(1))),
		zap.Float64("float64_key", 123.456),
		zap.Float32("float32_key", 123.456),
		zap.Int("int_key", 123),
		zap.Int64("int64_key", 123),
		zap.Int32("int32_key", 123),
		zap.Int16("int16_key", 123),
		zap.Int8("int8_key", 123),
		zap.String("string_key", "my string with spaces"),
		zap.Uint("uint_key", 123),
		zap.Uint64("uint64_key", 123),
		zap.Uint32("uint32_key", 123),
		zap.Uint16("uint16_key", 123),
		zap.Uint8("uint8_key", 123),
		zap.Reflect("reflect_key", map[string]interface{}{"object": []string{"a", "b"}}),
		zap.Any("any_key", map[string]interface{}{"object": []string{"a", "b"}}),
		zap.Strings("strings", []string{"a", "b", "c"}),
		zap.Int64s("numbers", []int64{1, 2, 3, 4}),
	)

	logger.Debug("my debug message")

	lvl, msg := deconstructLogLine(buffer.String())

	const expectedMsg = `[msg:my debug message][binary_key:Vg==][bool_key:true][bytestring_key:æ][complex128_key:1+1i][complex64_key:1+1i][float64_key:123.456][float32_key:123.45600128173828][int_key:123][int64_key:123][int32_key:123][int16_key:123][int8_key:123][string_key:my string with spaces][uint_key:123][uint64_key:123][uint32_key:123][uint16_key:123][uint8_key:123][reflect_key:{"object":["a","b"]}][any_key:{"object":["a","b"]}][strings:[a][b][c]][numbers:[1][2][3][4]]`
	requireEqual(t, "debug", lvl)
	requireEqual(t, expectedMsg, msg)
}

func TestAppendFloatSpecialCases(t *testing.T) {
	logger, buffer := testLogger()

	logger.Debug("my debug message",
		zap.Float64("nan", math.NaN()),
		zap.Float64("pos_inf", math.Inf(1)),
		zap.Float64("neg_inf", math.Inf(-1)),
	)

	lvl, msg := deconstructLogLine(buffer.String())
	requireEqual(t, "debug", lvl)
	requireEqual(t, "[msg:my debug message][nan:NaN][pos_inf:+Inf][neg_inf:-Inf]", msg)
}

func TestNamedLogger(t *testing.T) {
	logger, buffer := testLogger()
	logger = logger.Named("my_name")
	logger = logger.Named("another")

	logger.Debug("my debug message")

	requireEqual(t, "[ts:1970-01-01T00:00:00Z][level:debug][logger:my_name.another][msg:my debug message]\n", buffer.String())
}

var logRegex = regexp.MustCompile(`\[ts:1970-01-01T00:00:00Z\]\[level:([a-z]+)\](\[msg:.*)`)

func deconstructLogLine(line string) (level, content string) {
	matches := logRegex.FindAllStringSubmatch(line, -1)
	return matches[0][1], matches[0][2]
}

func testLogger() (*zap.Logger, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	writer := zapcore.Lock(zapcore.AddSync(buf))

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(time.Unix(0, 0).UTC().Format(time.RFC3339))
	}

	encoder := encoders.NewKeyValueEncoder(encoderConfig)

	core := zapcore.NewCore(encoder, writer, zap.NewAtomicLevelAt(zap.DebugLevel))

	return zap.New(core), buf
}

func requireEqual(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
