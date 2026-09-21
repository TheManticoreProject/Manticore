// Package attributes implements AD attribute wire formats that are not part of the
// standard NDR/RPC layer — packed binary structures stored in or returned by LDAP
// attribute values.
package attributes

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

// ManagedPasswordBlob is the MSDS-MANAGEDPASSWORD_BLOB structure returned by the
// msDS-ManagedPassword LDAP attribute on group Managed Service Accounts (gMSA).
//
// Reference: [MS-ADTS] 2.2.19
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/a9019740-3d73-46ef-a9ae-3ea8eb86ac2e
type ManagedPasswordBlob struct {
	// CurrentPassword is the cleartext current password as a UTF-16LE string.
	CurrentPassword string

	// PreviousPassword is the cleartext previous password as a UTF-16LE string.
	// Empty when no previous password is available.
	PreviousPassword string

	// QueryPasswordIntervalTicks is the time (100-nanosecond intervals) after which
	// the client should re-query the password from the directory.
	QueryPasswordIntervalTicks int64

	// UnchangedPasswordIntervalTicks is the time (100-nanosecond intervals) during
	// which queries are guaranteed to return the same password.
	UnchangedPasswordIntervalTicks int64
}

// managedPasswordHeaderSize is the fixed 16-byte header of MSDS-MANAGEDPASSWORD_BLOB.
const managedPasswordHeaderSize = 16

// ParseManagedPasswordBlob decodes an MSDS-MANAGEDPASSWORD_BLOB from raw bytes.
func ParseManagedPasswordBlob(data []byte) (*ManagedPasswordBlob, error) {
	if len(data) < managedPasswordHeaderSize {
		return nil, fmt.Errorf("managed password blob too short: %d bytes, need at least %d", len(data), managedPasswordHeaderSize)
	}

	version := binary.LittleEndian.Uint16(data[0:2])
	if version != 1 {
		return nil, fmt.Errorf("unsupported managed password blob version: %d", version)
	}

	// reserved := binary.LittleEndian.Uint16(data[2:4])
	length := binary.LittleEndian.Uint32(data[4:8])
	if uint32(len(data)) < length {
		return nil, fmt.Errorf("managed password blob length field %d exceeds data size %d", length, len(data))
	}

	currentOff := binary.LittleEndian.Uint16(data[8:10])
	previousOff := binary.LittleEndian.Uint16(data[10:12])
	queryIntervalOff := binary.LittleEndian.Uint16(data[12:14])
	unchangedIntervalOff := binary.LittleEndian.Uint16(data[14:16])

	blob := &ManagedPasswordBlob{}

	if currentOff != 0 {
		s, err := readUTF16NullTerm(data, int(currentOff))
		if err != nil {
			return nil, fmt.Errorf("current password: %w", err)
		}
		blob.CurrentPassword = s
	}

	if previousOff != 0 {
		s, err := readUTF16NullTerm(data, int(previousOff))
		if err != nil {
			return nil, fmt.Errorf("previous password: %w", err)
		}
		blob.PreviousPassword = s
	}

	if queryIntervalOff != 0 {
		if int(queryIntervalOff)+8 > len(data) {
			return nil, fmt.Errorf("query password interval at offset %d out of bounds", queryIntervalOff)
		}
		blob.QueryPasswordIntervalTicks = int64(binary.LittleEndian.Uint64(data[queryIntervalOff:]))
	}

	if unchangedIntervalOff != 0 {
		if int(unchangedIntervalOff)+8 > len(data) {
			return nil, fmt.Errorf("unchanged password interval at offset %d out of bounds", unchangedIntervalOff)
		}
		blob.UnchangedPasswordIntervalTicks = int64(binary.LittleEndian.Uint64(data[unchangedIntervalOff:]))
	}

	return blob, nil
}

// MarshalManagedPasswordBlob encodes an MSDS-MANAGEDPASSWORD_BLOB.
func MarshalManagedPasswordBlob(blob *ManagedPasswordBlob) []byte {
	currentUTF16 := encodeUTF16NullTerm(blob.CurrentPassword)
	var previousUTF16 []byte
	if blob.PreviousPassword != "" {
		previousUTF16 = encodeUTF16NullTerm(blob.PreviousPassword)
	}

	// Layout: header(16) | currentPassword | previousPassword | padding | queryInterval(8) | unchangedInterval(8)
	currentOff := managedPasswordHeaderSize
	previousOff := currentOff + len(currentUTF16)
	afterPasswords := previousOff + len(previousUTF16)

	// Align to 8-byte boundary for the LARGE_INTEGER fields.
	padding := (8 - afterPasswords%8) % 8
	queryIntervalOff := afterPasswords + padding
	unchangedIntervalOff := queryIntervalOff + 8
	totalLen := unchangedIntervalOff + 8

	buf := make([]byte, totalLen)
	binary.LittleEndian.PutUint16(buf[0:2], 1) // Version
	// buf[2:4] reserved = 0
	binary.LittleEndian.PutUint32(buf[4:8], uint32(totalLen))
	binary.LittleEndian.PutUint16(buf[8:10], uint16(currentOff))
	if len(previousUTF16) > 0 {
		binary.LittleEndian.PutUint16(buf[10:12], uint16(previousOff))
	}
	binary.LittleEndian.PutUint16(buf[12:14], uint16(queryIntervalOff))
	binary.LittleEndian.PutUint16(buf[14:16], uint16(unchangedIntervalOff))

	copy(buf[currentOff:], currentUTF16)
	if len(previousUTF16) > 0 {
		copy(buf[previousOff:], previousUTF16)
	}
	binary.LittleEndian.PutUint64(buf[queryIntervalOff:], uint64(blob.QueryPasswordIntervalTicks))
	binary.LittleEndian.PutUint64(buf[unchangedIntervalOff:], uint64(blob.UnchangedPasswordIntervalTicks))

	return buf
}

// readUTF16NullTerm decodes a null-terminated UTF-16LE string starting at off in data.
func readUTF16NullTerm(data []byte, off int) (string, error) {
	if off < 0 || off >= len(data) {
		return "", fmt.Errorf("offset %d out of bounds (len %d)", off, len(data))
	}
	var units []uint16
	for i := off; i+1 < len(data); i += 2 {
		u := binary.LittleEndian.Uint16(data[i:])
		if u == 0 {
			break
		}
		units = append(units, u)
	}
	return string(utf16.Decode(units)), nil
}

// encodeUTF16NullTerm encodes s as null-terminated UTF-16LE bytes.
func encodeUTF16NullTerm(s string) []byte {
	runes := []rune(s)
	units := utf16.Encode(runes)
	buf := make([]byte, (len(units)+1)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(buf[2*i:], u)
	}
	// trailing null already zero from make
	return buf
}
