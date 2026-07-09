package ldap

import (
	"bytes"
	"crypto/x509"
	"encoding/binary"
	"testing"
)

// serverOffer builds an RFC 4752 §3.1 server security-layer offer: the supported
// bitmask followed by the three-octet max receive buffer.
func serverOffer(supported byte, maxBuf int) []byte {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(maxBuf))
	return []byte{supported, b[1], b[2], b[3]}
}

func TestSelectSASLLayerParsesServerMax(t *testing.T) {
	offer := serverOffer(saslLayerNone|saslLayerIntegrity|saslLayerConfidentiality, 0x00A00000)
	_, chosen, serverMax, err := selectSASLLayer(offer, saslLayerConfidentiality, "")
	if err != nil {
		t.Fatalf("selectSASLLayer: %v", err)
	}
	if chosen != saslLayerConfidentiality {
		t.Errorf("chosen = 0x%02x, want confidentiality 0x%02x", chosen, saslLayerConfidentiality)
	}
	if serverMax != 0x00A00000 {
		t.Errorf("serverMax = %#x, want 0xA00000", serverMax)
	}
}

func TestSelectSASLLayerResponseEncoding(t *testing.T) {
	offer := serverOffer(saslLayerNone|saslLayerIntegrity|saslLayerConfidentiality, 0x100000)

	// Integrity selected: response advertises the client max receive buffer.
	resp, chosen, _, err := selectSASLLayer(offer, saslLayerIntegrity, "u@REALM")
	if err != nil {
		t.Fatalf("selectSASLLayer: %v", err)
	}
	if chosen != saslLayerIntegrity {
		t.Fatalf("chosen = 0x%02x, want integrity", chosen)
	}
	if resp[0] != saslLayerIntegrity {
		t.Errorf("response layer octet = 0x%02x, want 0x%02x", resp[0], saslLayerIntegrity)
	}
	gotMax := int(resp[1])<<16 | int(resp[2])<<8 | int(resp[3])
	if gotMax != saslClientMaxRecv {
		t.Errorf("client max recv = %#x, want %#x", gotMax, saslClientMaxRecv)
	}
	if string(resp[4:]) != "u@REALM" {
		t.Errorf("authzid = %q, want %q", resp[4:], "u@REALM")
	}
}

func TestSelectSASLLayerNoneZeroesMaxBuffer(t *testing.T) {
	offer := serverOffer(saslLayerNone|saslLayerIntegrity, 0x100000)
	resp, chosen, _, err := selectSASLLayer(offer, saslLayerNone, "")
	if err != nil {
		t.Fatalf("selectSASLLayer: %v", err)
	}
	if chosen != saslLayerNone {
		t.Fatalf("chosen = 0x%02x, want none", chosen)
	}
	// RFC 4752: the max buffer field MUST be 0 when no security layer is selected.
	if resp[1] != 0 || resp[2] != 0 || resp[3] != 0 {
		t.Errorf("no-layer response must zero the max-buffer octets, got % x", resp[1:4])
	}
}

func TestSelectSASLLayerDefaultsToNone(t *testing.T) {
	offer := serverOffer(saslLayerNone|saslLayerIntegrity, 0x1000)
	_, chosen, _, err := selectSASLLayer(offer, 0, "")
	if err != nil {
		t.Fatalf("selectSASLLayer: %v", err)
	}
	if chosen != saslLayerNone {
		t.Errorf("unset desired layer should default to none, got 0x%02x", chosen)
	}
}

func TestSelectSASLLayerRejectsUnofferedLayer(t *testing.T) {
	// Server offers only integrity; requesting confidentiality must fail.
	offer := serverOffer(saslLayerNone|saslLayerIntegrity, 0x1000)
	if _, _, _, err := selectSASLLayer(offer, saslLayerConfidentiality, ""); err == nil {
		t.Error("expected error when server does not offer the requested layer")
	}
}

func TestSelectSASLLayerShortOffer(t *testing.T) {
	if _, _, _, err := selectSASLLayer([]byte{0x01, 0x02}, saslLayerNone, ""); err == nil {
		t.Error("expected error on a short (<4 byte) offer")
	}
}

func TestTLSServerEndPointCBT(t *testing.T) {
	// A minimal certificate with a SHA-256 signature: the CBT is
	// "tls-server-end-point:" || SHA-256(cert.Raw).
	cert := &x509.Certificate{
		Raw:                []byte("dummy-der-bytes-for-hashing"),
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	cbt := tlsServerEndPointCBT(cert)
	if !bytes.HasPrefix(cbt, []byte(tlsServerEndPointPrefix)) {
		t.Fatalf("CBT missing %q prefix: % x", tlsServerEndPointPrefix, cbt)
	}
	hash := cbt[len(tlsServerEndPointPrefix):]
	if len(hash) != 32 {
		t.Errorf("SHA-256 CBT hash length = %d, want 32", len(hash))
	}
	want := certificateHash(cert)
	if !bytes.Equal(hash, want) {
		t.Errorf("CBT hash does not match certificateHash")
	}
}

func TestCertificateHashAlgorithmSelection(t *testing.T) {
	raw := []byte("some-certificate-der")
	cases := []struct {
		alg    x509.SignatureAlgorithm
		length int // RFC 5929: SHA-1/MD5 upgrade to SHA-256
	}{
		{x509.SHA1WithRSA, 32},
		{x509.MD5WithRSA, 32},
		{x509.SHA256WithRSA, 32},
		{x509.SHA384WithRSA, 48},
		{x509.SHA512WithRSA, 64},
		{x509.ECDSAWithSHA384, 48},
	}
	for _, c := range cases {
		cert := &x509.Certificate{Raw: raw, SignatureAlgorithm: c.alg}
		if got := len(certificateHash(cert)); got != c.length {
			t.Errorf("%v: hash length = %d, want %d", c.alg, got, c.length)
		}
	}
}
