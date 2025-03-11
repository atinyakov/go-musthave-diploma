package middleware

import (
	"net/http"
	"strings"

	"github.com/atinyakov/go-musthave-diploma/pkg/auth"
)

// AuthMiddleware checks JWT token in Authorization header
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := auth.ParseJWT(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add username to request context (optional)
		r.Header.Set("X-User", claims.Username)

		next.ServeHTTP(w, r)
	})
}
