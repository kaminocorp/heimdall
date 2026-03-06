package middleware

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

// jwkEntry represents a single key in a JWKS response.
type jwkEntry struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
}

// jwksResponse represents the response from a JWKS endpoint.
type jwksResponse struct {
	Keys []jwkEntry `json:"keys"`
}

// JWKSClient fetches and caches ECDSA public keys from a Supabase JWKS endpoint.
type JWKSClient struct {
	endpoint string
	client   *http.Client

	mu      sync.RWMutex
	keys    map[string]*ecdsa.PublicKey // kid → public key
	fetched time.Time
}

const jwksCacheDuration = 10 * time.Minute

// NewJWKSClient creates a client that fetches signing keys from the given JWKS URL.
func NewJWKSClient(supabaseURL string) *JWKSClient {
	endpoint := strings.TrimRight(supabaseURL, "/") + "/auth/v1/.well-known/jwks.json"
	return &JWKSClient{
		endpoint: endpoint,
		client:   &http.Client{Timeout: 10 * time.Second},
		keys:     make(map[string]*ecdsa.PublicKey),
	}
}

// Fetch retrieves the current signing keys from the JWKS endpoint.
// Must be called at least once at startup before verifying tokens.
func (j *JWKSClient) Fetch(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, j.endpoint, nil)
	if err != nil {
		return fmt.Errorf("building JWKS request: %w", err)
	}

	resp, err := j.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetching JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned %d", resp.StatusCode)
	}

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("decoding JWKS response: %w", err)
	}

	keys := make(map[string]*ecdsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "EC" || k.Crv != "P-256" {
			continue
		}
		pub, err := parseECKey(k.X, k.Y)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}

	if len(keys) == 0 {
		return fmt.Errorf("no usable ES256 keys in JWKS response")
	}

	j.mu.Lock()
	j.keys = keys
	j.fetched = time.Now()
	j.mu.Unlock()

	return nil
}

// GetKey returns the public key for the given kid. If the cache is stale, it
// refreshes in the background. If the kid is unknown, it forces an immediate
// refresh in case Supabase rotated keys.
func (j *JWKSClient) GetKey(kid string) (*ecdsa.PublicKey, error) {
	j.mu.RLock()
	key, ok := j.keys[kid]
	stale := time.Since(j.fetched) > jwksCacheDuration
	j.mu.RUnlock()

	// Background refresh if cache is stale but we have a matching key.
	if ok && stale {
		go j.Fetch(context.Background()) //nolint:errcheck
		return key, nil
	}

	if ok {
		return key, nil
	}

	// Unknown kid — force refresh and retry.
	if err := j.Fetch(context.Background()); err != nil {
		return nil, fmt.Errorf("JWKS refresh failed: %w", err)
	}

	j.mu.RLock()
	key, ok = j.keys[kid]
	j.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown signing key %q", kid)
	}
	return key, nil
}

func parseECKey(xB64, yB64 string) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(xB64)
	if err != nil {
		return nil, err
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(yB64)
	if err != nil {
		return nil, err
	}

	pub := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}
	if !pub.Curve.IsOnCurve(pub.X, pub.Y) {
		return nil, fmt.Errorf("point not on P-256 curve")
	}
	return pub, nil
}

// ValidateJWT parses and validates a Supabase JWT signed with ES256, returning the user UUID.
// Used by both the HTTP auth middleware and WebSocket auth.
func ValidateJWT(tokenStr string, jwks *JWKSClient) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, fmt.Errorf("missing kid header")
		}
		return jwks.GetKey(kid)
	}, jwt.WithValidMethods([]string{"ES256"}))

	if err != nil || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid or expired token")
	}

	sub, err := token.Claims.GetSubject()
	if err != nil || sub == "" {
		return uuid.Nil, fmt.Errorf("missing subject claim")
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid subject claim")
	}

	return userID, nil
}

// Auth returns middleware that validates Supabase-issued JWTs using ES256 via JWKS.
func Auth(jwks *JWKSClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			userID, err := ValidateJWT(tokenStr, jwks)
			if err != nil {
				jsonError(w, err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ContextWithUserID returns a new context with the given user ID set.
// This is primarily useful for testing, where JWT validation is bypassed.
func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
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
