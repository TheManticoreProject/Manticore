package keycredentiallink

import (
	"github.com/TheManticoreProject/Manticore/network/ldap"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/customkeyinformation"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/source"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/usage"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/utils"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"

	"encoding/hex"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// KeyCredentialLink represents a key credential structure used for authentication and authorization.
//
// Fields:
// - Version: A KeyCredentialLinkVersion object representing the version of the key credential.
// - Identifier: A string representing the unique identifier of the key credential.
// - KeyHash: A byte slice containing the hash of the key material.
// - KeyMaterial: A KeyMaterial object representing the key material of the key credential.
// - Usage: A KeyUsage object representing the usage of the key credential.
// - LegacyUsage: A string representing the legacy usage of the key credential.
// - Source: A KeySource object representing the source of the key credential.
// - LastLogonTime: A DateTime object representing the last logon time associated with the key credential.
// - CreationTime: A DateTime object representing the creation time of the key credential.
//
// Methods:
// - ParseDNWithBinary: Parses the provided DNWithBinary object into the KeyCredentialLink structure.
//
// Note:
// The KeyCredentialLink structure is used to store and manage key credentials, which are used for authentication and authorization purposes.
// The structure includes fields for version, identifier, key hash, raw key material, usage, legacy usage, source, last logon time, creation time, owner, and raw binary data.
// The ParseDNWithBinary method is used to parse a DNWithBinary object and populate the fields of the KeyCredentialLink structure.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/f3f01e95-6d0c-4fe6-8b43-d585167658fa
type KeyCredentialLink struct {
	// A KeyCredentialLinkVersion object representing the version of the key credential.
	// This field is MANDATORY.
	Version version.KeyCredentialLinkVersion
	// A string representing the unique identifier of the key credential.
	// This field is MANDATORY.
	Identifier string
	// A byte slice containing the hash of the key material.
	// This field is OPTIONAL.
	KeyHash []byte
	// A KeyMaterial object representing the key material of the key credential.
	// This field is MANDATORY.
	KeyMaterial bcrypt.KeyMaterial
	// A KeyUsage object representing the usage of the key credential.
	// This field is MANDATORY.
	Usage usage.KeyUsage
	// A string representing the legacy usage of the key credential.
	// This field is OPTIONAL.
	LegacyUsage string
	// A KeySource object representing the source of the key credential.
	// This field is OPTIONAL.
	Source *source.KeySource
	// A CustomKeyInformation object representing the custom key information of the key credential.
	// This field is OPTIONAL.
	CustomKeyInfo *customkeyinformation.CustomKeyInformation
	// A GUID object representing the device ID of the key credential.
	// This field is OPTIONAL.
	DeviceId *guid.GUID
	// A DateTime object representing the last logon time associated with the key credential.
	// This field is OPTIONAL.
	LastLogonTime *utils.DateTime
	// A DateTime object representing the creation time of the key credential.
	// This field is OPTIONAL.
	CreationTime *utils.DateTime
}

// NewKeyCredentialLink creates a new KeyCredentialLink structure.
//
// Parameters:
//
// - version: A KeyCredentialLinkVersion object representing the version of the key credential.
//
// - Identifier: A string representing the unique identifier of the key credential.
//
// - KeyHash: A byte slice containing the hash of the key material.
//
// - KeyMaterial: An RSAKeyMaterial object representing the raw RSA key material.
//
// - Usage: A KeyUsage object representing the usage of the key credential.
//
// - LegacyUsage: A string representing the legacy usage of the key credential.
//
// - Source: A KeySource object representing the source of the key credential.
//
// - CustomKeyInfo: A CustomKeyInformation object representing the custom key information of the key credential.
//
// - DeviceId: A GUID object representing the device ID of the key credential.
//
// - LastLogonTime: A DateTime object representing the last logon time associated with the key credential.
//
// - CreationTime: A DateTime object representing the creation time of the key credential.
//
// Returns:
//
// - A pointer to a KeyCredentialLink object.
func NewKeyCredentialLink(
	version version.KeyCredentialLinkVersion,
	identifier string,
	keyMaterial bcrypt.KeyMaterial,
	deviceId *guid.GUID,
	lastLogonTime *utils.DateTime,
	creationTime *utils.DateTime,
) *KeyCredentialLink {
	kc := &KeyCredentialLink{
		Version:     version,
		Identifier:  identifier,
		KeyHash:     []byte{},
		KeyMaterial: keyMaterial,
		Usage:       usage.KeyUsage{Value: usage.KeyUsage_NGC},
		LegacyUsage: "",
		Source:      &source.KeySource{},
		CustomKeyInfo: &customkeyinformation.CustomKeyInformation{
			Version: 1,
			Flags:   customkeyinformation.NewCUSTOMKEYINFO_FLAGS(customkeyinformation.CUSTOMKEYINFO_FLAGS_NONE),
		},
		DeviceId:      deviceId,
		LastLogonTime: lastLogonTime,
		CreationTime:  creationTime,
	}

	kc.Source.Value = source.KeySource_AD

	kc.KeyHash = kc.ComputeKeyHash()

	return kc
}

// ParseDNWithBinary parses the provided DNWithBinary object into the KeyCredentialLink structure.
//
// Parameters:
// - dnWithBinary: A DNWithBinary object containing the distinguished name and binary data to be parsed.
//
// Returns:
// - An error if the parsing fails, otherwise nil.
//
// Note:
// The function performs the following steps:
// 1. Sets the RawBytes and RawBytesSize fields to the provided binary data and its length, respectively.
// 2. Sets the Owner field to the distinguished name from the DNWithBinary object.
// 3. Parses the version information from the binary data and updates the RawBytesSize and remainder accordingly.
// 4. Iterates through the remaining binary data, parsing each entry based on its type and length.
// 5. Updates the corresponding fields of the KeyCredentialLink structure based on the parsed entry type and data.
//
// The function handles various entry types, including key identifier, key hash, key material, key usage, legacy usage, key source, last logon time, and creation time.
// Unsupported entry types, such as device ID and custom key information, are commented out for future implementation.
func (kc *KeyCredentialLink) ParseDNWithBinary(dnWithBinary ldap.DNWithBinary) error {
	_, err := kc.Unmarshal(dnWithBinary.BinaryData)
	if err != nil {
		return err
	}
	return nil
}

// Unmarshal parses the provided binary data into the KeyCredentialLink structure.
//
// Parameters:
// - data: A byte slice containing the binary data to be parsed.
//
// Returns:
// - bytesRead: The number of bytes read from the data.
// - error: An error if the parsing fails, otherwise nil.
//
// Note:
// The function performs the following steps:
// 1. Sets the RawBytes and RawBytesSize fields to the provided binary data and its length.
// 2. Parses the version information from the binary data.
// 3. Iterates through the remaining binary data, parsing each entry based on its type and length.
// 4. Updates the corresponding fields of the KeyCredentialLink structure based on the parsed entry type and data.
//
// The function handles various entry types, including:
// - Key identifier
// - Key hash
// - Key material
// - Key usage (both V2 enum and legacy string formats)
// - Key source
// - Device ID
// - Custom key information
// - Last logon timestamp
// - Creation time
func (kc *KeyCredentialLink) Unmarshal(data []byte) (int, error) {
	bytesRead := 0

	blob := KEYCREDENTIALLINK_BLOB{}
	bytesRead, err := blob.Unmarshal(data)
	if err != nil {
		return bytesRead, err
	}

	// The version governs how the entries below are interpreted (identifier form,
	// timestamp encoding), so it has to be carried over from the parsed blob before
	// the entries are processed, not left at its zero value.
	kc.Version = blob.Version

	for _, entry := range blob.Entries {
		// Process the entry data based on its type.
		switch entry.Identifier {

		case KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyID:
			kc.Identifier = utils.ConvertFromBinaryIdentifier(entry.Value, kc.Version)

		case KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyHash:
			kc.KeyHash = entry.Value

		case KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyMaterial:
			kc.KeyMaterial, bytesRead, err = bcrypt.UnmarshalKeyMaterial(entry.Value)
			if err != nil {
				return bytesRead, fmt.Errorf("failed to unmarshal KeyCredentialLink key material: %w", err)
			}

		case KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyUsage:
			if len(entry.Value) == 1 {
				// This is apparently a V2 structure (single byte enum).
				_, err := kc.Usage.Unmarshal(entry.Value)
				if err != nil {
					return bytesRead, fmt.Errorf("failed to unmarshal KeyCredentialLink usage: %w", err)
				}
			} else {
				// This is a legacy structure that contains a string-encoded key usage.
				kc.LegacyUsage = string(entry.Value)
			}

		case KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeySource:
			if kc.Source == nil {
				kc.Source = &source.KeySource{}
			}
			_, err := kc.Source.Unmarshal(entry.Value)
			if err != nil {
				return bytesRead, fmt.Errorf("failed to unmarshal KeyCredentialLink source: %w", err)
			}

		case KEYCREDENTIALLINK_ENTRY_IDENTIFIER_DeviceId:
			if kc.DeviceId == nil {
				kc.DeviceId = &guid.GUID{}
			}
			kc.DeviceId.FromRawBytes(entry.Value)

		case KEYCREDENTIALLINK_ENTRY_IDENTIFIER_CustomKeyInformation:
			if kc.CustomKeyInfo == nil {
				kc.CustomKeyInfo = &customkeyinformation.CustomKeyInformation{}
			}
			_, err := kc.CustomKeyInfo.Unmarshal(entry.Value, kc.Version)
			if err != nil {
				return bytesRead, fmt.Errorf("failed to unmarshal KeyCredentialLink custom key information: %w", err)
			}

		case KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyApproximateLastLogonTimeStamp:
			t := utils.ConvertFromBinaryTime(entry.Value, *kc.Source, kc.Version)
			kc.LastLogonTime = &t

		case KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyCreationTime:
			t := utils.ConvertFromBinaryTime(entry.Value, *kc.Source, kc.Version)
			kc.CreationTime = &t
		}
	}

	return bytesRead, nil
}

// CheckIntegrity checks the integrity of the key credential.
//
// Returns:
// - A boolean value indicating the integrity of the key credential.
func (kc *KeyCredentialLink) CheckIntegrity() bool {
	hash := kc.ComputeKeyHash()

	if len(hash) != len(kc.KeyHash) {
		return false
	}

	for i := range hash {
		if hash[i] != kc.KeyHash[i] {

			fmt.Printf("computedhash : %s\n", hex.EncodeToString(hash))
			fmt.Printf("kc.KeyHash   : %s\n", hex.EncodeToString(kc.KeyHash))

			return false
		}
	}

	return true
}

// ComputeKeyHash computes the key hash of the key credential.
//
// Returns:
// - A byte slice containing the key hash.
func (kc *KeyCredentialLink) ComputeKeyHash() []byte {
	// Src: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/a99409ea-4f72-b7ef-8596013a36c7
	blob, err := kc.ToKeyCredentialLinkBlob()
	if err != nil {
		return nil
	}
	// Build the hash input from entries AFTER KeyHash:
	// KeyMaterial, KeyUsage, KeySource, DeviceId, CustomKeyInformation, LastLogonTime, CreationTime.
	binaryProperties := make([]byte, 0)
	for _, entry := range blob.Entries {
		if entry.Identifier == KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyMaterial ||
			entry.Identifier == KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyUsage ||
			entry.Identifier == KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeySource ||
			entry.Identifier == KEYCREDENTIALLINK_ENTRY_IDENTIFIER_DeviceId ||
			entry.Identifier == KEYCREDENTIALLINK_ENTRY_IDENTIFIER_CustomKeyInformation ||
			entry.Identifier == KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyApproximateLastLogonTimeStamp ||
			entry.Identifier == KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyCreationTime {
			entryBytes, err := entry.Marshal()
			if err != nil {
				return nil
			}
			binaryProperties = append(binaryProperties, entryBytes...)
		} // else: Skip Version/KeyID/KeyHash
	}
	return utils.ComputeHash(binaryProperties)
}

// Marshal returns the raw bytes of the KeyCredentialLink structure.
//
// Returns:
// - A byte slice representing the raw bytes of the KeyCredentialLink structure.
// - An error if the conversion fails.
func (kc *KeyCredentialLink) Marshal() ([]byte, error) {
	blob, err := kc.ToKeyCredentialLinkBlob()
	if err != nil {
		return nil, err
	}
	return blob.Marshal()
}

// ToKeyCredentialLinkBlob converts the KeyCredentialLink structure to a KEYCREDENTIALLINK_BLOB structure.
//
// Returns:
// - A pointer to a KEYCREDENTIALLINK_BLOB structure.
// - An error if the conversion fails.
//
// Note:
// The function converts the KeyCredentialLink structure to a KEYCREDENTIALLINK_BLOB structure.
// The function sorts the entries by their Identifier fields in increasing order.
// The function returns the KEYCREDENTIALLINK_BLOB structure.
func (kc *KeyCredentialLink) ToKeyCredentialLinkBlob() (*KEYCREDENTIALLINK_BLOB, error) {
	blob := KEYCREDENTIALLINK_BLOB{}
	blob.Version = kc.Version
	blob.Entries = make([]KEYCREDENTIALLINK_ENTRY, 0)

	// The KEYCREDENTIALLINK_ENTRY structure MUST be sorted by their Identifier fields in increasing order.
	// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/f3f01e95-6d0c-4fe6-8b43-d585167658fa

	// Key Identifier (entry type 0x01) [MANDATORY]
	identifierBytes, err := utils.ConvertToBinaryIdentifier(kc.Identifier, kc.Version)
	if err != nil {
		return nil, err
	}
	blob.Entries = append(
		blob.Entries,
		KEYCREDENTIALLINK_ENTRY{
			Identifier: KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyID,
			Value:      identifierBytes,
			Length:     uint16(len(identifierBytes)),
		},
	)

	// Key Hash (entry type 0x02) [OPTIONAL]
	if len(kc.KeyHash) > 0 {
		blob.Entries = append(
			blob.Entries,
			KEYCREDENTIALLINK_ENTRY{
				Identifier: KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyHash,
				Value:      kc.KeyHash,
				Length:     uint16(len(kc.KeyHash)),
			},
		)
	}

	// Key Material (entry type 0x03) [MANDATORY]
	keyMaterialBytes, err := kc.KeyMaterial.Marshal()
	if err != nil {
		return nil, err
	}
	blob.Entries = append(
		blob.Entries,
		KEYCREDENTIALLINK_ENTRY{
			Identifier: KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyMaterial,
			Value:      keyMaterialBytes,
			Length:     uint16(len(keyMaterialBytes)),
		},
	)

	// Key Usage (entry type 0x04) [MANDATORY]
	usageBytes, err := kc.Usage.Marshal()
	if err != nil {
		return nil, err
	}
	blob.Entries = append(
		blob.Entries,
		KEYCREDENTIALLINK_ENTRY{
			Identifier: KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyUsage,
			Value:      usageBytes,
			Length:     uint16(len(usageBytes)),
		},
	)

	// Legacy Key Usage (entry type 0x04) [OPTIONAL]
	if len(kc.LegacyUsage) > 0 {
		legacyUsageBytes := []byte(kc.LegacyUsage)
		blob.Entries = append(
			blob.Entries,
			KEYCREDENTIALLINK_ENTRY{
				Identifier: KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyUsage,
				Value:      legacyUsageBytes,
				Length:     uint16(len(legacyUsageBytes)),
			},
		)
	}

	// Key Source (entry type 0x05) [OPTIONAL]
	if kc.Source != nil {
		sourceBytes, err := kc.Source.Marshal()
		if err != nil {
			return nil, err
		}
		blob.Entries = append(
			blob.Entries,
			KEYCREDENTIALLINK_ENTRY{
				Identifier: KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeySource,
				Value:      sourceBytes,
				Length:     uint16(len(sourceBytes)),
			},
		)
	}

	// Device Identifier (entry type 0x06) [OPTIONAL]
	if kc.DeviceId != nil {
		deviceIdBytes := kc.DeviceId.ToBytes()
		blob.Entries = append(
			blob.Entries,
			KEYCREDENTIALLINK_ENTRY{
				Identifier: KEYCREDENTIALLINK_ENTRY_IDENTIFIER_DeviceId,
				Value:      deviceIdBytes,
				Length:     uint16(len(deviceIdBytes)),
			},
		)
	}

	// Custom Key Information (entry type 0x07) [OPTIONAL]
	if kc.CustomKeyInfo != nil {
		customKeyInfoBytes, err := kc.CustomKeyInfo.Marshal()
		if err != nil {
			return nil, err
		}
		blob.Entries = append(
			blob.Entries,
			KEYCREDENTIALLINK_ENTRY{
				Identifier: KEYCREDENTIALLINK_ENTRY_IDENTIFIER_CustomKeyInformation,
				Value:      customKeyInfoBytes,
				Length:     uint16(len(customKeyInfoBytes)),
			},
		)
	}

	// Last Logon Time (entry type 0x08) [OPTIONAL]
	if kc.LastLogonTime != nil {
		lastLogonTimeBytes, err := kc.LastLogonTime.Marshal()
		if err != nil {
			return nil, err
		}
		blob.Entries = append(
			blob.Entries,
			KEYCREDENTIALLINK_ENTRY{
				Identifier: KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyApproximateLastLogonTimeStamp,
				Value:      lastLogonTimeBytes,
				Length:     uint16(len(lastLogonTimeBytes)),
			},
		)
	}

	// Creation Time (entry type 0x09) [OPTIONAL]
	if kc.CreationTime != nil {
		creationTimeBytes, err := kc.CreationTime.Marshal()
		if err != nil {
			return nil, err
		}
		blob.Entries = append(
			blob.Entries,
			KEYCREDENTIALLINK_ENTRY{
				Identifier: KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyCreationTime,
				Value:      creationTimeBytes,
				Length:     uint16(len(creationTimeBytes)),
			},
		)
	}

	return &blob, nil
}

// Describe prints a detailed description of the KeyCredentialLink structure.
//
// Parameters:
// - indent: An integer value specifying the indentation level for the output.
func (kc *KeyCredentialLink) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)

	fmt.Printf("%s<KeyCredentialLink structure>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mVersion\x1b[0m: %s (0x%x)\n", indentPrompt, kc.Version.String(), kc.Version.Value)
	fmt.Printf("%s │ \x1b[93mKeyID\x1b[0m: %s\n", indentPrompt, kc.Identifier)
	if len(kc.KeyHash) > 0 {
		if kc.CheckIntegrity() {
			fmt.Printf("%s │ \x1b[93mKeyHash\x1b[0m: %s (\x1b[92mvalid\x1b[0m)\n", indentPrompt, hex.EncodeToString(kc.KeyHash))
		} else {
			fmt.Printf("%s │ \x1b[93mKeyHash\x1b[0m: %s (\x1b[91minvalid\x1b[0m)\n", indentPrompt, hex.EncodeToString(kc.KeyHash))
		}
	}
	kc.KeyMaterial.Describe(indent + 1)
	fmt.Printf("%s │ \x1b[93mUsage\x1b[0m: %s\n", indentPrompt, kc.Usage.String())
	if len(kc.LegacyUsage) != 0 {
		fmt.Printf("%s │ \x1b[93mLegacyUsage\x1b[0m: %s\n", indentPrompt, kc.LegacyUsage)
	}
	fmt.Printf("%s │ \x1b[93mSource\x1b[0m: 0x%02x (%s)\n", indentPrompt, kc.Source.Value, kc.Source.String())
	fmt.Printf("%s │ \x1b[93mDeviceId\x1b[0m: %s\n", indentPrompt, kc.DeviceId.ToFormatD())
	kc.CustomKeyInfo.Describe(indent + 1)
	fmt.Printf("%s │ \x1b[93mLastLogonTime (UTC)\x1b[0m: %s\n", indentPrompt, kc.LastLogonTime.String())
	fmt.Printf("%s │ \x1b[93mCreationTime (UTC)\x1b[0m: %s\n", indentPrompt, kc.CreationTime.String())
	fmt.Printf("%s └───\n", indentPrompt)
}

