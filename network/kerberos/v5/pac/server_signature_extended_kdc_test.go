package pac

import (
	"bytes"
	"encoding/binary"
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// pacWithExtendedKDCSignature returns a PAC whose ticket (0x10) and extended
// KDC (0x13) signatures are already filled in, as a KDC leaves them before it
// computes the server signature, plus the offset of the server signature buffer.
func pacWithExtendedKDCSignature(t *testing.T) (*PAC, int) {
	t.Helper()
	ticketSig := sigBuf(sigHMACSHA1256, 12)
	copy(ticketSig[4:], bytes.Repeat([]byte{0x5A}, 12))
	extKDCSig := sigBuf(sigHMACSHA1256, 12)
	copy(extKDCSig[4:], bytes.Repeat([]byte{0xC3}, 12))
	p := &PAC{
		Buffers: []Buffer{
			{Type: BufferLogonInfo, Data: bytes.Repeat([]byte{0xAA}, 40)},
			{Type: BufferClientInfo, Data: []byte("client-info-bytes")},
			{Type: BufferServerChecksum, Data: sigBuf(sigHMACSHA1256, 12)},
			{Type: BufferKDCChecksum, Data: sigBuf(sigHMACSHA1256, 12)},
			{Type: BufferTicketChecksum, Data: ticketSig},
			{Type: BufferExtendedKDCChecksum, Data: extKDCSig},
		},
	}
	_, offsets := p.layout()
	return p, offsets[2]
}

// referenceServerSignature computes the [MS-PAC] 2.8.4 server signature: a
// keyed checksum of the whole PAC with only the server and KDC Signature
// fields zeroed (the ticket and extended KDC signatures are left intact).
func referenceServerSignature(t *testing.T, p *PAC, serverKey []byte) []byte {
	t.Helper()
	marshaled, offsets := p.layout()
	for i, b := range p.Buffers {
		if b.Type == BufferServerChecksum || b.Type == BufferKDCChecksum {
			copy(marshaled[offsets[i]+4:offsets[i]+len(b.Data)], make([]byte, len(b.Data)-4))
		}
	}
	sig, err := kerbcrypto.GetChecksum(cksumType(sigHMACSHA1256), serverKey, iana.KeyUsageKerbNonKerbCksumSalt, marshaled)
	if err != nil {
		t.Fatalf("GetChecksum: %v", err)
	}
	return sig
}

// TestServerSignatureCoversExtendedKDCSignature checks that a server signature
// computed per [MS-PAC] 2.8.4 over a PAC carrying a non-zero extended KDC
// signature verifies, and that a forged extended KDC signature then breaks it.
func TestServerSignatureCoversExtendedKDCSignature(t *testing.T) {
	serverKey := bytes.Repeat([]byte{0x11}, 32)
	p, srvOff := pacWithExtendedKDCSignature(t)
	want := referenceServerSignature(t, p, serverKey)

	wire, err := p.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	copy(wire[srvOff+4:], want)

	parsed, err := Parse(wire)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if err := parsed.VerifyServerSignature(serverKey); err != nil {
		t.Fatalf("VerifyServerSignature rejected a spec-conformant signature: %v", err)
	}

	// The extended KDC signature is covered by the server signature, so changing
	// it must invalidate the server signature.
	ext, _ := parsed.Buffer(BufferExtendedKDCChecksum)
	tampered := append([]byte(nil), wire...)
	tampered[int(ext.Offset)+4] ^= 0xFF
	reparsed, err := Parse(tampered)
	if err != nil {
		t.Fatalf("Parse tampered: %v", err)
	}
	if err := reparsed.VerifyServerSignature(serverKey); err == nil {
		t.Fatal("VerifyServerSignature accepted a PAC whose extended KDC signature was altered")
	}
}

// TestSignServerSignatureCoversExtendedKDCSignature checks that Sign leaves the
// extended KDC signature in place and produces the [MS-PAC] 2.8.4 server
// signature over it.
func TestSignServerSignatureCoversExtendedKDCSignature(t *testing.T) {
	serverKey := bytes.Repeat([]byte{0x11}, 32)
	kdcKey := bytes.Repeat([]byte{0x22}, 32)
	p, _ := pacWithExtendedKDCSignature(t)
	want := referenceServerSignature(t, p, serverKey)

	signed, err := p.Sign(serverKey, kdcKey)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	parsed, err := Parse(signed)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	srv, _ := parsed.Buffer(BufferServerChecksum)
	if got := srv.Data[4:]; !bytes.Equal(got, want) {
		t.Fatalf("server signature = %x, want %x", got, want)
	}
	ext, _ := parsed.Buffer(BufferExtendedKDCChecksum)
	if binary.LittleEndian.Uint32(ext.Data) != sigHMACSHA1256 || !bytes.Equal(ext.Data[4:], bytes.Repeat([]byte{0xC3}, 12)) {
		t.Fatalf("extended KDC signature changed by Sign: %x", ext.Data)
	}
}
