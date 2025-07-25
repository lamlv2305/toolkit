package provider

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/lamlv2305/toolkit/v2/mailer"
)

type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewSMTP(host string, port int, user, pass, from string) *SMTP {
	return &SMTP{Host: host, Port: port, Username: user, Password: pass, From: from}
}

func (s *SMTP) Name() string {
	return "SMTP"
}

func (s *SMTP) Send(ctx context.Context, email mailer.Email) error {
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", email.To, email.Subject, email.Text))

	return smtp.SendMail(addr, auth, s.From, []string{email.To}, msg)
}
