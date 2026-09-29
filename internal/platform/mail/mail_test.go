package mail

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net"
	netmail "net/mail"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var fixedNow = time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)

func TestRequiresTLS(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		// Loopback addresses do not require TLS
		{"localhost", false},
		{"127.0.0.1", false},
		{"127.5.6.7", false},
		{"::1", false},
		// Non-loopback addresses require TLS
		{"smtp.ionos.mx", true},
		{"10.0.0.5", true},
		{"192.168.1.10", true},
		{"mail.example.com", true},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := requiresTLS(tt.host)
			if got != tt.want {
				t.Fatalf("requiresTLS(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestCleanAddress(t *testing.T) {
	good := []string{"flota@empresa.mx", "  flota@empresa.mx  ", "a.b+c@sub.empresa.com.mx"}
	for _, s := range good {
		got, err := CleanAddress(s)
		if err != nil {
			t.Fatalf("%q: %v", s, err)
		}
		if got != strings.TrimSpace(s) {
			t.Fatalf("%q: got %q", s, got)
		}
	}

	bad := []string{
		"",
		"no-at-sign",
		"Flota <flota@empresa.mx>",
		"a@b.mx, c@d.mx",
		"a@b.mx\r\nBcc: x@y.mx",
		"a@b.mx\nBcc: x@y.mx",
	}
	for _, s := range bad {
		if _, err := CleanAddress(s); err == nil {
			t.Fatalf("%q: accepted, want an error", s)
		}
	}
}

func TestBuild(t *testing.T) {
	pdf := bytes.Repeat([]byte("%PDF-1.4 contenido "), 200)
	raw, err := Build("reportes@empresa.mx", Message{
		To:      []string{"a@cliente.mx", "b@cliente.mx"},
		Subject: "Reporte semanal de combustible — Transportes Durán",
		Body:    "Adjunto encontrará el reporte.\nPeriodo: 21/09/2026 a 27/09/2026.\n",
		Attachments: []Attachment{{
			Filename: "combustible-semanal-2026-09-21.pdf", ContentType: "application/pdf", Data: pdf,
		}},
	}, fixedNow)
	if err != nil {
		t.Fatal(err)
	}

	msg, err := netmail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("the message does not parse: %v", err)
	}
	if got := msg.Header.Get("To"); got != "a@cliente.mx, b@cliente.mx" {
		t.Fatalf("To = %q", got)
	}
	subject, err := new(mime.WordDecoder).DecodeHeader(msg.Header.Get("Subject"))
	if err != nil || subject != "Reporte semanal de combustible — Transportes Durán" {
		t.Fatalf("Subject = %q, err = %v", subject, err)
	}
	if msg.Header.Get("Message-ID") == "" || msg.Header.Get("Date") == "" {
		t.Fatal("Message-ID and Date are required by most receiving servers")
	}

	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/mixed" {
		t.Fatalf("Content-Type = %q, err = %v", mediaType, err)
	}
	mr := multipart.NewReader(msg.Body, params["boundary"])

	text, err := mr.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	// multipart.Reader decodes quoted-printable transparently.
	body, _ := io.ReadAll(text)
	if !strings.Contains(string(body), "Adjunto encontrará el reporte.") {
		t.Fatalf("body = %q", body)
	}

	att, err := mr.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if att.FileName() != "combustible-semanal-2026-09-21.pdf" {
		t.Fatalf("filename = %q", att.FileName())
	}
	encoded, _ := io.ReadAll(att)
	for _, line := range strings.Split(strings.TrimSpace(string(encoded)), "\r\n") {
		if len(line) > 76 {
			t.Fatalf("base64 line of %d characters, the limit is 76", len(line))
		}
	}
}

func TestBuildRefusesHeaderInjection(t *testing.T) {
	base := Message{To: []string{"a@cliente.mx"}, Subject: "s", Body: "b"}

	injected := base
	injected.To = []string{"a@cliente.mx\r\nBcc: espia@otro.mx"}
	if _, err := Build("reportes@empresa.mx", injected, fixedNow); err == nil {
		t.Fatal("a recipient with a line break was accepted")
	}

	injected = base
	injected.Subject = "hola\r\nBcc: espia@otro.mx"
	if _, err := Build("reportes@empresa.mx", injected, fixedNow); err == nil {
		t.Fatal("a subject with a line break was accepted")
	}

	if _, err := Build("reportes@empresa.mx", Message{Subject: "s"}, fixedNow); err == nil {
		t.Fatal("a message with no recipients was accepted")
	}
	if _, err := Build("", base, fixedNow); err == nil {
		t.Fatal("an empty sender was accepted")
	}
}

// fakeSMTP is the smallest server net/smtp will complete a delivery against.
// It offers no STARTTLS and no AUTH, so the client sends in the clear, which
// is what a test on the loopback interface wants.
type fakeSMTP struct {
	ln net.Listener

	mu       sync.Mutex
	from     string
	rcpts    []string
	data     string
	silent   bool // accept the connection and never greet
	rejectTo string
	// hangUpAfterData closes the connection right after replying 250 to the
	// end of DATA, without ever answering QUIT — the message was accepted,
	// but the following QUIT then fails.
	hangUpAfterData bool
}

func newFakeSMTP(t *testing.T) *fakeSMTP {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeSMTP{ln: ln}
	t.Cleanup(func() { ln.Close() })
	go f.serve()
	return f
}

