package mgmt

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/pdu"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/transport"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

var testActivity = guid.GUID{A: 0xa1a2a3a4, B: 0xb1b2, C: 0xc1c2, D: 0xd1d2, E: 0xe1e2e3e4e5e6}

type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

type mockConn struct{ events [][]byte }

var _ transport.Transport = (*mockConn)(nil)

func (m *mockConn) Connect() error           { return nil }
func (m *mockConn) Send([]byte) (int, error) { return 0, nil }
func (m *mockConn) Recv() ([]byte, error) {
	if len(m.events) == 0 {
		return nil, timeoutError{}
	}
	e := m.events[0]
	m.events = m.events[1:]
	return e, nil
}
func (m *mockConn) SetDeadline(time.Time) error { return nil }
func (m *mockConn) MaxPDUSize() int             { return transport.MaxPDUSizeDefault }
func (m *mockConn) RemoteAddr() net.Addr        { return nil }
func (m *mockConn) IsConnected() bool           { return true }
func (m *mockConn) Close() error                { return nil }

func newMgmtOverMock(t *testing.T, respStub []byte) *Client {
	t.Helper()
	h := pdu.NewHeader(pdu.PacketTypeResponse)
	h.ActivityID = testActivity
	h.SequenceNumber = 0
	h.ServerBoot = 0x12345678
	raw, err := (&pdu.PDU{Header: h, Body: respStub}).Marshal()
	if err != nil {
		t.Fatalf("build response PDU: %v", err)
	}
	rpc := client.New(&mockConn{events: [][]byte{raw}}, client.WithActivityID(testActivity))
	return New(rpc)
}

// statusRetval builds an is_server_listening response: [out] status then the
// boolean32 return value.
func statusRetval(status, retval uint32) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint32(b[0:4], status)
	binary.LittleEndian.PutUint32(b[4:8], retval)
	return b
}

func TestIsServerListeningTrue(t *testing.T) {
	c := newMgmtOverMock(t, statusRetval(0, 1))
	listening, err := c.IsServerListening()
	if err != nil {
		t.Fatalf("IsServerListening: %v", err)
	}
	if !listening {
		t.Fatal("expected listening = true")
	}
}

func TestIsServerListeningFalse(t *testing.T) {
	c := newMgmtOverMock(t, statusRetval(0, 0))
	listening, err := c.IsServerListening()
	if err != nil {
		t.Fatalf("IsServerListening: %v", err)
	}
	if listening {
		t.Fatal("expected listening = false")
	}
}

func TestIsServerListeningStatusError(t *testing.T) {
	c := newMgmtOverMock(t, statusRetval(0x1c010002, 0))
	if _, err := c.IsServerListening(); err == nil {
		t.Fatal("expected an error when status is non-zero")
	}
}

// inqIfIdsResponse builds an inq_if_ids response carrying ids and a trailing status.
func inqIfIdsResponse(ids []IfID, status uint32) []byte {
	var buf []byte
	put32 := func(v uint32) {
		var t [4]byte
		binary.LittleEndian.PutUint32(t[:], v)
		buf = append(buf, t[:]...)
	}
	put16 := func(v uint16) {
		var t [2]byte
		binary.LittleEndian.PutUint16(t[:], v)
		buf = append(buf, t[:]...)
	}
	put32(0x00000001)       // if_id_vector referent (non-null)
	put32(uint32(len(ids))) // hoisted conformant maximum_count
	put32(uint32(len(ids))) // count
	for i := range ids {    // element referent ids
		put32(uint32(0x00020000 + i))
	}
	for _, id := range ids { // deferred rpc_if_id_t pointees (20 octets each)
		buf = append(buf, id.UUID.ToBytes()...)
		put16(id.VersionMajor)
		put16(id.VersionMinor)
	}
	put32(status)
	return buf
}

func TestParseInqIfIdsResponseRoundTrip(t *testing.T) {
	ids := []IfID{
		{UUID: guid.GUID{A: 0x11111111, B: 0x2222, C: 0x3333, D: 0x4444, E: 0x555566667777}, VersionMajor: 1, VersionMinor: 0},
		{UUID: guid.GUID{A: 0x88888888, B: 0x9999, C: 0xaaaa, D: 0xbbbb, E: 0xccccddddeeee}, VersionMajor: 3, VersionMinor: 2},
	}
	got, status, err := parseInqIfIdsResponse(inqIfIdsResponse(ids, 0))
	if err != nil {
		t.Fatalf("parseInqIfIdsResponse: %v", err)
	}
	if status != 0 {
		t.Fatalf("status = 0x%08x, want 0", status)
	}
	if len(got) != len(ids) {
		t.Fatalf("got %d ids, want %d", len(got), len(ids))
	}
	for i := range ids {
		if !got[i].UUID.Equal(&ids[i].UUID) || got[i].VersionMajor != ids[i].VersionMajor || got[i].VersionMinor != ids[i].VersionMinor {
			t.Errorf("id %d = %s, want %s", i, got[i], ids[i])
		}
	}
}

func TestParseInqIfIdsEmptyVector(t *testing.T) {
	got, status, err := parseInqIfIdsResponse(inqIfIdsResponse(nil, 0))
	if err != nil {
		t.Fatalf("parseInqIfIdsResponse: %v", err)
	}
	if status != 0 || len(got) != 0 {
		t.Fatalf("got %d ids status 0x%08x, want 0 ids status 0", len(got), status)
	}
}

func TestParseInqIfIdsRejectsHugeCount(t *testing.T) {
	var buf []byte
	put32 := func(v uint32) {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], v)
		buf = append(buf, b[:]...)
	}
	put32(0x00000001) // referent
	put32(0xffffffff) // max_count
	put32(0xffffffff) // count -> implausible
	if _, _, err := parseInqIfIdsResponse(buf); err == nil {
		t.Fatal("expected error on implausible if_id count")
	}
}

func TestInqIfIdsEndToEnd(t *testing.T) {
	ids := []IfID{
		{UUID: Interface, VersionMajor: 1, VersionMinor: 0},
	}
	c := newMgmtOverMock(t, inqIfIdsResponse(ids, 0))
	got, err := c.InqIfIds()
	if err != nil {
		t.Fatalf("InqIfIds: %v", err)
	}
	if len(got) != 1 || !got[0].UUID.Equal(&Interface) {
		t.Fatalf("got %v, want the mgmt interface id", got)
	}
}

func TestInterfaceIdentity(t *testing.T) {
	// afa8bd80-7d8a-11c9-bef4-08002b102989, version 1.0.
	if Interface.ToFormatD() != "afa8bd80-7d8a-11c9-bef4-08002b102989" {
		t.Fatalf("mgmt interface UUID = %s", Interface.ToFormatD())
	}
	if InterfaceMajorVersion != 1 || OpnumInqIfIds != 0 || OpnumIsServerListening != 2 {
		t.Fatal("mgmt interface constants do not match [C706] Appendix Q")
	}
}
