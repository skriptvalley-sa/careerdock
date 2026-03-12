package email

import (
	"fmt"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// VerificationEmail builds an email verification message.
func VerificationEmail(to, token, baseURL string) *domain.EmailMessage {
	link := fmt.Sprintf("%s/verify-email/%s", baseURL, token)
	return &domain.EmailMessage{
		To:      to,
		Subject: "Verify your CareerDock email",
		HTML: fmt.Sprintf(`<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
<h2>Welcome to CareerDock!</h2>
<p>Please verify your email address by clicking the link below:</p>
<p><a href="%s" style="display: inline-block; padding: 12px 24px; background-color: #2563eb; color: white; text-decoration: none; border-radius: 6px;">Verify Email</a></p>
<p style="color: #6b7280; font-size: 14px;">This link expires in 24 hours. If you didn't create a CareerDock account, you can safely ignore this email.</p>
</div>`, link),
	}
}

// PasswordResetEmail builds a password reset message.
func PasswordResetEmail(to, token, baseURL string) *domain.EmailMessage {
	link := fmt.Sprintf("%s/reset-password/%s", baseURL, token)
	return &domain.EmailMessage{
		To:      to,
		Subject: "Reset your CareerDock password",
		HTML: fmt.Sprintf(`<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
<h2>Password Reset</h2>
<p>You requested a password reset. Click the link below to set a new password:</p>
<p><a href="%s" style="display: inline-block; padding: 12px 24px; background-color: #2563eb; color: white; text-decoration: none; border-radius: 6px;">Reset Password</a></p>
<p style="color: #6b7280; font-size: 14px;">This link expires in 1 hour. If you didn't request a password reset, you can safely ignore this email.</p>
</div>`, link),
	}
}
