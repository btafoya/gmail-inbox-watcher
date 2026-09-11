package mail

import (
	"fmt"
	"net/mail"
	"strings"
)

type Message struct {
	UID        uint32
	From       string
	FromName   string
	To         string
	Subject    string
	MessageID  string
	References string
}

func ReplySubject(subject string) string {
	trimmed := strings.TrimSpace(subject)
	if strings.HasPrefix(strings.ToLower(trimmed), "re:") {
		return trimmed
	}
	if trimmed == "" {
		return "Re:"
	}
	return "Re: " + trimmed
}

func ReplyHeaders(msg Message, from string) (map[string]string, error) {
	if _, err := mail.ParseAddress(msg.From); err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}
	headers := map[string]string{
		"From":    from,
		"To":      msg.From,
		"Subject": ReplySubject(msg.Subject),
	}
	if msg.MessageID != "" {
		headers["In-Reply-To"] = msg.MessageID
		if msg.References != "" {
			headers["References"] = msg.References + " " + msg.MessageID
		} else {
			headers["References"] = msg.MessageID
		}
	}
	return headers, nil
}
