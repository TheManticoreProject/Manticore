package keycredentiallink

import (
	"github.com/TheManticoreProject/Manticore/windows/keycredential/crypto"
	"github.com/TheManticoreProject/Manticore/windows/keycredential/key"
	"github.com/TheManticoreProject/Manticore/windows/keycredential/utils"

	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// KeyCredential represents a key credential structure used for authentication and authorization.
//
// Fields:
// - Version: A KeyCredentialVersion object representing the version of the key credential.
// - Identifier: A string representing the unique identifier of the key credential.
// - KeyHash: A byte slice containing the hash of the key material.
// - RawKeyMaterial: An RSAKeyMaterial object representing the raw RSA key material.
// - Usage: A KeyUsage object representing the usage of the key credential.
// - LegacyUsage: A string representing the legacy usage of the key credential.
// - Source: A KeySource object representing the source of the key credential.
// - LastLogonTime: A DateTime object representing the last logon time associated with the key credential.
// - CreationTime: A DateTime object representing the creation time of the key credential.
// - Owner: A string representing the owner of the key credential.
// - RawBytes: A byte slice containing the raw binary data of the key credential.
// - RawBytesSize: A uint32 value representing the size of the raw binary data.
//
// Methods:
// - ParseDNWithBinary: Parses the provided DNWithBinary object into the KeyCredential structure.
//
// Note:
// The KeyCredential structure is used to store and manage key credentials, which are used for authentication and authorization purposes.
// The structure includes fields for version, identifier, key hash, raw key material, usage, legacy usage, source, last logon time, creation time, owner, and raw binary data.
// The ParseDNWithBinary method is used to parse a DNWithBinary object and populate the fields of the KeyCredential structure.
type KeyCredential struct {
	Version        key.KeyCredentialVersion
	Identifier     string
	KeyHash        []byte
	RawKeyMaterial crypto.RSAKeyMaterial
	Usage          key.KeyUsage
	LegacyUsage    string
	Source         key.KeySource
	CustomKeyInfo  key.CustomKeyInformation
	DeviceId       guid.GUID
	LastLogonTime  utils.DateTime
	CreationTime   utils.DateTime

	// Internal
	RawBytes     []byte
	RawBytesSize uint32
}

// NewKeyCredential creates a new KeyCredential structure.
//
// Parameters:
//
// - version: A KeyCredentialVersion object representing the version of the key credential.
//
// - Identifier: A string representing the unique identifier of the key credential.
//
// - KeyHash: A byte slice containing the hash of the key material.
//
// - RawKeyMaterial: An RSAKeyMaterial object representing the raw RSA key material.
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
// - Owner: A string representing the owner of the key credential.
//
// Returns:
//
// - A pointer to a KeyCredential object.
func NewKeyCredential(
	Version key.KeyCredentialVersion,
	Identifier string,
	RawKeyMaterial crypto.RSAKeyMaterial,
	DeviceId guid.GUID,
	LastLogonTime utils.DateTime,
	CreationTime utils.DateTime,
) *KeyCredential {
	kc := &KeyCredential{
		Version:        Version,
		Identifier:     Identifier,
		KeyHash:        []byte{},
		RawKeyMaterial: RawKeyMaterial,
		Usage:          key.KeyUsage{Value: key.KeyUsage_NGC},
		LegacyUsage:    "",
		Source:         key.KeySource_AD,
		CustomKeyInfo: key.CustomKeyInformation{
			Version: 1,
			Flags: key.CustomKeyInformationFlags{
				Value: 0,
			},
		},
		DeviceId:      DeviceId,
		LastLogonTime: LastLogonTime,
		CreationTime:  CreationTime,
		RawBytes:      []byte{},
		RawBytesSize:  0,
	}

	kc.KeyHash = kc.ComputeKeyHash()

	return kc
}

// ParseDNWithBinary parses the provided DNWithBinary object into the KeyCredential structure.
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
// 5. Updates the corresponding fields of the KeyCredential structure based on the parsed entry type and data.
//
// The function handles various entry types, including key identifier, key hash, key material, key usage, legacy usage, key source, last logon time, and creation time.
// Unsupported entry types, such as device ID and custom key information, are commented out for future implementation.
func (kc *KeyCredential) ParseDNWithBinary(dnWithBinary DNWithBinary) error {
	err := kc.FromBytes(dnWithBinary.BinaryData)
	if err != nil {
		return err
	}
	return nil
}

// ParseDNWithBinary parses the provided DNWithBinary object into the KeyCredential structure.
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
// 5. Updates the corresponding fields of the KeyCredential structure based on the parsed entry type and data.
//
// The function handles various entry types, including key identifier, key hash, key material, key usage, legacy usage, key source, last logon time, and creation time.
// Unsupported entry types, such as device ID and custom key information, are commented out for future implementation.
func (kc *KeyCredential) FromBytes(rawBytes []byte) error {
	kc.RawBytes = rawBytes
	remainder := rawBytes
	kc.RawBytesSize = 0

	kc.Version.FromBytes(kc.RawBytes)
	kc.RawBytesSize += kc.Version.RawBytesSize
	remainder = remainder[kc.Version.RawBytesSize:]

	// Read all entries corresponding to the KEYCREDENTIALLINK_ENTRY structure:
	for len(remainder) > 3 {
		// A 16-bit unsigned integer that specifies the length of the Value field.
		length := binary.LittleEndian.Uint16(remainder[:2])

		entryType := key.KeyCredentialEntryType{}
		entryType.FromBytes(remainder[2])

		remainder = remainder[3:]

		// An 8-bit unsigned integer that specifies the type of data that is stored in the Value field.
		entryData := remainder[:length]
		remainder = remainder[length:]

		switch entryType.Value {
		case key.KeyCredentialEntryType_KeyID:
			kc.Identifier = utils.ConvertFromBinaryIdentifier(entryData, kc.Version)
		case key.KeyCredentialEntryType_KeyHash:
			kc.KeyHash = entryData
		case key.KeyCredentialEntryType_KeyMaterial:
			kc.RawKeyMaterial.FromBytes(entryData)
		case key.KeyCredentialEntryType_KeyUsage:
			if len(entryData) == 1 {
				// This is apparently a V2 structure
				kc.Usage.FromBytes(entryData[0])
			} else {
				// This is a legacy structure that contains a string-encoded key usage instead of enum.
				kc.LegacyUsage = string(entryData)
			}
		case key.KeyCredentialEntryType_KeySource:
			kc.Source = key.KeySource(entryData[0])
		case key.KeyCredentialEntryType_DeviceId:
			kc.DeviceId.FromRawBytes(entryData)
		case key.KeyCredentialEntryType_CustomKeyInformation:
			kc.CustomKeyInfo.FromBytes(entryData, kc.Version)
		case key.KeyCredentialEntryType_KeyApproximateLastLogonTimeStamp:
			kc.LastLogonTime = utils.ConvertFromBinaryTime(entryData, kc.Source, kc.Version)
		case key.KeyCredentialEntryType_KeyCreationTime:
			kc.CreationTime = utils.ConvertFromBinaryTime(entryData, kc.Source, kc.Version)
		}
	}

	return nil
}

// CheckIntegrity checks the integrity of the key credential.
//
// Returns:
// - A boolean value indicating the integrity of the key credential.
func (kc *KeyCredential) CheckIntegrity() bool {
	hash := kc.ComputeKeyHash()

	if len(hash) != len(kc.KeyHash) {
		return false
	}

	for i := range hash {
		if hash[i] != kc.KeyHash[i] {
			return false
		}
	}

	return true
}

// ComputeKeyHash computes the key hash of the key credential.
//
// Returns:
// - A byte slice containing the key hash.
func (kc *KeyCredential) ComputeKeyHash() []byte {
	// Src: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/a99409ea-4f72-b7ef-8596013a36c7
	data := []byte{}

	if len(kc.RawBytes) < 4 {
		rawBytes, err := kc.ToBytes()
		if err != nil {
			return nil
		}
		kc.RawBytes = rawBytes
	}

	remainder := kc.RawBytes[4:]

	// Read all entries corresponding to the KEYCREDENTIALLINK_ENTRY structure:
	for len(remainder) > 3 {
		// A 16-bit unsigned integer that specifies the length of the Value field.
		length := binary.LittleEndian.Uint16(remainder[:2])
		entryType := key.KeyCredentialEntryType{}
		entryType.FromBytes(remainder[2])

		remainder = remainder[3:]
		remainder = remainder[length:]

		switch entryType.Value {
		case key.KeyCredentialEntryType_KeyHash:
			data = append(data, remainder...)
		}
	}

	hash := utils.ComputeHash(data)

	return hash
}

// writeEntry writes a typed KeyCredentialEntry to the buffer.
//
// Parameters:
// - buffer: A pointer to a bytes.Buffer object.
// - entryType: A KeyCredentialEntryType object representing the type of the entry.
// - data: A byte slice representing the data to be written.
func writeEntry(buffer *bytes.Buffer, entryType key.KeyCredentialEntryType, data []byte) {
	binary.Write(buffer, binary.LittleEndian, uint16(len(data)))
	buffer.Write(entryType.ToBytes())
	buffer.Write(data)
}

// ToBytes returns the raw bytes of the KeyCredential structure.
//
// Returns:
// - A byte slice representing the raw bytes of the KeyCredential structure.
// - An error if the conversion fails.
func (kc *KeyCredential) ToBytes() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)

	// kc.Version
	buffer.Write(kc.Version.ToBytes())

	// kc.Identifier
	if len(kc.Identifier) > 0 {
		entryType := key.KeyCredentialEntryType{Value: key.KeyCredentialEntryType_KeyID}
		identifierBytes, err := utils.ConvertToBinaryIdentifier(kc.Identifier, kc.Version)
		if err != nil {
			return nil, err
		}
		writeEntry(buffer, entryType, identifierBytes)
	}

	// kc.KeyHash
	if len(kc.KeyHash) > 0 {
		entryType := key.KeyCredentialEntryType{Value: key.KeyCredentialEntryType_KeyHash}
		writeEntry(buffer, entryType, kc.KeyHash)
	} else {
		entryType := key.KeyCredentialEntryType{Value: key.KeyCredentialEntryType_KeyHash}
		writeEntry(buffer, entryType, make([]byte, 32))
	}

	// kc.RawKeyMaterial
	entryType := key.KeyCredentialEntryType{Value: key.KeyCredentialEntryType_KeyMaterial}
	writeEntry(buffer, entryType, kc.RawKeyMaterial.ToBytes())

	// kc.Usage
	entryType = key.KeyCredentialEntryType{Value: key.KeyCredentialEntryType_KeyUsage}
	writeEntry(buffer, entryType, []byte{kc.Usage.Value})

	// kc.LegacyUsage
	if len(kc.LegacyUsage) > 0 {
		entryType = key.KeyCredentialEntryType{Value: key.KeyCredentialEntryType_KeyUsage}
		data := []byte(kc.LegacyUsage)
		writeEntry(buffer, entryType, data)
	}

	// kc.Source
	entryType = key.KeyCredentialEntryType{Value: key.KeyCredentialEntryType_KeySource}
	writeEntry(buffer, entryType, []byte{byte(kc.Source)})

	// kc.DeviceId
	entryType = key.KeyCredentialEntryType{Value: key.KeyCredentialEntryType_DeviceId}
	data := kc.DeviceId.ToBytes()
	writeEntry(buffer, entryType, data)

	// kc.CustomKeyInfo
	customKeyInfoBytes := kc.CustomKeyInfo.ToBytes()
	if len(customKeyInfoBytes) > 0 {
		entryType := key.KeyCredentialEntryType{Value: key.KeyCredentialEntryType_CustomKeyInformation}
		writeEntry(buffer, entryType, customKeyInfoBytes)
	}

	// kc.LastLogonTime
	entryType = key.KeyCredentialEntryType{Value: key.KeyCredentialEntryType_KeyApproximateLastLogonTimeStamp}
	data = kc.LastLogonTime.ToBytes()
	writeEntry(buffer, entryType, data)

	// kc.CreationTime
	entryType = key.KeyCredentialEntryType{Value: key.KeyCredentialEntryType_KeyCreationTime}
	data = kc.CreationTime.ToBytes()
	writeEntry(buffer, entryType, data)

	return buffer.Bytes(), nil
}

