package nbns

import (
	"net"
	"testing"
)

// TestNBNSPacketNameQueryRequestRoundTrip marshals a NAME QUERY REQUEST and
// unmarshals it back, asserting the header fields, flags/counts, and the
// question section survive the wire round-trip.
func TestNBNSPacketNameQueryRequestRoundTrip(t *testing.T) {
	req := &NBNSPacket{
		Header: NBNSHeader{
			TransactionID: 0x1234,
			Flags:         OpNameQuery | FlagRecursion | FlagBroadcast,
			Questions:     1,
		},
		Questions: []NBNSQuestion{
			{
				Name:  &NetBIOSName{Name: "FRED"},
				Type:  QuestionTypeNB,
				Class: QuestionClassIn,
			},
		},
	}

	data, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	var got NBNSPacket
	consumed, err := got.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if consumed != len(data) {
		t.Fatalf("Unmarshal consumed %d bytes, want %d", consumed, len(data))
	}

	if got.Header.TransactionID != req.Header.TransactionID {
		t.Errorf("TransactionID = 0x%04x, want 0x%04x", got.Header.TransactionID, req.Header.TransactionID)
	}
	if got.Header.Flags != req.Header.Flags {
		t.Errorf("Flags = 0x%04x, want 0x%04x", got.Header.Flags, req.Header.Flags)
	}
	if got.Header.Flags&FlagResponse != 0 {
		t.Errorf("request should not have the response flag set: 0x%04x", got.Header.Flags)
	}
	if got.Header.Questions != 1 {
		t.Errorf("Questions = %d, want 1", got.Header.Questions)
	}
	if len(got.Questions) != 1 {
		t.Fatalf("decoded %d questions, want 1", len(got.Questions))
	}

	q := got.Questions[0]
	if q.Name == nil || q.Name.Name != "FRED" {
		t.Errorf("question name = %+v, want FRED", q.Name)
	}
	if q.Type != QuestionTypeNB {
		t.Errorf("question type = 0x%04x, want 0x%04x", q.Type, QuestionTypeNB)
	}
	if q.Class != QuestionClassIn {
		t.Errorf("question class = 0x%04x, want 0x%04x", q.Class, QuestionClassIn)
	}
}

// TestNBNSPacketNameQueryResponseRoundTrip marshals a POSITIVE NAME QUERY
// RESPONSE carrying an NB answer resource record whose RDATA is an ADDR_ENTRY
// (NB_FLAGS + IPv4 address), then unmarshals it and asserts the flags, counts
// and decoded IP address survive.
func TestNBNSPacketNameQueryResponseRoundTrip(t *testing.T) {
	const wantIP = "192.168.1.50"

	entry := ADDR_ENTRY{
		Flags:   0x0000, // unique name, B-node owner
		Address: ipToUint32(t, wantIP),
	}
	rdata := entry.Marshal()

	resp := &NBNSPacket{
		Header: NBNSHeader{
			TransactionID: 0xBEEF,
			Flags:         FlagResponse | OpNameQuery | FlagAuthoritative | FlagRecursion | FlagRecursionAvailable,
			Answers:       1,
		},
		Answers: []NBNSResourceRecord{
			{
				Name:     &NetBIOSName{Name: "FRED"},
				Type:     QuestionTypeNB,
				Class:    QuestionClassIn,
				TTL:      300,
				RDLength: uint16(len(rdata)),
				RData:    rdata,
			},
		},
	}

	data, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	var got NBNSPacket
	if _, err := got.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if got.Header.Flags&FlagResponse == 0 {
		t.Errorf("response flag not set: 0x%04x", got.Header.Flags)
	}
	if got.Header.Flags&FlagAuthoritative == 0 {
		t.Errorf("authoritative flag not set: 0x%04x", got.Header.Flags)
	}
	if got.Header.Flags != resp.Header.Flags {
		t.Errorf("Flags = 0x%04x, want 0x%04x", got.Header.Flags, resp.Header.Flags)
	}
	if got.Header.Answers != 1 {
		t.Errorf("Answers = %d, want 1", got.Header.Answers)
	}
	if len(got.Answers) != 1 {
		t.Fatalf("decoded %d answers, want 1", len(got.Answers))
	}

	rr := got.Answers[0]
	if rr.Name == nil || rr.Name.Name != "FRED" {
		t.Errorf("answer name = %+v, want FRED", rr.Name)
	}
	if rr.Type != QuestionTypeNB {
		t.Errorf("answer type = 0x%04x, want 0x%04x", rr.Type, QuestionTypeNB)
	}
	if rr.TTL != 300 {
		t.Errorf("answer TTL = %d, want 300", rr.TTL)
	}

	ip, err := ParseIPFromRData(rr.RData)
	if err != nil {
		t.Fatalf("ParseIPFromRData returned error: %v", err)
	}
	if !ip.Equal(net.ParseIP(wantIP)) {
		t.Errorf("decoded IP = %v, want %v", ip, wantIP)
	}
}

