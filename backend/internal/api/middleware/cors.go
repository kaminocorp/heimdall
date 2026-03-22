package middleware

import (
	"net/http"
	"os"
	"strings"
)

// allowedOrigins returns the set of origins permitted to make cross-origin requests.
// Reads CORS_ALLOWED_ORIGINS env var (comma-separated). Falls back to localhost dev
// origins when unset.
func allowedOrigins() map[string]bool {
	if env := os.Getenv("CORS_ALLOWED_ORIGINS"); env != "" {
		origins := make(map[string]bool)
		for _, o := range strings.Split(env, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				origins[o] = true
			}
		}
		return origins
	}
	// Dev defaults — override in production via CORS_ALLOWED_ORIGINS.
	return map[string]bool{
		"http://localhost:5173": true,
		"http://localhost:4173": true,
	}
}


func CORS(next http.Handler) http.Handler {
	origins := allowedOrigins()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
