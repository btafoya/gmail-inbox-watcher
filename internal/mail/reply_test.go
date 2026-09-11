package mail

import "testing"

func TestReplySubject(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Hello", "Re: Hello"},
		{"Re: Hello", "Re: Hello"},
		{"re: Hello", "re: Hello"},
		{"", "Re:"},
	}
	for _, tt := range tests {
		if got := ReplySubject(tt.in); got != tt.want {
			t.Errorf("ReplySubject(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestReplyHeaders(t *testing.T) {
	h, err := ReplyHeaders(Message{
		From:       "alice@example.com",
		Subject:    "Hello",
		MessageID:  "<abc@example.com>",
		References: "<old@example.com>",
	}, "me@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if h["To"] != "alice@example.com" {
		t.Fatal("wrong To")
	}
	if h["In-Reply-To"] != "<abc@example.com>" {
		t.Fatal("missing In-Reply-To")
	}
	if h["References"] != "<old@example.com> <abc@example.com>" {
		t.Fatalf("wrong References: %q", h["References"])
	}
}
