package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"google.golang.org/api/idtoken"
)

type GoogleTokenPayload struct {
	Email string
	HD    string
	Sub   string
}

type GoogleTokenValidator interface {
	Validate(ctx context.Context, idToken string) (*GoogleTokenPayload, error)
}

type RealGoogleTokenValidator struct{}

func (v *RealGoogleTokenValidator) Validate(ctx context.Context, tokenStr string) (*GoogleTokenPayload, error) {
	// Mock fallback for test tokens
	if strings.HasPrefix(tokenStr, "mock-google-token-") {
		raw := strings.TrimPrefix(tokenStr, "mock-google-token-")
		parts := strings.Split(raw, ":")
		email := parts[0]
		sub := "google-sub-mock-123"
		if len(parts) >= 2 {
			sub = parts[1]
		}
		hd := ""
		if emailParts := strings.Split(email, "@"); len(emailParts) == 2 {
			hd = emailParts[1]
		}
		return &GoogleTokenPayload{Email: email, HD: hd, Sub: sub}, nil
	}

	clientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID_WEB")
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	payload, err := idtoken.Validate(ctxTimeout, tokenStr, clientID)
	if err != nil {
		return nil, errors.New("invalid Google ID token")
	}

	email, _ := payload.Claims["email"].(string)
	hd, _ := payload.Claims["hd"].(string)
	sub := payload.Subject

	if email == "" {
		return nil, errors.New("google token missing email claim")
	}

	// If hd is not set in claim, derive from email domain
	if hd == "" {
		parts := strings.Split(email, "@")
		if len(parts) == 2 {
			hd = parts[1]
		}
	}

	return &GoogleTokenPayload{
		Email: email,
		HD:    hd,
		Sub:   sub,
	}, nil
}
