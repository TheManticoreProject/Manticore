package msdrsr

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const (
	userPropertiesFixedSize = 110
	kerberosNewHeaderSize   = 24
	kerberosNewKeySize      = 24
	kerberosOldHeaderSize   = 16
	kerberosOldKeySize      = 20
)

// KerberosKeyCategory identifies the KERB_STORED_CREDENTIAL[_NEW] array from which a
// key was extracted.
type KerberosKeyCategory string

const (
	KerberosKeyCurrent KerberosKeyCategory = "current"
	KerberosKeyService KerberosKeyCategory = "service"
	KerberosKeyOld     KerberosKeyCategory = "old"
	KerberosKeyOlder   KerberosKeyCategory = "older"
)

// KerberosKey is one key extracted from Primary:Kerberos or
// Primary:Kerberos-Newer-Keys. KeyType is the Kerberos encryption type, Value is the raw
// key, Salt is the default UTF-16 salt, and IterationCount is zero for legacy records.
type KerberosKey struct {
	KeyType        uint32
	Value          []byte
	Salt           string
	IterationCount uint32
	Category       KerberosKeyCategory
}

// SupplementalCredentialsInfo is the parsed credential material in an MS-SAMR
// USER_PROPERTIES blob.
type SupplementalCredentialsInfo struct {
	KerberosKeys         []KerberosKey
	CleartextPassword    string
	CleartextPasswordRaw []byte
	WDigestHashes        [][16]byte
}

// ParseSupplementalCredentials parses an MS-SAMR USER_PROPERTIES blob. Unknown
// properties are ignored; recognized properties are bounds-checked and malformed data
// is rejected.
func ParseSupplementalCredentials(blob []byte) (*SupplementalCredentialsInfo, error) {
	if len(blob) < userPropertiesFixedSize+1 {
		return nil, fmt.Errorf("supplementalCredentials: USER_PROPERTIES is %d bytes, want at least %d", len(blob), userPropertiesFixedSize+1)
	}
	if signature := binary.LittleEndian.Uint16(blob[108:110]); signature != 0x50 {
		return nil, fmt.Errorf("supplementalCredentials: invalid property signature 0x%x", signature)
	}
	propertiesEnd := 12 + int(binary.LittleEndian.Uint32(blob[4:8]))
	if propertiesEnd < userPropertiesFixedSize || propertiesEnd >= len(blob) {
		return nil, fmt.Errorf("supplementalCredentials: invalid USER_PROPERTIES length %d", propertiesEnd-12)
	}
	info := &SupplementalCredentialsInfo{}
	if propertiesEnd == userPropertiesFixedSize {
		return info, nil // PropertyCount is omitted when it is zero.
	}
	if propertiesEnd < userPropertiesFixedSize+2 {
		return nil, fmt.Errorf("supplementalCredentials: USER_PROPERTIES is missing PropertyCount")
	}

	propertyCount := int(binary.LittleEndian.Uint16(blob[110:112]))
	offset := 112
	for i := 0; i < propertyCount; i++ {
		if propertiesEnd-offset < 6 {
			return nil, fmt.Errorf("supplementalCredentials: property %d header exceeds USER_PROPERTIES", i)
		}
		nameLength := int(binary.LittleEndian.Uint16(blob[offset : offset+2]))
		valueLength := int(binary.LittleEndian.Uint16(blob[offset+2 : offset+4]))
		offset += 6
		if nameLength%2 != 0 || nameLength > propertiesEnd-offset || valueLength > propertiesEnd-offset-nameLength {
			return nil, fmt.Errorf("supplementalCredentials: property %d has invalid name/value lengths %d/%d", i, nameLength, valueLength)
		}
		name := utf16leToString(blob[offset : offset+nameLength])
		offset += nameLength
		encodedValue := blob[offset : offset+valueLength]
		offset += valueLength
		value := make([]byte, hex.DecodedLen(len(encodedValue)))
		n, err := hex.Decode(value, encodedValue)
		if err != nil {
			return nil, fmt.Errorf("supplementalCredentials: property %q: invalid hexadecimal value: %w", name, err)
		}
		value = value[:n]

		switch name {
		case "Primary:Kerberos-Newer-Keys":
			keys, err := parseKerberosNew(value)
			if err != nil {
				return nil, fmt.Errorf("supplementalCredentials: %s: %w", name, err)
			}
			info.KerberosKeys = append(info.KerberosKeys, keys...)
		case "Primary:Kerberos":
			keys, err := parseKerberosLegacy(value)
			if err != nil {
				return nil, fmt.Errorf("supplementalCredentials: %s: %w", name, err)
			}
			info.KerberosKeys = append(info.KerberosKeys, keys...)
		case "Primary:CLEARTEXT":
			if len(value)%2 != 0 {
				return nil, fmt.Errorf("supplementalCredentials: %s: odd UTF-16 byte count %d", name, len(value))
			}
			info.CleartextPasswordRaw = append([]byte(nil), value...)
			info.CleartextPassword = utf16leToString(value)
		case "Primary:WDigest":
			hashes, err := parseWDigest(value)
			if err != nil {
				return nil, fmt.Errorf("supplementalCredentials: %s: %w", name, err)
			}
			info.WDigestHashes = hashes
		}
	}
	if offset != propertiesEnd {
		return nil, fmt.Errorf("supplementalCredentials: %d unparsed property bytes", propertiesEnd-offset)
	}
	return info, nil
}

