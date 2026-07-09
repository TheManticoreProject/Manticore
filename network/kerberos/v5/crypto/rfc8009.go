package kerbcrypto

import (
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"hash"

	"github.com/TheManticoreProject/Manticore/crypto/aes/cts"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// aes8009DefaultIterCount is the default PBKDF2 iteration count for the RFC 8009
// AES-SHA2 string-to-key (decimal 32768), distinct from RFC 3962's 4096.
const aes8009DefaultIterCount = 32768

// aes2Params bundles the per-enctype parameters of the RFC 8009 profile.
//
//   - keyBytes: AES key length in bytes (the length of Ke), 16 or 32.
//   - macBytes: truncated HMAC length h in bytes, which is also the length of
//     the integrity key Ki and checksum key Kc: 16 (128 bits) or 24 (192 bits).
//   - newHash:  the SHA-2 variant (SHA-256 for etype 19, SHA-384 for etype 20).
//   - name:     the enctype name prepended to the salt in string-to-key.
type aes2Params struct {
	keyBytes int
	macBytes int
	newHash  func() hash.Hash
	name     string
}

// aes2ParamsFor returns the RFC 8009 parameters for etype 19/20.
func aes2ParamsFor(etype int) (aes2Params, bool) {
	switch etype {
	case iana.ETypeAES128CTSHMACSHA256:
		return aes2Params{keyBytes: 16, macBytes: 16, newHash: sha256.New, name: "aes128-cts-hmac-sha256-128"}, true
	case iana.ETypeAES256CTSHMACSHA384:
		return aes2Params{keyBytes: 32, macBytes: 24, newHash: sha512.New384, name: "aes256-cts-hmac-sha384-192"}, true
	default:
		return aes2Params{}, false
	}
}

// kdfHMACSHA2 implements KDF-HMAC-SHA2 from RFC 8009 Section 3 (the SP800-108
// counter-mode KDF keyed with HMAC-SHA-2):
//
//	K1 = HMAC-Hash(key, i | label | 0x00 | k)
//
// with the 4-byte big-endian counter i starting at 1 and k the requested
// output length in bits (also big-endian, 4 bytes). It returns the leftmost
// kBits/8 bytes. For every derivation this specification needs, kBits never
// exceeds one hash output, so a single iteration is used; the loop keeps the
// function correct should a longer output ever be requested.
func kdfHMACSHA2(newHash func() hash.Hash, key, label []byte, kBits int) []byte {
	need := kBits / 8
	var kbuf [4]byte
	binary.BigEndian.PutUint32(kbuf[:], uint32(kBits))

	out := make([]byte, 0, need)
	for i := uint32(1); len(out) < need; i++ {
		mac := hmac.New(newHash, key)
		var ibuf [4]byte
		binary.BigEndian.PutUint32(ibuf[:], i)
		mac.Write(ibuf[:])
		mac.Write(label)
		mac.Write([]byte{0x00})
		mac.Write(kbuf[:])
		out = mac.Sum(out)
	}
	return out[:need]
}

// aes2StringToKey derives an RFC 8009 base-key from a password and salt.
//
//	saltp    = enctype-name | 0x00 | salt
//	tkey     = PBKDF2(HMAC-Hash, password, saltp, iter_count, keyBytes)
//	base-key = KDF-HMAC-SHA2(tkey, "kerberos", keyBytes*8)
func aes2StringToKey(password, salt string, iterCount int, p aes2Params) ([]byte, error) {
	saltp := make([]byte, 0, len(p.name)+1+len(salt))
	saltp = append(saltp, p.name...)
	saltp = append(saltp, 0x00)
	saltp = append(saltp, salt...)

	tkey, err := pbkdf2.Key(p.newHash, password, saltp, iterCount, p.keyBytes)
	if err != nil {
		return nil, err
	}
	return kdfHMACSHA2(p.newHash, tkey, []byte("kerberos"), p.keyBytes*8), nil
}

// aes2Encrypt encrypts plaintext using aes{128,256}-cts-hmac-sha{256,384} per
// RFC 8009 Section 5 (encrypt-then-MAC over the cipher state | ciphertext):
//
//	Ke = KDF-HMAC-SHA2(base-key, usage | 0xAA, keyBits)
//	Ki = KDF-HMAC-SHA2(base-key, usage | 0x55, macBits)
//	C  = AES-CBC-CS3(Ke, IV=0, confounder(16) | plaintext)
//	H  = HMAC(Ki, IV | C)[:h]
//	ciphertext = C | H
func aes2Encrypt(baseKey []byte, etype, usage int, plaintext []byte) ([]byte, error) {
	p, ok := aes2ParamsFor(etype)
	if !ok {
		return nil, ErrUnsupportedEType
	}
	ke := kdfHMACSHA2(p.newHash, baseKey, usageConstant(usage, 0xAA), p.keyBytes*8)
	ki := kdfHMACSHA2(p.newHash, baseKey, usageConstant(usage, 0x55), p.macBytes*8)

	conf := make([]byte, 16)
	if _, err := randRead(conf); err != nil {
		return nil, err
	}
	ptc := make([]byte, 16+len(plaintext))
	copy(ptc[:16], conf)
	copy(ptc[16:], plaintext)

	ivz := make([]byte, 16) // initial cipher state is all zero
	c, err := cts.Encrypt(ke, ivz, ptc)
	if err != nil {
		return nil, err
	}

	mac := hmac.New(p.newHash, ki)
	mac.Write(ivz)
	mac.Write(c)
	h := mac.Sum(nil)[:p.macBytes]

	result := make([]byte, len(c)+p.macBytes)
	copy(result, c)
	copy(result[len(c):], h)
	return result, nil
}

// aes2Decrypt decrypts and verifies an aes{128,256}-cts-hmac-sha{256,384}
// ciphertext per RFC 8009 Section 5, verifying the HMAC before decrypting and
// stripping the 16-byte confounder.
func aes2Decrypt(baseKey []byte, etype, usage int, ciphertext []byte) ([]byte, error) {
	p, ok := aes2ParamsFor(etype)
	if !ok {
		return nil, ErrUnsupportedEType
	}
	// Minimum ciphertext: one AES block (confounder) + truncated HMAC.
	if len(ciphertext) < 16+p.macBytes {
		return nil, ErrCiphertextTooShort
	}
	ke := kdfHMACSHA2(p.newHash, baseKey, usageConstant(usage, 0xAA), p.keyBytes*8)
	ki := kdfHMACSHA2(p.newHash, baseKey, usageConstant(usage, 0x55), p.macBytes*8)

	c := ciphertext[:len(ciphertext)-p.macBytes]
	recvMAC := ciphertext[len(ciphertext)-p.macBytes:]

	ivz := make([]byte, 16)
	mac := hmac.New(p.newHash, ki)
	mac.Write(ivz)
	mac.Write(c)
	if !hmac.Equal(recvMAC, mac.Sum(nil)[:p.macBytes]) {
		return nil, ErrIntegrityCheckFailed
	}

	ptc, err := cts.Decrypt(ke, ivz, c)
	if err != nil {
		return nil, err
	}
	if len(ptc) < 16 {
		return nil, ErrCiphertextTooShort
	}
	return ptc[16:], nil
}

// aes2Checksum computes the RFC 8009 Section 6 keyed checksum
// (hmac-sha{256,384} paired with the AES-SHA2 enctypes):
//
//	Kc  = KDF-HMAC-SHA2(base-key, usage | 0x99, macBits)
//	MIC = HMAC(Kc, message)[:h]
func aes2Checksum(baseKey []byte, etype, usage int, message []byte) ([]byte, error) {
	p, ok := aes2ParamsFor(etype)
	if !ok {
		return nil, ErrUnsupportedEType
	}
	kc := kdfHMACSHA2(p.newHash, baseKey, usageConstant(usage, 0x99), p.macBytes*8)
	mac := hmac.New(p.newHash, kc)
	mac.Write(message)
	return mac.Sum(nil)[:p.macBytes], nil
}
