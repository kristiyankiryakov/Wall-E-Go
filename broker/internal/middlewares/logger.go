package middlewares

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"time"
)

func Logger(log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			wrappedWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			defer func() {
				log.WithFields(logrus.Fields{
					"request-id": GetRequestID(r.Context()),
					"status":     wrappedWriter.Status(),
					"method":     r.Method,
					"path":       r.URL.Path,
					"query":      r.URL.RawQuery,
					"ip":         r.RemoteAddr,
					"trace-id":   trace.SpanFromContext(r.Context()).SpanContext().TraceID().String(),
					"latency":    time.Since(start).String(),
				}).Info("request completed")
			}()

			next.ServeHTTP(wrappedWriter, r.WithContext(r.Context()))
		}
		return http.HandlerFunc(fn)
	}
}
