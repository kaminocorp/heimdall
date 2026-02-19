package middleware

import "net/http"

// Auth is a placeholder authentication middleware.
// TODO: implement actual auth (JWT or session-based).
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
