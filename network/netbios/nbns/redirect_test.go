package nbns

import (
	"net"
	"testing"
)

// TestRedirectResponseLayout builds a REDIRECT NAME QUERY RESPONSE, sends it
// through the wire codec and confirms the RFC 1002 4.2.15 layout: R=1 with the
// query OPCODE and only RD set, no answers, one NS record in the authority
// section whose RDATA is the NSD_NAME label sequence, and one A record in the
// additional section named NSD_NAME whose 4-byte RDATA is the NBNS address.
func TestRedirectResponseLayout(t *testing.T) {
	rm := NewRedirectManager()
	rm.AddRedirect("corp", net.ParseIP("10.0.0.5"), 137)

	question := &NetBIOSName{Name: "HOST", ScopeID: "corp"}
	request := &NBNSPacket{
		Header:    NBNSHeader{TransactionID: 0x5150, Flags: OpNameQuery | FlagRecursion, Questions: 1},
		Questions: []NBNSQuestion{{Name: question, Type: QuestionTypeNB, Class: QuestionClassIn}},
	}
	response := &NBNSPacket{Header: NBNSHeader{TransactionID: 0x5150, Flags: FlagResponse | FlagAuthoritative}}
	if !rm.HandleRedirect(request, response) {
		t.Fatal("HandleRedirect did not redirect a query for a configured scope")
	}

	data, err := response.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got NBNSPacket
	if _, err := got.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if want := FlagResponse | OpNameQuery | FlagRecursion; got.Header.Flags != want {
		t.Errorf("flags = 0x%04x, want 0x%04x", got.Header.Flags, want)
	}
	if got.Header.Questions != 0 || got.Header.Answers != 0 || got.Header.Authority != 1 || got.Header.Additional != 1 {
		t.Fatalf("counts QD/AN/NS/AR = %d/%d/%d/%d, want 0/0/1/1",
			got.Header.Questions, got.Header.Answers, got.Header.Authority, got.Header.Additional)
	}

	ns := got.Authority[0]
	if ns.Type != 0x0002 || ns.Class != QuestionClassIn {
		t.Errorf("authority record type/class = 0x%04x/0x%04x, want NS 0x0002/IN 0x0001", ns.Type, ns.Class)
	}
	nsdName, consumed, err := unmarshalName(ns.RData, 0)
	if err != nil {
		t.Fatalf("NS RDATA is not a name: %v", err)
	}
	if consumed != len(ns.RData) || *nsdName != *question {
		t.Errorf("NSD_NAME = %+v (%d of %d bytes), want %+v", *nsdName, consumed, len(ns.RData), *question)
	}

	a := got.Additional[0]
	if a.Type != 0x0001 || a.Class != QuestionClassIn {
		t.Errorf("additional record type/class = 0x%04x/0x%04x, want A 0x0001/IN 0x0001", a.Type, a.Class)
	}
	if *a.Name != *nsdName {
		t.Errorf("additional RR_NAME = %+v, want NSD_NAME %+v", *a.Name, *nsdName)
	}
	if a.RDLength != 4 || !net.IP(a.RData).Equal(net.ParseIP("10.0.0.5")) {
		t.Errorf("NSD_IP_ADDR rdlength=%d rdata=%x, want 4 bytes 0a000005", a.RDLength, a.RData)
	}
}

// TestRedirectRequiresIPv4Server confirms a redirect whose server has no IPv4
// address is not applied, since NSD_IP_ADDR is a 4-byte field.
func TestRedirectRequiresIPv4Server(t *testing.T) {
	rm := NewRedirectManager()
	rm.AddRedirect("", net.ParseIP("2001:db8::1"), 137)

	request := &NBNSPacket{
		Header:    NBNSHeader{Flags: OpNameQuery, Questions: 1},
		Questions: []NBNSQuestion{{Name: &NetBIOSName{Name: "HOST"}, Type: QuestionTypeNB, Class: QuestionClassIn}},
	}
	if rm.HandleRedirect(request, &NBNSPacket{}) {
		t.Error("HandleRedirect redirected to an IPv6-only server")
	}
}
