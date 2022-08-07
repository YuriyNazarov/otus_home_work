package internalhttp

import (
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter { //nolint:revive
	return &loggingResponseWriter{w, http.StatusOK}
}

func loggingMiddleware(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := NewLoggingResponseWriter(w)
		next.ServeHTTP(ww, r)
		dt := time.Since(start).String()
		logger.Info(
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			r.Proto,
			ww.statusCode,
			dt,
			r.Header.Get("User-Agent"),
		)
	})
}
