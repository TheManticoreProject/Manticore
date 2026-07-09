package kerbcrypto

import (
	"crypto/aes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// Key-usage numbers introduced by RFC 6113 (Kerberos FAST). They live here
// alongside the FAST cryptographic primitives so callers pass them to Encrypt /
// GetChecksum without re-declaring the magic numbers.
const (
	// KeyUsageFASTReqChksum keys the req-checksum in KrbFastArmoredReq
	// (RFC 6113 §5.4.2, KEY_USAGE_FAST_REQ_CHKSUM = 50).
	KeyUsageFASTReqChksum = 50
	// KeyUsageFASTEnc keys the enc-fast-req EncryptedData in KrbFastArmoredReq
	// (RFC 6113 §5.4.2, KEY_USAGE_FAST_ENC = 51).
	KeyUsageFASTEnc = 51
	// KeyUsageFASTRep keys the enc-fast-rep EncryptedData in KrbFastArmoredRep
	// (RFC 6113 §5.4.3, KEY_USAGE_FAST_REP = 52).
	KeyUsageFASTRep = 52
	// KeyUsageFASTFinished keys the ticket-checksum in KrbFastFinished
	// (RFC 6113 §5.4.3, KEY_USAGE_FAST_FINISHED = 53).
	KeyUsageFASTFinished = 53
	// KeyUsageEncChallengeClient keys the client's PA-ENCRYPTED-CHALLENGE
	// (RFC 6113 §5.4.6, KEY_USAGE_ENC_CHALLENGE_CLIENT = 54).
	KeyUsageEncChallengeClient = 54
	// KeyUsageEncChallengeKDC keys the KDC's PA-ENCRYPTED-CHALLENGE reply
	// (RFC 6113 §5.4.6, KEY_USAGE_ENC_CHALLENGE_KDC = 55).
	KeyUsageEncChallengeKDC = 55
)

// PRF implements the RFC 3961 pseudo-random function for the given encryption
// type, as required by the KRB-FX-CF2 key combination of RFC 6113. Each enctype
// binds a specific construction:
//
//   - AES-CTS-HMAC-SHA1-96 (17/18), RFC 3962 §6:
//     prf(key, s) = E(DK(key, "prf"), truncate-128(SHA-1(s)))
//     i.e. AES-encrypt the first 16 bytes of SHA-1(s) under the "prf"-derived key.
//   - AES-CTS-HMAC-SHA2 (19/20), RFC 8009 §5:
//     PRF(key, s) = KDF-HMAC-SHA2(key, "prf" | s, k) with k = 256 (SHA-256)
//     or 384 (SHA-384).
//   - RC4-HMAC (23): HMAC-SHA1(key, s), matching the widely deployed
//     arcfour-hmac PRF.
//
// The output length is fixed by the enctype (16 bytes for AES-SHA1 and RC4, the
// full SHA-2 output for the RFC 8009 types).
func PRF(etype int, key, input []byte) ([]byte, error) {
	switch etype {
	case iana.ETypeAES128CTSHMACSHA196, iana.ETypeAES256CTSHMACSHA196:
		dk := deriveKey(key, []byte("prf"), aesKeyLen(etype))
		block, err := aes.NewCipher(dk)
		if err != nil {
			return nil, err
		}
		h := sha1.Sum(input)
		out := make([]byte, 16)
		block.Encrypt(out, h[:16]) // single 16-byte block (CBC with zero IV == ECB)
		return out, nil

	case iana.ETypeAES128CTSHMACSHA256, iana.ETypeAES256CTSHMACSHA384:
		// RFC 8009 §5: PRF(key, s) = KDF-HMAC-SHA2(key, "prf", s, k), i.e. the
		// SP800-108 counter-mode KDF with label "prf" and s in the context
		// position: HMAC-Hash(key, 0x00000001 | "prf" | 0x00 | s | k), where k is
		// the hash output length in bits. k equals one full hash output, so a
		// single iteration suffices.
		p, _ := aes2ParamsFor(etype)
		kBits := p.newHash().Size() * 8
		var kbuf [4]byte
		binary.BigEndian.PutUint32(kbuf[:], uint32(kBits))
		mac := hmac.New(p.newHash, key)
		mac.Write([]byte{0x00, 0x00, 0x00, 0x01}) // counter i = 1
		mac.Write([]byte("prf"))
		mac.Write([]byte{0x00})
		mac.Write(input)
		mac.Write(kbuf[:])
		return mac.Sum(nil), nil

	case iana.ETypeRC4HMAC:
		return hmacSHA1(key, input), nil

	default:
		return nil, fmt.Errorf("%w: PRF for etype %d", ErrUnsupportedEType, etype)
	}
}

// prfPlus implements the RFC 6113 §5.1 PRF+ expansion:
//
//	PRF+(key, pepper) = PRF(key, 1 | pepper) | PRF(key, 2 | pepper) | ...
//
// where the counter is a single octet starting at 1. It returns the first n
// output bytes.
func prfPlus(etype int, key []byte, pepper string, n int) ([]byte, error) {
	out := make([]byte, 0, n)
	for i := 1; len(out) < n; i++ {
		if i > 255 {
			return nil, fmt.Errorf("kerbcrypto: PRF+ counter overflow producing %d bytes", n)
		}
		block, err := PRF(etype, key, append([]byte{byte(i)}, pepper...))
		if err != nil {
			return nil, err
		}
		out = append(out, block...)
	}
	return out[:n], nil
}

// KRBFXCF2 implements the RFC 6113 §5.1 KRB-FX-CF2 key-combination function:
//
//	KRB-FX-CF2(K1, K2, pepper1, pepper2) =
//	    random-to-key( PRF+(K1, pepper1) ^ PRF+(K2, pepper2) )
//
// The two input keys may have different enctypes; the result adopts K1's
// enctype (its key length drives the PRF+ output length and random-to-key).
// For every enctype implemented here random-to-key is the identity function, so
// the XOR of the two PRF+ streams is the combined key. The returned int is the
// result enctype (= k1Etype).
func KRBFXCF2(k1 []byte, k1Etype int, k2 []byte, k2Etype int, pepper1, pepper2 string) ([]byte, int, error) {
	n := KeyLen(k1Etype)
	if n == 0 {
		return nil, 0, fmt.Errorf("%w: KRB-FX-CF2 output etype %d", ErrUnsupportedEType, k1Etype)
	}
	o1, err := prfPlus(k1Etype, k1, pepper1, n)
	if err != nil {
		return nil, 0, fmt.Errorf("kerbcrypto: KRB-FX-CF2 PRF+ on K1: %w", err)
	}
	o2, err := prfPlus(k2Etype, k2, pepper2, n)
	if err != nil {
		return nil, 0, fmt.Errorf("kerbcrypto: KRB-FX-CF2 PRF+ on K2: %w", err)
	}
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = o1[i] ^ o2[i]
	}
	return out, k1Etype, nil
}
