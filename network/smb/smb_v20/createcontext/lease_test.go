package createcontext

import (
	"encoding/binary"
	"testing"
)

func TestMarshalLeaseV1(t *testing.T) {
	key := [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	state := SMB2_LEASE_READ_CACHING | SMB2_LEASE_HANDLE_CACHING

	ctx := MarshalLeaseV1(key, state)
	if string(ctx.Name) != string(NameRequestLease) {
		t.Errorf("Name = %q, want %q", ctx.Name, NameRequestLease)
	}
	if len(ctx.Data) != leaseV1DataSize {
		t.Fatalf("Data length = %d, want %d", len(ctx.Data), leaseV1DataSize)
	}

	parsed, err := ParseLeaseV1(ctx.Data)
	if err != nil {
		t.Fatalf("ParseLeaseV1: %v", err)
	}
	if parsed.LeaseKey != key {
		t.Errorf("LeaseKey = %x, want %x", parsed.LeaseKey, key)
	}
	if parsed.LeaseState != state {
		t.Errorf("LeaseState = 0x%x, want 0x%x", parsed.LeaseState, state)
	}
}

func TestMarshalLeaseV2(t *testing.T) {
	key := [16]byte{0xAA, 0xBB}
	parentKey := [16]byte{0xCC, 0xDD}
	state := SMB2_LEASE_READ_CACHING | SMB2_LEASE_WRITE_CACHING
	epoch := uint16(3)

	ctx := MarshalLeaseV2(key, state, parentKey, epoch)
	if len(ctx.Data) != leaseV2DataSize {
		t.Fatalf("Data length = %d, want %d", len(ctx.Data), leaseV2DataSize)
	}

	parsed, err := ParseLeaseV2(ctx.Data)
	if err != nil {
		t.Fatalf("ParseLeaseV2: %v", err)
	}
	if parsed.LeaseKey != key {
		t.Errorf("LeaseKey = %x, want %x", parsed.LeaseKey, key)
	}
	if parsed.LeaseState != state {
		t.Errorf("LeaseState = 0x%x, want 0x%x", parsed.LeaseState, state)
	}
	if parsed.ParentLeaseKey != parentKey {
		t.Errorf("ParentLeaseKey = %x, want %x", parsed.ParentLeaseKey, parentKey)
	}
	if parsed.Epoch != epoch {
		t.Errorf("Epoch = %d, want %d", parsed.Epoch, epoch)
	}
	if parsed.LeaseFlags&SMB2_LEASE_FLAG_PARENT_LEASE_KEY_SET == 0 {
		t.Error("expected PARENT_LEASE_KEY_SET flag to be set for non-zero parent key")
	}
}

func TestMarshalLeaseV2NoParent(t *testing.T) {
	key := [16]byte{0xAA}
	ctx := MarshalLeaseV2(key, SMB2_LEASE_READ_CACHING, [16]byte{}, 1)
	parsed, err := ParseLeaseV2(ctx.Data)
	if err != nil {
		t.Fatalf("ParseLeaseV2: %v", err)
	}
	if parsed.LeaseFlags&SMB2_LEASE_FLAG_PARENT_LEASE_KEY_SET != 0 {
		t.Error("PARENT_LEASE_KEY_SET should not be set for zero parent key")
	}
}

func TestParseLeaseResponseDetectsVersion(t *testing.T) {
	v1Data := make([]byte, leaseV1DataSize)
	binary.LittleEndian.PutUint32(v1Data[16:20], SMB2_LEASE_READ_CACHING)
	parsed, err := ParseLeaseResponse(v1Data)
	if err != nil {
		t.Fatalf("ParseLeaseResponse(V1): %v", err)
	}
	if _, ok := parsed.(*LeaseV1); !ok {
		t.Errorf("expected *LeaseV1, got %T", parsed)
	}

	v2Data := make([]byte, leaseV2DataSize)
	binary.LittleEndian.PutUint32(v2Data[16:20], SMB2_LEASE_WRITE_CACHING)
	binary.LittleEndian.PutUint16(v2Data[48:50], 7)
	parsed, err = ParseLeaseResponse(v2Data)
	if err != nil {
		t.Fatalf("ParseLeaseResponse(V2): %v", err)
	}
	v2, ok := parsed.(*LeaseV2)
	if !ok {
		t.Fatalf("expected *LeaseV2, got %T", parsed)
	}
	if v2.Epoch != 7 {
		t.Errorf("Epoch = %d, want 7", v2.Epoch)
	}
}

func TestLeaseV1ContextRoundTrip(t *testing.T) {
	key := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	state := SMB2_LEASE_READ_CACHING | SMB2_LEASE_HANDLE_CACHING | SMB2_LEASE_WRITE_CACHING

	ctx := MarshalLeaseV1(key, state)
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
	if string(parsed[0].Name) != string(NameRequestLease) {
		t.Errorf("Name = %q, want %q", parsed[0].Name, NameRequestLease)
	}

	v1, err := ParseLeaseV1(parsed[0].Data)
	if err != nil {
		t.Fatalf("ParseLeaseV1: %v", err)
	}
	if v1.LeaseKey != key {
		t.Errorf("LeaseKey round-trip mismatch: %x", v1.LeaseKey)
	}
	if v1.LeaseState != state {
		t.Errorf("LeaseState = 0x%x, want 0x%x", v1.LeaseState, state)
	}
}
