// Package ntds implements the secret-decryption primitives for offline NTDS.dit
// (Active Directory database) analysis: decrypting the Password Encryption Key (PEK)
// list with a SYSTEM boot key, and decrypting the per-account hash blobs stored in
// NTDS attributes (unicodePwd, dBCSPwd, ntPwdHistory, lmPwdHistory, ...).
//
// The decryption matches the reference behaviour of impacket's secretsdump
// (NTDSHashes): a PEK-keyed outer layer (RC4 or AES depending on the format version)
// followed by the RID-keyed DES layer of [MS-SAMR] 2.2.11.1. It is independent of how
// the encrypted bytes are obtained — an offline ESE reader or a remote DRSUAPI reply
// can both feed these functions.
//
// References:
//   - [MS-SAMR] 2.2.11.1 Encrypting/Decrypting an NT or LM Hash Value
//   - impacket secretsdump.py NTDSHashes (__removeRC4Layer / __removeDESLayer, PEKLIST_*)
package ntds

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
	"fmt"
)

// PEK is a 16-byte Password Encryption Key recovered from the pekList attribute.
type PEK []byte

// transformKey expands a 7-byte key into the 8-byte odd-parity DES key used by the
// RID-based hash encryption ([MS-SAMR] 2.2.11.1.2 / 5.1.3).
func transformKey(in []byte) []byte {
	out := make([]byte, 8)
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

// deriveDESKeys derives the two DES keys (Key1, Key2) from a RID, per [MS-SAMR]
// 2.2.11.1.3: the RID is taken little-endian and its bytes reshuffled into two 7-byte
// keys that transformKey expands.
func deriveDESKeys(rid uint32) (key1, key2 []byte) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], rid)
	key1 = transformKey([]byte{b[0], b[1], b[2], b[3], b[0], b[1], b[2]})
	key2 = transformKey([]byte{b[3], b[0], b[1], b[2], b[3], b[0], b[1]})
	return key1, key2
}

// removeDESLayer strips the RID-keyed DES layer from a 16-byte hash: the first 8 bytes
// are DES-decrypted with Key1 and the second 8 with Key2.
func removeDESLayer(hash []byte, rid uint32) ([]byte, error) {
	if len(hash) < 16 {
		return nil, fmt.Errorf("ntds: hash too short for DES layer: %d bytes", len(hash))
	}
	key1, key2 := deriveDESKeys(rid)
	c1, err := des.NewCipher(key1)
	if err != nil {
		return nil, fmt.Errorf("ntds: DES key1: %w", err)
	}
	c2, err := des.NewCipher(key2)
	if err != nil {
		return nil, fmt.Errorf("ntds: DES key2: %w", err)
	}
	out := make([]byte, 16)
	c1.Decrypt(out[0:8], hash[0:8])
	c2.Decrypt(out[8:16], hash[8:16])
	return out, nil
}

// decryptAESCBC decrypts value with key under AES-CBC using iv, zero-padding a trailing
// partial block (matching impacket's decryptAES). A 16-byte key selects AES-128.
func decryptAESCBC(key, value, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("ntds: AES key: %w", err)
	}
	if len(iv) < aes.BlockSize {
		return nil, fmt.Errorf("ntds: AES IV too short: %d bytes", len(iv))
	}
	data := value
	if r := len(data) % aes.BlockSize; r != 0 {
		data = append(append([]byte(nil), data...), make([]byte, aes.BlockSize-r)...)
	}
	out := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, iv[:aes.BlockSize]).CryptBlocks(out, data)
	return out, nil
}

// decryptBlob removes the PEK-keyed outer layer of an encrypted attribute blob,
// returning the inner DES-layered hash bytes (16 for a single hash, a multiple of 16
// for a history). The blob's header selects the PEK index (byte 4) and the algorithm
// (byte 0 == 0x13 means AES, otherwise RC4):
//
//	[0:8]   Header (Header[0]=algorithm tag, Header[4]=PEK index)
//	[8:24]  KeyMaterial (RC4 salt, or AES IV)
//	RC4:    [24:]    EncryptedHash
//	AES:    [24:28]  Unknown, [28:] EncryptedHash
func decryptBlob(peks []PEK, blob []byte) ([]byte, error) {
	if len(blob) < 24 {
		return nil, fmt.Errorf("ntds: encrypted blob too short: %d bytes", len(blob))
	}
	header := blob[0:8]
	keyMaterial := blob[8:24]
	idx := int(header[4])
	if idx < 0 || idx >= len(peks) {
		return nil, fmt.Errorf("ntds: PEK index %d out of range (have %d)", idx, len(peks))
	}

	if header[0] == 0x13 {
		if len(blob) < 28 {
			return nil, fmt.Errorf("ntds: AES blob too short: %d bytes", len(blob))
		}
		return decryptAESCBC(peks[idx], blob[28:], keyMaterial)
	}

	h := md5.New()
	h.Write(peks[idx])
	h.Write(keyMaterial)
	c, err := rc4.NewCipher(h.Sum(nil))
	if err != nil {
		return nil, fmt.Errorf("ntds: RC4 key: %w", err)
	}
	enc := blob[24:]
	out := make([]byte, len(enc))
	c.XORKeyStream(out, enc)
	return out, nil
}

// DecryptHash decrypts a single-hash attribute blob (e.g. unicodePwd / dBCSPwd) using
// the PEK list and the account's RID, returning the raw 16-byte NT or LM hash. The PEK
// index is taken from the blob header.
func DecryptHash(peks []PEK, rid uint32, encrypted []byte) ([]byte, error) {
	plain, err := decryptBlob(peks, encrypted)
	if err != nil {
		return nil, err
	}
	if len(plain) < 16 {
		return nil, fmt.Errorf("ntds: decrypted hash too short: %d bytes", len(plain))
	}
	return removeDESLayer(plain[:16], rid)
}

// DecryptHashHistory decrypts a hash-history attribute blob (ntPwdHistory /
// lmPwdHistory) using the PEK list and the account's RID, returning the sequence of
// raw 16-byte hashes (most recent first, as stored).
func DecryptHashHistory(peks []PEK, rid uint32, encrypted []byte) ([][]byte, error) {
	plain, err := decryptBlob(peks, encrypted)
	if err != nil {
		return nil, err
	}
	var out [][]byte
	for i := 0; i+16 <= len(plain); i += 16 {
		h, err := removeDESLayer(plain[i:i+16], rid)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}
