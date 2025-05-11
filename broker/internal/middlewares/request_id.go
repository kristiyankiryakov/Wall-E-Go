package middlewares

import (
	"context"
	"github.com/google/uuid"
	"net/http"
)

type ctxKey int

const ridKey ctxKey = ctxKey(0)

func GetRequestID(ctx context.Context) string {
	return ctx.Value(ridKey).(string)
}

func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			uuid, err := uuid.NewV6()
			rid = uuid.String()
			if err != nil {
				http.Error(w, "failed to generate request ID", http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), ridKey, rid)
			w.Header().Set("X-Request-ID", rid)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
	return http.HandlerFunc(fn)
}
