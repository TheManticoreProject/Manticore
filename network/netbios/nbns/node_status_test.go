package nbns

import (
	"encoding/binary"
	"encoding/hex"
	"strings"
	"testing"
)

// TestBuildNodeStatusRequestWire asserts the wire layout of a NODE STATUS
// REQUEST (RFC 1002 4.2.17): the transaction ID is placed in the header, the
// flags are a plain OPCODE query (node status is distinguished only by its
// NBSTAT question type, so no RD/B bits are set), QDCOUNT is 1, and the single
// question asks for the reserved "*" wildcard name (first-level encoded) with
// type NBSTAT (0x0021) in class IN (0x0001).
func TestBuildNodeStatusRequestWire(t *testing.T) {
	const trnID uint16 = 0x270f
	data := buildNodeStatusRequest(trnID)

	// Header: TRN_ID, flags (OPCODE query, no RD/B), QDCOUNT=1.
	if got := binary.BigEndian.Uint16(data[0:2]); got != trnID {
		t.Errorf("wire TRN_ID = 0x%04x, want 0x%04x", got, trnID)
	}
	if got := binary.BigEndian.Uint16(data[2:4]); got != OpNameQuery {
		t.Errorf("flags = 0x%04x, want 0x%04x (OPCODE query, no RD/B)", got, OpNameQuery)
	}
	if got := binary.BigEndian.Uint16(data[4:6]); got != 1 {
		t.Errorf("QDCOUNT = %d, want 1", got)
	}

	// Question name: a single 32-byte first-level label terminated by the root
	// label, carrying the "*" wildcard encoding.
	if got := data[12]; got != EncodedNameLength {
		t.Fatalf("question label length = %d, want %d", got, EncodedNameLength)
	}
	label := string(data[13 : 13+EncodedNameLength])
	if label != encodedWildcardName() {
		t.Errorf("question label = %q, want %q", label, encodedWildcardName())
	}
	// The "*" (0x2A) first byte first-level-encodes to "CK", and every NUL pad
	// byte to "AA".
	if !strings.HasPrefix(label, "CK") {
		t.Errorf("wildcard label = %q, want it to begin with the '*' encoding \"CK\"", label)
	}
	off := 13 + EncodedNameLength
	if got := data[off]; got != 0x00 {
		t.Errorf("root label terminator = 0x%02x, want 0x00", got)
	}
	off++

	// Question type and class.
	if got := binary.BigEndian.Uint16(data[off : off+2]); got != QuestionTypeNBSTAT {
		t.Errorf("question type = 0x%04x, want NBSTAT 0x%04x", got, QuestionTypeNBSTAT)
	}
	if got := binary.BigEndian.Uint16(data[off+2 : off+4]); got != QuestionClassIn {
		t.Errorf("question class = 0x%04x, want IN 0x%04x", got, QuestionClassIn)
	}
}

// nodeStatusResponseKAT is a real 193-byte NODE STATUS RESPONSE captured from a
// Windows host (10.7.0.10, "TMP-W-2016") answering a query with transaction ID
// 0x270f. Its RDATA carries NUM_NAMES=5 NODE_NAME entries followed by the
// STATISTICS block whose UNIT_ID is the adapter MAC 08:00:27:3d:ec:2f. Frozen
// here as a known-answer test for the RDATA parser.
const nodeStatusResponseKAT = "270f8400000000010000000020434b414141414141414141414141414141414141414141414141414141414141000021000100000000008905544d502d572d323031362020202020004400544d502d572d32303136302020202000c400544d502d572d3230313630202020201cc400544d502d572d323031362020202020204400544d502d572d3230313630202020201b44000800273dec2f00000000000000000000000000000000000000000000000000000000000000000000000000000000"

