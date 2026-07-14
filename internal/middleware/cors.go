package middleware

import (
	"net/http"
	"os"
	"strings"
)


func Cors(next http.Handler) http.Handler {
	allowedOriginsValue := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsValue == "" {
		allowedOriginsValue = "http://localhost:3000,http://127.0.0.1:3000"
	}

	allowedOrigins := make(map[string]bool)
	for _, origin := range strings.Split(allowedOriginsValue, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowedOrigins[origin] = true
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		w.Header().Add("Vary", "Origin")

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			if origin != "" && !allowedOrigins[origin] {
				http.Error(w, "origin not allowed", http.StatusForbidden)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
