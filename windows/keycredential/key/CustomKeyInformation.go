package key

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

// CustomKeyInformation represents the CUSTOM_KEY_INFORMATION structure.
//
// See: https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/701a55dc-d062-4032-a2da-dbdfc384c8cf
type CustomKeyInformation struct {
	Version                 int
	Flags                   CustomKeyInformationFlags
	CurrentVersion          int
	ShortRepresentationSize int
	ReservedSize            int
	VolumeType              CustomKeyInformationVolumeType
	SupportsNotification    bool
	FekKeyVersion           uint8
	Strength                KeyStrength
	Reserved                []byte
	EncodedExtendedCKI      []byte

	// Internal
	RawBytes     []byte
	RawBytesSize uint32
}

// Parse parses the provided byte slice into the CustomKeyInformation structure.
//
// Parameters:
// - blob: A byte slice containing the raw custom key information to be parsed.
//
// Returns:
// - An error if the parsing fails, otherwise nil.
//
// Note:
// The function expects the byte slice to follow the CUSTOM_KEY_INFORMATION structure format.
// It extracts the version, flags, volume type, supports notification, FEK key version, strength, reserved, and encoded extended CKI fields from the byte slice.
// The parsed values are stored in the CustomKeyInformation structure.
func (cki *CustomKeyInformation) FromBytes(blob []byte, version KeyCredentialVersion) error {
	cki.RawBytes = blob
	cki.RawBytesSize = uint32(len(blob))

	if len(blob) < 2 {
		return fmt.Errorf("invalid blob size: %d", len(blob))
	}

	// An 8-bit unsigned integer that must be set to 1:
	cki.Version = int(blob[0])
	if cki.Version != 1 {
		return fmt.Errorf("invalid CustomKeyInformation version: %d", cki.Version)
	}

	// An 8-bit unsigned integer that specifies zero or more bit-flag values.
	cki.Flags.FromBytes(blob[1])

	// An 8-bit unsigned integer that specifies one of the volume types.
	if 2 < cki.RawBytesSize && cki.RawBytesSize >= 3 {
		cki.VolumeType.FromBytes(blob[2])
	} else {
		return nil
	}

	// An 8-bit unsigned integer that specifies whether the device associated with this credential supports notification.
	if 3 < cki.RawBytesSize && cki.RawBytesSize >= 4 {
		cki.SupportsNotification = (blob[3] != 0)
	} else {
		return nil
	}

	// An 8-bit unsigned integer that specifies the version of the File Encryption Key (FEK). This field must be set to 1.
	if 4 < cki.RawBytesSize && cki.RawBytesSize >= 5 {
		cki.FekKeyVersion = blob[4]
	} else {
		return nil
	}

	// An 32-bit unsigned integer that specifies the strength of the NGC key.
	if 5 < cki.RawBytesSize && cki.RawBytesSize >= 9 {
		cki.Strength.FromBytes(blob[5:9])
	} else {
		return nil
	}

	// 10 bytes reserved for future use.
	// Note: With FIDO, Azure incorrectly puts here 9 bytes instead of 10.
	if 9 < cki.RawBytesSize && cki.RawBytesSize >= 19 {
		cki.Reserved = make([]byte, 10)
		copy(cki.Reserved, blob[9:19])
	} else {
		return nil
	}

	// Extended custom key information.
	if 19 < cki.RawBytesSize {
		cki.EncodedExtendedCKI = make([]byte, cki.RawBytesSize-19)
		copy(cki.EncodedExtendedCKI, blob[19:])
	} else {
		return nil
	}

	return nil
}

