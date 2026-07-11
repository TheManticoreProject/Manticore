package pkinit

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/asn1"
	"encoding/binary"
	"hash"
)

// OctetString2Key implements the RFC 4556 §3.2.3.1 key-derivation function
// used to turn the PKINIT Diffie-Hellman shared secret into the AS reply key:
//
//	octetstring2key(x) == random-to-key(K-truncate(
//	    SHA1(0x00 | x) | SHA1(0x01 | x) | SHA1(0x02 | x) | ... ))
//
// where x = DHSharedSecret | n_c | n_k, K-truncate keeps the first K bits, and
// random-to-key() derives a protocol key of length keyLen bytes from that
// bitstring. For the AES and RC4 enctypes used by Windows KDCs random-to-key is
// the identity function, so the truncated SHA-1 stream is the key directly.
//
// keyLen is the key-generation seed length in bytes for the AS reply key's
// enctype (16 for AES128/RC4, 32 for AES256).
func OctetString2Key(x []byte, keyLen int) []byte {
	out := make([]byte, 0, keyLen)
	for counter := 0; len(out) < keyLen; counter++ {
		h := sha1.New()
		h.Write([]byte{byte(counter)})
		h.Write(x)
		digest := h.Sum(nil)
		need := keyLen - len(out)
		if need < len(digest) {
			out = append(out, digest[:need]...)
		} else {
			out = append(out, digest...)
		}
	}
	return out
}

// AgilityKDF implements the RFC 8636 §3 PKINIT algorithm-agility key-derivation
// function: the NIST SP800-56A single-step (concatenation) KDF instantiated with
// a SHA-2 (or SHA-1) hash. It derives a reply key of keyLen bytes from the DH
// shared secret z and the DER-encoded OtherInfo structure (see BuildKDFOtherInfo):
//
//	reps    = ceil(keyLen*8 / H_outputBits)
//	counter = 0x00000001                       (32-bit big-endian, incrementing)
//	Hash_i  = H(counter || z || OtherInfo)     for i = 1..reps
//	key     = (Hash_1 || Hash_2 || ...) truncated to keyLen bytes
//
// newHash selects the hash primitive per the negotiated kdfID (see kdfHash).
// Unlike the RFC 4556 OctetString2Key above, the counter is a four-octet
// big-endian value that *precedes* the secret, and the OtherInfo (party info and
// suppPubInfo) is appended after z.
func AgilityKDF(newHash func() hash.Hash, z, otherInfo []byte, keyLen int) []byte {
	out := make([]byte, 0, keyLen)
	var counter [4]byte
	for i := uint32(1); len(out) < keyLen; i++ {
		binary.BigEndian.PutUint32(counter[:], i)
		h := newHash()
		h.Write(counter[:])
		h.Write(z)
		h.Write(otherInfo)
		out = append(out, h.Sum(nil)...)
	}
	return out[:keyLen]
}

// kdfHash maps a negotiated RFC 8636 kdfID OID to its hash constructor, reporting
// whether the OID names a KDF this client implements. The SHA-256 and SHA-384
// agility KDFs derive the AES-SHA2 (etype 19/20) reply keys; the SHA-1 agility
// KDF is distinct from the legacy RFC 4556 OctetString2Key (which is used only
// when the KDC returns no kdfID at all).
func kdfHash(oid asn1.ObjectIdentifier) (func() hash.Hash, bool) {
	switch {
	case oid.Equal(oidPKINITKDFSHA1):
		return sha1.New, true
	case oid.Equal(oidPKINITKDFSHA256):
		return sha256.New, true
	case oid.Equal(oidPKINITKDFSHA384):
		return sha512.New384, true
	}
	return nil, false
}
