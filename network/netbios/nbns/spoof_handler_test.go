package nbns

import (
	"net"
	"regexp"
	"testing"
)

// receivedQuery builds a NAME QUERY REQUEST for name+suffix the way a victim on
// the link would, then marshals and unmarshals it so the returned packet carries
// the decoded, padded NetBIOS name the server actually sees on the wire. broadcast
// selects the B-node form (B + RD set), matching NewClient's default.
func receivedQuery(t *testing.T, name string, suffix byte, broadcast bool) *NBNSPacket {
	t.Helper()

	c := NewClientWithServer("127.0.0.1")
	if broadcast {
		c = NewClient()
	}

	req, err := c.BuildNameQueryRequest(name, suffix, "")
	if err != nil {
		t.Fatalf("BuildNameQueryRequest(%q): %v", name, err)
	}

	data, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal request: %v", err)
	}

	var got NBNSPacket
	if _, err := got.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal request: %v", err)
	}
	return &got
}

// TestSpoofHandlerAnswerWireCorrectness verifies that a spoofed positive NAME
// QUERY RESPONSE is wire-correct: it echoes the transaction ID, sets R + AA with
// B clear, copies RD from the request, echoes the question, and carries a single
// NB answer RR whose ADDR_ENTRY holds the configured IP + TTL with the Group bit
// clear (unique).
func TestSpoofHandlerAnswerWireCorrectness(t *testing.T) {
	spoof := net.ParseIP("10.7.0.8")
	h, err := NewSpoofHandler(SpoofConfig{
		Mode:    MatchAll,
		SpoofIP: spoof,
		TTL:     165,
	})
	if err != nil {
		t.Fatalf("NewSpoofHandler: %v", err)
	}

	req := receivedQuery(t, "FAKEHOST", SuffixWorkstation, true)
	resp, ok := h.BuildResponse(req)
	if !ok {
		t.Fatal("expected the query to be answered")
	}

	if resp.Header.TransactionID != req.Header.TransactionID {
		t.Errorf("TRN_ID = 0x%04x, want 0x%04x", resp.Header.TransactionID, req.Header.TransactionID)
	}
	if resp.Header.Flags&FlagResponse == 0 {
		t.Error("R (response) flag not set")
	}
	if resp.Header.Flags&FlagAuthoritative == 0 {
		t.Error("AA (authoritative) flag not set")
	}
	if resp.Header.Flags&FlagBroadcast != 0 {
		t.Error("B (broadcast) flag must be clear in a response")
	}
	if resp.Header.Flags&FlagRecursion == 0 {
		t.Error("RD flag should be copied from the (broadcast) request")
	}
	if rcode := resp.Header.Flags & RcodeMask; rcode != RcodeSuccess {
		t.Errorf("RCODE = 0x%x, want SUCCESS", rcode)
	}

	// Question echo.
	if resp.Header.Questions != 1 || len(resp.Questions) != 1 {
		t.Fatalf("question echo: Header.Questions=%d len(Questions)=%d, want 1/1", resp.Header.Questions, len(resp.Questions))
	}
	if resp.Questions[0].Name.Name != req.Questions[0].Name.Name {
		t.Errorf("echoed question name = %q, want %q", resp.Questions[0].Name.Name, req.Questions[0].Name.Name)
	}

	// Single NB answer RR with the configured TTL.
	if resp.Header.Answers != 1 || len(resp.Answers) != 1 {
		t.Fatalf("answers: Header.Answers=%d len(Answers)=%d, want 1/1", resp.Header.Answers, len(resp.Answers))
	}
	ans := resp.Answers[0]
	if ans.Type != QuestionTypeNB {
		t.Errorf("answer type = 0x%04x, want NB 0x%04x", ans.Type, QuestionTypeNB)
	}
	if ans.TTL != 165 {
		t.Errorf("answer TTL = %d, want 165", ans.TTL)
	}

	// ADDR_ENTRY: unique (Group bit clear) + the configured spoof IP.
	var entry ADDR_ENTRY
	if err := entry.Unmarshal(ans.RData); err != nil {
		t.Fatalf("ADDR_ENTRY.Unmarshal: %v", err)
	}
	if entry.Flags&NBFlagGroup != 0 {
		t.Errorf("NB_FLAGS Group bit set (0x%04x); a poisoned name must be unique", entry.Flags)
	}
	if !entry.IP().Equal(spoof.To4()) {
		t.Errorf("answer IP = %s, want %s", entry.IP(), spoof)
	}
}

// TestSpoofHandlerEndToEndWithClient proves the poisoner and this package's own
// resolver agree on the wire: the handler answers a name it does not own and the
// client parses that response back to the configured spoof IP.
func TestSpoofHandlerEndToEndWithClient(t *testing.T) {
	spoof := net.ParseIP("10.7.0.8")
	h, err := NewSpoofHandler(SpoofConfig{Mode: MatchAll, SpoofIP: spoof})
	if err != nil {
		t.Fatalf("NewSpoofHandler: %v", err)
	}

	req := receivedQuery(t, "FAKEHOST", SuffixWorkstation, true)
	resp, ok := h.BuildResponse(req)
	if !ok {
		t.Fatal("expected the query to be answered")
	}

	data, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal response: %v", err)
	}

	ips, matched, err := parseNameQueryResponse(data, req.Header.TransactionID)
	if err != nil {
		t.Fatalf("parseNameQueryResponse: %v", err)
	}
	if !matched {
		t.Fatal("client did not match the spoofed response")
	}
	if len(ips) != 1 || !ips[0].Equal(spoof.To4()) {
		t.Fatalf("client resolved %v, want [%s]", ips, spoof)
	}
}

