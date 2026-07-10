package nbns

import (
	"encoding/binary"
	"encoding/hex"
	"net"
	"testing"
	"time"
)

// TestBuildNameQueryRequestWireUnicast asserts the wire layout of a unicast NAME
// QUERY REQUEST (RFC 1002 4.2.12): the transaction ID is echoed into the header,
// the flags carry OPCODE query with neither RD nor B set (a plain end node will
// not answer a unicast query with RD set), QDCOUNT is 1, and the single question
// is an NB (0x0020) question in class IN whose name decodes back to the base
// name and 16th-byte suffix that were asked for.
func TestBuildNameQueryRequestWireUnicast(t *testing.T) {
	c := NewClientWithServer("10.7.0.10")

	req, err := c.BuildNameQueryRequest("TMP-W-2016", SuffixServer, "")
	if err != nil {
		t.Fatalf("BuildNameQueryRequest returned error: %v", err)
	}

	data, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	// Header: TRN_ID, flags, and QDCOUNT.
	if got := binary.BigEndian.Uint16(data[0:2]); got != req.Header.TransactionID {
		t.Errorf("wire TRN_ID = 0x%04x, want 0x%04x", got, req.Header.TransactionID)
	}
	if got := binary.BigEndian.Uint16(data[2:4]); got != OpNameQuery {
		t.Errorf("unicast flags = 0x%04x, want 0x%04x (OPCODE query, no RD/B)", got, OpNameQuery)
	}
	if got := binary.BigEndian.Uint16(data[4:6]); got != 1 {
		t.Errorf("QDCOUNT = %d, want 1", got)
	}

	// Question section: type NB (0x0020), class IN (0x0001).
	q := req.Questions[0]
	if q.Type != QuestionTypeNB {
		t.Errorf("question type = 0x%04x, want NB 0x%04x", q.Type, QuestionTypeNB)
	}
	if q.Class != QuestionClassIn {
		t.Errorf("question class = 0x%04x, want IN 0x%04x", q.Class, QuestionClassIn)
	}

	// The encoded name must carry the base name and the suffix in its 16th byte.
	if len(q.Name.Name) != NetBIOSNameLength {
		t.Fatalf("question name length = %d, want %d", len(q.Name.Name), NetBIOSNameLength)
	}
	if base := string([]byte(q.Name.Name)[:10]); base != "TMP-W-2016" {
		t.Errorf("question base name = %q, want %q", base, "TMP-W-2016")
	}
	if suffix := q.Name.Name[NetBIOSNameLength-1]; suffix != SuffixServer {
		t.Errorf("question suffix = 0x%02x, want 0x%02x", suffix, SuffixServer)
	}
}

// TestBuildNameQueryRequestWireBroadcast asserts that a broadcast client (no
// server configured) sets both the RD and B flags, matching the RFC 1002 4.2.12
// B-node NAME QUERY REQUEST.
func TestBuildNameQueryRequestWireBroadcast(t *testing.T) {
	c := NewClient()

	req, err := c.BuildNameQueryRequest("WORKGROUP", SuffixMasterBrowser, "")
	if err != nil {
		t.Fatalf("BuildNameQueryRequest returned error: %v", err)
	}

	want := OpNameQuery | FlagRecursion | FlagBroadcast
	if req.Header.Flags != want {
		t.Errorf("broadcast flags = 0x%04x, want 0x%04x (OPCODE query, RD+B)", req.Header.Flags, want)
	}
}

// positiveNameQueryResponseKAT is a real POSITIVE NAME QUERY RESPONSE captured
// from a Windows host (10.7.0.10, name "TMP-W-2016", suffix 0x20). It answers a
// query with transaction ID 0xc47e and carries two ADDR_ENTRY owner addresses in
// a single NB record's RDATA (10.7.0.10 and its 169.254.x APIPA address),
// exercising the multi-entry RDATA walk. Frozen here as a known-answer test.
const positiveNameQueryResponseKAT = "c47e85000000000100000000204645454e4641434e4648434e44434441444244474341434143414341434143410000200001000493e0000c60000a07000a6000a9fefd4f"

