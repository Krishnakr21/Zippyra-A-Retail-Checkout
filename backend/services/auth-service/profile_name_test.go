package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zippyra/backend/shared/jwt"
)

func TestUpdateName_Success(t *testing.T) {
	jwtSecret := "zippyra-dev-jwt-secret-key-32bytes"
	os.Setenv("JWT_SECRET", jwtSecret)
	defer os.Unsetenv("JWT_SECRET")

	repo := NewMemoryRepository()
	user, err := repo.CreateUserWithPhone(nil, "+919876543210")
	assert.NoError(t, err)

	handler := NewAuthHandler(repo, nil, nil)
	router := SetupRoutes(handler)

	token, err := jwt.GenerateToken(&jwt.Claims{UserID: user.ID, Role: "CUSTOMER"}, jwtSecret, time.Hour)
	assert.NoError(t, err)

	body, _ := json.Marshal(UpdateNameRequest{Name: "Anita Sharma"})
	req := httptest.NewRequest(http.MethodPut, "/v1/auth/me/name", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var dto UserDTO
	err = json.Unmarshal(rr.Body.Bytes(), &dto)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, dto.ID)
	assert.NotNil(t, dto.Name)
	assert.Equal(t, "Anita Sharma", *dto.Name)
}

func TestUpdateName_ValidationError_EmptyName(t *testing.T) {
	jwtSecret := "zippyra-dev-jwt-secret-key-32bytes"
	os.Setenv("JWT_SECRET", jwtSecret)
	defer os.Unsetenv("JWT_SECRET")

	repo := NewMemoryRepository()
	user, _ := repo.CreateUserWithPhone(nil, "+919876543211")

	handler := NewAuthHandler(repo, nil, nil)
	router := SetupRoutes(handler)

	token, _ := jwt.GenerateToken(&jwt.Claims{UserID: user.ID, Role: "CUSTOMER"}, jwtSecret, time.Hour)

	body, _ := json.Marshal(UpdateNameRequest{Name: "   "})
	req := httptest.NewRequest(http.MethodPut, "/v1/auth/me/name", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
