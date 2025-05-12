package version

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// Version 15 of the NTLMSSP is in use.
	NTLMSSP_REVISION_W2K3 = 15
)

// The VERSION structure contains operating system version information that SHOULD be ignored. This structure is used for debugging purposes only and its value does not affect NTLM message processing. It is populated in the NEGOTIATE_MESSAGE, CHALLENGE_MESSAGE, and AUTHENTICATE_MESSAGE messages only if NTLMSSP_NEGOTIATE_VERSION is negotiated; otherwise, it MUST be set to all zero.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/b1a6ceb2-f8ad-462b-b5af-f18527c48175
type Version struct {
	// ProductMajorVersion (1 byte): An 8-bit unsigned integer that SHOULD
	// contain the major version number of the operating system in use.
	ProductMajorVersion byte

	// ProductMinorVersion (1 byte): An 8-bit unsigned integer that SHOULD
	// contain the minor version number of the operating system in use.
	ProductMinorVersion byte

	// ProductBuild (2 bytes): A 16-bit unsigned integer that contains the
	// build number of the operating system in use. This field SHOULD be set
	// to a 16-bit quantity that identifies the operating system build number.
	ProductBuild uint16

	// Reserved (3 bytes): A 24-bit data area that SHOULD be set to zero and
	// MUST be ignored by the recipient.
	Reserved [3]byte

	// NTLMRevisionCurrent (1 byte): An 8-bit unsigned integer that contains
	// a value indicating the current revision of the NTLMSSP in use.
	// This field SHOULD contain the following value:
	// NTLMSSP_REVISION_W2K3 = 15
	NTLMRevision byte
}

// DefaultVersion returns the default NTLM version (Windows 10.0.18362)
//
// Returns: A Version struct representing the default NTLM version (Windows 10)
func DefaultVersion() Version {
	return Version{
		ProductMajorVersion: 10,
		ProductMinorVersion: 0,
		ProductBuild:        18362,
		NTLMRevision:        NTLMSSP_REVISION_W2K3,
	}
}

// NewVersion creates a new Version with the specified parameters
//
// Parameters:
//   - major: The major version number
//   - minor: The minor version number
//   - build: The build number
//   - ntlmRevision: The NTLM revision number
func NewVersion(major, minor byte, build uint16, ntlmRevision byte) Version {
	return Version{
		ProductMajorVersion: major,
		ProductMinorVersion: minor,
		ProductBuild:        build,
		Reserved:            [3]byte{0, 0, 0},
		NTLMRevision:        ntlmRevision,
	}
}

// String returns a string representation of the Version
// Returns: "Version <major>.<minor> (Build <build>) NTLM Revision <ntlmRevision>"
func (v Version) String() string {
	return fmt.Sprintf("Version %d.%d (Build %d) NTLM Revision %d",
		v.ProductMajorVersion, v.ProductMinorVersion, v.ProductBuild, v.NTLMRevision)
}

// Marshal serializes the Version structure to a byte slice
// Returns: A byte slice containing the serialized Version and an error if the data is too short
func (v Version) Marshal() ([]byte, error) {
	data := make([]byte, 8)

	data[0] = v.ProductMajorVersion

	data[1] = v.ProductMinorVersion

	binary.LittleEndian.PutUint16(data[2:4], v.ProductBuild)

	copy(data[4:7], v.Reserved[:])

	data[7] = v.NTLMRevision

	return data, nil
}

// Unmarshal deserializes a byte slice into a Version structure
// Returns: The number of bytes read and an error if the data is too short
func (v *Version) Unmarshal(data []byte) (int, error) {
	if len(data) < 8 {
		return 0, errors.New("version data too short")
	}

	v.ProductMajorVersion = data[0]

	v.ProductMinorVersion = data[1]

	v.ProductBuild = binary.LittleEndian.Uint16(data[2:4])

	copy(v.Reserved[:], data[4:7])

	v.NTLMRevision = data[7]

	return 8, nil
}
