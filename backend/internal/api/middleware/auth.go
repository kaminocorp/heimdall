package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

// Auth returns middleware that validates Supabase-issued JWTs using HMAC-SHA256.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	secretBytes := []byte(jwtSecret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Bearer token from Authorization header.
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				jsonError(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok || tokenStr == "" {
				jsonError(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			// Parse and validate the JWT.
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return secretBytes, nil
			}, jwt.WithValidMethods([]string{"HS256"}))

			if err != nil || !token.Valid {
				jsonError(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Extract user ID from the "sub" claim.
			sub, err := token.Claims.GetSubject()
			if err != nil || sub == "" {
				jsonError(w, "missing subject claim", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(sub)
			if err != nil {
				jsonError(w, "invalid subject claim", http.StatusUnauthorized)
				return
			}

			// Store user ID in request context.
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the authenticated user's UUID from the request context.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
