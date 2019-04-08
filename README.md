
# ZAPFMT

[![GoDoc](https://godoc.org/github.com/emiguens/zapfmt?status.svg)](https://godoc.org/github.com/emiguens/zapfmt)

ZAP fmt is a small wrapper around [Uber log package](https://godoc.org/go.uber.org/zap). This package is supposed to be used with ZAP field functions, so importing zap package is required.

This package provides:

- An implementation of a key value encoder (instead of JSON).
- Method for creating a child logger (inheriting attributed) with a different log level.
- Helper methods for contextualizing a logger by injecting it into a `context.Context` and latter consuming it directly through it.

## Key Value Encoding

Logging using this package will result in the following format.

```log
[ts:2019-04-08T20:21:32.375067Z][level:info][caller:zapfmt/main.go:44][msg:calling thisImportantCall][uuid:34d4fb89-c27b-4c7c-bb51-4e46fba614dd][v:15646231]
[ts:2019-04-08T20:21:32.375073Z][level:debug][caller:zapfmt/main.go:37][msg:calling thisCall][uuid:34d4fb89-c27b-4c7c-bb51-4e46fba614dd][v:3]
[ts:2019-04-08T20:21:32.375079Z][level:error][caller:zapfmt/main.go:44][msg:calling thisImportantCall][uuid:34d4fb89-c27b-4c7c-bb51-4e46fba614dd][v:15646231]
```

## Dynamic Log Level

Instantiating a logger requires a `zap.AtomicLevel` reference. If you keep the reference to the given object you can then modify the logging level at runtime dynamically. Keep in mind that using the `WithLevel` method for instantiating a child logger on another level will lock that child logger into the new level.

Zaps AtomicLevel object complies with the `ServeHTTP` interface, which means that one can hook it directly into a webserver and expose an HTTP method for changing the log level.

Example:

```go
func main() {
    lvl := zap.NewAtomicLevelAt(zap.ErrorLevel)
    logger := log.NewProductionLogger(&lvl)

    // Wrap logger inside given context
    ctx := log.Context(context.Background(), logger)

    go func() {
        for {
            // Use logger within this context to log
            log.Debug(ctx, "my log line", zap.String("key", "value"))
            time.Sleep(1 * time.Second)
        }
    }()

    // Expose dynamic log level endpoint
    http.HandleFunc("/debug/log", lvl)
    http.ListenAndServe(":8080", nil)
}
```

```bash
# Get current log level
curl -X GET http://localhost:8080/debug/log
{"level":"error"}

# Change log level
curl -X PUT http://localhost:8080/debug/log -d '{"level":"debug"}'
{"level":"debug"}
```