// TestParseNameQueryResponsePositive decodes the frozen positive-response vector
// and asserts the owner addresses are recovered in order.
func TestParseNameQueryResponsePositive(t *testing.T) {
	data, err := hex.DecodeString(positiveNameQueryResponseKAT)
	if err != nil {
		t.Fatalf("failed to decode KAT: %v", err)
	}

	ips, matched, err := parseNameQueryResponse(data, 0xc47e)
	if err != nil {
		t.Fatalf("parseNameQueryResponse returned error: %v", err)
	}
	if !matched {
		t.Fatal("parseNameQueryResponse did not match the response to the query")
	}

	want := []string{"10.7.0.10", "169.254.253.79"}
	if len(ips) != len(want) {
		t.Fatalf("parsed %d addresses (%v), want %d", len(ips), ips, len(want))
	}
	for i, w := range want {
		if ips[i].String() != w {
			t.Errorf("address[%d] = %s, want %s", i, ips[i], w)
		}
	}
}

// TestParseNameQueryResponseTRNMismatch confirms a response carrying a different
// transaction ID is not accepted as ours.
func TestParseNameQueryResponseTRNMismatch(t *testing.T) {
	data, err := hex.DecodeString(positiveNameQueryResponseKAT)
	if err != nil {
		t.Fatalf("failed to decode KAT: %v", err)
	}

	_, matched, err := parseNameQueryResponse(data, 0x0000)
	if err != nil {
		t.Fatalf("parseNameQueryResponse returned error: %v", err)
	}
	if matched {
		t.Error("parseNameQueryResponse matched a response with a mismatched transaction ID")
	}
}

// TestParseNameQueryResponseNameError checks that a NAME_ERROR (RCODE 0x03)
// negative response is reported as a definitive not-found: matched, no owners,
// and no error.
func TestParseNameQueryResponseNameError(t *testing.T) {
	// A minimal negative NAME QUERY RESPONSE: R+AA set, RD echoed, RCODE 0x03,
	// with empty question/answer sections.
	var hdr [12]byte
	binary.BigEndian.PutUint16(hdr[0:2], 0x1234)
	binary.BigEndian.PutUint16(hdr[2:4], FlagResponse|FlagAuthoritative|RcodeNameError)

	ips, matched, err := parseNameQueryResponse(hdr[:], 0x1234)
	if err != nil {
		t.Fatalf("parseNameQueryResponse returned error: %v, want nil for NAME_ERROR", err)
	}
	if !matched {
		t.Fatal("parseNameQueryResponse did not match the NAME_ERROR response")
	}
	if ips == nil {
		t.Fatal("parseNameQueryResponse returned a nil slice, want an empty non-nil slice")
	}
	if len(ips) != 0 {
		t.Errorf("parseNameQueryResponse returned %v, want an empty slice", ips)
	}
}

// TestResolveAgainstLocalResponder drives the full Resolve path against an
// in-process UDP responder: the responder echoes the request's transaction ID
// and answers with an NB record owning 127.0.0.1, and Resolve must return that
// address.
func TestResolveAgainstLocalResponder(t *testing.T) {
	server, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("failed to start local responder: %v", err)
	}
	defer server.Close()

	go func() {
		buf := make([]byte, 512)
		n, from, err := server.ReadFromUDP(buf)
		if err != nil {
			return
		}

		var req NBNSPacket
		if _, err := req.Unmarshal(buf[:n]); err != nil {
			return
		}

		owner := ADDR_ENTRY{Address: binary.BigEndian.Uint32(net.IPv4(127, 0, 0, 1).To4())}
		resp := &NBNSPacket{
			Header: NBNSHeader{
				TransactionID: req.Header.TransactionID,
				Flags:         FlagResponse | FlagAuthoritative,
				Answers:       1,
			},
			Answers: []NBNSResourceRecord{
				{
					Name:     req.Questions[0].Name,
					Type:     QuestionTypeNB,
					Class:    QuestionClassIn,
					TTL:      300,
					RDLength: owner.Length(),
					RData:    owner.Marshal(),
				},
			},
		}
		data, err := resp.Marshal()
		if err != nil {
			return
		}
		_, _ = server.WriteToUDP(data, from)
	}()

	c := NewClientWithServer(server.LocalAddr().String())
	c.Timeout = 2 * time.Second

	ips, err := c.Resolve("TMP-W-2016", SuffixServer)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if len(ips) != 1 || ips[0].String() != "127.0.0.1" {
		t.Fatalf("Resolve = %v, want [127.0.0.1]", ips)
	}
}
