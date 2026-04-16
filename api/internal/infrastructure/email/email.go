package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

// Config holds SMTP configuration
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// Service handles sending emails
type Service struct {
	config *Config
}

// NewService creates a new email service
func NewService(config *Config) *Service {
	return &Service{config: config}
}

// IsConfigured returns true if SMTP is properly configured
func (s *Service) IsConfigured() bool {
	return s.config.Host != "" && s.config.From != ""
}

// SendInvite sends an invitation email
func (s *Service) SendInvite(to, inviterName, tenantName, role, acceptURL string) error {
	subject := fmt.Sprintf("You've been invited to join %s on FlagFlash", tenantName)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #0f0f23; color: #e2e8f0; padding: 40px 20px;">
  <div style="max-width: 520px; margin: 0 auto; background: #1a1a2e; border-radius: 12px; padding: 32px; border: 1px solid #2d2d44;">
    <div style="text-align: center; margin-bottom: 24px;">
      <span style="font-size: 24px; font-weight: bold; color: #a855f7;">⚡ FlagFlash</span>
    </div>
    <h2 style="color: #f1f5f9; margin-bottom: 16px;">You're Invited!</h2>
    <p style="color: #94a3b8; line-height: 1.6;">
      <strong style="color: #e2e8f0;">%s</strong> has invited you to join
      <strong style="color: #e2e8f0;">%s</strong> as a <strong style="color: #a855f7;">%s</strong>.
    </p>
    <div style="text-align: center; margin: 32px 0;">
      <a href="%s" style="display: inline-block; background: linear-gradient(135deg, #a855f7, #7c3aed); color: white; text-decoration: none; padding: 12px 32px; border-radius: 8px; font-weight: 600; font-size: 16px;">
        Accept Invitation
      </a>
    </div>
    <p style="color: #64748b; font-size: 13px; text-align: center;">
      This invitation expires in 7 days. If you didn't expect this, you can safely ignore this email.
    </p>
  </div>
</body>
</html>`, inviterName, tenantName, role, acceptURL)

	return s.sendHTML(to, subject, body)
}

func (s *Service) sendHTML(to, subject, htmlBody string) error {
	headers := []string{
		fmt.Sprintf("From: %s", s.config.From),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}

	msg := []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + htmlBody)

	addr := s.config.Host + ":" + s.config.Port

	var auth smtp.Auth
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	return smtp.SendMail(addr, auth, s.config.From, []string{to}, msg)
}
