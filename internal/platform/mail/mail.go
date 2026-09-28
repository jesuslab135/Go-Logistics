// Package mail sends a message with attachments over SMTP, using only the
// standard library. It knows nothing about reports.
package mail

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

// sendTimeout bounds one whole delivery. A mail server that accepts the
// connection and then says nothing must not hold the scheduler's tick open.
const sendTimeout = 60 * time.Second

// implicitTLSPort is the one port where TLS starts before the first byte.
// Every other port is upgraded with STARTTLS.
const implicitTLSPort = 465

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// Enabled reports whether a mail server is configured at all. When it is not,
// reports are still built and their runs recorded as not sent.
func (c Config) Enabled() bool { return c.Host != "" }

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type Message struct {
	To          []string
	Subject     string
	Body        string
	Attachments []Attachment
}

// Sender is what the scheduler depends on, so a test can record messages
// instead of sending them.
type Sender interface {
	Send(ctx context.Context, m Message) error
}

// CleanAddress returns the bare address of s, or an error if s is anything
// other than one plain address.
//
// A display name or a line break is refused rather than stripped. The value is
// written into a message header, where a line break starts a new header: an
// address of "a@b.mx\r\nBcc: x@y.mx" would otherwise add a recipient.
func CleanAddress(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("mail: address is empty")
	}
	if strings.ContainsAny(s, "\r\n") {
		return "", errors.New("mail: address contains a line break")
	}
	a, err := mail.ParseAddress(s)
	if err != nil {
		return "", fmt.Errorf("mail: %q is not an email address", s)
	}
	if a.Name != "" || a.Address != s {
		return "", fmt.Errorf("mail: %q must be a bare address, with no name", s)
	}
	return a.Address, nil
}

// Build renders m as an RFC 5322 message: a text body, then each attachment
// base64-encoded.
func Build(from string, m Message, now time.Time) ([]byte, error) {
	from, err := CleanAddress(from)
	if err != nil {
		return nil, fmt.Errorf("mail: from: %w", err)
	}
	if len(m.To) == 0 {
		return nil, errors.New("mail: no recipients")
	}
	to := make([]string, 0, len(m.To))
	for _, addr := range m.To {
		clean, err := CleanAddress(addr)
		if err != nil {
			return nil, err
		}
		to = append(to, clean)
	}
	if strings.ContainsAny(m.Subject, "\r\n") {
		return nil, errors.New("mail: subject contains a line break")
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	text, err := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/plain; charset=utf-8"},
		"Content-Transfer-Encoding": {"quoted-printable"},
	})
	if err != nil {
		return nil, err
	}
	qp := quotedprintable.NewWriter(text)
	if _, err := qp.Write([]byte(m.Body)); err != nil {
		return nil, err
	}
	if err := qp.Close(); err != nil {
		return nil, err
	}

	for _, a := range m.Attachments {
		name := strings.NewReplacer("\r", "", "\n", "", `"`, "").Replace(a.Filename)
		part, err := mw.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {a.ContentType},
			"Content-Transfer-Encoding": {"base64"},
			"Content-Disposition":       {mime.FormatMediaType("attachment", map[string]string{"filename": name})},
		})
		if err != nil {
			return nil, err
		}
		if err := writeBase64Lines(part, a.Data); err != nil {
			return nil, err
		}
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	domain := from[strings.LastIndexByte(from, '@')+1:]
	var out bytes.Buffer
	fmt.Fprintf(&out, "From: %s\r\n", from)
	fmt.Fprintf(&out, "To: %s\r\n", strings.Join(to, ", "))
	fmt.Fprintf(&out, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", m.Subject))
	fmt.Fprintf(&out, "Date: %s\r\n", now.Format(time.RFC1123Z))
	fmt.Fprintf(&out, "Message-ID: <%s@%s>\r\n", randomID(), domain)
	fmt.Fprintf(&out, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&out, "Content-Type: multipart/mixed; boundary=%q\r\n\r\n", mw.Boundary())
	out.Write(body.Bytes())
	return out.Bytes(), nil
}

// writeBase64Lines wraps at 76 characters, the line length RFC 2045 allows.
func writeBase64Lines(w interface{ Write([]byte) (int, error) }, data []byte) error {
	const width = 76
	encoded := base64.StdEncoding.EncodeToString(data)
	for len(encoded) > 0 {
		n := min(width, len(encoded))
		if _, err := w.Write([]byte(encoded[:n] + "\r\n")); err != nil {
			return err
		}
		encoded = encoded[n:]
	}
	return nil
}

func randomID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// requiresTLS reports whether TLS is required for the given host. Loopback
// addresses (localhost, 127.0.0.0/8, ::1) are allowed to send without TLS,
// to support local relay servers and testing. All other addresses must use TLS
// to protect fleet cost data in transit.
func requiresTLS(host string) bool {
	if host == "localhost" {
		return false
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return false
	}
	return true
}

// SMTP is the Sender that talks to a real server.
type SMTP struct {
	cfg Config
	now func() time.Time
}

func NewSMTP(cfg Config) *SMTP {
	return &SMTP{cfg: cfg, now: time.Now}
}

func (s *SMTP) Send(ctx context.Context, m Message) error {
	raw, err := Build(s.cfg.From, m, s.now())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("mail: connect to %s: %w", addr, err)
	}
	defer conn.Close()
	// net/smtp takes no context, so the deadline on the connection is what
	// enforces the timeout for everything after the dial.
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	tlsConfig := &tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}
	if s.cfg.Port == implicitTLSPort {
		conn = tls.Client(conn, tlsConfig)
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("mail: greeting from %s: %w", addr, err)
	}
	defer client.Close()

	if s.cfg.Port != implicitTLSPort {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("mail: starttls: %w", err)
			}
		} else if requiresTLS(s.cfg.Host) {
			return fmt.Errorf("mail: %s does not offer STARTTLS; refusing to send in cleartext (use port 465 for implicit TLS)", s.cfg.Host)
		}
	}
	if s.cfg.User != "" {
		// PlainAuth itself refuses to send the password over a connection
		// that is neither TLS nor localhost.
		auth := smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("mail: authenticate: %w", err)
		}
	}

	if err := client.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("mail: sender refused: %w", err)
	}
	for _, to := range m.To {
		if err := client.Rcpt(strings.TrimSpace(to)); err != nil {
			return fmt.Errorf("mail: recipient %s refused: %w", to, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("mail: data: %w", err)
	}
	if _, err := w.Write(raw); err != nil {
		return fmt.Errorf("mail: write message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("mail: message refused: %w", err)
	}
	return client.Quit()
}
