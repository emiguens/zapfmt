package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	log "github.com/emiguens/zapfmt"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

func greet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	log.Debug(r.Context(), "handling request",
		zap.Stringer("url", r.URL),
		zap.Reflect("headers", r.Header), // You can print maps or objects using Reflect
	)

	fmt.Fprintf(w, "Hello World!")

	log.Debug(r.Context(), "request time", zap.Duration("elapsed", time.Since(start)))
}

// changeLogLevel randomly changes log level to Debug for the given percentage of requests.
func changeLogLevel(next http.HandlerFunc, percentage int64) http.HandlerFunc {
	rand.Seed(time.Now().UnixNano())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rand.Int63n(100) < percentage {
			// Get a child logger with a different level from their parent
			ctx := log.WithLevel(r.Context(), zap.DebugLevel)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

// withRequestID extracts the request id from the header, or generates a new one.
// It decorates the attached logger with the request id field and injects
// it to the request object.
func withRequestID(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("x-request-id")
		if requestID == "" {
			requestID = uuid.Must(uuid.NewV4()).String()
		}

		ctx := log.With(r.Context(), zap.String("request_id", requestID))
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// attachLogger decorated the request context with the given logger.
func attachLogger(next http.HandlerFunc, logger log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := log.Context(r.Context(), logger)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func main() {
	lvl := zap.NewAtomicLevelAt(zap.ErrorLevel)
	logger := log.NewProductionLogger(&lvl)

	// Decorate request with logging middlewares.
	handler := attachLogger(withRequestID(changeLogLevel(greet, 10)), logger)
	http.Handle("/", handler)

	// AtomicLevel is an http.Handler supporting GET and PUT actions.
	http.Handle("/debug/log", lvl)
	http.ListenAndServe(":8080", nil)
}
