package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logging(next http.Handler, message string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts := time.Now().Format(time.RFC3339)
		log.Printf("%s %s %s %s", ts, r.Method, r.URL.Path, message)
		next.ServeHTTP(w, r)
	})
}
