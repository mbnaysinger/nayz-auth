package middlewares

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
		start := time.Now()

		wrapper := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK, 
		}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)

		// O VERDADEIRO PODER DO SLOG: Logging Estruturado.
		// Ao invés de uma string de texto corrido, passamos chaves e valores!
		slog.Info("Requisição HTTP processada",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("ip", r.RemoteAddr),
			slog.Int("status", wrapper.statusCode),
			slog.String("duration", duration.String()),
		)
	})
}
