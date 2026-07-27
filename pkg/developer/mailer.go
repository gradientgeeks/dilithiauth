package developer

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
)

// Mailer handles sending emails via OCI Email Delivery (SMTP).
type Mailer struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// NewMailerFromEnv reads OCI Email Delivery SMTP config from environment variables.
func NewMailerFromEnv() *Mailer {
	host := os.Getenv("OCI_SMTP_HOST")
	if host == "" {
		return nil // Mailer disabled — codes logged to stdout
	}

	port, _ := strconv.Atoi(os.Getenv("OCI_SMTP_PORT"))
	if port == 0 {
		port = 587 // default OCI STARTTLS port
	}

	return &Mailer{
		Host:     host,
		Port:     port,
		Username: os.Getenv("OCI_SMTP_USER"),
		Password: os.Getenv("OCI_SMTP_PASS"),
		From:     os.Getenv("OCI_SMTP_FROM"),
	}
}

// SendVerificationOTP sends a 6-digit verification code email via OCI Email Delivery.
func (m *Mailer) SendVerificationOTP(toEmail, code string) error {
	if m == nil || m.Host == "" {
		log.Printf("[DEV MODE - OCI MAIL DISABLED] Verification code for %s: %s", toEmail, code)
		return nil
	}

	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)

	subject := "Your DilithiAuth Verification Code: " + code
	body := fmt.Sprintf(`Hi,

Your DilithiAuth 6-digit email verification code is:

    %s

This code is valid for 15 minutes. If you did not request this, please ignore this email.

Secured by DilithiAuth · Post-Quantum Identity Provider (ML-DSA-44)
`, code)

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", m.From, toEmail, subject, body))

	err := smtp.SendMail(addr, auth, m.From, []string{toEmail}, msg)
	if err != nil {
		log.Printf("ERROR sending OCI email to %s: %v", toEmail, err)
		return fmt.Errorf("oci_email: %w", err)
	}

	log.Printf("[OCI EMAIL DELIVERED] Verification code sent to %s via %s", toEmail, m.Host)
	return nil
}
