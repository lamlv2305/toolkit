package provider

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/lamlv2305/toolkit/v2/mailer"
)

type SES struct {
	SMTPHost string // e.g., "email-smtp.us-east-1.amazonaws.com"
	SMTPPort int    // usually 587 or 465
	Username string // SMTP username from SES
	Password string // SMTP password from SES
	From     string
}

func NewSES(host string, port int, user, pass, from string) *SES {
	return &SES{
		SMTPHost: host,
		SMTPPort: port,
		Username: user,
		Password: pass,
		From:     from,
	}
}

func (s *SES) Name() string {
	return "Amazon SES"
}

func (s *SES) Send(ctx context.Context, email mailer.Email) error {
	auth := smtp.PlainAuth("", s.Username, s.Password, s.SMTPHost)
	addr := fmt.Sprintf("%s:%d", s.SMTPHost, s.SMTPPort)

	msg := []byte(fmt.Sprintf(
		"To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		email.To, email.Subject, email.Text,
	))

	return smtp.SendMail(addr, auth, s.From, []string{email.To}, msg)
}
