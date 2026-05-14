// mailer.go
// SMTP mailer (verification + API key email)

package email

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// Config holds SMTP credentials loaded from environment variables.
type Config struct {
	Host     string // e.g. mail.infomaniak.com
	Port     string // e.g. 587
	Username string // full email address used to send
	Password string
	From     string // display sender address
}

// LoadConfig reads SMTP settings from environment variables.
func LoadConfig() Config {
	return Config{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}
}

// SendVerification sends the email verification link to the new user.
func SendVerification(cfg Config, toEmail, name, token string) error {
	gatewayURL := os.Getenv("GATEWAY_URL") // e.g. https://gateway.mythic.com
	link := fmt.Sprintf("%s/verify?token=%s", gatewayURL, token)

	subject := "Verify your Mythic Access Gateway email"

	body := fmt.Sprintf(`Hello %s, 

Thanks for registering with the Mythic Access Gateway.
		
Please verify your email address by clicking the link below:
%s

This Link expires in 24 hours.
		
If you did not register, you can safely ignore this email.
		
-- Mythic Access`, name, link)

	return send(cfg, toEmail, subject, body)
}

// SendAPIKey sends the generated API key to the verified user.
// The key is shown only once - it is never stored in plain text.
func SendAPIKey(cfg Config, toEmail, name, apiKey string) error {
	subject := "Your Mythic Access Gateway API Key"
	body := fmt.Sprintf(`Hello %s,

Your email has been verified. Here is your API key:

%s

Keep it safe - this is the only time it will be sent.
If you lose it you will need to request a new one.

Usage: add the following header to every request:
	X-API-Key: %s

-- Mythic Access`, name, apiKey, apiKey)

	return send(cfg, toEmail, subject, body)
}

// send is the low-level SMTP helper used by all public functions.
func send(cfg Config, to, subject, body string) error {
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)

	msg := strings.Join([]string{
		"From: " + cfg.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	if err := smtp.SendMail(addr, auth, cfg.From, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("SMTP send to %s: %w", to, err)
	}
	return nil
}
