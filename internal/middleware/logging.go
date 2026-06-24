package middleware

import(
	"net/http"
	"log/slog"
	"time"
)


func Logging(next http.Handler)(http.Handler){
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){

		start := time.Now()

		next.ServeHTTP(w,r)

		slog.Info(
			"request compelted",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	})
}