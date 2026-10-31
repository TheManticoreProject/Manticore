package nbns

import (
	"net"
	"testing"
	"time"
)

// TestRegisteredZeroSuffixNamePreserved registers FOO<00> from a decoded wire
// packet and confirms the 0x00 service suffix survives in the name table: the
// NODE STATUS name table reports it as FOO<00> (not FOO<20>), the name resolves
// as FOO<00>, and FOO<20> stays a distinct, unregistered name.
func TestRegisteredZeroSuffixNamePreserved(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	h := NewPacketHandler(nbns)

	// Round-trip a FOO<00> question through the wire codec to obtain the name
	// exactly as a received packet carries it.
	query, err := NewClient().BuildNameQueryRequest("FOO", 0x00, "")
	if err != nil {
		t.Fatalf("BuildNameQueryRequest: %v", err)
	}
	data, err := query.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded NBNSPacket
	if _, err := decoded.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	name := decoded.Questions[0].Name

	entry := ADDR_ENTRY{Address: 0x0A000001}
	request := &NBNSPacket{
		Header: NBNSHeader{Flags: OpRegistration, Answers: 1},
		Answers: []NBNSResourceRecord{{
			Name:     name,
			Type:     QuestionTypeNB,
			Class:    QuestionClassIn,
			TTL:      300,
			RDLength: entry.Length(),
			RData:    entry.Marshal(),
		}},
	}
	h.dispatch(request, &NBNSPacket{}, nil)

	table := nbns.NameTable()
	if len(table) != 1 {
		t.Fatalf("name table has %d entries, want 1", len(table))
	}
	if table[0].Name != "FOO" || table[0].Suffix != 0x00 {
		t.Errorf("name table entry = %q<%02X>, want \"FOO\"<00>", table[0].Name, table[0].Suffix)
	}

	owners, _, _, err := nbns.QueryName(name.Name, name.ScopeID)
	if err != nil {
		t.Fatalf("QueryName FOO<00>: %v", err)
	}
	if len(owners) != 1 || !owners[0].Equal(net.IPv4(10, 0, 0, 1)) {
		t.Errorf("QueryName FOO<00> owners = %v, want [10.0.0.1]", owners)
	}
	if _, _, _, err := nbns.QueryName("FOO", ""); err == nil {
		t.Error("QueryName FOO<20> succeeded, want only FOO<00> registered")
	}
}

// TestShortNameTrailingNULsStillNormalized confirms a name shorter than 16 bytes
// keeps having its trailing NUL bytes stripped, so it is stored and looked up
// under its bare form.
func TestShortNameTrailingNULsStillNormalized(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	if err := nbns.RegisterName("BAR\x00\x00", "", Unique, net.ParseIP("10.0.0.2"), time.Hour); err != nil {
		t.Fatalf("RegisterName: %v", err)
	}
	if _, _, _, err := nbns.QueryName("BAR", ""); err != nil {
		t.Errorf("QueryName BAR: %v", err)
	}
}
