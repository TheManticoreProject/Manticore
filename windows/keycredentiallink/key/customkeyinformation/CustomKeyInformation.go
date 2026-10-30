package customkeyinformation

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/strength"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"
)

// CustomKeyInformation represents the CUSTOM_KEY_INFORMATION structure.
//
// Note: This structure has two possible representations. In the first representation,
// only the Version and Flags fields are present; in this case the structure has a total
// size of two bytes. In the second representation, all additional fields shown below
// are also present; in this case, the structure's total size is variable.
// Differentiating between the two representations MUST be inferred using only
// the total size.
//
// See: https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/701a55dc-d062-4032-a2da-dbdfc384c8cf
type CustomKeyInformation struct {
	// Version (1 byte): An 8-bit unsigned integer that MUST be set to 1.
	Version uint8

	// Flags (1 byte): An 8-bit unsigned integer that specifies zero or more bit-flag values
	Flags *CUSTOMKEYINFO_FLAGS

	// VolType (1 byte): An 8-bit unsigned integer that specifies one of the following volume types:
	VolumeType *CustomKeyInformationVolumeType

	// SupportsNotification (1 byte): An 8-bit unsigned integer that specifies whether the device
	// associated with this credential supports notification.
	SupportsNotification *SupportsNotification

	// FekKeyVersion (1 byte): An 8-bit unsigned integer that specifies the version of the buffer
	// stored in KEY_USAGE_FEK (section 2.2.20.5.3). This field MUST be set to 1.
	FekKeyVersion *FekKeyVersion

	// KeyStrength (1 byte): An 8-bit unsigned integer that specifies the strength of the NGC key.
	KeyStrength *strength.KeyStrength

	// Reserved (10 bytes): Reserved for future use.
	Reserved *Reserved

	// EncodedExtendedCKI (variable): Extended custom key information.
	// The contents of this field are defined in section 2.2.20.4.1.
	// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/b2c0cb9b-e49e-4907-9235-f9fd7eee8c13
	EncodedExtendedCKI *EncodedExtendedCKI
}

