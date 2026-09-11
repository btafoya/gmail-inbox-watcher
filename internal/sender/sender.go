package sender

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/smtp"

	"github.com/emersion/go-message/mail"
)

type Sender struct {
	Host     string
	Port     int
	Username string
	Password string
}

func (s Sender) Send(to string, headers map[string]string, body string) error {
	addr := net.JoinHostPort(s.Host, fmt.Sprint(s.Port))
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: s.Host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return fmt.Errorf("connect SMTP: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return fmt.Errorf("create SMTP client: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication: %w", err)
	}
	if err := client.Mail(s.Username); err != nil {
		return fmt.Errorf("SMTP MAIL FROM: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT TO: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA: %w", err)
	}

	if err := writeMessage(w, headers, body); err != nil {
		_ = w.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("finish SMTP message: %w", err)
	}
	return client.Quit()
}

func writeMessage(w io.Writer, headers map[string]string, body string) error {
	var h mail.Header
	for _, key := range []string{"From", "To", "In-Reply-To", "References"} {
		if value := headers[key]; value != "" {
			h.Set(key, value)
		}
	}
	h.SetSubject(headers["Subject"])
	h.SetContentType("text/plain", map[string]string{"charset": "utf-8"})

	pw, err := mail.CreateSingleInlineWriter(w, h)
	if err != nil {
		return fmt.Errorf("create MIME writer: %w", err)
	}
	if _, err := io.WriteString(pw, body); err != nil {
		_ = pw.Close()
		return fmt.Errorf("write message body: %w", err)
	}
	return pw.Close()
}
