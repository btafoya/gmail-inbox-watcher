# Gmail Inbox Watcher

A small Go service that watches a Gmail INBOX over IMAPS and automatically replies to messages from addresses listed in `senders.txt`.

Why? I was in need of a away notification outside of the default vacation function in gMail that didn't send replies to everyone. And this worked!

## Design

- Gmail IMAP over TLS (`imap.gmail.com:993`)
- IMAP IDLE for near-real-time new-message detection
- Gmail SMTP over implicit TLS (`smtp.gmail.com:465`) for replies
- Gmail app-password authentication
- Exact, case-insensitive sender matching
- `senders.txt`: one sender address per line
- `reply.txt`: response body
- Persistent UID state to prevent duplicate replies
- Reconnects after connection failures
- Dry-run mode
- Structured tests around sender matching, subject handling, and state

Google currently supports IMAP for Gmail accounts; IMAP access is always enabled for personal Gmail accounts, and Google documents app passwords as an option for compatible clients using 2-Step Verification. Gmail IMAP uses TLS on port 993 and SMTP supports TLS on ports 465/587.

## Requirements

- Go 1.24+
- Gmail account with 2-Step Verification
- Gmail app password
- IMAP access available for the account

## Quick start

```bash
git clone https://github.com/btafoya/gmail-inbox-watcher.git
cd gmail-inbox-watcher

cp .env.example .env
cp senders.txt.example senders.txt
cp reply.txt.example reply.txt

# Edit .env, senders.txt, and reply.txt
go build -o gmail-watcher ./cmd/gmail-watcher

./gmail-watcher
```

Dry run:

```bash
./gmail-watcher --dry-run
```

Validate configuration without connecting:

```bash
./gmail-watcher --check-config
```

## Configuration

Environment variables:

| Variable | Default | Description |
|---|---|---|
| `GMAIL_USERNAME` | required | Gmail address |
| `GMAIL_APP_PASSWORD` | required | 16-character Google app password |
| `IMAP_HOST` | `imap.gmail.com` | IMAP server |
| `IMAP_PORT` | `993` | IMAPS port |
| `SMTP_HOST` | `smtp.gmail.com` | SMTP server |
| `SMTP_PORT` | `465` | SMTP implicit TLS port |
| `SENDERS_FILE` | `senders.txt` | Sender allowlist |
| `REPLY_FILE` | `reply.txt` | Reply body |
| `STATE_FILE` | `state.json` | Persistent processed UID state |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `DRY_RUN` | `false` | Do not send or mark messages processed |
| `REPLY_SUBJECT_PREFIX` | `Re:` | Prefix added when needed |
| `INITIAL_SCAN` | `true` | Process unread/new messages present when starting |
| `RECONNECT_DELAY` | `10s` | Delay between reconnect attempts |

A `.env` file is supported by the project, but the program intentionally does not log its contents.

## Sender file

`senders.txt`:

```text
# Exact email addresses
alice@example.com
bob@example.org
```

Blank lines and `#` comments are ignored.

The initial implementation intentionally uses exact address matching. It does not automatically match whole domains or wildcard patterns.

## Reply file

`reply.txt`:

```text
Thanks for your message.

I received your email and will get back to you as soon as possible.

-- Brian
```

The reply is sent as plain text.

## Duplicate protection

The watcher records processed Gmail message UIDs in `state.json`. A message is only recorded after the reply succeeds.

If the process crashes before the state update, a retry may occur. The implementation also sets the `\Seen` flag after successful processing. It does not claim exactly-once delivery because SMTP/IMAP cannot provide an atomic transaction spanning both systems.

## Gmail setup

1. Enable 2-Step Verification on the Google account.
2. Create an App Password for this watcher.
3. Put the app password in `.env`.
4. Do not use the normal Gmail password.

Do not commit `.env`.

## systemd

See `deploy/gmail-watcher.service`.

Install:

```bash
sudo install -m 0755 gmail-watcher /usr/local/bin/gmail-watcher
sudo install -d -m 0750 /etc/gmail-watcher
sudo cp .env senders.txt reply.txt /etc/gmail-watcher/
sudo cp deploy/gmail-watcher.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now gmail-watcher
```

Review and adjust the service's `User`, `Group`, and `WorkingDirectory` for your host.

## Security

- Never commit `.env`.
- Restrict `.env` permissions to the service account.
- Do not put the app password on the command line.
- Logs never print authentication credentials.
- TLS certificate verification is enabled by default.
- The SMTP client uses Gmail's implicit TLS endpoint on port 465.

## Limitations

This is deliberately a simple responder, not a full mail client.

It does not:
- send HTML mail
- download attachments
- interpret arbitrary MIME content beyond the fields needed for matching/reply threading
- support OAuth2
- support multiple Gmail accounts
- support wildcard/domain sender rules
- guarantee exactly-once email delivery

The architecture keeps those features easy to add later.

## License

[MIT](LICENSE)
