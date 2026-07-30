package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

// RequireStepUp validates that the authenticated admin session has completed 2FA step-up within maxAge.
func RequireStepUp(maxAge time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var claims *jwt.Claims

			if val := r.Context().Value("user_claims"); val != nil {
				if c, ok := val.(*jwt.Claims); ok {
					claims = c
				}
			}
			if claims == nil {
				if val := r.Context().Value("claims"); val != nil {
					if c, ok := val.(*jwt.Claims); ok {
						claims = c
					}
				}
			}

			if claims == nil || claims.UserType != "ADMIN" {
				errors.WriteError(w, http.StatusForbidden, "STEP_UP_REQUIRED", "Admin authentication required", nil)
				return
			}

			if claims.StepUpAt <= 0 {
				errors.WriteError(w, http.StatusForbidden, "STEP_UP_REQUIRED", "2FA step-up verification required", nil)
				return
			}

			elapsed := time.Now().Unix() - claims.StepUpAt
			if elapsed < 0 || elapsed > int64(maxAge.Seconds()) {
				errors.WriteError(w, http.StatusForbidden, "STEP_UP_REQUIRED", "2FA step-up verification expired", nil)
				return
			}

			ctx := context.WithValue(r.Context(), "step_up_verified", true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
