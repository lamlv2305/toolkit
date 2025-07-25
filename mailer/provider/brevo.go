package provider

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2" // Import resty v2
	"github.com/lamlv2305/toolkit/v2/mailer"
)

type Brevo struct {
	APIKey string
	client *resty.Client // Add a resty client
}

func NewBrevo(apiKey string) *Brevo {
	return &Brevo{
		APIKey: apiKey,
		client: resty.New(), // Initialize the resty client
	}
}

func (b *Brevo) Name() string {
	return "Brevo"
}

func (b *Brevo) Send(ctx context.Context, email mailer.Email) error {
	body := map[string]any{
		"sender": map[string]string{
			"name":  "OTP Service",
			"email": email.From,
		},
		"to": []map[string]string{
			{"email": email.To},
		},
		"subject":     email.Subject,
		"textContent": email.Text,
		"htmlContent": email.HTML,
	}

	resp, err := b.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("api-key", b.APIKey).
		SetBody(body).
		Post("https://api.brevo.com/v3/smtp/email")
	if err != nil {
		return fmt.Errorf("failed to send email via Brevo: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("brevo send failed with status %s: %s", resp.Status(), resp.String())
	}

	return nil
}
