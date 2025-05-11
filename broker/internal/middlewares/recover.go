package middlewares

import (
	"fmt"
	"github.com/go-stack/stack"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func Recover(log logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					err, ok := p.(error)
					if !ok {
						err = fmt.Errorf("panic: %v", p)
					}

					var stackTrace stack.CallStack
					traces := stack.Trace().TrimRuntime()

					// Format the stack trace
					for i := 0; i < len(traces); i++ {
						t := traces[i]
						tFunc := t.Frame().Function

						if tFunc == "runtime.gopanic" || tFunc == "go.opentelemtry.io/otel/sdk/trace.(*Span).End" {
							continue
						}

						if tFunc == "net/http.HandlerFunc.ServeHTTP" {
							break
						}

						stackTrace = append(stackTrace, t)
					}
					log.WithFields(logrus.Fields{
						"trace-id":   trace.SpanFromContext(r.Context()).SpanContext().TraceID().String(),
						"request-id": GetRequestID(r.Context()),
						"stack":      fmt.Sprintf("%+v", stackTrace),
					}).WithError(err).Panic("panic")
				}
			}()
		}
		return http.HandlerFunc(fn)
	}
}
