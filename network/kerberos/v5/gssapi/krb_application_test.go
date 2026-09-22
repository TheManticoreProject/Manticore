package gssapi

import (
	"bytes"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

func TestKRBSafeAndPrivRoundTrip(t *testing.T) {
	ctx := &SecContext{SessionKey: bytes.Repeat([]byte{0x44}, 32), SessionEType: iana.ETypeAES256CTSHMACSHA196}
	now := time.Now().UTC()
	seq := 17
	sender := messages.HostAddress{AddrType: 2, Address: []byte{192, 0, 2, 1}}
	recipient := messages.HostAddress{AddrType: 2, Address: []byte{192, 0, 2, 2}}
	send := KRBApplicationOptions{Timestamp: now, SequenceNumber: &seq, SenderAddress: sender, RecipientAddress: &recipient}
	recv := KRBApplicationReceiveOptions{Now: now, SequenceNumber: &seq, SenderAddress: &sender, RecipientAddress: &recipient}
	payload := []byte("native kerberos application data")

	safe, err := ctx.MakeKRBSafe(payload, send)
	if err != nil {
		t.Fatalf("MakeKRBSafe: %v", err)
	}
	got, err := ctx.ReadKRBSafe(safe, recv)
	if err != nil {
		t.Fatalf("ReadKRBSafe: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("safe payload = %q", got)
	}
	tamperedSafe := append([]byte(nil), safe...)
	tamperedSafe[len(tamperedSafe)-1] ^= 1
	if _, err := ctx.ReadKRBSafe(tamperedSafe, recv); err == nil {
		t.Fatal("tampered KRB-SAFE accepted")
	}

	priv, err := ctx.MakeKRBPriv(payload, send)
	if err != nil {
		t.Fatalf("MakeKRBPriv: %v", err)
	}
	got, err = ctx.ReadKRBPriv(priv, recv)
	if err != nil {
		t.Fatalf("ReadKRBPriv: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("priv payload = %q", got)
	}
	tamperedPriv := append([]byte(nil), priv...)
	tamperedPriv[len(tamperedPriv)-1] ^= 1
	if _, err := ctx.ReadKRBPriv(tamperedPriv, recv); err == nil {
		t.Fatal("tampered KRB-PRIV accepted")
	}
}

func TestKRBApplicationMetadataValidation(t *testing.T) {
	ctx := &SecContext{SessionKey: bytes.Repeat([]byte{0x55}, 16), SessionEType: iana.ETypeAES128CTSHMACSHA196}
	now := time.Now().UTC()
	seq := 9
	sender := messages.HostAddress{AddrType: 2, Address: []byte{203, 0, 113, 1}}
	wire, err := ctx.MakeKRBSafe([]byte("data"), KRBApplicationOptions{Timestamp: now, SequenceNumber: &seq, SenderAddress: sender})
	if err != nil {
		t.Fatal(err)
	}
	wrongSeq := 10
	if _, err := ctx.ReadKRBSafe(wire, KRBApplicationReceiveOptions{Now: now, SequenceNumber: &wrongSeq}); err == nil {
		t.Fatal("wrong sequence accepted")
	}
	wrongSender := messages.HostAddress{AddrType: 2, Address: []byte{203, 0, 113, 2}}
	if _, err := ctx.ReadKRBSafe(wire, KRBApplicationReceiveOptions{Now: now, SenderAddress: &wrongSender}); err == nil {
		t.Fatal("wrong sender address accepted")
	}
	if _, err := ctx.ReadKRBSafe(wire, KRBApplicationReceiveOptions{Now: now.Add(2 * DefaultClockSkew)}); err == nil {
		t.Fatal("stale KRB-SAFE accepted")
	}
}