// Unmarshal parses the provided byte slice into the CustomKeyInformation structure.
//
// Parameters:
// - data: A byte slice containing the raw custom key information to be parsed.
// - kcv: The version of the key credential link.
//
// Returns:
// - The number of bytes read from the data.
// - An error if the parsing fails, otherwise nil.
//
// Note:
// The function expects the byte slice to follow the CUSTOM_KEY_INFORMATION structure format.
// It extracts the version, flags, volume type, supports notification, FEK key version, strength, reserved, and encoded extended CKI fields from the byte slice.
// The parsed values are stored in the CustomKeyInformation structure.
func (cki *CustomKeyInformation) Unmarshal(data []byte, kcv version.KeyCredentialLinkVersion) (int, error) {
	if len(data) < 1 {
		return 0, fmt.Errorf("invalid data size: %d", len(data))
	}

	cki.Flags = nil
	cki.VolumeType = nil
	cki.SupportsNotification = nil
	cki.FekKeyVersion = nil
	cki.KeyStrength = nil
	cki.Reserved = nil
	cki.EncodedExtendedCKI = nil

	bytesRead := 0

	// An 8-bit unsigned integer that must be set to 1:
	cki.Version = data[bytesRead]
	if cki.Version != 1 {
		return 0, fmt.Errorf("invalid CustomKeyInformation version: %d", cki.Version)
	}
	bytesRead += 1

	// An 8-bit unsigned integer that specifies zero or more bit-flag values.
	if len(data[bytesRead:]) >= 1 {
		cki.Flags = new(CUSTOMKEYINFO_FLAGS)
		bytesReadFlags, err := cki.Flags.Unmarshal(data[bytesRead:])
		if err != nil {
			return 0, err
		}
		bytesRead += bytesReadFlags
	} else {
		return bytesRead, nil
	}

	// An 8-bit unsigned integer that specifies one of the volume types.
	if len(data[bytesRead:]) >= 1 {
		cki.VolumeType = new(CustomKeyInformationVolumeType)
		bytesReadVolumeType, err := cki.VolumeType.Unmarshal(data[bytesRead:])
		if err != nil {
			return 0, err
		}
		bytesRead += bytesReadVolumeType
	} else {
		return bytesRead, nil
	}

	// An 8-bit unsigned integer that specifies whether the device associated with this credential supports notification.
	if len(data[bytesRead:]) >= 1 {
		cki.SupportsNotification = new(SupportsNotification)
		bytesReadSupportsNotification, err := cki.SupportsNotification.Unmarshal(data[bytesRead:])
		if err != nil {
			return 0, err
		}
		bytesRead += bytesReadSupportsNotification
	} else {
		return bytesRead, nil
	}

	// An 8-bit unsigned integer that specifies the version of the File Encryption Key (FEK). This field must be set to 1.
	if len(data[bytesRead:]) >= 1 {
		cki.FekKeyVersion = new(FekKeyVersion)
		bytesReadFekKeyVersion, err := cki.FekKeyVersion.Unmarshal(data[bytesRead:])
		if err != nil {
			return 0, err
		}
		bytesRead += bytesReadFekKeyVersion
	} else {
		return bytesRead, nil
	}

	// An 32-bit unsigned integer that specifies the strength of the NGC key.
	if len(data[bytesRead:]) >= 4 {
		cki.KeyStrength = new(strength.KeyStrength)
		bytesReadKeyStrength, err := cki.KeyStrength.Unmarshal(data[bytesRead:])
		if err != nil {
			return 0, err
		}
		bytesRead += bytesReadKeyStrength
	} else {
		return bytesRead, nil
	}

	// Reserved (10 bytes) precedes the extended custom key information.
	if len(data[bytesRead:]) >= 10 {
		cki.Reserved = new(Reserved)
		bytesReadReserved, err := cki.Reserved.Unmarshal(data[bytesRead:])
		if err != nil {
			return 0, err
		}
		bytesRead += bytesReadReserved
	} else {
		return bytesRead, nil
	}

	// Extended custom key information.
	if len(data[bytesRead:]) >= 2 {
		cki.EncodedExtendedCKI = new(EncodedExtendedCKI)
		bytesReadEncodedExtendedCKI, err := cki.EncodedExtendedCKI.Unmarshal(data[bytesRead:])
		if err != nil {
			return 0, err
		}
		bytesRead += bytesReadEncodedExtendedCKI
	} else {
		return bytesRead, nil
	}

	return bytesRead, nil
}

