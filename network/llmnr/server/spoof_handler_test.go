package server_test

import (
	"net"
	"regexp"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/llmnr/class"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
	"github.com/TheManticoreProject/Manticore/network/llmnr/server"
)

// newQuery builds a single-question LLMNR query for name/qtype, mirroring what a
// victim on the link would multicast.
func newQuery(t *testing.T, name string, qtype llmnr_type.Type) *message.Message {
	t.Helper()
	msg := message.NewMessage()
	msg.SetQuery()
	if err := msg.AddQuestion(name, qtype, class.ClassIN); err != nil {
		t.Fatalf("AddQuestion(%q): %v", name, err)
	}
	return msg
}

// captureWriter is a ResponseWriter that records the message written to it
// instead of transmitting it, so Run can be exercised without a socket.
type captureWriter struct {
	written *message.Message
	remote  net.Addr
}

func (w *captureWriter) WriteMessage(msg *message.Message) error {
	msg.SetResponse()
	w.written = msg
	return nil
}

func (w *captureWriter) GetRemoteAddr() net.Addr { return w.remote }

// TestSpoofHandlerAnswerWireCorrectnessA verifies that a spoofed A response is
// wire-correct: it copies the transaction ID, sets the QR (response) flag,
// echoes the question, and carries a single A answer with the configured spoof
// IP and TTL.
func TestSpoofHandlerAnswerWireCorrectnessA(t *testing.T) {
	spoof := net.ParseIP("10.7.0.8")
	h, err := server.NewSpoofHandler(server.SpoofConfig{
		Mode:      server.MatchAll,
		AnswerA:   true,
		SpoofIPv4: spoof,
		TTL:       30,
	})
	if err != nil {
		t.Fatalf("NewSpoofHandler: %v", err)
	}

	query := newQuery(t, "wpad", llmnr_type.TypeA)
	resp, ok := h.BuildResponse(query)
	if !ok {
		t.Fatal("expected the query to be answered")
	}

	if !resp.IsResponse() {
		t.Error("response does not have the QR flag set")
	}
	if resp.Header.Identifier != query.Header.Identifier {
		t.Errorf("response ID = %#04x, want %#04x (query ID must be copied)", resp.Header.Identifier, query.Header.Identifier)
	}

	if len(resp.Questions) != 1 {
		t.Fatalf("response has %d questions, want 1 (question must be echoed)", len(resp.Questions))
	}
	if string(resp.Questions[0].Name) != "wpad" || resp.Questions[0].Type != llmnr_type.TypeA {
		t.Errorf("echoed question = %q/%s, want wpad/A", resp.Questions[0].Name, resp.Questions[0].Type.String())
	}

	if len(resp.Answers) != 1 {
		t.Fatalf("response has %d answers, want 1", len(resp.Answers))
	}
	ans := resp.Answers[0]
	if ans.Type != llmnr_type.TypeA {
		t.Errorf("answer type = %s, want A", ans.Type.String())
	}
	if string(ans.Name) != "wpad" {
		t.Errorf("answer name = %q, want wpad", ans.Name)
	}
	if ans.TTL != 30 {
		t.Errorf("answer TTL = %d, want 30", ans.TTL)
	}
	got, err := ans.AsA()
	if err != nil {
		t.Fatalf("AsA: %v", err)
	}
	if !got.Equal(spoof) {
		t.Errorf("answer address = %s, want %s", got, spoof)
	}

	// The response must survive a marshal/unmarshal round-trip unchanged, which
	// is what a real resolver on the link parses.
	encoded, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	decoded := message.Message{}
	if _, err := decoded.Unmarshal(encoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !decoded.IsResponse() || len(decoded.Answers) != 1 {
		t.Fatalf("round-tripped response malformed: QR=%v answers=%d", decoded.IsResponse(), len(decoded.Answers))
	}
	rtIP, err := decoded.Answers[0].AsA()
	if err != nil || !rtIP.Equal(spoof) {
		t.Errorf("round-tripped answer address = %v (err %v), want %s", rtIP, err, spoof)
	}
}

// TestSpoofHandlerAnswerWireCorrectnessAAAA verifies the same wire correctness
// for an AAAA answer.
func TestSpoofHandlerAnswerWireCorrectnessAAAA(t *testing.T) {
	spoof := net.ParseIP("fe80::dead:beef")
	h, err := server.NewSpoofHandler(server.SpoofConfig{
		Mode:       server.MatchAll,
		AnswerAAAA: true,
		SpoofIPv6:  spoof,
	})
	if err != nil {
		t.Fatalf("NewSpoofHandler: %v", err)
	}

	resp, ok := h.BuildResponse(newQuery(t, "fileserver", llmnr_type.TypeAAAA))
	if !ok {
		t.Fatal("expected the query to be answered")
	}
	if len(resp.Answers) != 1 || resp.Answers[0].Type != llmnr_type.TypeAAAA {
		t.Fatalf("expected a single AAAA answer, got %d answers", len(resp.Answers))
	}
	if resp.Answers[0].TTL != server.DefaultSpoofTTL {
		t.Errorf("answer TTL = %d, want default %d", resp.Answers[0].TTL, server.DefaultSpoofTTL)
	}
	got, err := resp.Answers[0].AsAAAA()
	if err != nil {
		t.Fatalf("AsAAAA: %v", err)
	}
	if !got.Equal(spoof) {
		t.Errorf("answer address = %s, want %s", got, spoof)
	}
}

// TestSpoofHandlerMatchModes covers the three name-matching modes and
// own-hostname suppression.
func TestSpoofHandlerMatchModes(t *testing.T) {
	v4 := net.ParseIP("10.0.0.1")

	t.Run("all", func(t *testing.T) {
		h, err := server.NewSpoofHandler(server.SpoofConfig{Mode: server.MatchAll, AnswerA: true, SpoofIPv4: v4})
		if err != nil {
			t.Fatalf("NewSpoofHandler: %v", err)
		}
		for _, name := range []string{"wpad", "anything", "SomeHost"} {
			if !h.Matches(name) {
				t.Errorf("MatchAll should match %q", name)
			}
		}
	})

	t.Run("list-case-insensitive", func(t *testing.T) {
		h, err := server.NewSpoofHandler(server.SpoofConfig{
			Mode:      server.MatchList,
			Names:     []string{"WPAD", "proxy"},
			AnswerA:   true,
			SpoofIPv4: v4,
		})
		if err != nil {
			t.Fatalf("NewSpoofHandler: %v", err)
		}
		if !h.Matches("wpad") {
			t.Error("allowlist should match wpad case-insensitively")
		}
		if !h.Matches("PROXY") {
			t.Error("allowlist should match PROXY case-insensitively")
		}
		if h.Matches("fileserver") {
			t.Error("allowlist should not match an unlisted name")
		}
		// A query for an unlisted name yields no response.
		if _, ok := h.BuildResponse(newQuery(t, "fileserver", llmnr_type.TypeA)); ok {
			t.Error("BuildResponse should decline an unlisted name")
		}
	})

	t.Run("regex", func(t *testing.T) {
		h, err := server.NewSpoofHandler(server.SpoofConfig{
			Mode:      server.MatchRegex,
			Regex:     regexp.MustCompile(`^wpad.*$`),
			AnswerA:   true,
			SpoofIPv4: v4,
		})
		if err != nil {
			t.Fatalf("NewSpoofHandler: %v", err)
		}
		if !h.Matches("wpad") || !h.Matches("wpadproxy") {
			t.Error("regex should match wpad-prefixed names")
		}
		if h.Matches("fileserver") {
			t.Error("regex should not match a non-matching name")
		}
	})

	t.Run("own-hostname-suppression", func(t *testing.T) {
		h, err := server.NewSpoofHandler(server.SpoofConfig{
			Mode:                server.MatchAll,
			AnswerA:             true,
			SpoofIPv4:           v4,
			SuppressOwnHostname: true,
			OwnHostname:         "MyHost",
		})
		if err != nil {
			t.Fatalf("NewSpoofHandler: %v", err)
		}
		if h.Matches("myhost") {
			t.Error("own hostname should be suppressed (case-insensitively)")
		}
		if !h.Matches("otherhost") {
			t.Error("a non-own name should still match")
		}
	})
}

// TestSpoofHandlerTypeGating verifies that a query type the handler is not
// configured to answer is left unanswered.
func TestSpoofHandlerTypeGating(t *testing.T) {
	// IPv4-only handler must not answer AAAA queries.
	h, err := server.NewSpoofHandler(server.SpoofConfig{
		Mode:      server.MatchAll,
		AnswerA:   true,
		SpoofIPv4: net.ParseIP("10.0.0.1"),
	})
	if err != nil {
		t.Fatalf("NewSpoofHandler: %v", err)
	}
	if _, ok := h.BuildResponse(newQuery(t, "wpad", llmnr_type.TypeAAAA)); ok {
		t.Error("IPv4-only handler should not answer an AAAA query")
	}
	if _, ok := h.BuildResponse(newQuery(t, "wpad", llmnr_type.TypeA)); !ok {
		t.Error("IPv4-only handler should answer an A query")
	}
}

// TestSpoofHandlerRun exercises the Handler.Run path end to end through a
// capturing ResponseWriter, confirming a matched query produces a response that
// stops the handler chain, while an unmatched query passes through.
func TestSpoofHandlerRun(t *testing.T) {
	h, err := server.NewSpoofHandler(server.SpoofConfig{
		Mode:      server.MatchList,
		Names:     []string{"wpad"},
		AnswerA:   true,
		SpoofIPv4: net.ParseIP("10.7.0.8"),
	})
	if err != nil {
		t.Fatalf("NewSpoofHandler: %v", err)
	}

	remote := &net.UDPAddr{IP: net.ParseIP("10.7.0.20"), Port: 54321}

	// Matched query: Run answers and returns false (stop the chain).
	w := &captureWriter{remote: remote}
	if cont := h.Run(nil, remote, w, newQuery(t, "wpad", llmnr_type.TypeA)); cont {
		t.Error("Run should return false after answering a matched query")
	}
	if w.written == nil {
		t.Fatal("Run did not write a response for a matched query")
	}
	if !w.written.IsResponse() || len(w.written.Answers) != 1 {
		t.Errorf("written response malformed: QR=%v answers=%d", w.written.IsResponse(), len(w.written.Answers))
	}

	// Unmatched query: Run writes nothing and returns true (continue the chain).
	w2 := &captureWriter{remote: remote}
	if cont := h.Run(nil, remote, w2, newQuery(t, "notlisted", llmnr_type.TypeA)); !cont {
		t.Error("Run should return true when no name matched")
	}
	if w2.written != nil {
		t.Error("Run should not write a response for an unmatched query")
	}
}

// TestNewSpoofHandlerValidation verifies the constructor rejects unusable
// configurations.
func TestNewSpoofHandlerValidation(t *testing.T) {
	cases := []struct {
		name string
		cfg  server.SpoofConfig
	}{
		{"no-address-family", server.SpoofConfig{Mode: server.MatchAll}},
		{"list-without-names", server.SpoofConfig{Mode: server.MatchList, AnswerA: true, SpoofIPv4: net.ParseIP("10.0.0.1")}},
		{"regex-without-regex", server.SpoofConfig{Mode: server.MatchRegex, AnswerA: true, SpoofIPv4: net.ParseIP("10.0.0.1")}},
		{"answerA-without-ip", server.SpoofConfig{Mode: server.MatchAll, AnswerA: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := server.NewSpoofHandler(tc.cfg); err == nil {
				t.Errorf("expected an error for config %q", tc.name)
			}
		})
	}
}
