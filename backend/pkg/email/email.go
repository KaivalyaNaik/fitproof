package email

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wneessen/go-mail"
)

// Sender sends transactional emails via SMTP.
// When SMTP_HOST is empty, sending is disabled — all calls log and return nil.
type Sender struct {
	host     string
	port     int
	username string
	password string
	from     string
	enabled  bool
}

func NewSender(host string, port int, username, password, from string) *Sender {
	return &Sender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		enabled:  host != "",
	}
}

func (s *Sender) SendVerificationCode(ctx context.Context, toEmail, displayName, code string) error {
	if !s.enabled {
		slog.Info("email disabled — skipping verification send", slog.String("to", toEmail), slog.String("code", code))
		return nil
	}

	m := mail.NewMsg()
	if err := m.From(s.from); err != nil {
		return fmt.Errorf("email from: %w", err)
	}
	if err := m.To(toEmail); err != nil {
		return fmt.Errorf("email to: %w", err)
	}
	m.Subject("Your FitProof verification code")
	m.SetBodyString(mail.TypeTextPlain, fmt.Sprintf(
		"Hi %s,\n\nYour verification code is: %s\n\nThis code expires in 15 minutes.\n\nIf you didn't request this, you can safely ignore this email.\n\n— FitProof",
		displayName, code,
	))
	m.AddAlternativeString(mail.TypeTextHTML, fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;max-width:480px;margin:0 auto;padding:32px;color:#09090b">
  <p style="font-size:14px;color:#71717a;margin-bottom:4px">FitProof</p>
  <h2 style="font-size:20px;font-weight:600;margin:0 0 24px">Verify your email</h2>
  <p>Hi %s,</p>
  <p>Enter this code to verify your email address:</p>
  <div style="background:#f4f4f5;border-radius:12px;padding:24px;text-align:center;margin:24px 0">
    <span style="font-size:36px;font-weight:700;letter-spacing:8px;font-family:monospace">%s</span>
  </div>
  <p style="color:#71717a;font-size:13px">This code expires in 15 minutes. If you didn't request this, you can safely ignore this email.</p>
</body>
</html>`, displayName, code))

	c, err := mail.NewClient(s.host,
		mail.WithPort(s.port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(s.username),
		mail.WithPassword(s.password),
		mail.WithTLSPolicy(mail.TLSOpportunistic),
	)
	if err != nil {
		return fmt.Errorf("email client: %w", err)
	}
	return c.DialAndSendWithContext(ctx, m)
}