func parseKerberosNew(value []byte) ([]KerberosKey, error) {
	if len(value) < kerberosNewHeaderSize {
		return nil, fmt.Errorf("KERB_STORED_CREDENTIAL_NEW is %d bytes", len(value))
	}
	if revision := binary.LittleEndian.Uint16(value[0:2]); revision != 4 {
		return nil, fmt.Errorf("invalid revision %d", revision)
	}
	counts := [4]int{
		int(binary.LittleEndian.Uint16(value[4:6])),
		int(binary.LittleEndian.Uint16(value[6:8])),
		int(binary.LittleEndian.Uint16(value[8:10])),
		int(binary.LittleEndian.Uint16(value[10:12])),
	}
	total := counts[0] + counts[1] + counts[2] + counts[3]
	if total > (len(value)-kerberosNewHeaderSize)/kerberosNewKeySize {
		return nil, fmt.Errorf("%d key records exceed property length", total)
	}
	salt, err := parseCredentialSalt(value, binary.LittleEndian.Uint16(value[12:14]), binary.LittleEndian.Uint32(value[16:20]))
	if err != nil {
		return nil, err
	}
	categories := [...]KerberosKeyCategory{KerberosKeyCurrent, KerberosKeyService, KerberosKeyOld, KerberosKeyOlder}
	keys := make([]KerberosKey, 0, total)
	offset := kerberosNewHeaderSize
	for categoryIndex, count := range counts {
		for range count {
			record := value[offset : offset+kerberosNewKeySize]
			key, err := parseCredentialKey(value, binary.LittleEndian.Uint32(record[12:16]), binary.LittleEndian.Uint32(record[16:20]), binary.LittleEndian.Uint32(record[20:24]))
			if err != nil {
				return nil, err
			}
			keys = append(keys, KerberosKey{
				KeyType:        binary.LittleEndian.Uint32(record[12:16]),
				Value:          key,
				Salt:           salt,
				IterationCount: binary.LittleEndian.Uint32(record[8:12]),
				Category:       categories[categoryIndex],
			})
			offset += kerberosNewKeySize
		}
	}
	return keys, nil
}

func parseKerberosLegacy(value []byte) ([]KerberosKey, error) {
	if len(value) < kerberosOldHeaderSize {
		return nil, fmt.Errorf("KERB_STORED_CREDENTIAL is %d bytes", len(value))
	}
	if revision := binary.LittleEndian.Uint16(value[0:2]); revision != 3 {
		return nil, fmt.Errorf("invalid revision %d", revision)
	}
	currentCount := int(binary.LittleEndian.Uint16(value[4:6]))
	oldCount := int(binary.LittleEndian.Uint16(value[6:8]))
	total := currentCount + oldCount
	if total > (len(value)-kerberosOldHeaderSize)/kerberosOldKeySize {
		return nil, fmt.Errorf("%d key records exceed property length", total)
	}
	salt, err := parseCredentialSalt(value, binary.LittleEndian.Uint16(value[8:10]), binary.LittleEndian.Uint32(value[12:16]))
	if err != nil {
		return nil, err
	}
	keys := make([]KerberosKey, 0, total)
	offset := kerberosOldHeaderSize
	for i := 0; i < total; i++ {
		record := value[offset : offset+kerberosOldKeySize]
		keyType := binary.LittleEndian.Uint32(record[8:12])
		key, err := parseCredentialKey(value, keyType, binary.LittleEndian.Uint32(record[12:16]), binary.LittleEndian.Uint32(record[16:20]))
		if err != nil {
			return nil, err
		}
		category := KerberosKeyCurrent
		if i >= currentCount {
			category = KerberosKeyOld
		}
		keys = append(keys, KerberosKey{KeyType: keyType, Value: key, Salt: salt, Category: category})
		offset += kerberosOldKeySize
	}
	return keys, nil
}

func parseCredentialSalt(value []byte, length uint16, offset uint32) (string, error) {
	if length == 0 {
		return "", nil
	}
	if length%2 != 0 || uint64(offset)+uint64(length) > uint64(len(value)) {
		return "", fmt.Errorf("invalid salt offset/length %d/%d", offset, length)
	}
	return utf16leToString(value[int(offset) : int(offset)+int(length)]), nil
}

func parseCredentialKey(value []byte, keyType, length, offset uint32) ([]byte, error) {
	if uint64(offset)+uint64(length) > uint64(len(value)) {
		return nil, fmt.Errorf("key type %d has invalid offset/length %d/%d", keyType, offset, length)
	}
	return append([]byte(nil), value[int(offset):int(offset)+int(length)]...), nil
}

func parseWDigest(value []byte) ([][16]byte, error) {
	if len(value) < 16 {
		return nil, fmt.Errorf("WDIGEST_CREDENTIALS is %d bytes", len(value))
	}
	if value[2] != 1 {
		return nil, fmt.Errorf("invalid version %d", value[2])
	}
	count := int(value[3])
	if count != 29 || count > (len(value)-16)/16 {
		return nil, fmt.Errorf("invalid hash count %d for %d-byte value", count, len(value))
	}
	hashes := make([][16]byte, count)
	for i := range hashes {
		copy(hashes[i][:], value[16+i*16:16+(i+1)*16])
	}
	return hashes, nil
}