// TestParseNodeStatusResponse decodes the frozen response vector and asserts the
// full name table, the per-entry suffix/group/label decoding, and the adapter
// MAC recovered from the STATISTICS block are all correct.
func TestParseNodeStatusResponse(t *testing.T) {
	data, err := hex.DecodeString(nodeStatusResponseKAT)
	if err != nil {
		t.Fatalf("failed to decode KAT: %v", err)
	}

	result, matched, err := parseNodeStatusResponse(data, 0x270f)
	if err != nil {
		t.Fatalf("parseNodeStatusResponse returned error: %v", err)
	}
	if !matched {
		t.Fatal("parseNodeStatusResponse did not match the response to the query")
	}

	if len(result.Names) != 5 {
		t.Fatalf("parsed %d names, want 5: %v", len(result.Names), result.Names)
	}

	// The workstation (0x00, unique) and file server (0x20, unique) entries must
	// carry the host's base name.
	if n := result.Names[0]; n.Name != "TMP-W-2016" || n.Suffix != SuffixWorkstation || n.IsGroup() {
		t.Errorf("names[0] = %q suffix 0x%02x group=%v, want TMP-W-2016 / 0x00 / unique", n.Name, n.Suffix, n.IsGroup())
	}
	if n := result.Names[0]; n.SuffixLabel() != "Workstation Service" {
		t.Errorf("names[0] label = %q, want Workstation Service", n.SuffixLabel())
	}
	if n := result.Names[3]; n.Name != "TMP-W-2016" || n.Suffix != SuffixServer || n.IsGroup() {
		t.Errorf("names[3] = %q suffix 0x%02x group=%v, want TMP-W-2016 / 0x20 / unique", n.Name, n.Suffix, n.IsGroup())
	}
	if n := result.Names[3]; n.SuffixLabel() != "File Server Service" {
		t.Errorf("names[3] label = %q, want File Server Service", n.SuffixLabel())
	}

	// The 0x1c entry is the domain-controllers group name.
	if n := result.Names[2]; n.Suffix != SuffixDomainControllers || !n.IsGroup() {
		t.Errorf("names[2] suffix 0x%02x group=%v, want 0x1c / group", n.Suffix, n.IsGroup())
	}

	// Every entry in this reply is active.
	for i, n := range result.Names {
		if !n.IsActive() {
			t.Errorf("names[%d] (%s) is not marked active", i, n.Name)
		}
	}

	// The STATISTICS UNIT_ID must decode to the captured adapter MAC.
	wantMAC := "08:00:27:3d:ec:2f"
	if result.MAC == nil {
		t.Fatal("parseNodeStatusResponse returned a nil MAC")
	}
	if result.MAC.String() != wantMAC {
		t.Errorf("MAC = %s, want %s", result.MAC, wantMAC)
	}
}

// TestParseNodeStatusResponseTRNMismatch confirms a response carrying a
// different transaction ID is not accepted as ours.
func TestParseNodeStatusResponseTRNMismatch(t *testing.T) {
	data, err := hex.DecodeString(nodeStatusResponseKAT)
	if err != nil {
		t.Fatalf("failed to decode KAT: %v", err)
	}

	_, matched, err := parseNodeStatusResponse(data, 0x0000)
	if err != nil {
		t.Fatalf("parseNodeStatusResponse returned error: %v", err)
	}
	if matched {
		t.Error("parseNodeStatusResponse matched a response with a mismatched transaction ID")
	}
}

// TestSuffixLabelGroupVsUnique checks that the 0x00 suffix is labelled
// according to the G bit: a unique 0x00 is the Workstation Service, a group
// 0x00 is the domain/workgroup name.
func TestSuffixLabelGroupVsUnique(t *testing.T) {
	if got := SuffixLabel(SuffixWorkstation, false); got != "Workstation Service" {
		t.Errorf("SuffixLabel(0x00, unique) = %q, want Workstation Service", got)
	}
	if got := SuffixLabel(SuffixWorkstation, true); got != "Domain/Workgroup Name" {
		t.Errorf("SuffixLabel(0x00, group) = %q, want Domain/Workgroup Name", got)
	}
}