// TestSpoofHandlerNonMatchingYieldsNoAnswer verifies that a name outside the
// allowlist is left unanswered so legitimate resolution can proceed.
func TestSpoofHandlerNonMatchingYieldsNoAnswer(t *testing.T) {
	h, err := NewSpoofHandler(SpoofConfig{
		Mode:    MatchList,
		Names:   []string{"FAKEHOST"},
		SpoofIP: net.ParseIP("10.7.0.8"),
	})
	if err != nil {
		t.Fatalf("NewSpoofHandler: %v", err)
	}

	if _, ok := h.BuildResponse(receivedQuery(t, "REALHOST", SuffixWorkstation, true)); ok {
		t.Error("expected no answer for a name outside the allowlist")
	}
	if _, ok := h.BuildResponse(receivedQuery(t, "FAKEHOST", SuffixWorkstation, true)); !ok {
		t.Error("expected an answer for a name on the allowlist")
	}
}

// TestSpoofHandlerMatchModes exercises MatchAll, MatchList, MatchRegex and the
// deny-list, using the handler's exported Matches predicate. Names are compared
// after stripping the service suffix/padding and upper-casing.
func TestSpoofHandlerMatchModes(t *testing.T) {
	spoof := net.ParseIP("10.7.0.8")

	// The decoded, padded 16-byte names the server sees for a workstation query.
	fakehost := receivedQuery(t, "FAKEHOST", SuffixWorkstation, true).Questions[0].Name.Name
	wpad := receivedQuery(t, "WPAD", SuffixWorkstation, true).Questions[0].Name.Name

	all, _ := NewSpoofHandler(SpoofConfig{Mode: MatchAll, SpoofIP: spoof})
	if !all.Matches(fakehost) || !all.Matches(wpad) {
		t.Error("MatchAll should match every name")
	}

	// Deny-list suppresses WPAD even under MatchAll.
	allDeny, _ := NewSpoofHandler(SpoofConfig{Mode: MatchAll, SpoofIP: spoof, Deny: []string{"wpad"}})
	if !allDeny.Matches(fakehost) {
		t.Error("MatchAll should still match FAKEHOST with WPAD denied")
	}
	if allDeny.Matches(wpad) {
		t.Error("deny-list should suppress WPAD")
	}

	list, _ := NewSpoofHandler(SpoofConfig{Mode: MatchList, Names: []string{"fakehost"}, SpoofIP: spoof})
	if !list.Matches(fakehost) {
		t.Error("MatchList should match a case-insensitive allowlist entry")
	}
	if list.Matches(wpad) {
		t.Error("MatchList should not match a name off the allowlist")
	}

	re, _ := NewSpoofHandler(SpoofConfig{Mode: MatchRegex, Regex: regexp.MustCompile(`^(WPAD|PROXY)$`), SpoofIP: spoof})
	if !re.Matches(wpad) {
		t.Error("MatchRegex should match WPAD")
	}
	if re.Matches(fakehost) {
		t.Error("MatchRegex should not match FAKEHOST")
	}
}

// TestSpoofHandlerSuffixFilter verifies the optional service-suffix filter using
// the fully recoverable 0x00 (workstation) suffix.
func TestSpoofHandlerSuffixFilter(t *testing.T) {
	spoof := net.ParseIP("10.7.0.8")
	h, err := NewSpoofHandler(SpoofConfig{
		Mode:     MatchAll,
		SpoofIP:  spoof,
		Suffixes: []byte{SuffixWorkstation},
	})
	if err != nil {
		t.Fatalf("NewSpoofHandler: %v", err)
	}

	if _, ok := h.BuildResponse(receivedQuery(t, "FAKEHOST", SuffixWorkstation, true)); !ok {
		t.Error("expected a workstation (0x00) query to be answered")
	}
	if _, ok := h.BuildResponse(receivedQuery(t, "FAKEHOST", SuffixDomainMasterBrows, true)); ok {
		t.Error("expected a non-workstation suffix to be filtered out")
	}
}

// TestNewSpoofHandlerValidation checks that a misconfigured poisoner fails fast.
func TestNewSpoofHandlerValidation(t *testing.T) {
	if _, err := NewSpoofHandler(SpoofConfig{Mode: MatchAll}); err == nil {
		t.Error("expected an error when no SpoofIP is configured")
	}
	if _, err := NewSpoofHandler(SpoofConfig{Mode: MatchList, SpoofIP: net.ParseIP("10.7.0.8")}); err == nil {
		t.Error("expected an error for MatchList with an empty allowlist")
	}
	if _, err := NewSpoofHandler(SpoofConfig{Mode: MatchRegex, SpoofIP: net.ParseIP("10.7.0.8")}); err == nil {
		t.Error("expected an error for MatchRegex without a regex")
	}
	if _, err := NewSpoofHandler(SpoofConfig{Mode: MatchAll, SpoofIP: net.ParseIP("::1")}); err == nil {
		t.Error("expected an error for a non-IPv4 SpoofIP")
	}
}

// TestSpoofHandlerDefaultTTL verifies the default TTL is applied when unset.
func TestSpoofHandlerDefaultTTL(t *testing.T) {
	h, err := NewSpoofHandler(SpoofConfig{Mode: MatchAll, SpoofIP: net.ParseIP("10.7.0.8")})
	if err != nil {
		t.Fatalf("NewSpoofHandler: %v", err)
	}
	resp, ok := h.BuildResponse(receivedQuery(t, "FAKEHOST", SuffixWorkstation, true))
	if !ok {
		t.Fatal("expected an answer")
	}
	if resp.Answers[0].TTL != DefaultSpoofTTL {
		t.Errorf("TTL = %d, want DefaultSpoofTTL %d", resp.Answers[0].TTL, DefaultSpoofTTL)
	}
}
