package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func buildOplockBreakWire(t *testing.T) []byte {
	t.Helper()
	notify := commands.NewOplockBreakResponse()
	notify.OplockLevel = commands.SMB2_OPLOCK_LEVEL_II
	notify.FileId = types.SMB2_FILEID{Persistent: 0xAAAA, Volatile: 0xBBBB}

	m := message.NewMessage()
	m.Header.AddFlags(flags.SMB2_FLAGS_SERVER_TO_REDIR)
	m.Header.MessageId = types.UINT64(unsolicitedMessageId)
	m.SetCommand(notify)
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("building oplock break notification: %v", err)
	}
	return wire
}

func buildLeaseBreakWire(t *testing.T, key [16]byte, currentState, newState uint32, epoch uint16, ackRequired bool) []byte {
	t.Helper()
	notify := commands.NewLeaseBreakNotification()
	notify.LeaseKey = key
	notify.CurrentLeaseState = currentState
	notify.NewLeaseState = newState
	notify.NewEpoch = epoch
	if ackRequired {
		notify.Flags = commands.SMB2_NOTIFY_BREAK_LEASE_FLAG_ACK_REQUIRED
	}

	m := message.NewMessage()
	m.Header.AddFlags(flags.SMB2_FLAGS_SERVER_TO_REDIR)
	m.Header.MessageId = types.UINT64(unsolicitedMessageId)
	m.SetCommand(notify)
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("building lease break notification: %v", err)
	}
	return wire
}

func TestWaitBreakNotificationOplock(t *testing.T) {
	wire := buildOplockBreakWire(t)
	ft := &fakeTransport{connected: true, responses: [][]byte{wire}}
	c := newTestClient(ft)

	bn, err := c.WaitBreakNotification()
	if err != nil {
		t.Fatalf("WaitBreakNotification: %v", err)
	}
	if bn.OplockBreak == nil {
		t.Fatal("expected an oplock break, got nil")
	}
	if bn.LeaseBreak != nil {
		t.Fatal("expected no lease break")
	}
	if bn.OplockBreak.NewLevel != commands.SMB2_OPLOCK_LEVEL_II {
		t.Errorf("NewLevel = 0x%02x, want 0x%02x", bn.OplockBreak.NewLevel, commands.SMB2_OPLOCK_LEVEL_II)
	}
	if bn.OplockBreak.FileId.Persistent != 0xAAAA {
		t.Errorf("Persistent = 0x%x, want 0xAAAA", bn.OplockBreak.FileId.Persistent)
	}
}

func TestWaitBreakNotificationLease(t *testing.T) {
	key := [16]byte{0xDE, 0xAD, 0xBE, 0xEF}
	wire := buildLeaseBreakWire(t, key, 0x07, 0x01, 3, true)
	ft := &fakeTransport{connected: true, responses: [][]byte{wire}}
	c := newTestClient(ft)

	bn, err := c.WaitBreakNotification()
	if err != nil {
		t.Fatalf("WaitBreakNotification: %v", err)
	}
	if bn.LeaseBreak == nil {
		t.Fatal("expected a lease break, got nil")
	}
	if bn.OplockBreak != nil {
		t.Fatal("expected no oplock break")
	}
	lb := bn.LeaseBreak
	if lb.LeaseKey != key {
		t.Errorf("LeaseKey = %x, want %x", lb.LeaseKey, key)
	}
	if lb.CurrentLeaseState != 0x07 {
		t.Errorf("CurrentLeaseState = 0x%x, want 0x07", lb.CurrentLeaseState)
	}
	if lb.NewLeaseState != 0x01 {
		t.Errorf("NewLeaseState = 0x%x, want 0x01", lb.NewLeaseState)
	}
	if lb.NewEpoch != 3 {
		t.Errorf("NewEpoch = %d, want 3", lb.NewEpoch)
	}
	if !lb.AckRequired() {
		t.Error("AckRequired should be true")
	}
}

func TestLeaseBreakAckMarshals(t *testing.T) {
	ft := &fakeTransport{connected: true}
	c := newTestClient(ft)
	c.Session = &Session{Client: c, SessionId: 0x99, TreeId: 0x5}
	c.Connection.SessionTable[0x99] = c.Session

	key := [16]byte{0x11, 0x22, 0x33, 0x44}

	ackResp := commands.NewOplockBreakResponse()
	m := message.NewMessage()
	m.Header.AddFlags(flags.SMB2_FLAGS_SERVER_TO_REDIR)
	m.SetCommand(ackResp)
	ackWire, _ := m.Marshal()
	ft.responses = [][]byte{ackWire}

	err := c.AcknowledgeLeaseBreak(key, 0x01)
	if err != nil {
		t.Fatalf("AcknowledgeLeaseBreak: %v", err)
	}
	if len(ft.sent) != 1 {
		t.Fatalf("expected 1 sent frame, got %d", len(ft.sent))
	}
	sent := ft.sent[0]
	if len(sent) < 64+36 {
		t.Fatalf("sent frame too short: %d bytes", len(sent))
	}
	structSize := uint16(sent[64]) | uint16(sent[65])<<8
	if structSize != commands.LeaseBreakAckStructureSize {
		t.Errorf("StructureSize = %d, want %d", structSize, commands.LeaseBreakAckStructureSize)
	}
}
