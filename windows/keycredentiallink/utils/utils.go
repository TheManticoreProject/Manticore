package utils

import (
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/source"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"

	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
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
		return decodeBase64Identifier(keyIdentifier)
	default:
		return decodeBase64Identifier(keyIdentifier)
	}
}

// decodeBase64Identifier decodes a base64 key identifier, accepting both the padded
// form ConvertFromBinaryIdentifier produces and an unpadded one.
//
// The padding is left as it is rather than rewritten. Base64 uses two padding
// characters when the encoded length is 1 mod 3, one when it is 2 mod 3 and none
// otherwise, so stripping the padding and appending a fixed single '=' only ever
// produced a decodable string for the 2-mod-3 case and corrupted the length of
// every other one.
//
// Parameters:
// - keyIdentifier: A string containing the base64 representation of the key identifier.
//
// Returns:
// - A byte slice containing the decoded key identifier.
// - An error if the string is not valid base64.
func decodeBase64Identifier(keyIdentifier string) ([]byte, error) {
	if strings.HasSuffix(keyIdentifier, "=") {
		return base64.StdEncoding.DecodeString(keyIdentifier)
	}

	return base64.RawStdEncoding.DecodeString(keyIdentifier)
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

// timeEncoding is how a KeyCredentialLink timestamp entry is encoded.
type timeEncoding int

const (
	// fileTimeEncoding is a Windows FILETIME: 100-nanosecond ticks since
	// 1601-01-01 UTC (DateTime.ToFileTime).
	fileTimeEncoding timeEncoding = iota
	// dotNetTicksEncoding is .NET DateTime.Ticks: 100-nanosecond ticks since
	// 0001-01-01.
	dotNetTicksEncoding
	// dotNetBinaryEncoding is .NET DateTime.ToBinary: ticks since 0001-01-01 in
	// the low 62 bits and the DateTimeKind in the top two.
	dotNetBinaryEncoding
)

// timeEncodingFor returns the timestamp encoding of a KeyCredentialLink entry, as
// DSInternals (KeyCredential.ConvertFromBinaryTime) and Windows use it: version 0
// stores .NET ticks, version 1 DateTime.ToBinary, and version 2 and later a
// FILETIME for keys whose source is AD and DateTime.ToBinary for Azure AD keys.
func timeEncodingFor(ksrc source.KeySource, kcv version.KeyCredentialLinkVersion) timeEncoding {
	switch kcv.Value {
	case version.KeyCredentialLinkVersion_0:
		return dotNetTicksEncoding
	case version.KeyCredentialLinkVersion_1:
		return dotNetBinaryEncoding
	default:
		if ksrc.Value == source.KeySource_AD {
			return fileTimeEncoding
		}
		return dotNetBinaryEncoding
	}
}

const (
	// ticksBetweenYear1And1601 is the number of 100-nanosecond ticks from
	// 0001-01-01 (the .NET epoch) to 1601-01-01 (the FILETIME epoch).
	ticksBetweenYear1And1601 uint64 = 504911232000000000

	// unixSecondsOfYear1 is 0001-01-01 00:00:00 UTC in Unix seconds.
	unixSecondsOfYear1 int64 = -62135596800

	// dotNetTicksMask selects the tick count of a DateTime.ToBinary value; the
	// top two bits hold the DateTimeKind.
	dotNetTicksMask uint64 = 0x3FFFFFFFFFFFFFFF
	// dotNetKindUtc is DateTimeKind.Utc in a DateTime.ToBinary value.
	dotNetKindUtc uint64 = 0x4000000000000000
	// dotNetKindLocal is set for DateTimeKind.Local, whose ToBinary value holds the
	// UTC tick count, possibly wrapped below zero by the local offset.
	dotNetKindLocal uint64 = 0x8000000000000000
	// dotNetTicksCeiling and ticksPerDay bound the wrapped UTC tick count of a
	// local DateTime (DateTime.FromBinary).
	dotNetTicksCeiling uint64 = 0x4000000000000000
	ticksPerDay        uint64 = 864000000000
)

// dateTimeFromDotNetTicks builds a DateTime from 100-nanosecond ticks since
// 0001-01-01 UTC. An instant before 1601 has no FILETIME tick count, so its
// DateTime holds the time with a tick count of 0.
func dateTimeFromDotNetTicks(ticks uint64) DateTime {
	if ticks >= ticksBetweenYear1And1601 {
		return dateTimeFromWireTicks(ticks - ticksBetweenYear1And1601)
	}
	seconds := int64(ticks / ticksPerSecond)
	nanos := int64(ticks%ticksPerSecond) * 100
	return DateTime{time: time.Unix(unixSecondsOfYear1+seconds, nanos).UTC()}
}

// dateTimeFromDotNetBinary decodes a DateTime.ToBinary value the way
// DateTime.FromBinary does, returning the instant it denotes. A local-kind value
// already holds UTC ticks; an unspecified-kind value carries no zone and is read
// as UTC.
func dateTimeFromDotNetBinary(value uint64) DateTime {
	ticks := value & dotNetTicksMask
	if value&dotNetKindLocal != 0 && ticks > dotNetTicksCeiling-ticksPerDay {
		// A local time within a day of 0001-01-01 whose UTC equivalent fell below
		// zero; FromBinary unwraps it to a negative tick count, which is
		// before any representable instant, so it is clamped to the epoch.
		ticks = 0
	}
	return dateTimeFromDotNetTicks(ticks)
}

// dotNetMaxTicks is DateTime.MaxValue.Ticks, 9999-12-31 23:59:59.9999999.
const dotNetMaxTicks uint64 = 3155378975999999999

// dotNetTicksFromTime returns the 100-nanosecond ticks from 0001-01-01 UTC to t,
// clamped to the range a .NET DateTime can hold (0001-01-01 to 9999-12-31), so
// an out-of-range time cannot wrap into another date or into the kind bits.
func dotNetTicksFromTime(t time.Time) uint64 {
	t = t.UTC()
	if t.Unix() < unixSecondsOfYear1 {
		return 0
	}
	seconds := uint64(t.Unix() - unixSecondsOfYear1)
	if seconds > dotNetMaxTicks/ticksPerSecond {
		return dotNetMaxTicks
	}
	return min(seconds*ticksPerSecond+uint64(t.Nanosecond())/100, dotNetMaxTicks)
}

// ConvertFromBinaryTime converts the binary value of a KeyCredentialLink timestamp
// entry (KeyApproximateLastLogonTimeStamp or KeyCreationTime) to a DateTime.
//
// The 64-bit little-endian value is decoded per the entry's version and key
// source, as DSInternals and Windows encode it:
//   - Version 0: .NET DateTime.Ticks, 100-nanosecond ticks since 0001-01-01.
//   - Version 1: .NET DateTime.ToBinary, ticks since 0001-01-01 with the
//     DateTimeKind in the top two bits.
//   - Version 2 and later: a FILETIME (ticks since 1601-01-01 UTC) when the key
//     source is AD, and DateTime.ToBinary when it is Azure AD.
//
// Parameters:
// - rawBinaryTime: A byte slice containing the binary representation of the time.
// - source: The source of the key, which selects the encoding for version 2 and later.
// - version: The version of the KeyCredentialLink, which selects the encoding.
//
// Returns:
// - A DateTime object representing the converted time.
// - An error if the buffer cannot hold a 64-bit timestamp.
//
// Src : https://github.com/MichaelGrafnetter/DSInternals/blob/master/Src/DSInternals.Common/Data/Hello/KeyCredential.cs
func ConvertFromBinaryTime(rawBinaryTime []byte, ksrc source.KeySource, kcv version.KeyCredentialLinkVersion) (DateTime, error) {
	// The value comes from a KEYCREDENTIALLINK_ENTRY whose length is whatever the
	// entry declared, so a blob can parse into an entry too short to hold the
	// field. Reading it unchecked indexes past the end of the slice.
	if len(rawBinaryTime) < 8 {
		return DateTime{}, fmt.Errorf("insufficient bytes for a timestamp (expected at least 8, got %d)", len(rawBinaryTime))
	}

	timeStamp := binary.LittleEndian.Uint64(rawBinaryTime)

	var dt DateTime
	switch timeEncodingFor(ksrc, kcv) {
	case dotNetTicksEncoding:
		// Version 0 has no kind bits: the whole value is the tick count.
		dt = dateTimeFromDotNetTicks(timeStamp)
	case dotNetBinaryEncoding:
		dt = dateTimeFromDotNetBinary(timeStamp)
	default:
		dt = dateTimeFromWireTicks(timeStamp)
	}
	copy(dt.wire[:], rawBinaryTime[:8])
	dt.hasWire = true
	return dt, nil
}

// EncodeBinaryTime returns the binary value of a KeyCredentialLink timestamp entry
// for dt: the bytes dt was decoded from when it came from ConvertFromBinaryTime
// and has not been changed since, and otherwise ConvertToBinaryTime of its time
// for the given version and key source.
//
// Re-emitting the decoded bytes keeps a parsed blob byte-identical when it is
// marshalled again, which its KeyHash requires: a DateTime.ToBinary value carries
// a DateTimeKind (local, unspecified) that re-encoding from the instant would
// replace. The decoded bytes are tied to the version and key source they were
// read under; a caller that changes either on the credential should SetTime the
// timestamps so they are re-encoded.
func EncodeBinaryTime(dt DateTime, ksrc source.KeySource, kcv version.KeyCredentialLinkVersion) []byte {
	if dt.hasWire {
		return append([]byte(nil), dt.wire[:]...)
	}
	return ConvertToBinaryTime(dt.GetTime(), ksrc, kcv)
}

// ConvertToBinaryTime converts a time.Time object to the binary value of a
// KeyCredentialLink timestamp entry, little-endian, in the encoding
// ConvertFromBinaryTime decodes for the same version and key source. A
// DateTime.ToBinary value is written with DateTimeKind.Utc.
//
// Parameters:
// - date: A time.Time object representing the time to be converted.
// - source: The source of the key, which selects the encoding for version 2 and later.
// - version: The version of the KeyCredentialLink, which selects the encoding.
//
// Returns:
// - A byte slice containing the binary representation of the time in little-endian format.
func ConvertToBinaryTime(date time.Time, ksrc source.KeySource, kcv version.KeyCredentialLinkVersion) []byte {
	switch timeEncodingFor(ksrc, kcv) {
	case dotNetTicksEncoding:
		return binary.LittleEndian.AppendUint64(nil, dotNetTicksFromTime(date))
	case dotNetBinaryEncoding:
		return binary.LittleEndian.AppendUint64(nil, dotNetTicksFromTime(date)|dotNetKindUtc)
	default:
		// The field holds ticks since 1601, not nanoseconds since 1970. Deriving
		// them through NewDateTimeFromTime keeps this function the exact inverse
		// of ConvertFromBinaryTime instead of an encoder for a different
		// representation.
		return binary.LittleEndian.AppendUint64(nil, NewDateTimeFromTime(date).ToTicks())
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