// Describe prints a detailed description of the KeyCredential structure.
//
// Parameters:
// - indent: An integer value specifying the indentation level for the output.
func (kc *KeyCredential) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)

	fmt.Printf("%s<KeyCredential structure>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mVersion\x1b[0m: %s (0x%x)\n", indentPrompt, kc.Version.String(), kc.Version.Value)
	fmt.Printf("%s │ \x1b[93mKeyID\x1b[0m: %s\n", indentPrompt, kc.Identifier)
	if kc.CheckIntegrity() {
		fmt.Printf("%s │ \x1b[93mKeyHash\x1b[0m: %s (\x1b[92mvalid\x1b[0m)\n", indentPrompt, hex.EncodeToString(kc.KeyHash))
	} else {
		fmt.Printf("%s │ \x1b[93mKeyHash\x1b[0m: %s (\x1b[91minvalid\x1b[0m)\n", indentPrompt, hex.EncodeToString(kc.KeyHash))
	}
	kc.RawKeyMaterial.Describe(indent + 1)
	fmt.Printf("%s │ \x1b[93mUsage\x1b[0m: %s\n", indentPrompt, kc.Usage.String())
	if len(kc.LegacyUsage) != 0 {
		fmt.Printf("%s │ \x1b[93mLegacyUsage\x1b[0m: %s\n", indentPrompt, kc.LegacyUsage)
	} else {
		fmt.Printf("%s │ \x1b[93mLegacyUsage\x1b[0m: None\n", indentPrompt)
	}
	fmt.Printf("%s │ \x1b[93mSource\x1b[0m: %s\n", indentPrompt, kc.Source)
	fmt.Printf("%s │ \x1b[93mDeviceId\x1b[0m: %s\n", indentPrompt, kc.DeviceId.ToFormatD())
	kc.CustomKeyInfo.Describe(indent + 1)
	fmt.Printf("%s │ \x1b[93mLastLogonTime (UTC)\x1b[0m: %s\n", indentPrompt, kc.LastLogonTime.String())
	fmt.Printf("%s │ \x1b[93mCreationTime (UTC)\x1b[0m: %s\n", indentPrompt, kc.CreationTime.String())
	fmt.Printf("%s └───\n", indentPrompt)
}