// Marshal returns the raw bytes of the CustomKeyInformation structure.
//
// Returns:
// - A byte slice representing the raw bytes of the CustomKeyInformation structure.
// - An error if the conversion fails.
func (cki *CustomKeyInformation) Marshal() ([]byte, error) {
	data := make([]byte, 0)

	// Version: An 8-bit unsigned integer that must be set to 1.
	data = append(data, byte(cki.Version))

	// Flags: An 8-bit unsigned integer that specifies zero or more bit-flag values.
	if cki.Flags != nil {
		flagsBytes, err := cki.Flags.Marshal()
		if err != nil {
			return nil, err
		}
		data = append(data, flagsBytes...)
	} else {
		return data, nil
	}

	// VolumeType: An 8-bit unsigned integer that specifies one of the following volume types.
	if cki.VolumeType != nil {
		volumeTypeBytes, err := cki.VolumeType.Marshal()
		if err != nil {
			return nil, err
		}
		data = append(data, volumeTypeBytes...)
	} else {
		return data, nil
	}

	// SupportsNotification: An 8-bit unsigned integer that specifies whether the device
	// associated with this credential supports notification.
	if cki.SupportsNotification != nil {
		supportsNotificationBytes, err := cki.SupportsNotification.Marshal()
		if err != nil {
			return nil, err
		}
		data = append(data, supportsNotificationBytes...)
	} else {
		return data, nil
	}

	// FekKeyVersion: An 8-bit unsigned integer that specifies the version of the buffer
	// stored in KEY_USAGE_FEK (section 2.2.20.5.3). This field MUST be set to 1.
	if cki.FekKeyVersion != nil {
		fekKeyVersionBytes, err := cki.FekKeyVersion.Marshal()
		if err != nil {
			return nil, err
		}
		data = append(data, fekKeyVersionBytes...)
	} else {
		return data, nil
	}

	// KeyStrength: An 8-bit unsigned integer that specifies the strength of the NGC key.
	if cki.KeyStrength != nil {
		strengthBytes, err := cki.KeyStrength.Marshal()
		if err != nil {
			return nil, err
		}
		data = append(data, strengthBytes...)
	} else {
		return data, nil
	}

	// Reserved: 10 bytes reserved for future use.
	if cki.Reserved != nil {
		reservedBytes, err := cki.Reserved.Marshal()
		if err != nil {
			return nil, err
		}
		data = append(data, reservedBytes...)
	} else {
		return data, nil
	}

	// EncodedExtendedCKI: Extended custom key information.
	if cki.EncodedExtendedCKI != nil {
		encodedExtendedCKIBytes, err := cki.EncodedExtendedCKI.Marshal()
		if err != nil {
			return nil, err
		}
		data = append(data, encodedExtendedCKIBytes...)
	} else {
		return data, nil
	}

	return data, nil
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

	if cki.Flags != nil {
		fmt.Printf("%s │ \x1b[93mFlags\x1b[0m: [%s] (0x%02x)\n", indentPrompt, cki.Flags.String(), uint8(*cki.Flags))
	} else {
		// fmt.Printf("%s │ \x1b[93mFlags\x1b[0m: None\n", indentPrompt)
	}

	if cki.VolumeType != nil {
		fmt.Printf("%s │ \x1b[93mVolumeType\x1b[0m: %s (0x%02x)\n", indentPrompt, cki.VolumeType.String(), cki.VolumeType)
	} else {
		// fmt.Printf("%s │ \x1b[93mVolumeType\x1b[0m: None\n", indentPrompt)
	}

	if cki.SupportsNotification != nil {
		fmt.Printf("%s │ \x1b[93mSupportsNotification\x1b[0m: %s\n", indentPrompt, cki.SupportsNotification.String())
	} else {
		// fmt.Printf("%s │ \x1b[93mSupportsNotification\x1b[0m: None\n", indentPrompt)
	}

	if cki.FekKeyVersion != nil {
		fmt.Printf("%s │ \x1b[93mFekKeyVersion\x1b[0m: %d\n", indentPrompt, cki.FekKeyVersion)
	} else {
		// fmt.Printf("%s │ \x1b[93mFekKeyVersion\x1b[0m: None\n", indentPrompt)
	}

	if cki.KeyStrength != nil {
		fmt.Printf("%s │ \x1b[93mStrength\x1b[0m: %s (0x%02x)\n", indentPrompt, cki.KeyStrength.String(), cki.KeyStrength)
	} else {
		// fmt.Printf("%s │ \x1b[93mStrength\x1b[0m: None\n", indentPrompt)
	}

	if cki.Reserved != nil {
		fmt.Printf("%s │ \x1b[93mReserved\x1b[0m: %s\n", indentPrompt, hex.EncodeToString(cki.Reserved[:]))
	} else {
		// fmt.Printf("%s │ \x1b[93mReserved\x1b[0m: None\n", indentPrompt)
	}

	if cki.EncodedExtendedCKI != nil {
		cki.EncodedExtendedCKI.Describe(indent + 1)
	} else {
		// fmt.Printf("%s │ \x1b[93mEncodedExtendedCKI\x1b[0m: None\n", indentPrompt)
	}

	fmt.Printf("%s └───\n", indentPrompt)
}
