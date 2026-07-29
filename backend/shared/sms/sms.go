package sms

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/zippyra/backend/shared/logger"
)

type SmsSender interface {
	SendSMS(ctx context.Context, phone, code string) error
}

// LogSmsSender logs OTP to stdout/log for dev/test
type LogSmsSender struct{}

func (l *LogSmsSender) SendSMS(ctx context.Context, phone, code string) error {
	logger.Info("SMS OTP delivered to %s (Dev Log Only)", logger.MaskPhone(phone))
	return nil
}

// TwilioSmsSender sends SMS via Twilio API
type TwilioSmsSender struct {
	AccountSID string
	AuthToken  string
	FromPhone  string
}

func NewTwilioSmsSender() *TwilioSmsSender {
	return &TwilioSmsSender{
		AccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
		AuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		FromPhone:  os.Getenv("TWILIO_FROM_PHONE"),
	}
}

func (t *TwilioSmsSender) SendSMS(ctx context.Context, phone, code string) error {
	if t.AccountSID == "" || t.AuthToken == "" {
		// Fallback to log sender if credentials missing
		logger.Info("Twilio credentials not configured. SMS code for %s (Dev Log Only)", logger.MaskPhone(phone))
		return nil
	}

	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.AccountSID)
	data := url.Values{}
	data.Set("To", phone)
	data.Set("From", t.FromPhone)
	data.Set("Body", fmt.Sprintf("Your Zippyra login code is %s", code))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(t.AccountSID, t.AuthToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("twilio API returned status %d", resp.StatusCode)
	}

	logger.Info("Twilio SMS sent to %s", logger.MaskPhone(phone))
	return nil
}
