package sender

import (
	"strings"
	"testing"
)

func TestWriteMessage(t *testing.T) {
	var buf strings.Builder
	err := writeMessage(&buf, map[string]string{
		"From": "me@example.com", "To": "you@example.com",
		"Subject": "Re: Test", "In-Reply-To": "<abc@example.com>",
	}, "Hello")
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()

	for _, want := range []string{
		"From: me@example.com\r\n",
		"To: you@example.com\r\n",
		"Subject: Re: Test\r\n",
		"In-Reply-To: <abc@example.com>\r\n",
		"Mime-Version: 1.0\r\n",
		`Content-Type: text/plain; charset=utf-8` + "\r\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("message missing %q, got:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "Hello") {
		t.Errorf("message missing body, got:\n%s", got)
	}
}