// ToBytes returns the raw bytes of the CustomKeyInformation structure.
func (cki *CustomKeyInformation) ToBytes() []byte {
	blob := make([]byte, 0)

	blob = append(blob, byte(cki.Version))

	if 2 < cki.RawBytesSize && cki.RawBytesSize >= 3 {
		blob = append(blob, byte(cki.Flags.Value))
	}

	if 3 < cki.RawBytesSize && cki.RawBytesSize >= 4 {
		blob = append(blob, byte(cki.VolumeType.Value))
	}

	if 4 < cki.RawBytesSize && cki.RawBytesSize >= 5 {
		if cki.SupportsNotification {
			blob = append(blob, 1)
		} else {
			blob = append(blob, 0)
		}
	}

	if 5 < cki.RawBytesSize && cki.RawBytesSize >= 6 {
		blob = append(blob, byte(cki.FekKeyVersion))
	}

	if 6 < cki.RawBytesSize && cki.RawBytesSize >= 10 {
		buffer := make([]byte, 4)
		binary.LittleEndian.PutUint32(buffer, cki.Strength.Value)
		blob = append(blob, buffer...)
	}

	if 10 < cki.RawBytesSize && cki.RawBytesSize >= 20 {
		blob = append(blob, cki.Reserved...)
	}

	if 20 < cki.RawBytesSize {
		blob = append(blob, cki.EncodedExtendedCKI...)
	}

	return blob
}

// Describe prints a detailed description of the CustomKeyInformation instance.
//
// Parameters:
// - indent: An integer representing the indentation level for the printed output.
//
// Note:
// This function prints the Flags, VolumeType, SupportsNotification, FekKeyVersion, Strength, Reserved, and EncodedExtendedCKI values of the CustomKeyInformation instance.
// The output is formatted with the specified indentation level to improve readability.
func (cki *CustomKeyInformation) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<\x1b[93mCustomKeyInformation\x1b[0m>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mVersion\x1b[0m: %d\n", indentPrompt, cki.Version)
	fmt.Printf("%s │ \x1b[93mFlags\x1b[0m: [%s] (%d)\n", indentPrompt, strings.Join(cki.Flags.Name, ", "), cki.Flags.Value)
	if 2 < cki.RawBytesSize && cki.RawBytesSize >= 3 {
		fmt.Printf("%s │ \x1b[93mVolumeType\x1b[0m: %s (%d)\n", indentPrompt, cki.VolumeType.String(), cki.VolumeType.Value)
	} else {
		fmt.Printf("%s │ \x1b[93mVolumeType\x1b[0m: None\n", indentPrompt)
	}
	if 3 < cki.RawBytesSize && cki.RawBytesSize >= 4 {
		fmt.Printf("%s │ \x1b[93mSupportsNotification\x1b[0m: %t\n", indentPrompt, cki.SupportsNotification)
	} else {
		fmt.Printf("%s │ \x1b[93mSupportsNotification\x1b[0m: None\n", indentPrompt)
	}
	if 4 < cki.RawBytesSize && cki.RawBytesSize >= 5 {
		fmt.Printf("%s │ \x1b[93mFekKeyVersion\x1b[0m: %d\n", indentPrompt, cki.FekKeyVersion)
	} else {
		fmt.Printf("%s │ \x1b[93mFekKeyVersion\x1b[0m: None\n", indentPrompt)
	}
	if 5 < cki.RawBytesSize && cki.RawBytesSize >= 9 {
		fmt.Printf("%s │ \x1b[93mStrength\x1b[0m: %s (0x%x)\n", indentPrompt, cki.Strength.Name, cki.Strength.Value)
	} else {
		fmt.Printf("%s │ \x1b[93mStrength\x1b[0m: None\n", indentPrompt)
	}
	if 6 < cki.RawBytesSize && cki.RawBytesSize >= 16 {
		fmt.Printf("%s │ \x1b[93mReserved\x1b[0m: %s\n", indentPrompt, hex.EncodeToString(cki.Reserved))
	} else {
		fmt.Printf("%s │ \x1b[93mReserved\x1b[0m: None\n", indentPrompt)
	}
	if 16 < cki.RawBytesSize {
		fmt.Printf("%s │ \x1b[93mEncodedExtendedCKI\x1b[0m: %s\n", indentPrompt, hex.EncodeToString(cki.EncodedExtendedCKI))
	} else {
		fmt.Printf("%s │ \x1b[93mEncodedExtendedCKI\x1b[0m: None\n", indentPrompt)
	}
	fmt.Printf("%s └───\n", indentPrompt)
}
