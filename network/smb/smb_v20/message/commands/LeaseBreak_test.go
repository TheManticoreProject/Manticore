package commands

import (
	"bytes"
	"testing"
)

func TestLeaseBreakNotificationRoundTrip(t *testing.T) {
	orig := NewLeaseBreakNotification()
	orig.NewEpoch = 5
	orig.Flags = SMB2_NOTIFY_BREAK_LEASE_FLAG_ACK_REQUIRED
	orig.LeaseKey = [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	orig.CurrentLeaseState = 0x07
	orig.NewLeaseState = 0x01

	wire, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 44 {
		t.Fatalf("marshalled size = %d, want 44", len(wire))
	}

	parsed := NewLeaseBreakNotification()
	n, err := parsed.Unmarshal(wire)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != 44 {
		t.Errorf("consumed = %d, want 44", n)
	}
	if parsed.StructureSize != LeaseBreakNotificationStructureSize {
		t.Errorf("StructureSize = %d, want %d", parsed.StructureSize, LeaseBreakNotificationStructureSize)
	}
	if parsed.NewEpoch != orig.NewEpoch {
		t.Errorf("NewEpoch = %d, want %d", parsed.NewEpoch, orig.NewEpoch)
	}
	if parsed.Flags != orig.Flags {
		t.Errorf("Flags = 0x%x, want 0x%x", parsed.Flags, orig.Flags)
	}
	if parsed.LeaseKey != orig.LeaseKey {
		t.Errorf("LeaseKey = %x, want %x", parsed.LeaseKey, orig.LeaseKey)
	}
	if parsed.CurrentLeaseState != orig.CurrentLeaseState {
		t.Errorf("CurrentLeaseState = 0x%x, want 0x%x", parsed.CurrentLeaseState, orig.CurrentLeaseState)
	}
	if parsed.NewLeaseState != orig.NewLeaseState {
		t.Errorf("NewLeaseState = 0x%x, want 0x%x", parsed.NewLeaseState, orig.NewLeaseState)
	}
	if !parsed.AckRequired() {
		t.Error("AckRequired should be true")
	}
}

func TestLeaseBreakAckRoundTrip(t *testing.T) {
	orig := NewLeaseBreakAckRequest()
	orig.LeaseKey = [16]byte{0xAA, 0xBB, 0xCC, 0xDD}
	orig.LeaseState = 0x01

	wire, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(wire) != 36 {
		t.Fatalf("marshalled size = %d, want 36", len(wire))
	}

	parsed := NewLeaseBreakAckRequest()
	n, err := parsed.Unmarshal(wire)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != 36 {
		t.Errorf("consumed = %d, want 36", n)
	}
	if parsed.StructureSize != LeaseBreakAckStructureSize {
		t.Errorf("StructureSize = %d, want %d", parsed.StructureSize, LeaseBreakAckStructureSize)
	}
	if !bytes.Equal(parsed.LeaseKey[:], orig.LeaseKey[:]) {
		t.Errorf("LeaseKey = %x, want %x", parsed.LeaseKey, orig.LeaseKey)
	}
	if parsed.LeaseState != orig.LeaseState {
		t.Errorf("LeaseState = 0x%x, want 0x%x", parsed.LeaseState, orig.LeaseState)
	}
}

func TestLeaseBreakNotificationAckNotRequired(t *testing.T) {
	n := NewLeaseBreakNotification()
	n.Flags = 0
	if n.AckRequired() {
		t.Error("AckRequired should be false when flag is not set")
	}
}