// ComposeKeyCredentialLinkForComputer builds a DN-Binary for msDS-KeyCredentialLink
// following the algorithm described in:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-drsr/f81a5aa6-6cc1-4320-9eb4-1443ad8f4b7c
//
// Parameters:
// - obj: Distinguished Name of the directory object (computer)
// - keyValue: the key material bytes (CNG key blob) to embed
//
// Returns:
// - ldap.DNWithBinary containing DN and composed KEYCREDENTIALLINK_BLOB
// - error if composition fails
func ComposeKeyCredentialLinkForComputer(obj string, keyValue []byte) (*ldap.DNWithBinary, error) {
	kcv := version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}

	// Compute KeyID from keyValue as SHA256(keyValue) represented per version
	keyID := utils.ComputeKeyIdentifier(keyValue, kcv)

	// Parse provided keyValue into a KeyMaterial
	keyMaterial, _, err := bcrypt.UnmarshalKeyMaterial(keyValue)
	if err != nil {
		return nil, fmt.Errorf("invalid key material: %w", err)
	}

	now := utils.NewDateTimeFromTicks(0)

	// Construct KCL with required fields; constructor also sets Usage=NGC and Source=AD
	kc := NewKeyCredentialLink(
		kcv,
		keyID,
		keyMaterial,
		nil,  // DeviceId (optional)
		&now, // LastLogonTime
		&now, // CreationTime
	)

	// Ensure KeyHash present (constructor computes it)
	if len(kc.KeyHash) == 0 {
		kc.KeyHash = kc.ComputeKeyHash()
	}

	// Marshal to blob bytes and wrap in DNWithBinary
	raw, err := kc.Marshal()
	if err != nil {
		return nil, err
	}

	return &ldap.DNWithBinary{
		DistinguishedName: obj,
		BinaryData:        raw,
	}, nil
}
