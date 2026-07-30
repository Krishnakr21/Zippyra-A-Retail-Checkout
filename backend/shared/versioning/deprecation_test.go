package versioning

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDeprecationMiddleware(t *testing.T) {
	sunset := time.Date(2027, 8, 2, 0, 0, 0, 0, time.UTC)
	migrationURL := "https://docs.zippyra.com/api/v2/migration-guide"

	handler := Deprecated(sunset, migrationURL)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/store/admin/legacy", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", res.StatusCode)
	}

	if rec.Header().Get("Deprecation") != "true" {
		t.Errorf("Expected Deprecation: true, got '%s'", rec.Header().Get("Deprecation"))
	}

	expectedSunset := sunset.Format(http.TimeFormat)
	if rec.Header().Get("Sunset") != expectedSunset {
		t.Errorf("Expected Sunset: %s, got '%s'", expectedSunset, rec.Header().Get("Sunset"))
	}

	expectedLink := "<https://docs.zippyra.com/api/v2/migration-guide>; rel=\"deprecation\""
	if rec.Header().Get("Link") != expectedLink {
		t.Errorf("Expected Link: %s, got '%s'", expectedLink, rec.Header().Get("Link"))
	}
}

func TestDeprecatedHandlerFunc(t *testing.T) {
	sunset := time.Date(2027, 8, 2, 0, 0, 0, 0, time.UTC)

	handler := DeprecatedHandler(sunset, "", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/legacy", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Deprecation") != "true" {
		t.Errorf("Expected Deprecation: true")
	}

	if rec.Header().Get("Sunset") == "" {
		t.Errorf("Expected non-empty Sunset header")
	}

	if rec.Header().Get("Link") != "" {
		t.Errorf("Expected empty Link header when migrationURL is empty")
	}
}
