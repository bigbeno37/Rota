package main

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"os"
)

type Middleware func(http.Handler) http.Handler

func CreateStack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			x := xs[i]
			next = x(next)
		}

		return next
	}
}

func getCookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idCookie, err := r.Cookie("id")

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("ID cookie is not present. Connect to the WebSocket server first!"))
			return
		}

		modifiedRequest := r.WithContext(context.WithValue(r.Context(), "id", idCookie.Value))
		next.ServeHTTP(w, modifiedRequest)
	})
}

func withLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var id *string = nil
		if r.Context().Value("id") != nil {
			id = r.Context().Value("id").(*string)
		}

		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

		request := slog.Group("request", slog.String("method", r.Method), slog.String("path", r.URL.Path))

		logger := slog.New(handler).With(request, "traceId", uuid.NewString())

		if id != nil {
			logger = logger.With("id", *id)
		}

		modifiedRequest := r.WithContext(context.WithValue(r.Context(), "logger", logger))
		next.ServeHTTP(w, modifiedRequest)
	})
}
