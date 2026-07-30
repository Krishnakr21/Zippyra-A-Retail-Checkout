package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"

	"github.com/zippyra/backend/shared/logger"
)

// NOTE: A personal Gmail account is rate-limited (~500 msgs/day) and is fine for pilot/dev only.
// TODO: Migrate to AWS SES / Google Workspace SMTP relay before scaling past a handful of stores.

type EmailSender interface {
	SendEmailOTP(ctx context.Context, email, code string) error
}

type GmailEmailSender struct {
	User        string
	AppPassword string
	Host        string
	Port        string
}

func NewGmailEmailSender() *GmailEmailSender {
	return &GmailEmailSender{
		User:        os.Getenv("GMAIL_SMTP_USER"),
		AppPassword: os.Getenv("GMAIL_SMTP_APP_PASSWORD"),
		Host:        "smtp.gmail.com",
		Port:        "587",
	}
}

func (g *GmailEmailSender) SendEmailOTP(ctx context.Context, email, code string) error {
	// Always log OTP to server console in dev mode for easy testing
	logger.Info("==================================================")
	logger.Info("[DEV OTP LOG] Email OTP for %s is >>> %s <<<", email, code)
	logger.Info("==================================================")

	if g.User == "" || g.AppPassword == "" || g.User == "otp@zippyra.com" {
		logger.Info("Using placeholder Gmail SMTP credentials. OTP logged above for dev/testing.")
		return nil
	}

	addr := net.JoinHostPort(g.Host, g.Port)
	auth := smtp.PlainAuth("", g.User, g.AppPassword, g.Host)

	subject := "Subject: Your Zippyra login code\r\n"
	contentType := "Content-Type: text/plain; charset=UTF-8\r\n"
	fromHeader := fmt.Sprintf("From: Zippyra Auth <%s>\r\n", g.User)
	toHeader := fmt.Sprintf("To: %s\r\n", email)
	body := fmt.Sprintf("\r\nYour Zippyra login code is %s. It expires in 5 minutes.\r\n", code)

	message := []byte(strings.Join([]string{fromHeader, toHeader, subject, contentType, body}, ""))

	// Dial with STARTTLS
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, g.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	tlsConfig := &tls.Config{
		ServerName: g.Host,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed SMTP auth: %w", err)
	}

	if err = client.Mail(g.User); err != nil {
		return fmt.Errorf("failed mail command: %w", err)
	}

	if err = client.Rcpt(email); err != nil {
		return fmt.Errorf("failed rcpt command: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed data command: %w", err)
	}

	_, err = w.Write(message)
	if err != nil {
		return fmt.Errorf("failed writing message body: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed closing data writer: %w", err)
	}

	logger.Info("Email OTP delivered via Gmail SMTP to %s", logger.MaskEmail(email))
	return nil
}
