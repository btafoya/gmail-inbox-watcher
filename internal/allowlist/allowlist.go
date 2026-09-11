package allowlist

import (
	"bufio"
	"fmt"
	"net/mail"
	"os"
	"strings"
)

type List struct {
	addresses map[string]struct{}
}

func Load(path string) (List, error) {
	f, err := os.Open(path)
	if err != nil {
		return List{}, fmt.Errorf("open sender allowlist: %w", err)
	}
	defer f.Close()

	result := List{addresses: make(map[string]struct{})}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	line := 0
	for scanner.Scan() {
		line++
		value := strings.TrimSpace(scanner.Text())
		if value == "" || strings.HasPrefix(value, "#") {
			continue
		}
		if _, err := mail.ParseAddress(value); err != nil {
			return List{}, fmt.Errorf("invalid sender address on line %d: %w", line, err)
		}
		// The file is intentionally restricted to an exact address, not a display name.
		parsed, _ := mail.ParseAddress(value)
		if parsed.Address != value {
			return List{}, fmt.Errorf("line %d must contain an email address only", line)
		}
		result.addresses[strings.ToLower(strings.TrimSpace(parsed.Address))] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return List{}, fmt.Errorf("read sender allowlist: %w", err)
	}
	return result, nil
}

func (l List) Contains(address string) bool {
	_, ok := l.addresses[strings.ToLower(strings.TrimSpace(address))]
	return ok
}

func (l List) Len() int {
	return len(l.addresses)
}
