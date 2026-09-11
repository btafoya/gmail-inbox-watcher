package config

import "testing"

func TestValidateRequiresCredentials(t *testing.T) {
	c := Config{
		IMAPHost: "imap.gmail.com", IMAPPort: 993,
		SMTPHost: "smtp.gmail.com", SMTPPort: 465,
		SendersFile: "senders.txt", ReplyFile: "reply.txt", StateFile: "state.json",
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected missing credentials error")
	}
}
