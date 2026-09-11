package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Username           string
	AppPassword        string
	IMAPHost           string
	IMAPPort           int
	SMTPHost           string
	SMTPPort           int
	SendersFile        string
	ReplyFile          string
	StateFile          string
	LogLevel           string
	DryRun             bool
	InitialScan        bool
	ReconnectDelay     time.Duration
	ReplySubjectPrefix string
}

func Load() (Config, error) {
	c := Config{
		Username:           os.Getenv("GMAIL_USERNAME"),
		AppPassword:        os.Getenv("GMAIL_APP_PASSWORD"),
		IMAPHost:           env("IMAP_HOST", "imap.gmail.com"),
		SMTPHost:           env("SMTP_HOST", "smtp.gmail.com"),
		SendersFile:        env("SENDERS_FILE", "senders.txt"),
		ReplyFile:          env("REPLY_FILE", "reply.txt"),
		StateFile:          env("STATE_FILE", "state.json"),
		LogLevel:           env("LOG_LEVEL", "info"),
		ReplySubjectPrefix: env("REPLY_SUBJECT_PREFIX", "Re:"),
	}
	var err error
	c.IMAPPort, err = intEnv("IMAP_PORT", 993)
	if err != nil {
		return Config{}, err
	}
	c.SMTPPort, err = intEnv("SMTP_PORT", 465)
	if err != nil {
		return Config{}, err
	}
	c.DryRun, err = boolEnv("DRY_RUN", false)
	if err != nil {
		return Config{}, err
	}
	c.InitialScan, err = boolEnv("INITIAL_SCAN", true)
	if err != nil {
		return Config{}, err
	}
	c.ReconnectDelay, err = durationEnv("RECONNECT_DELAY", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	return c, c.Validate()
}

func (c Config) Validate() error {
	if c.Username == "" {
		return errors.New("GMAIL_USERNAME is required")
	}
	if c.AppPassword == "" {
		return errors.New("GMAIL_APP_PASSWORD is required")
	}
	if c.IMAPHost == "" || c.IMAPPort <= 0 || c.SMTPHost == "" || c.SMTPPort <= 0 {
		return errors.New("invalid mail server configuration")
	}
	if c.SendorsFile() == "" || c.ReplyFile == "" || c.StateFile == "" {
		return errors.New("mail file paths must not be empty")
	}
	return nil
}

func (c Config) SendorsFile() string { return c.SendersFile }

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func intEnv(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return n, nil
}

func boolEnv(key string, fallback bool) (bool, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("%s must be boolean: %w", key, err)
	}
	return b, nil
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	return d, nil
}
