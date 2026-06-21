package msdrsr

import (
	"crypto/des"
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
	"fmt"
)

// decryptReplicatedValue removes the DRSUAPI transport encryption layer from a secret
// attribute value ([MS-DRSR] 4.1.10.5, EncryptValuesIfNecessary; the wire structure is
// ENCRYPTED_PAYLOAD). The layout is:
//
//	Salt[16]              cleartext
//	CheckSum[4]           RC4-encrypted CRC32 of the value
//	EncryptedData[...]    RC4-encrypted value
//
// The RC4 key is MD5(sessionKey || Salt) and covers CheckSum||EncryptedData. The 4-byte
// checksum is stripped (its verification is advisory and omitted, matching impacket).
func decryptReplicatedValue(sessionKey, payload []byte) ([]byte, error) {
	if len(payload) < 20 {
		return nil, fmt.Errorf("drsuapi: encrypted payload too short: %d bytes (need >=20)", len(payload))
	}
	salt := payload[:16]

	h := md5.New()
	h.Write(sessionKey)
	h.Write(salt)
	c, err := rc4.NewCipher(h.Sum(nil))
	if err != nil {
		return nil, fmt.Errorf("drsuapi: rc4 key: %w", err)
	}
	plain := make([]byte, len(payload)-16)
	c.XORKeyStream(plain, payload[16:])
	// plain[0:4] is the CRC32 checksum; the value follows.
	return plain[4:], nil
}

// deriveKeysFromRID derives the two DES keys from an account RID
// ([MS-SAMR] 2.2.11.1.3): the little-endian RID bytes are reordered into two 7-byte
// sequences which are each expanded to an 8-byte DES key.
func deriveKeysFromRID(rid uint32) (key1, key2 [8]byte) {
	var k [4]byte
	binary.LittleEndian.PutUint32(k[:], rid)
	s1 := [7]byte{k[0], k[1], k[2], k[3], k[0], k[1], k[2]}
	s2 := [7]byte{k[3], k[0], k[1], k[2], k[3], k[0], k[1]}
	return transformKey(s1), transformKey(s2)
}

// transformKey expands a 7-byte (56-bit) value into an 8-byte DES key by spreading the
// bits across the high 7 bits of each output byte and leaving the low bit for parity
// ([MS-SAMR] 2.2.11.1.2). Go's DES ignores the parity bit, so it is left zero.
func transformKey(in [7]byte) [8]byte {
	var out [8]byte
	out[0] = in[0] >> 1
	out[1] = ((in[0] & 0x01) << 6) | (in[1] >> 2)
	out[2] = ((in[1] & 0x03) << 5) | (in[2] >> 3)
	out[3] = ((in[2] & 0x07) << 4) | (in[3] >> 4)
	out[4] = ((in[3] & 0x0F) << 3) | (in[4] >> 5)
	out[5] = ((in[4] & 0x1F) << 2) | (in[5] >> 6)
	out[6] = ((in[5] & 0x3F) << 1) | (in[6] >> 7)
	out[7] = in[6] & 0x7F
	for i := range out {
		out[i] = (out[i] << 1) & 0xFE
	}
	return out
}

// removeDESLayer removes the RID-keyed DES layer from a 16-byte encrypted hash
// ([MS-SAMR] 2.2.11.1.1): the two halves are DES-ECB-decrypted with the RID-derived
// keys, yielding the plaintext NT or LM hash.
func removeDESLayer(encryptedHash []byte, rid uint32) ([16]byte, error) {
	var out [16]byte
	if len(encryptedHash) != 16 {
		return out, fmt.Errorf("drsuapi: encrypted hash must be 16 bytes, got %d", len(encryptedHash))
	}
	k1, k2 := deriveKeysFromRID(rid)
	c1, err := des.NewCipher(k1[:])
	if err != nil {
		return out, fmt.Errorf("drsuapi: des key1: %w", err)
	}
	c2, err := des.NewCipher(k2[:])
	if err != nil {
		return out, fmt.Errorf("drsuapi: des key2: %w", err)
	}
	c1.Decrypt(out[0:8], encryptedHash[0:8])
	c2.Decrypt(out[8:16], encryptedHash[8:16])
	return out, nil
}

// decryptHash decrypts a single-hash secret attribute value (unicodePwd / dBCSPwd):
// the transport layer is removed, then the RID/DES layer, yielding the 16-byte hash.
func decryptHash(sessionKey, value []byte, rid uint32) ([16]byte, error) {
	inner, err := decryptReplicatedValue(sessionKey, value)
	if err != nil {
		return [16]byte{}, err
	}
	if len(inner) != 16 {
		return [16]byte{}, fmt.Errorf("drsuapi: decrypted hash is %d bytes, want 16", len(inner))
	}
	return removeDESLayer(inner, rid)
}

// decryptHashHistory decrypts a password-history secret value (ntPwdHistory /
// lmPwdHistory): the transport layer yields a concatenation of 16-byte DES-encrypted
// hashes, each of which is then RID/DES-decrypted.
func decryptHashHistory(sessionKey, value []byte, rid uint32) ([][16]byte, error) {
	inner, err := decryptReplicatedValue(sessionKey, value)
	if err != nil {
		return nil, err
	}
	if len(inner)%16 != 0 {
		return nil, fmt.Errorf("drsuapi: history length %d is not a multiple of 16", len(inner))
	}
	out := make([][16]byte, 0, len(inner)/16)
	for off := 0; off < len(inner); off += 16 {
		h, err := removeDESLayer(inner[off:off+16], rid)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}