func (f *fakeSMTP) config() Config {
	host, port, _ := net.SplitHostPort(f.ln.Addr().String())
	p, _ := strconv.Atoi(port)
	return Config{Host: host, Port: p, From: "reportes@empresa.mx"}
}

func (f *fakeSMTP) serve() {
	for {
		conn, err := f.ln.Accept()
		if err != nil {
			return
		}
		go f.handle(conn)
	}
}

func (f *fakeSMTP) handle(conn net.Conn) {
	defer conn.Close()
	f.mu.Lock()
	silent, rejectTo, hangUp := f.silent, f.rejectTo, f.hangUpAfterData
	f.mu.Unlock()
	if silent {
		_, _ = io.Copy(io.Discard, conn)
		return
	}

	r := bufio.NewReader(conn)
	say := func(s string) { _, _ = conn.Write([]byte(s + "\r\n")) }
	say("220 fake ESMTP")
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
			say("250 fake")
		case strings.HasPrefix(upper, "MAIL FROM:"):
			f.mu.Lock()
			f.from = strings.Trim(line[len("MAIL FROM:"):], "<> ")
			f.mu.Unlock()
			say("250 ok")
		case strings.HasPrefix(upper, "RCPT TO:"):
			addr := strings.Trim(line[len("RCPT TO:"):], "<> ")
			if addr == rejectTo {
				say("550 no such user")
				continue
			}
			f.mu.Lock()
			f.rcpts = append(f.rcpts, addr)
			f.mu.Unlock()
			say("250 ok")
		case upper == "DATA":
			say("354 go ahead")
			var data strings.Builder
			for {
				l, err := r.ReadString('\n')
				if err != nil {
					return
				}
				if l == ".\r\n" {
					break
				}
				data.WriteString(l)
			}
			f.mu.Lock()
			f.data = data.String()
			f.mu.Unlock()
			say("250 queued")
			if hangUp {
				// The message was accepted; hanging up now, instead of
				// answering QUIT, is what a Send that treats a failed
				// QUIT as a send failure would wrongly report as failed.
				return
			}
		case upper == "QUIT":
			say("221 bye")
			return
		default:
			say("500 unknown")
		}
	}
}

func TestSMTPSend(t *testing.T) {
	srv := newFakeSMTP(t)
	sender := NewSMTP(srv.config())

	err := sender.Send(context.Background(), Message{
		To:          []string{"a@cliente.mx", "b@cliente.mx"},
		Subject:     "Reporte",
		Body:        "Adjunto.",
		Attachments: []Attachment{{Filename: "r.pdf", ContentType: "application/pdf", Data: []byte("%PDF-1.4")}},
	})
	if err != nil {
		t.Fatal(err)
	}

	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.from != "reportes@empresa.mx" {
		t.Fatalf("envelope sender = %q", srv.from)
	}
	if len(srv.rcpts) != 2 {
		t.Fatalf("envelope recipients = %v, want both", srv.rcpts)
	}
	if !strings.Contains(srv.data, "Subject: ") || !strings.Contains(srv.data, "filename=r.pdf") {
		t.Fatalf("the server did not receive the message: %q", srv.data)
	}
}

func TestSMTPSendReportsARefusedRecipient(t *testing.T) {
	srv := newFakeSMTP(t)
	srv.rejectTo = "b@cliente.mx"

	err := NewSMTP(srv.config()).Send(context.Background(), Message{
		To: []string{"a@cliente.mx", "b@cliente.mx"}, Subject: "s", Body: "b",
	})
	if err == nil || !strings.Contains(err.Error(), "b@cliente.mx") {
		t.Fatalf("err = %v, want it to name the refused recipient", err)
	}
}

// A server that accepts the connection and then says nothing must make Send
// return, not hang.
func TestSMTPSendTimesOutOnASilentServer(t *testing.T) {
	srv := newFakeSMTP(t)
	srv.silent = true

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- NewSMTP(srv.config()).Send(ctx, Message{To: []string{"a@cliente.mx"}, Subject: "s", Body: "b"})
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Send succeeded against a server that never answered")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Send did not return; the deadline is not enforced on the connection")
	}
}

func TestSMTPSendFailsWhenNothingListens(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	host, port, _ := net.SplitHostPort(ln.Addr().String())
	ln.Close()
	p, _ := strconv.Atoi(port)

	err = NewSMTP(Config{Host: host, Port: p, From: "r@empresa.mx"}).
		Send(context.Background(), Message{To: []string{"a@cliente.mx"}, Subject: "s", Body: "b"})
	if err == nil {
		t.Fatal("Send succeeded with nothing listening")
	}
}

// Once w.Close() reports the message accepted, a QUIT that then fails must
// not turn a delivered email into a "failed" run that gets retried.
func TestSMTPSendIgnoresAQuitFailureAfterTheMessageWasAccepted(t *testing.T) {
	srv := newFakeSMTP(t)
	srv.hangUpAfterData = true

	err := NewSMTP(srv.config()).Send(context.Background(), Message{
		To: []string{"a@cliente.mx"}, Subject: "s", Body: "b",
	})
	if err != nil {
		t.Fatalf("Send = %v, want nil: the server already accepted the message before QUIT failed", err)
	}
}