// TestNBNSPacketGroupFlag confirms the NB_FLAGS Group bit in an ADDR_ENTRY
// round-trips through the RR RDATA.
func TestNBNSPacketGroupFlag(t *testing.T) {
	entry := ADDR_ENTRY{
		Flags:   NBFlagGroup,
		Address: ipToUint32(t, "10.0.0.1"),
	}
	rdata := entry.Marshal()

	var decoded ADDR_ENTRY
	if err := decoded.Unmarshal(rdata); err != nil {
		t.Fatalf("ADDR_ENTRY.Unmarshal returned error: %v", err)
	}
	if decoded.Flags&NBFlagGroup == 0 {
		t.Errorf("group flag lost: 0x%04x", decoded.Flags)
	}
	if !decoded.IP().Equal(net.ParseIP("10.0.0.1")) {
		t.Errorf("decoded IP = %v, want 10.0.0.1", decoded.IP())
	}
}

// TestNBNSPacketUnmarshalErrors asserts that malformed or truncated inputs
// return an error rather than panicking.
func TestNBNSPacketUnmarshalErrors(t *testing.T) {
	// A valid response we can truncate at various points.
	entry := ADDR_ENTRY{Address: ipToUint32(t, "192.168.1.50")}
	rdata := entry.Marshal()
	valid := &NBNSPacket{
		Header: NBNSHeader{
			TransactionID: 1,
			Flags:         FlagResponse | OpNameQuery,
			Answers:       1,
		},
		Answers: []NBNSResourceRecord{
			{
				Name:     &NetBIOSName{Name: "FRED"},
				Type:     QuestionTypeNB,
				Class:    QuestionClassIn,
				TTL:      300,
				RDLength: uint16(len(rdata)),
				RData:    rdata,
			},
		},
	}
	full, err := valid.Marshal()
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"short header", []byte{0x00, 0x01, 0x02}},
		{"header only, claims 1 question", func() []byte {
			b := make([]byte, 12)
			b[5] = 0x01 // Questions = 1, but no question bytes follow
			return b
		}()},
		{"header only, claims 1 answer", func() []byte {
			b := make([]byte, 12)
			b[7] = 0x01 // Answers = 1, but no RR bytes follow
			return b
		}()},
		{"truncated mid-rdata", full[:len(full)-2]},
		{"truncated mid-record", full[:20]},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var p NBNSPacket
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Unmarshal panicked on %s: %v", tc.name, r)
				}
			}()
			if _, err := p.Unmarshal(tc.data); err == nil {
				t.Fatalf("Unmarshal(%s) expected error, got nil", tc.name)
			}
		})
	}
}

// TestNBNSPacketScopedNameRoundTrip marshals a NAME QUERY REQUEST whose
// question name carries a multi-component scope ID and asserts the name and
// scope survive the RFC 1002 4.2.1.2 label-sequence round-trip.
func TestNBNSPacketScopedNameRoundTrip(t *testing.T) {
	req := &NBNSPacket{
		Header: NBNSHeader{
			TransactionID: 0x2222,
			Flags:         OpNameQuery | FlagRecursion,
			Questions:     1,
		},
		Questions: []NBNSQuestion{
			{
				Name:  &NetBIOSName{Name: "NAME", ScopeID: "SCOPE.COM"},
				Type:  QuestionTypeNB,
				Class: QuestionClassIn,
			},
		},
	}

	data, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	// The name must encode as length-prefixed labels: 0x20 + 32-byte encoded
	// name, then 0x05 "SCOPE", 0x03 "COM", terminated by 0x00.
	if data[12] != 0x20 {
		t.Fatalf("first label length = 0x%02x, want 0x20", data[12])
	}
	if got := data[12+1+32]; got != 0x05 {
		t.Errorf("scope label length = 0x%02x, want 0x05 (SCOPE)", got)
	}

	var got NBNSPacket
	consumed, err := got.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if consumed != len(data) {
		t.Fatalf("Unmarshal consumed %d bytes, want %d", consumed, len(data))
	}
	if len(got.Questions) != 1 {
		t.Fatalf("decoded %d questions, want 1", len(got.Questions))
	}

	q := got.Questions[0]
	if q.Name == nil || q.Name.Name != "NAME" {
		t.Errorf("question name = %+v, want NAME", q.Name)
	}
	if q.Name.ScopeID != "SCOPE.COM" {
		t.Errorf("question scope = %q, want %q", q.Name.ScopeID, "SCOPE.COM")
	}
}

