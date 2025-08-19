package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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

func WithIdMiddleware(next http.Handler) http.Handler {
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

func GetIdFromContext(ctx context.Context) string {
	id, ok := ctx.Value("id").(string)
	if !ok {
		panic("Id in context is not present. Something has gone wrong!")
	}

	return id
}

func AddIdToLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetIdFromContext(r.Context())
		logger := GetLoggerFromContext(r.Context())

		modifiedRequest := r.WithContext(context.WithValue(r.Context(), "logger", logger.With(slog.String("id", id))))
		next.ServeHTTP(w, modifiedRequest)
	})
}

func WithLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

		request := slog.Group("request", slog.String("method", r.Method), slog.String("path", r.URL.Path))

		logger := slog.New(handler).With(request, "traceId", uuid.NewString())

		modifiedRequest := r.WithContext(context.WithValue(r.Context(), "logger", logger))
		next.ServeHTTP(w, modifiedRequest)
	})
}

func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value("logger").(*slog.Logger)
	if !ok {
		panic("Logger in context is not present. Something has gone wrong!")
	}

	return logger
}

func WithRedisMiddleware(redis *redis.Client) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			modifiedRequest := r.WithContext(context.WithValue(r.Context(), "redis", redis))
			next.ServeHTTP(w, modifiedRequest)
		})
	}
}

func GetRedisFromContext(ctx context.Context) *redis.Client {
	rdb, ok := ctx.Value("redis").(*redis.Client)
	if !ok {
		panic("Redis in context is not present. Something has gone wrong!")
	}

	return rdb
}
