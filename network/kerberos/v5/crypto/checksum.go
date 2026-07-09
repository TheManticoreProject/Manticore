package kerbcrypto

import (
	"crypto/hmac"
	"crypto/md5"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// ErrUnsupportedCksumType is returned when a checksum type is not supported.
var ErrUnsupportedCksumType = fmt.Errorf("kerbcrypto: unsupported checksum type")

// GetChecksum computes a keyed Kerberos checksum of data for the given checksum
// type, protocol key, and key usage. The supported types are the ones paired
// with the enctypes this package implements:
//
//	15  hmac-sha1-96-aes128     (RFC 3961/3962, etype 17) — 12-byte output
//	16  hmac-sha1-96-aes256     (RFC 3961/3962, etype 18) — 12-byte output
//	19  hmac-sha256-128-aes128  (RFC 8009, etype 19)      — 16-byte output
//	20  hmac-sha384-192-aes256  (RFC 8009, etype 20)      — 24-byte output
//	-138 hmac-md5 (KERB_CHECKSUM_HMAC_MD5, RFC 4757, RC4) — 16-byte output
func GetChecksum(cksumType int, key []byte, usage int, data []byte) ([]byte, error) {
	switch cksumType {
	case iana.CksumTypeHMACSHA196AES128:
		return aesSHA1Checksum(key, 16, usage, data), nil
	case iana.CksumTypeHMACSHA196AES256:
		return aesSHA1Checksum(key, 32, usage, data), nil
	case iana.CksumTypeHMACSHA256128AES128:
		return aes2Checksum(key, iana.ETypeAES128CTSHMACSHA256, usage, data)
	case iana.CksumTypeHMACSHA384192AES256:
		return aes2Checksum(key, iana.ETypeAES256CTSHMACSHA384, usage, data)
	case iana.CksumTypeHMACMD5:
		return rc4HMACChecksum(key, usage, data), nil
	default:
		return nil, fmt.Errorf("%w: %d", ErrUnsupportedCksumType, cksumType)
	}
}

// VerifyChecksum recomputes the checksum of data and compares it, in constant
// time, against want. It returns false (never an error) for unsupported types.
func VerifyChecksum(cksumType int, key []byte, usage int, data, want []byte) bool {
	got, err := GetChecksum(cksumType, key, usage, data)
	if err != nil {
		return false
	}
	return hmac.Equal(got, want)
}

// ChecksumTypeForEType returns the checksum type paired with an encryption type
// (RFC 3961/3962/8009/4757), i.e. the checksum an authenticator or PA-FOR-USER
// should use when keyed with a key of that etype. Reports false for unsupported
// etypes.
func ChecksumTypeForEType(etype int) (int, bool) {
	switch etype {
	case iana.ETypeRC4HMAC:
		return iana.CksumTypeHMACMD5, true
	case iana.ETypeAES128CTSHMACSHA196:
		return iana.CksumTypeHMACSHA196AES128, true
	case iana.ETypeAES256CTSHMACSHA196:
		return iana.CksumTypeHMACSHA196AES256, true
	case iana.ETypeAES128CTSHMACSHA256:
		return iana.CksumTypeHMACSHA256128AES128, true
	case iana.ETypeAES256CTSHMACSHA384:
		return iana.CksumTypeHMACSHA384192AES256, true
	default:
		return 0, false
	}
}

// aesSHA1Checksum implements the RFC 3961 simplified-profile checksum used by
// the AES-SHA1 enctypes (cksumtype 15/16): Kc = DK(base-key, usage | 0x99),
// then HMAC-SHA1(Kc, data) truncated to 96 bits (12 bytes). keyLen is the AES
// key length (16 for aes128, 32 for aes256).
func aesSHA1Checksum(key []byte, keyLen, usage int, data []byte) []byte {
	kc := deriveKey(key, usageConstant(usage, 0x99), keyLen)
	return hmacSHA1(kc, data)[:12]
}

// rc4HMACChecksum implements KERB_CHECKSUM_HMAC_MD5 (cksumtype -138) per
// RFC 4757 Section 4:
//
//	Ksign = HMAC-MD5(key, "signaturekey\0")
//	tmp   = MD5(T | data)                     where T = usage as an MS msg-type LE uint32
//	CHKSUM = HMAC-MD5(Ksign, tmp)
func rc4HMACChecksum(key []byte, usage int, data []byte) []byte {
	ksign := hmacMD5(key, []byte("signaturekey\x00"))

	h := md5.New()
	h.Write(usageMsgType(usage)) // T, the little-endian mapped message type
	h.Write(data)
	tmp := h.Sum(nil)

	return hmacMD5(ksign, tmp)
}
