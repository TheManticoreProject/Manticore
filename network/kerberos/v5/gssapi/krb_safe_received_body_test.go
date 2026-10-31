package gssapi

import (
	"bytes"
	"encoding/asn1"
	"testing"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// peerSafeBody is a KRB-SAFE safe-body as a peer that always sends usec and
// seq-number encodes it (RFC 4120 section 5.6.1), including when they are 0.
type peerSafeBody struct {
	UserData      []byte               `asn1:"explicit,tag:0"`
	Timestamp     time.Time            `asn1:"explicit,tag:1,generalized"`
	Usec          int                  `asn1:"explicit,tag:2"`
	SeqNumber     int                  `asn1:"explicit,tag:3"`
	SenderAddress messages.HostAddress `asn1:"explicit,tag:4"`
}

type peerKRBSafe struct {
	PVNO     int               `asn1:"explicit,tag:0"`
	MsgType  int               `asn1:"explicit,tag:1"`
	SafeBody asn1.RawValue     // [2] wrapper built explicitly
	Checksum messages.Checksum `asn1:"explicit,tag:3"`
}

// TestReadKRBSafeVerifiesReceivedBody checks that the KRB-SAFE checksum is
// verified over the safe-body bytes as received: a peer message carrying an
// explicit seq-number 0 and usec 0 must be accepted.
func TestReadKRBSafeVerifiesReceivedBody(t *testing.T) {
	ctx := &SecContext{SessionKey: bytes.Repeat([]byte{0x44}, 32), SessionEType: iana.ETypeAES256CTSHMACSHA196}
	now := time.Now().UTC().Truncate(time.Second)
	sender := messages.HostAddress{AddrType: 2, Address: []byte{192, 0, 2, 1}}

	body, err := asn1.Marshal(peerSafeBody{UserData: []byte("payload"), Timestamp: now, SenderAddress: sender})
	if err != nil {
		t.Fatal(err)
	}
	cksumType, _ := kerbcrypto.ChecksumTypeForEType(ctx.SessionEType)
	sum, err := kerbcrypto.GetChecksum(cksumType, ctx.SessionKey, iana.KeyUsageKRBSafeCksum, body)
	if err != nil {
		t.Fatal(err)
	}
	seq, err := asn1.Marshal(peerKRBSafe{
		PVNO: messages.KerberosV5, MsgType: messages.MsgTypeKRBSafe,
		SafeBody: asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 2, IsCompound: true, Bytes: body},
		Checksum: messages.Checksum{CKSumType: cksumType, Checksum: sum},
	})
	if err != nil {
		t.Fatal(err)
	}
	wire, err := asn1.Marshal(asn1.RawValue{Class: asn1.ClassApplication, Tag: messages.MsgTypeKRBSafe, IsCompound: true, Bytes: seq})
	if err != nil {
		t.Fatal(err)
	}

	wantSeq := 0
	got, err := ctx.ReadKRBSafe(wire, KRBApplicationReceiveOptions{Now: now, SequenceNumber: &wantSeq, SenderAddress: &sender})
	if err != nil {
		t.Fatalf("ReadKRBSafe rejected a peer KRB-SAFE with seq-number 0: %v", err)
	}
	if string(got) != "payload" {
		t.Fatalf("payload = %q", got)
	}
}

// TestKRBApplicationSequenceZeroIsSentAndRequired checks that a sequence number
// of 0 is put on the wire by MakeKRBSafe/MakeKRBPriv, and that a message with
// no seq-number does not satisfy an expected sequence number of 0.
func TestKRBApplicationSequenceZeroIsSentAndRequired(t *testing.T) {
	ctx := &SecContext{SessionKey: bytes.Repeat([]byte{0x55}, 16), SessionEType: iana.ETypeAES128CTSHMACSHA196}
	now := time.Now().UTC()
	sender := messages.HostAddress{AddrType: 2, Address: []byte{203, 0, 113, 1}}
	zero := 0

	withSeq := KRBApplicationOptions{Timestamp: now, SequenceNumber: &zero, SenderAddress: sender}
	safe, err := ctx.MakeKRBSafe([]byte("data"), withSeq)
	if err != nil {
		t.Fatal(err)
	}
	var msg messages.KRBSafe
	if _, err := msg.Unmarshal(safe); err != nil {
		t.Fatal(err)
	}
	if msg.SafeBody.SequenceNumber == nil || *msg.SafeBody.SequenceNumber != 0 {
		t.Fatalf("KRB-SAFE seq-number = %v, want present 0", msg.SafeBody.SequenceNumber)
	}

	priv, err := ctx.MakeKRBPriv([]byte("data"), withSeq)
	if err != nil {
		t.Fatal(err)
	}
	var privMsg messages.KRBPriv
	if _, err := privMsg.Unmarshal(priv); err != nil {
		t.Fatal(err)
	}
	plain, err := kerbcrypto.Decrypt(ctx.SessionEType, ctx.SessionKey, iana.KeyUsageKRBPrivEncPart, privMsg.EncPart.Cipher)
	if err != nil {
		t.Fatal(err)
	}
	var part messages.EncKRBPrivPart
	if _, err := part.Unmarshal(plain); err != nil {
		t.Fatal(err)
	}
	if part.SequenceNumber == nil || *part.SequenceNumber != 0 {
		t.Fatalf("KRB-PRIV seq-number = %v, want present 0", part.SequenceNumber)
	}

	noSeq := KRBApplicationOptions{Timestamp: now, SenderAddress: sender}
	expectZero := KRBApplicationReceiveOptions{Now: now, SequenceNumber: &zero}
	safe, err = ctx.MakeKRBSafe([]byte("data"), noSeq)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ctx.ReadKRBSafe(safe, expectZero); err == nil {
		t.Fatal("KRB-SAFE without seq-number accepted for expected sequence 0")
	}
	priv, err = ctx.MakeKRBPriv([]byte("data"), noSeq)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ctx.ReadKRBPriv(priv, expectZero); err == nil {
		t.Fatal("KRB-PRIV without seq-number accepted for expected sequence 0")
	}
}
