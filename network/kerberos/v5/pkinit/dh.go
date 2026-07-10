// Package pkinit implements the PKINIT (RFC 4556) Diffie-Hellman AS exchange for
// Kerberos v5: building a PA-PK-AS-REQ (a CMS SignedData wrapping an AuthPack
// with the client's ephemeral DH public value) and parsing the corresponding
// PA-PK-AS-REP (dhInfo variant) to recover the KDC's DH public value, compute
// the shared secret, and derive the AS reply key via RFC 4556 §3.2.3.1
// octetstring2key.
//
// It underpins certificate-based authentication and the Shadow Credentials
// technique (writing a public key to a target's msDS-KeyCredentialLink and then
// PKINIT-authenticating as that target). All cryptography is native Go stdlib
// (crypto/rsa, crypto/rand, math/big, crypto/sha1, encoding/asn1); the CMS
// SignedData is hand-built (see cms.go) with no external PKCS#7 dependency.
package pkinit

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// DHGroup describes a MODP Diffie-Hellman group (RFC 2412 / RFC 3526) used for
// PKINIT key agreement. Only the prime modulus P and generator G are needed to
// perform the exchange; the subgroup order Q is advertised to the KDC in the
// DomainParameters (a value of 0 means "unspecified", which the Windows KDC
// accepts, matching interoperable PKINIT clients).
type DHGroup struct {
	// ID is a human-readable label ("modp2", "modp14").
	ID string
	// P is the prime modulus.
	P *big.Int
	// G is the generator.
	G *big.Int
	// Q is the subgroup order advertised in DomainParameters (0 = unspecified).
	Q *big.Int
}

// hexToInt parses a whitespace-separated hexadecimal string into a big.Int.
func hexToInt(s string) *big.Int {
	clean := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\n' || c == '\t' || c == '\r' {
			continue
		}
		clean = append(clean, c)
	}
	n, ok := new(big.Int).SetString(string(clean), 16)
	if !ok {
		panic("pkinit: invalid hard-coded MODP prime")
	}
	return n
}

// modp2Prime is the 1024-bit MODP group (RFC 2412 Appendix E.2, "Second Oakley
// Group"; RFC 4556 well-known group 2).
const modp2Prime = `
	FFFFFFFF FFFFFFFF C90FDAA2 2168C234 C4C6628B 80DC1CD1
	29024E08 8A67CC74 020BBEA6 3B139B22 514A0879 8E3404DD
	EF9519B3 CD3A431B 302B0A6D F25F1437 4FE1356D 6D51C245
	E485B576 625E7EC6 F44C42E9 A637ED6B 0BFF5CB6 F406B7ED
	EE386BFB 5A899FA5 AE9F2411 7C4B1FE6 49286651 ECE65381
	FFFFFFFF FFFFFFFF`

// modp14Prime is the 2048-bit MODP group (RFC 3526 Section 3; RFC 4556
// well-known group 14).
const modp14Prime = `
	FFFFFFFF FFFFFFFF C90FDAA2 2168C234 C4C6628B 80DC1CD1
	29024E08 8A67CC74 020BBEA6 3B139B22 514A0879 8E3404DD
	EF9519B3 CD3A431B 302B0A6D F25F1437 4FE1356D 6D51C245
	E485B576 625E7EC6 F44C42E9 A637ED6B 0BFF5CB6 F406B7ED
	EE386BFB 5A899FA5 AE9F2411 7C4B1FE6 49286651 ECE45B3D
	C2007CB8 A163BF05 98DA4836 1C55D39A 69163FA8 FD24CF5F
	83655D23 DCA3AD96 1C62F356 208552BB 9ED52907 7096966D
	670C354E 4ABC9804 F1746C08 CA18217C 32905E46 2E36CE3B
	E39E772C 180E8603 9B2783A2 EC07A28F B5C55DF0 6F4C52C9
	DE2BCBF6 95581718 3995497C EA956AE5 15D22618 98FA0510
	15728E5A 8AACAA68 FFFFFFFF FFFFFFFF`

// MODPGroup2 returns the 1024-bit MODP Diffie-Hellman group (RFC 4556 group 2).
// This is the group most widely interoperable with Windows KDCs.
func MODPGroup2() DHGroup {
	return DHGroup{ID: "modp2", P: hexToInt(modp2Prime), G: big.NewInt(2), Q: big.NewInt(0)}
}

// MODPGroup14 returns the 2048-bit MODP Diffie-Hellman group (RFC 4556 group 14).
func MODPGroup14() DHGroup {
	return DHGroup{ID: "modp14", P: hexToInt(modp14Prime), G: big.NewInt(2), Q: big.NewInt(0)}
}

// modulusLen returns the byte length of the group's prime modulus.
func (g DHGroup) modulusLen() int {
	return (g.P.BitLen() + 7) / 8
}

// DHKeyPair is an ephemeral Diffie-Hellman key pair for a PKINIT exchange.
type DHKeyPair struct {
	Group DHGroup
	// X is the private exponent.
	X *big.Int
	// Y is the public value G^X mod P.
	Y *big.Int
}

// GenerateDHKeyPair creates a fresh ephemeral DH key pair in the given group.
// The private exponent is drawn uniformly from [2, P-2].
func GenerateDHKeyPair(group DHGroup) (*DHKeyPair, error) {
	// max = P-3, so x = rand(0..P-4) + 2 lands in [2, P-2].
	max := new(big.Int).Sub(group.P, big.NewInt(3))
	if max.Sign() <= 0 {
		return nil, fmt.Errorf("pkinit: degenerate DH group")
	}
	r, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, fmt.Errorf("pkinit: generate DH private exponent: %w", err)
	}
	x := r.Add(r, big.NewInt(2))
	y := new(big.Int).Exp(group.G, x, group.P)
	return &DHKeyPair{Group: group, X: x, Y: y}, nil
}

// SharedSecret computes the DH shared secret peerY^X mod P and returns it as a
// big-endian octet string left-padded with zeros to the length of the prime
// modulus, as required by RFC 4556 §3.2.3.1 ("padded with leading zeros ... so
// its size in octets equals the modulus").
func (kp *DHKeyPair) SharedSecret(peerY *big.Int) ([]byte, error) {
	if peerY == nil || peerY.Sign() <= 0 || peerY.Cmp(kp.Group.P) >= 0 {
		return nil, fmt.Errorf("pkinit: peer DH public value out of range")
	}
	z := new(big.Int).Exp(peerY, kp.X, kp.Group.P)
	return leftPad(z.Bytes(), kp.Group.modulusLen()), nil
}

// leftPad returns b left-padded with zero bytes to exactly size bytes. If b is
// already >= size, it is returned unchanged.
func leftPad(b []byte, size int) []byte {
	if len(b) >= size {
		return b
	}
	out := make([]byte, size)
	copy(out[size-len(b):], b)
	return out
}
