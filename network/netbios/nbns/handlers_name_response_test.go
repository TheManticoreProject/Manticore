package nbns

import (
	"encoding/binary"
	"net"
	"testing"
	"time"
)

// nameRequest builds a registration, release or refresh request carrying a
// single NB resource record for name owned by ip.
func nameRequest(opcode uint16, name string, ip net.IP, ttl uint32) *NBNSPacket {
	entry := ADDR_ENTRY{Address: binary.BigEndian.Uint32(ip.To4())}
	return &NBNSPacket{
		Header: NBNSHeader{TransactionID: 0x4242, Flags: opcode | FlagRecursion, Answers: 1},
		Answers: []NBNSResourceRecord{{
			Name:     &NetBIOSName{Name: name},
			Type:     QuestionTypeNB,
			Class:    QuestionClassIn,
			TTL:      ttl,
			RDLength: entry.Length(),
			RData:    entry.Marshal(),
		}},
	}
}

// TestNameResponsesFollowRFC1002Layout drives registration, refresh and release
// requests through the shared dispatch and confirms each response carries the
// header and answer record RFC 1002 prescribes: R=1, the response OPCODE
// (registration for both registration and refresh, release for release), AA
// plus RD and RA on registration responses, the handler's RCODE, ANCOUNT=1 and
// an answer record echoing the request's NB record. The response is also
// marshalled and re-parsed to confirm the layout survives the wire codec.
func TestNameResponsesFollowRFC1002Layout(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	if err := nbns.RegisterName("TAKEN", "", Unique, net.ParseIP("10.0.0.9"), time.Hour); err != nil {
		t.Fatalf("RegisterName: %v", err)
	}
	h := NewPacketHandler(nbns)
	ip := net.IPv4(10, 0, 0, 7).To4()

	registrationFlags := FlagResponse | OpRegistration | FlagAuthoritative | FlagRecursion | FlagRecursionAvailable
	releaseFlags := FlagResponse | OpRelease | FlagAuthoritative

	tests := []struct {
		label     string
		request   *NBNSPacket
		wantFlags uint16
	}{
		{"positive registration (4.2.5)", nameRequest(OpRegistration, "HOST", ip, 300), registrationFlags},
		{"negative registration (4.2.6)", nameRequest(OpRegistration, "TAKEN", ip, 300), registrationFlags | RcodeConflict},
		{"positive refresh", nameRequest(OpRefresh, "HOST", ip, 300), registrationFlags},
		{"negative refresh", nameRequest(OpRefresh, "MISSING", ip, 300), registrationFlags | RcodeServerError},
		{"positive release (4.2.10)", nameRequest(OpRelease, "HOST", ip, 0), releaseFlags},
		{"negative release (4.2.11)", nameRequest(OpRelease, "MISSING", ip, 0), releaseFlags | RcodeServerError},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			response := &NBNSPacket{Header: NBNSHeader{
				TransactionID: tt.request.Header.TransactionID,
				Flags:         FlagResponse | FlagAuthoritative,
			}}
			h.dispatch(tt.request, response, nil)

			data, err := response.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			var got NBNSPacket
			if _, err := got.Unmarshal(data); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}

			if got.Header.Flags != tt.wantFlags {
				t.Errorf("flags = 0x%04x, want 0x%04x", got.Header.Flags, tt.wantFlags)
			}
			if got.Header.Questions != 0 || got.Header.Answers != 1 || got.Header.Authority != 0 || got.Header.Additional != 0 {
				t.Errorf("counts QD/AN/NS/AR = %d/%d/%d/%d, want 0/1/0/0",
					got.Header.Questions, got.Header.Answers, got.Header.Authority, got.Header.Additional)
			}
			if len(got.Answers) != 1 {
				t.Fatalf("answers = %d, want 1", len(got.Answers))
			}
			want := tt.request.Answers[0]
			rr := got.Answers[0]
			if rr.Name.Name != want.Name.Name || rr.Type != QuestionTypeNB || rr.Class != QuestionClassIn ||
				rr.TTL != want.TTL || rr.RDLength != 6 || string(rr.RData) != string(want.RData) {
				t.Errorf("answer = {%q type=0x%04x class=0x%04x ttl=%d rdlen=%d rdata=%x}, want an echo of the request record {%q ttl=%d rdata=%x}",
					rr.Name.Name, rr.Type, rr.Class, rr.TTL, rr.RDLength, rr.RData, want.Name.Name, want.TTL, want.RData)
			}
		})
	}
}
