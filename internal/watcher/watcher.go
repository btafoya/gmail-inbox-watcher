package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"strings"
	"time"

	"github.com/btafoya/gmail-inbox-watcher/internal/allowlist"
	"github.com/btafoya/gmail-inbox-watcher/internal/mail"
	"github.com/btafoya/gmail-inbox-watcher/internal/sender"
	"github.com/btafoya/gmail-inbox-watcher/internal/state"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/charset"
)

type Watcher struct {
	Username    string
	Password    string
	IMAPHost    string
	IMAPPort    int
	Allowlist   allowlist.List
	State       *state.State
	ReplyBody   string
	Sender      sender.Sender
	DryRun      bool
	InitialScan bool
	Reconnect   time.Duration
	Logger      *slog.Logger
}

func (w *Watcher) Run(ctx context.Context) error {
	for {
		err := w.runConnection(ctx)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			w.Logger.Error("watcher connection ended", "error", err)
		}
		timer := time.NewTimer(w.Reconnect)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}

func (w *Watcher) runConnection(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", w.IMAPHost, w.IMAPPort)
	options := &imapclient.Options{
		WordDecoder: &mime.WordDecoder{CharsetReader: charset.Reader},
	}
	c, err := imapclient.DialTLS(addr, options)
	if err != nil {
		return fmt.Errorf("dial IMAP: %w", err)
	}
	defer c.Close()

	if err := c.Login(w.Username, w.Password).Wait(); err != nil {
		return fmt.Errorf("IMAP login: %w", err)
	}
	defer func() { _ = c.Logout().Wait() }()

	if _, err := c.Select("INBOX", nil).Wait(); err != nil {
		return fmt.Errorf("select INBOX: %w", err)
	}

	w.Logger.Info("connected to Gmail IMAP", "mailbox", "INBOX")

	if w.InitialScan {
		if err := w.processNew(ctx, c, true); err != nil {
			return err
		}
	}

	for {
		if ctx.Err() != nil {
			return nil
		}

		idleCmd, err := c.Idle()
		if err != nil {
			return fmt.Errorf("start IMAP IDLE: %w", err)
		}

		done := make(chan error, 1)
		go func() { done <- idleCmd.Wait() }()

		select {
		case <-ctx.Done():
			_ = idleCmd.Close()
			<-done
			return nil
		case err := <-done:
			if err != nil {
				return fmt.Errorf("IMAP IDLE: %w", err)
			}
			continue
		case <-time.After(30 * time.Minute):
			if err := idleCmd.Close(); err != nil {
				return fmt.Errorf("stop IMAP IDLE: %w", err)
			}
			if err := <-done; err != nil {
				return fmt.Errorf("finish IMAP IDLE: %w", err)
			}
		}

		if err := w.processNew(ctx, c, false); err != nil {
			return err
		}
	}
}

func (w *Watcher) processNew(ctx context.Context, c *imapclient.Client, initial bool) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// UID SEARCH for messages without the Seen flag (equivalent to UNSEEN).
	criteria := &imap.SearchCriteria{NotFlag: []imap.Flag{imap.FlagSeen}}

	data, err := c.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return fmt.Errorf("search new mail: %w", err)
	}

	uids := data.AllUIDs()
	if len(uids) == 0 {
		return nil
	}

	fetchOpts := &imap.FetchOptions{
		UID:      true,
		Envelope: true,
		Flags:    true,
	}
	msgs, err := c.Fetch(imap.UIDSetNum(uids...), fetchOpts).Collect()
	if err != nil {
		return fmt.Errorf("fetch mail: %w", err)
	}

	for _, msg := range msgs {
		if err := w.handleMessage(ctx, c, msg); err != nil {
			w.Logger.Error("message processing failed", "error", err)
			// Continue with other messages; a transient failure for one mail
			// should not prevent the watcher from processing subsequent mail.
		}
	}
	_ = initial
	return nil
}

func (w *Watcher) handleMessage(ctx context.Context, c *imapclient.Client, msg *imapclient.FetchMessageBuffer) error {
	if msg.UID == 0 || w.State.Has(uint32(msg.UID)) {
		return nil
	}
	if msg.Envelope == nil || len(msg.Envelope.From) == 0 {
		return nil
	}

	from := msg.Envelope.From[0].Addr()
	if from == "" || !w.Allowlist.Contains(from) {
		return nil
	}

	headers := mail.Message{
		UID:        uint32(msg.UID),
		From:       from,
		Subject:    msg.Envelope.Subject,
		MessageID:  msg.Envelope.MessageID,
		References: strings.Join(msg.Envelope.InReplyTo, " "),
	}

	replyHeaders, err := mail.ReplyHeaders(headers, w.Username)
	if err != nil {
		return err
	}

	w.Logger.Info("matched sender", "from", from, "subject", msg.Envelope.Subject, "uid", msg.UID)

	if w.DryRun {
		w.Logger.Info("dry run: would send reply", "to", from)
		return nil
	}

	if err := w.Sender.Send(from, replyHeaders, w.ReplyBody); err != nil {
		return fmt.Errorf("send reply to %s: %w", from, err)
	}

	storeFlags := &imap.StoreFlags{Op: imap.StoreFlagsAdd, Silent: true, Flags: []imap.Flag{imap.FlagSeen}}
	if _, err := c.Store(imap.UIDSetNum(msg.UID), storeFlags, nil).Collect(); err != nil {
		return fmt.Errorf("mark UID %d seen: %w", msg.UID, err)
	}

	if err := w.State.Add(uint32(msg.UID)); err != nil {
		return fmt.Errorf("persist processed UID %d: %w", msg.UID, err)
	}

	w.Logger.Info("automatic reply sent", "to", from, "uid", msg.UID)
	_ = ctx
	return nil
}
