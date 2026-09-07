package utils

import (
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/source"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"

	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"time"
)

// ConvertToBinaryIdentifier converts a key identifier string to its binary representation.
//
// Parameters:
//
// - keyIdentifier: A string representing the key identifier to be converted.
//
// - version: A KeyCredentialLinkVersion object representing the version of the key credential.
//
// Returns:
//
// - A byte slice containing the binary representation of the key identifier.
//
// - An error if the conversion fails.
//
// Note:
//
// The function handles different versions of key credentials as follows:
//
// - For version 0 and 1, the key identifier is expected to be in hexadecimal format and is decoded using hex.DecodeString.
//
// - For version 2, the key identifier is expected to be in base64 format and is decoded using base64.StdEncoding.DecodeString with padding.
//
// - For any other version, the key identifier is treated as base64 format and is decoded using base64.StdEncoding.DecodeString with padding.
func ConvertToBinaryIdentifier(keyIdentifier string, kcv version.KeyCredentialLinkVersion) ([]byte, error) {
	switch kcv.Value {
	case version.KeyCredentialLinkVersion_0, version.KeyCredentialLinkVersion_1:
		return hex.DecodeString(keyIdentifier)
	case version.KeyCredentialLinkVersion_2:
		return base64.StdEncoding.DecodeString(strings.TrimRight(keyIdentifier, "=") + "=")
	default:
		return base64.StdEncoding.DecodeString(strings.TrimRight(keyIdentifier, "=") + "=")
	}
}

// ConvertFromBinaryIdentifier converts a binary key identifier to its string representation.
//
// Parameters:
// - keyIdentifier: A byte slice containing the binary representation of the key identifier.
// - version: A KeyCredentialLinkVersion object representing the version of the key credential.
//
// Returns:
// - A string representing the key identifier.
//
// Note:
// The function handles different versions of key credentials as follows:
// - For version 0 and 1, the key identifier is encoded to a hexadecimal string using hex.EncodeToString.
// - For version 2, the key identifier is encoded to a base64 string using base64.StdEncoding.EncodeToString.
// - For any other version, the key identifier is treated as base64 format and is encoded using base64.StdEncoding.EncodeToString.
func ConvertFromBinaryIdentifier(keyIdentifier []byte, kcv version.KeyCredentialLinkVersion) string {
	switch kcv.Value {
	case version.KeyCredentialLinkVersion_0, version.KeyCredentialLinkVersion_1:
		return hex.EncodeToString(keyIdentifier)
	case version.KeyCredentialLinkVersion_2:
		return base64.StdEncoding.EncodeToString(keyIdentifier)
	default:
		return base64.StdEncoding.EncodeToString(keyIdentifier)
	}
}

// ConvertFromBinaryTime converts a binary representation of time to a time.Time object.
// The binary representation is expected to be in little-endian format.
//
// Parameters:
// - rawBinaryTime: A byte slice containing the binary representation of the time.
// - source: The source of the key, which can affect the interpretation of the time.
// - version: The version of the KeyCredentialLink, which can affect the interpretation of the time.
//
// Returns:
// - A time.Time object representing the converted time.
//
// Note:
// The function currently treats all versions and sources the same way, decoding the binary
// timestamp as a count of 100-nanosecond intervals since 1601-01-01 00:00:00 UTC.
//
// Src : https://github.com/microsoft/referencesource/blob/master/mscorlib/system/datetime.cs
func ConvertFromBinaryTime(rawBinaryTime []byte, ksrc source.KeySource, kcv version.KeyCredentialLinkVersion) DateTime {
	timeStamp := binary.LittleEndian.Uint64(rawBinaryTime)

	switch kcv.Value {
	case version.KeyCredentialLinkVersion_0, version.KeyCredentialLinkVersion_1:
		return dateTimeFromWireTicks(timeStamp)
	case version.KeyCredentialLinkVersion_2:
		if ksrc.Value == source.KeySource_AD {
			return dateTimeFromWireTicks(timeStamp)
		} else {
			// This is not fully supported right now, you may encounter issues.
			return dateTimeFromWireTicks(timeStamp)
		}
	default:
		if ksrc.Value == source.KeySource_AD {
			return dateTimeFromWireTicks(timeStamp)
		} else {
			// This is not fully supported right now, you may encounter issues.
			return dateTimeFromWireTicks(timeStamp)
		}
	}
}

// ConvertToBinaryTime converts a time.Time object to its binary representation in little-endian format.
//
// Parameters:
// - date: A time.Time object representing the time to be converted.
// - source: The source of the key, which can affect the interpretation of the time.
// - version: The version of the KeyCredentialLink, which can affect the interpretation of the time.
//
// Returns:
// - A byte slice containing the binary representation of the time in little-endian format.
//
// Note:
// The function currently treats all versions and sources the same way, converting the time
// directly to a Unix time in nanoseconds and then encoding it in little-endian format.
func ConvertToBinaryTime(date time.Time, ksrc source.KeySource, kcv version.KeyCredentialLinkVersion) []byte {
	timeStamp := date.UnixNano()

	switch kcv.Value {
	case version.KeyCredentialLinkVersion_0, version.KeyCredentialLinkVersion_1:
		return binary.LittleEndian.AppendUint64(nil, uint64(timeStamp))
	case version.KeyCredentialLinkVersion_2:
		if ksrc.Value == source.KeySource_AD {
			return binary.LittleEndian.AppendUint64(nil, uint64(timeStamp))
		} else {
			return binary.LittleEndian.AppendUint64(nil, uint64(timeStamp))
		}
	default:
		if ksrc.Value == source.KeySource_AD {
			return binary.LittleEndian.AppendUint64(nil, uint64(timeStamp))
		} else {
			return binary.LittleEndian.AppendUint64(nil, uint64(timeStamp))
		}
	}
}

// ComputeHash calculates the SHA-256 hash of the provided data.
//
// Parameters:
// - data: A byte slice containing the input data to be hashed.
//
// Returns:
// - A byte slice containing the SHA-256 hash of the input data.
//
// Note:
// This function uses the SHA-256 hashing algorithm from the crypto/sha256 package to generate a fixed-size
// 32-byte hash. The resulting hash can be used for various purposes, such as data integrity verification
// and cryptographic operations.
func ComputeHash(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

// ComputeKeyIdentifier generates a key identifier based on the provided key material and version.
//
// Parameters:
// - keyMaterial: A byte slice containing the key material to be used for generating the key identifier.
// - version: A version.KeyCredentialLinkVersion value representing the version of the key credential.
//
// Returns:
// - A string representing the generated key identifier.
//
// Note:
// This function first computes the SHA-256 hash of the provided key material using the ComputeHash function.
// It then converts the resulting binary hash to a string representation based on the specified version
// using the ConvertFromBinaryIdentifier function. The generated key identifier can be used for various
// purposes, such as uniquely identifying cryptographic keys and credentials.
func ComputeKeyIdentifier(keyMaterial []byte, version version.KeyCredentialLinkVersion) string {
	binaryId := ComputeHash(keyMaterial)
	return ConvertFromBinaryIdentifier(binaryId, version)
}
