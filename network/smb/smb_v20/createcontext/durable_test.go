package createcontext

import (
	"testing"
)

func TestMarshalDurableHandleRequestV1(t *testing.T) {
	ctx := MarshalDurableHandleRequestV1()
	if string(ctx.Name) != string(NameDurableHandleReq) {
		t.Errorf("Name = %q, want %q", ctx.Name, NameDurableHandleReq)
	}
	if len(ctx.Data) != durableV1ReqDataSize {
		t.Fatalf("Data length = %d, want %d", len(ctx.Data), durableV1ReqDataSize)
	}
	for i, b := range ctx.Data {
		if b != 0 {
			t.Errorf("Data[%d] = 0x%02x, want 0x00 (reserved)", i, b)
		}
	}
}

func TestMarshalDurableHandleReconnectV1RoundTrip(t *testing.T) {
	ctx := MarshalDurableHandleReconnectV1(0xAABBCCDD, 0x11223344)
	if string(ctx.Name) != string(NameDurableHandleRecon) {
		t.Errorf("Name = %q, want %q", ctx.Name, NameDurableHandleRecon)
	}
	if len(ctx.Data) != durableV1ReconDataSize {
		t.Fatalf("Data length = %d, want %d", len(ctx.Data), durableV1ReconDataSize)
	}
	p, v, err := ParseDurableHandleReconnectV1(ctx.Data)
	if err != nil {
		t.Fatalf("ParseDurableHandleReconnectV1: %v", err)
	}
	if p != 0xAABBCCDD {
		t.Errorf("persistent = 0x%x, want 0xAABBCCDD", p)
	}
	if v != 0x11223344 {
		t.Errorf("volatile = 0x%x, want 0x11223344", v)
	}
}

func TestDurableHandleRequestV2RoundTrip(t *testing.T) {
	guid := [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}

	ctx := MarshalDurableHandleRequestV2(60000, SMB2_DHANDLE_FLAG_PERSISTENT, guid)
	if string(ctx.Name) != string(NameDurableHandleReqV2) {
		t.Errorf("Name = %q, want %q", ctx.Name, NameDurableHandleReqV2)
	}
	if len(ctx.Data) != durableV2ReqDataSize {
		t.Fatalf("Data length = %d, want %d", len(ctx.Data), durableV2ReqDataSize)
	}

	parsed, err := ParseDurableHandleRequestV2(ctx.Data)
	if err != nil {
		t.Fatalf("ParseDurableHandleRequestV2: %v", err)
	}
	if parsed.Timeout != 60000 {
		t.Errorf("Timeout = %d, want 60000", parsed.Timeout)
	}
	if parsed.Flags != SMB2_DHANDLE_FLAG_PERSISTENT {
		t.Errorf("Flags = 0x%x, want 0x%x", parsed.Flags, SMB2_DHANDLE_FLAG_PERSISTENT)
	}
	if parsed.CreateGuid != guid {
		t.Errorf("CreateGuid = %x, want %x", parsed.CreateGuid, guid)
	}
}

func TestDurableHandleReconnectV2RoundTrip(t *testing.T) {
	guid := [16]byte{0xAA, 0xBB, 0xCC, 0xDD}

	ctx := MarshalDurableHandleReconnectV2(0x1111, 0x2222, guid, SMB2_DHANDLE_FLAG_PERSISTENT)
	if string(ctx.Name) != string(NameDurableHandleReconV2) {
		t.Errorf("Name = %q, want %q", ctx.Name, NameDurableHandleReconV2)
	}
	if len(ctx.Data) != durableV2ReconDataSize {
		t.Fatalf("Data length = %d, want %d", len(ctx.Data), durableV2ReconDataSize)
	}

	parsed, err := ParseDurableHandleReconnectV2(ctx.Data)
	if err != nil {
		t.Fatalf("ParseDurableHandleReconnectV2: %v", err)
	}
	if parsed.Persistent != 0x1111 {
		t.Errorf("Persistent = 0x%x, want 0x1111", parsed.Persistent)
	}
	if parsed.Volatile != 0x2222 {
		t.Errorf("Volatile = 0x%x, want 0x2222", parsed.Volatile)
	}
	if parsed.CreateGuid != guid {
		t.Errorf("CreateGuid = %x, want %x", parsed.CreateGuid, guid)
	}
	if parsed.Flags != SMB2_DHANDLE_FLAG_PERSISTENT {
		t.Errorf("Flags = 0x%x, want 0x%x", parsed.Flags, SMB2_DHANDLE_FLAG_PERSISTENT)
	}
}

func TestDurableHandleResponseV2(t *testing.T) {
	data := make([]byte, 8)
	data[0] = 0x60 // timeout = 96
	data[4] = 0x02 // flags = SMB2_DHANDLE_FLAG_PERSISTENT
	resp, err := ParseDurableHandleResponseV2(data)
	if err != nil {
		t.Fatalf("ParseDurableHandleResponseV2: %v", err)
	}
	if resp.Timeout != 96 {
		t.Errorf("Timeout = %d, want 96", resp.Timeout)
	}
	if resp.Flags != SMB2_DHANDLE_FLAG_PERSISTENT {
		t.Errorf("Flags = 0x%x, want 0x%x", resp.Flags, SMB2_DHANDLE_FLAG_PERSISTENT)
	}
}

func TestDurableHandleV2ContextRoundTrip(t *testing.T) {
	guid := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	ctx := MarshalDurableHandleRequestV2(30000, 0, guid)

	buf, err := Marshal([]CreateContext{ctx})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	parsed, err := Parse(buf)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(parsed) != 1 {
		t.Fatalf("parsed %d contexts, want 1", len(parsed))
	}
	if string(parsed[0].Name) != string(NameDurableHandleReqV2) {
		t.Errorf("Name = %q, want %q", parsed[0].Name, NameDurableHandleReqV2)
	}

	dh2q, err := ParseDurableHandleRequestV2(parsed[0].Data)
	if err != nil {
		t.Fatalf("ParseDurableHandleRequestV2: %v", err)
	}
	if dh2q.CreateGuid != guid {
		t.Errorf("CreateGuid round-trip mismatch: %x", dh2q.CreateGuid)
	}
	if dh2q.Timeout != 30000 {
		t.Errorf("Timeout = %d, want 30000", dh2q.Timeout)
	}
}

func TestParseDurableHandleRequestV2TooShort(t *testing.T) {
	_, err := ParseDurableHandleRequestV2(make([]byte, 10))
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestParseDurableHandleReconnectV2TooShort(t *testing.T) {
	_, err := ParseDurableHandleReconnectV2(make([]byte, 10))
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestParseDurableHandleResponseV2TooShort(t *testing.T) {
	_, err := ParseDurableHandleResponseV2(make([]byte, 4))
	if err == nil {
		t.Error("expected error for short data")
	}
}
