package kerbcrypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
)

// rc4HMACUsageMap translates RFC 4120 key usage numbers to the Microsoft
// message-type values used in RC4-HMAC key derivation, per RFC 4757 Section 4
// and MS-KILE Section 3.1.5.7.
var rc4HMACUsageMap = map[int]uint32{
	3:  8,  // AS-REP enc-part
	9:  8,  // TGS-REP enc-part (sub-session key)
	23: 13, // AD-KDC-ISSUED checksum
}

// mapRC4HMACUsage converts an RFC 4120 key usage to the MS message-type value.
// Usages not in the map pass through unchanged.
func mapRC4HMACUsage(usage int) uint32 {
	if mapped, ok := rc4HMACUsageMap[usage]; ok {
		return mapped
	}
	return uint32(usage)
}

// usageMsgType encodes a mapped usage as a 4-byte little-endian slice.
// binary.PutUvarint is used for consistency with MS/gokrb5 reference behaviour.
func usageMsgType(usage int) []byte {
	mapped := mapRC4HMACUsage(usage)
	tb := make([]byte, 4)
	binary.PutUvarint(tb, uint64(mapped))
	return tb
}

// rc4HMACStringToKey derives an RC4-HMAC key from a password.
// For RC4-HMAC, the key is simply the NT hash of the password (MD4 of UTF-16LE).
// RFC 4757 Section 7.
func rc4HMACStringToKey(password string) []byte {
	h := nt.NTHash(password)
	key := make([]byte, 16)
	copy(key, h[:])
	return key
}

// hmacMD5 computes HMAC-MD5 of the data using the given key.
func hmacMD5(key, data []byte) []byte {
	mac := hmac.New(md5.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// rc4HMACEncrypt encrypts plaintext using RC4-HMAC (etype 23).
// Implements the algorithm from RFC 4757 Section 4 / MS-KILE:
//
//	K1  = key  (the base key)
//	K2  = HMAC-MD5(K1, UsageMsgType(usage))
//	conf = 8 random bytes
//	data = conf || plaintext
//	chksum = HMAC-MD5(K2, data)
//	K3  = HMAC-MD5(K2, chksum)
//	ciphertext = RC4(K3, data)
//	output = chksum(16) || ciphertext
func rc4HMACEncrypt(key []byte, usage int, plaintext []byte) ([]byte, error) {
	k1 := key
	k2 := hmacMD5(k1, usageMsgType(usage))

	// Generate 8-byte confounder
	conf := make([]byte, 8)
	if _, err := randRead(conf); err != nil {
		return nil, err
	}

	// data = confounder || plaintext
	data := make([]byte, 8+len(plaintext))
	copy(data[:8], conf)
	copy(data[8:], plaintext)

	// chksum = HMAC-MD5(K2, data)
	chksum := hmacMD5(k2, data)

	// K3 = HMAC-MD5(K2, chksum)
	k3 := hmacMD5(k2, chksum)

	// Encrypt with RC4(K3, data)
	cipher, err := rc4.NewCipher(k3)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(data))
	cipher.XORKeyStream(ciphertext, data)

	// output = chksum || ciphertext
	result := make([]byte, 16+len(ciphertext))
	copy(result[:16], chksum)
	copy(result[16:], ciphertext)
	return result, nil
}

// rc4HMACDecrypt decrypts ciphertext encrypted with RC4-HMAC (etype 23).
//
// Input format: chksum(16) || encrypted(confounder||plaintext)
// Decryption:
//
//	K2 = HMAC-MD5(key, UsageMsgType(usage))
//	K3 = HMAC-MD5(K2, chksum)
//	data = RC4(K3, ciphertext)
//	verify: HMAC-MD5(K2, data) == chksum
//	return data[8:] (skip 8-byte confounder)
func rc4HMACDecrypt(key []byte, usage int, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 24 { // 16 chksum + 8 confounder minimum
		return nil, ErrCiphertextTooShort
	}

	chksum := ciphertext[:16]
	encrypted := ciphertext[16:]

	k1 := key
	k2 := hmacMD5(k1, usageMsgType(usage))

	// K3 = HMAC-MD5(K2, chksum)
	k3 := hmacMD5(k2, chksum)

	// Decrypt with RC4(K3, encrypted)
	cipher, err := rc4.NewCipher(k3)
	if err != nil {
		return nil, err
	}
	data := make([]byte, len(encrypted))
	cipher.XORKeyStream(data, encrypted)

	// Verify integrity: HMAC-MD5(K2, data) must equal chksum
	expectedChksum := hmacMD5(k2, data)
	if !hmac.Equal(chksum, expectedChksum) {
		return nil, ErrIntegrityCheckFailed
	}

	// Strip the 8-byte confounder
	if len(data) < 8 {
		return nil, ErrCiphertextTooShort
	}
	return data[8:], nil
}
