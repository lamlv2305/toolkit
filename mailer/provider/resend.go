package provider

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2" // Import resty v2
	"github.com/lamlv2305/toolkit/v2/mailer"
)

type Resend struct {
	APIKey string
	client *resty.Client // Add a resty client
}

func NewResend(apiKey string) *Resend {
	return &Resend{
		APIKey: apiKey,
		client: resty.New(), // Initialize the resty client
	}
}

func (r *Resend) Name() string {
	return "Resend"
}

func (r *Resend) Send(ctx context.Context, email mailer.Email) error {
	body := map[string]any{
		"from":    email.From,
		"to":      []string{email.To},
		"subject": email.Subject,
		"text":    email.Text,
		"html":    email.HTML,
	}

	resp, err := r.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+r.APIKey).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("https://api.resend.com/emails")
	if err != nil {
		return fmt.Errorf("failed to send email via Resend: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("resend failed with status %s: %s", resp.Status(), resp.String())
	}

	return nil
}
