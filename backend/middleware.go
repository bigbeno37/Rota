package main

import (
	"context"
	"fmt"
	"net/http"
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

func MyMiddleware2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("The context is", r.Context().Value("foo"))
		next.ServeHTTP(w, r)
	})
}
