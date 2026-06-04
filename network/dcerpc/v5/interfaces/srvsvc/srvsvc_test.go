package srvsvc

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

func wstr(s string) *ndr.WSTR { w := ndr.WSTR(s); return &w }

func TestRemoteTOD_RequestBytes(t *testing.T) {
	raw, err := ndr.Request(&remoteTODRequest{})
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	// ServerName is a NULL unique pointer: a single zero referent.
	if !bytes.Equal(raw, []byte{0, 0, 0, 0}) {
		t.Errorf("RemoteTOD request = %x, want 00000000", raw)
	}
	if (&remoteTODRequest{}).Opnum() != 28 {
		t.Errorf("opnum = %d, want 28", (&remoteTODRequest{}).Opnum())
	}
}

func TestRemoteTOD_ResponseRoundTrip(t *testing.T) {
	in := &remoteTODResponse{
		Buffer:    &TimeOfDayInfo{Elapsedt: 0x66000000, Hours: 12, Mins: 34, Secs: 56, Day: 4, Month: 6, Year: 2026, Weekday: 4},
		ErrorCode: NERR_Success,
	}
	raw, err := ndr.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// referent id (4) + ErrorCode (4) + 12x DWORD struct (48) = 56 bytes.
	if len(raw) != 56 {
		t.Errorf("response length = %d, want 56", len(raw))
	}
	var out remoteTODResponse
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Buffer == nil || *out.Buffer != *in.Buffer {
		t.Errorf("round trip: got %+v, want %+v", out.Buffer, in.Buffer)
	}
}

func TestServerGetInfo_RequestBytes(t *testing.T) {
	raw, err := ndr.Request(&serverGetInfoRequest{Level: 101})
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	want := []byte{0, 0, 0, 0, 0x65, 0, 0, 0} // NULL ServerName + Level 101
	if !bytes.Equal(raw, want) {
		t.Errorf("ServerGetInfo request = %x, want %x", raw, want)
	}
}

func TestServerGetInfo101_ResponseRoundTrip(t *testing.T) {
	in := &serverGetInfo101Response{
		Discriminant: 101,
		Info: &ServerInfo101{
			PlatformID:   500,
			Name:         wstr("XPBOX"),
			VersionMajor: 5,
			VersionMinor: 1,
			Type:         0x00000003,
			Comment:      wstr("hello"),
		},
		ErrorCode: NERR_Success,
	}
	raw, err := ndr.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out serverGetInfo101Response
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Info == nil {
		t.Fatal("Info is nil")
	}
	if out.Discriminant != 101 || out.Info.PlatformID != 500 || out.Info.VersionMajor != 5 || out.Info.VersionMinor != 1 || out.Info.Type != 3 {
		t.Errorf("scalars: %+v", out.Info)
	}
	if out.Info.Name == nil || *out.Info.Name != "XPBOX" {
		t.Errorf("Name = %v, want XPBOX", out.Info.Name)
	}
	if out.Info.Comment == nil || *out.Info.Comment != "hello" {
		t.Errorf("Comment = %v, want hello", out.Info.Comment)
	}
}

func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if s.String() != "4b324fc8-1670-01d3-1278-5a47bf6ee188 v3.0" {
		t.Errorf("SyntaxID = %s", s.String())
	}
}
