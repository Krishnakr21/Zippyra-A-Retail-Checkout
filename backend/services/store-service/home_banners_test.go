package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHomeBanners_Success(t *testing.T) {
	handler := &StoreHandler{}
	router := SetupRoutes(handler, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/store/home-banners", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp HomeBannersResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Banners)
	assert.Equal(t, "banner-1", resp.Banners[0].ID)
}
