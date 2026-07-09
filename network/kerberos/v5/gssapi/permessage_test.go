package gssapi

import (
	"bytes"
	"encoding/binary"
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// newTestContext returns an initiator SecContext with a fixed AES256 session key
// (no subkey), so per-message tokens can be exercised directly.
func newTestContext(seq uint64) *SecContext {
	return &SecContext{
		SessionKey:   bytes.Repeat([]byte{0x33}, 32),
		SessionEType: iana.ETypeAES256CTSHMACSHA196,
		sendSeq:      seq,
	}
}

// acceptorMIC simulates the acceptor producing a MIC token over data (flag
// SentByAcceptor set, keyed with the acceptor-sign usage).
func acceptorMIC(t *testing.T, key []byte, etype int, seq uint64, data []byte) []byte {
	t.Helper()
	ct, _ := kerbcrypto.ChecksumTypeForEType(etype)
	hdr := micHeader(flagSentByAcceptor, seq)
	sum, err := kerbcrypto.GetChecksum(ct, key, kgUsageAcceptorSign, append(append([]byte{}, data...), hdr...))
	if err != nil {
		t.Fatal(err)
	}
	return append(hdr, sum...)
}

func TestMICInitiatorToAcceptor(t *testing.T) {
	ctx := newTestContext(100)
	data := []byte("integrity-protect-me")

	tok, err := ctx.MakeMIC(data)
	if err != nil {
		t.Fatalf("MakeMIC: %v", err)
	}
	if tok[0] != 0x04 || tok[1] != 0x04 {
		t.Errorf("MIC TOK_ID = %02x %02x", tok[0], tok[1])
	}
	if tok[2]&flagSentByAcceptor != 0 {
		t.Error("initiator MIC must not set SentByAcceptor")
	}
	if binary.BigEndian.Uint64(tok[8:16]) != 100 {
		t.Errorf("SND_SEQ = %d, want 100", binary.BigEndian.Uint64(tok[8:16]))
	}

	// The acceptor verifies with the initiator-sign usage.
	ct, _ := kerbcrypto.ChecksumTypeForEType(ctx.SessionEType)
	want, _ := kerbcrypto.GetChecksum(ct, ctx.SessionKey, kgUsageInitiatorSign, append(append([]byte{}, data...), tok[:16]...))
	if !bytes.Equal(tok[16:], want) {
		t.Error("initiator MIC checksum does not verify under initiator-sign usage")
	}

	// Tampered data must fail that verification.
	bad, _ := kerbcrypto.GetChecksum(ct, ctx.SessionKey, kgUsageInitiatorSign, append([]byte("tampered"), tok[:16]...))
	if bytes.Equal(tok[16:], bad) {
		t.Error("MIC did not depend on the data")
	}
}

func TestVerifyMICFromAcceptor(t *testing.T) {
	ctx := newTestContext(1)
	data := []byte("server-signed-response")
	tok := acceptorMIC(t, ctx.SessionKey, ctx.SessionEType, 5, data)

	if err := ctx.VerifyMIC(data, tok); err != nil {
		t.Errorf("VerifyMIC rejected a valid acceptor MIC: %v", err)
	}
	if err := ctx.VerifyMIC([]byte("other"), tok); err == nil {
		t.Error("VerifyMIC accepted a MIC over different data")
	}
	// An initiator-flagged token must be rejected by VerifyMIC.
	if err := ctx.VerifyMIC(data, func() []byte { tk, _ := newTestContext(5).MakeMIC(data); return tk }()); err == nil {
		t.Error("VerifyMIC accepted an initiator-sent MIC")
	}
}

func TestWrapSealRoundtrip(t *testing.T) {
	ctx := newTestContext(200)
	data := []byte("confidential-payload-of-some-length")

	tok, err := ctx.Wrap(data, true)
	if err != nil {
		t.Fatalf("Wrap: %v", err)
	}
	if tok[0] != 0x05 || tok[1] != 0x04 {
		t.Errorf("Wrap TOK_ID = %02x %02x", tok[0], tok[1])
	}
	if tok[2]&flagSealed == 0 {
		t.Error("sealed Wrap must set the Sealed flag")
	}
	// Ciphertext must not contain the plaintext.
	if bytes.Contains(tok, data) {
		t.Error("sealed Wrap token leaks plaintext")
	}

	// Acceptor side: decrypt with the initiator-seal usage, strip trailing header.
	pt, err := kerbcrypto.Decrypt(ctx.SessionEType, ctx.SessionKey, kgUsageInitiatorSeal, tok[16:])
	if err != nil {
		t.Fatalf("acceptor decrypt: %v", err)
	}
	if len(pt) < 16 || !bytes.Equal(pt[:len(pt)-16], data) {
		t.Errorf("unwrapped plaintext mismatch")
	}
}

func TestUnwrapSealFromAcceptor(t *testing.T) {
	ctx := newTestContext(1)
	data := []byte("server-sealed-data")

	// Acceptor builds a sealed Wrap token (SentByAcceptor + Sealed, acceptor-seal usage).
	hdr := wrapHeader(flagSentByAcceptor|flagSealed, 7)
	pt := append(append([]byte{}, data...), hdr...)
	ctext, err := kerbcrypto.Encrypt(ctx.SessionEType, ctx.SessionKey, kgUsageAcceptorSeal, pt)
	if err != nil {
		t.Fatal(err)
	}
	tok := append(hdr, ctext...)

	got, sealed, err := ctx.Unwrap(tok)
	if err != nil {
		t.Fatalf("Unwrap: %v", err)
	}
	if !sealed {
		t.Error("Unwrap reported not sealed")
	}
	if !bytes.Equal(got, data) {
		t.Errorf("Unwrap data = %q, want %q", got, data)
	}
}

func TestWrapIntegrityOnlyRoundtrip(t *testing.T) {
	ctx := newTestContext(300)
	data := []byte("signed-not-sealed")

	tok, err := ctx.Wrap(data, false)
	if err != nil {
		t.Fatalf("Wrap: %v", err)
	}
	if tok[2]&flagSealed != 0 {
		t.Error("integrity-only Wrap must not set the Sealed flag")
	}
	// Plaintext is present in the clear for integrity-only wrap.
	if !bytes.Contains(tok, data) {
		t.Error("integrity-only Wrap should carry plaintext")
	}

	// Acceptor verifies: recompute over payload | header(EC/RRC zeroed).
	ct, _ := kerbcrypto.ChecksumTypeForEType(ctx.SessionEType)
	ml := micLen(ctx.SessionEType)
	payload := tok[16 : len(tok)-ml]
	want, _ := kerbcrypto.GetChecksum(ct, ctx.SessionKey, kgUsageInitiatorSeal, append(append([]byte{}, payload...), tok[:16]...))
	if !bytes.Equal(tok[len(tok)-ml:], want) {
		t.Error("integrity-only Wrap checksum does not verify")
	}
	if !bytes.Equal(payload, data) {
		t.Error("integrity-only Wrap payload mismatch")
	}
}

func TestUnwrapRotation(t *testing.T) {
	ctx := newTestContext(1)
	data := []byte("rotated-sealed-data-payload")

	// Acceptor seals then right-rotates the body by RRC to exercise un-rotation.
	rrc := 5
	hdr := wrapHeader(flagSentByAcceptor|flagSealed, 9)
	binary.BigEndian.PutUint16(hdr[6:8], uint16(rrc))
	// The encrypted header must carry RRC=0 (per RFC 4121); build a separate one.
	encHdr := append([]byte{}, hdr...)
	binary.BigEndian.PutUint16(encHdr[6:8], 0)
	pt := append(append([]byte{}, data...), encHdr...)
	ctext, err := kerbcrypto.Encrypt(ctx.SessionEType, ctx.SessionKey, kgUsageAcceptorSeal, pt)
	if err != nil {
		t.Fatal(err)
	}
	// Right-rotate ctext by rrc.
	n := rrc % len(ctext)
	rotated := append(append([]byte{}, ctext[len(ctext)-n:]...), ctext[:len(ctext)-n]...)
	tok := append(hdr, rotated...)

	got, _, err := ctx.Unwrap(tok)
	if err != nil {
		t.Fatalf("Unwrap rotated: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("rotated Unwrap data mismatch: %q", got)
	}
}