// TestNBNSPacketCompressionPointer hand-builds a response whose single answer
// name is a 0xC0 compression pointer back to the question name (as real
// Windows NBNS replies commonly emit) and asserts the pointer-decode path
// resolves it to the correct name and consumes exactly two octets for it.
func TestNBNSPacketCompressionPointer(t *testing.T) {
	// Encode the question name into a label sequence at a known offset.
	qName := &NetBIOSName{Name: "FRED"}
	encodedName, err := marshalName(qName)
	if err != nil {
		t.Fatalf("marshalName returned error: %v", err)
	}

	entry := ADDR_ENTRY{Address: ipToUint32(t, "192.168.1.50")}
	rdata := entry.Marshal()

	var b []byte
	// Header: 1 question, 1 answer, response flag.
	hdr := make([]byte, 12)
	binaryPutUint16(hdr[0:2], 0x4444)
	binaryPutUint16(hdr[2:4], FlagResponse|OpNameQuery)
	binaryPutUint16(hdr[4:6], 1) // Questions
	binaryPutUint16(hdr[6:8], 1) // Answers
	b = append(b, hdr...)

	// Question section: name (offset 12) + type + class.
	questionNameOffset := len(b)
	b = append(b, encodedName...)
	b = binaryAppendUint16(b, QuestionTypeNB)
	b = binaryAppendUint16(b, QuestionClassIn)

	// Answer section: name is a compression pointer to the question name.
	answerNameOffset := len(b)
	b = append(b, byte(labelPointerFlag|(questionNameOffset>>8)), byte(questionNameOffset&0xFF))
	b = binaryAppendUint16(b, QuestionTypeNB)     // type
	b = binaryAppendUint16(b, QuestionClassIn)    // class
	b = append(b, 0x00, 0x00, 0x01, 0x2c)         // TTL = 300
	b = binaryAppendUint16(b, uint16(len(rdata))) // RDLength
	b = append(b, rdata...)

	var got NBNSPacket
	consumed, err := got.Unmarshal(b)
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if consumed != len(b) {
		t.Fatalf("Unmarshal consumed %d bytes, want %d", consumed, len(b))
	}
	if len(got.Answers) != 1 {
		t.Fatalf("decoded %d answers, want 1", len(got.Answers))
	}
	if got.Answers[0].Name == nil || got.Answers[0].Name.Name != "FRED" {
		t.Errorf("answer name via pointer = %+v, want FRED", got.Answers[0].Name)
	}

	// The pointer itself must consume exactly two octets at the answer name.
	name, n, err := unmarshalName(b, answerNameOffset)
	if err != nil {
		t.Fatalf("unmarshalName(pointer) returned error: %v", err)
	}
	if n != 2 {
		t.Errorf("pointer consumed %d bytes, want 2", n)
	}
	if name.Name != "FRED" {
		t.Errorf("pointer decoded name = %q, want FRED", name.Name)
	}
}

// TestNBNSNameUnmarshalPointerErrors asserts that malformed compression
// pointers (self-referential loops and out-of-range targets) return an error
// instead of looping forever.
func TestNBNSNameUnmarshalPointerErrors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		// Pointer at offset 0 targeting offset 0: an immediate self-loop.
		{"self-loop", []byte{labelPointerFlag, 0x00}},
		// Two pointers referencing each other: 0->2 and 2->0.
		{"mutual-loop", []byte{labelPointerFlag, 0x02, labelPointerFlag, 0x00}},
		// Pointer targeting an offset past the end of the buffer.
		{"out-of-range", []byte{labelPointerFlag, 0x7f}},
		// Pointer flag with no second octet.
		{"truncated-pointer", []byte{labelPointerFlag}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("unmarshalName panicked on %s: %v", tc.name, r)
				}
			}()
			if _, _, err := unmarshalName(tc.data, 0); err == nil {
				t.Fatalf("unmarshalName(%s) expected error, got nil", tc.name)
			}
		})
	}
}

// binaryPutUint16 and binaryAppendUint16 are tiny big-endian helpers used by
// the hand-built compression-pointer test buffer.
func binaryPutUint16(b []byte, v uint16) {
	b[0] = byte(v >> 8)
	b[1] = byte(v)
}

func binaryAppendUint16(b []byte, v uint16) []byte {
	return append(b, byte(v>>8), byte(v))
}

// ipToUint32 converts a dotted-quad IPv4 string into the big-endian uint32
// used by ADDR_ENTRY.Address.
func ipToUint32(t *testing.T, s string) uint32 {
	t.Helper()
	ip := net.ParseIP(s).To4()
	if ip == nil {
		t.Fatalf("invalid IPv4 address %q", s)
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}
