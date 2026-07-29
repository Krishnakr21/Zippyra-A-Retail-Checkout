package versioning

import (
	"fmt"
	"net/http"
	"time"
)

// Deprecated returns an HTTP middleware that injects IETF Deprecation, Sunset, and Link headers
// into route responses while preserving normal request execution.
func Deprecated(sunsetDate time.Time, migrationURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Deprecation", "true")
			if !sunsetDate.IsZero() {
				w.Header().Set("Sunset", sunsetDate.UTC().Format(http.TimeFormat))
			}
			if migrationURL != "" {
				w.Header().Set("Link", fmt.Sprintf("<%s>; rel=\"deprecation\"", migrationURL))
			}
			next.ServeHTTP(w, r)
		})
	}
}

// DeprecatedHandler wraps an http.HandlerFunc directly with deprecation headers.
func DeprecatedHandler(sunsetDate time.Time, migrationURL string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Deprecation", "true")
		if !sunsetDate.IsZero() {
			w.Header().Set("Sunset", sunsetDate.UTC().Format(http.TimeFormat))
		}
		if migrationURL != "" {
			w.Header().Set("Link", fmt.Sprintf("<%s>; rel=\"deprecation\"", migrationURL))
		}
		handler(w, r)
	}
}
