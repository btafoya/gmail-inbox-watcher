package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/btafoya/gmail-inbox-watcher/internal/allowlist"
	"github.com/btafoya/gmail-inbox-watcher/internal/config"
	"github.com/btafoya/gmail-inbox-watcher/internal/envfile"
	"github.com/btafoya/gmail-inbox-watcher/internal/sender"
	"github.com/btafoya/gmail-inbox-watcher/internal/state"
	"github.com/btafoya/gmail-inbox-watcher/internal/watcher"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "watch but do not send replies or persist processed messages")
	checkConfig := flag.Bool("check-config", false, "validate configuration and files, then exit")
	flag.Parse()

	if err := envfile.Load(".env"); err != nil {
		_, _ = os.Stderr.WriteString("failed to load .env: " + err.Error() + "\n")
		os.Exit(2)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration error", "error", err)
		os.Exit(2)
	}
	if *dryRun {
		cfg.DryRun = true
	}

	if *checkConfig {
		if _, err := allowlist.Load(cfg.SendersFile); err != nil {
			logger.Error("sender allowlist error", "error", err)
			os.Exit(2)
		}
		if _, err := os.ReadFile(cfg.ReplyFile); err != nil {
			logger.Error("reply file error", "error", err)
			os.Exit(2)
		}
		logger.Info("configuration is valid")
		return
	}

	logLevel := slog.LevelInfo
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))

	list, err := allowlist.Load(cfg.SendersFile)
	if err != nil {
		logger.Error("load sender allowlist", "error", err)
		os.Exit(2)
	}
	replyBody, err := os.ReadFile(cfg.ReplyFile)
	if err != nil {
		logger.Error("load reply body", "error", err)
		os.Exit(2)
	}
	st, err := state.Open(cfg.StateFile)
	if err != nil {
		logger.Error("open state", "error", err)
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	w := &watcher.Watcher{
		Username:  cfg.Username,
		Password:  cfg.AppPassword,
		IMAPHost:  cfg.IMAPHost,
		IMAPPort:  cfg.IMAPPort,
		Allowlist: list,
		State:     st,
		ReplyBody: string(replyBody),
		Sender: sender.Sender{
			Host: cfg.SMTPHost, Port: cfg.SMTPPort,
			Username: cfg.Username, Password: cfg.AppPassword,
		},
		DryRun:      cfg.DryRun,
		InitialScan: cfg.InitialScan,
		Reconnect:   cfg.ReconnectDelay,
		Logger:      logger,
	}

	logger.Info("starting Gmail inbox watcher", "dry_run", cfg.DryRun, "allowed_senders", list.Len())
	if err := w.Run(ctx); err != nil {
		logger.Error("watcher stopped", "error", err)
		os.Exit(1)
	}
	logger.Info("watcher stopped")
}
