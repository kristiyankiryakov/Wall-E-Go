package middlewares

import (
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"net/http"
)

func Tracer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		operation := r.Method + " " + r.URL.Path

		otelhttp.NewHandler(next, operation).ServeHTTP(w, r.WithContext(r.Context()))
	}

	return http.HandlerFunc(fn)
}
