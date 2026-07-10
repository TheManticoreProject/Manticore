package pkinit

import "crypto/sha1"

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
