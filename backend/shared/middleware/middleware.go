package middleware

import (
	"context"
	"net/http"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

const UserClaimsKey = "user_claims"

// MaxBytesMiddleware limits request body size to 1MB
func MaxBytesMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware handles basic CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RecoverMiddleware catches panics
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Internal server panic recovered", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates session & general JWT tokens
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
				errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authorization token required", nil)
				return
			}
			tokenStr := authHeader[7:]
			
			if claims, err := jwt.ParseAndVerifyToken(tokenStr, secret); err == nil && claims != nil {
				ctx := context.WithValue(r.Context(), "user_claims", claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if sessionClaims, err := jwt.ParseAndVerifySessionToken(tokenStr, secret); err == nil && sessionClaims != nil {
				ctx := context.WithValue(r.Context(), "user_claims", sessionClaims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid or expired token", nil)
		})
	}
}
