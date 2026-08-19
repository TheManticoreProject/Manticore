package msdrsr

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"
	"unicode/utf16"
)

func supplementalUTF16(s string) []byte {
	units := utf16.Encode([]rune(s))
	out := make([]byte, len(units)*2)
	for i, unit := range units {
		binary.LittleEndian.PutUint16(out[i*2:], unit)
	}
	return out
}

func supplementalProperty(name string, value []byte) []byte {
	nameBytes := supplementalUTF16(name)
	valueBytes := []byte(hex.EncodeToString(value))
	out := make([]byte, 6+len(nameBytes)+len(valueBytes))
	binary.LittleEndian.PutUint16(out[0:2], uint16(len(nameBytes)))
	binary.LittleEndian.PutUint16(out[2:4], uint16(len(valueBytes)))
	copy(out[6:], nameBytes)
	copy(out[6+len(nameBytes):], valueBytes)
	return out
}

func supplementalBlob(properties ...[]byte) []byte {
	out := make([]byte, 112)
	binary.LittleEndian.PutUint16(out[108:110], 0x50)
	binary.LittleEndian.PutUint16(out[110:112], uint16(len(properties)))
	for _, property := range properties {
		out = append(out, property...)
	}
	propertiesEnd := len(out)
	out = append(out, 0)
	binary.LittleEndian.PutUint32(out[4:8], uint32(propertiesEnd-12))
	return out
}

func kerberosNewValue() ([]byte, []byte, []byte, string) {
	salt := "LAB.LOCALtest"
	saltBytes := supplementalUTF16(salt)
	aes256 := bytes.Repeat([]byte{0x25}, 32)
	aes128 := bytes.Repeat([]byte{0x18}, 16)
	value := make([]byte, kerberosNewHeaderSize+2*kerberosNewKeySize)
	binary.LittleEndian.PutUint16(value[0:2], 4)
	binary.LittleEndian.PutUint16(value[4:6], 1)
	binary.LittleEndian.PutUint16(value[8:10], 1)
	binary.LittleEndian.PutUint16(value[12:14], uint16(len(saltBytes)))
	binary.LittleEndian.PutUint16(value[14:16], uint16(len(saltBytes)))
	binary.LittleEndian.PutUint32(value[16:20], uint32(len(value)))
	binary.LittleEndian.PutUint32(value[20:24], 4096)

	keyOffset := len(value) + len(saltBytes)
	binary.LittleEndian.PutUint32(value[24+8:24+12], 4096)
	binary.LittleEndian.PutUint32(value[24+12:24+16], 18)
	binary.LittleEndian.PutUint32(value[24+16:24+20], uint32(len(aes256)))
	binary.LittleEndian.PutUint32(value[24+20:24+24], uint32(keyOffset))
	second := 24 + kerberosNewKeySize
	binary.LittleEndian.PutUint32(value[second+8:second+12], 4096)
	binary.LittleEndian.PutUint32(value[second+12:second+16], 17)
	binary.LittleEndian.PutUint32(value[second+16:second+20], uint32(len(aes128)))
	binary.LittleEndian.PutUint32(value[second+20:second+24], uint32(keyOffset+len(aes256)))
	value = append(value, saltBytes...)
	value = append(value, aes256...)
	value = append(value, aes128...)
	return value, aes256, aes128, salt
}

func TestParseSupplementalCredentials(t *testing.T) {
	newer, aes256, aes128, salt := kerberosNewValue()
	cleartext := supplementalUTF16("P@ssw0rd!")
	wdigest := make([]byte, 16+29*16)
	wdigest[2], wdigest[3] = 1, 29
	for i := 0; i < 29; i++ {
		copy(wdigest[16+i*16:], bytes.Repeat([]byte{byte(i + 1)}, 16))
	}
	blob := supplementalBlob(
		supplementalProperty("Primary:Kerberos-Newer-Keys", newer),
		supplementalProperty("Primary:CLEARTEXT", cleartext),
		supplementalProperty("Primary:WDigest", wdigest),
	)

	info, err := ParseSupplementalCredentials(blob)
	if err != nil {
		t.Fatalf("ParseSupplementalCredentials: %v", err)
	}
	if len(info.KerberosKeys) != 2 {
		t.Fatalf("KerberosKeys = %d, want 2", len(info.KerberosKeys))
	}
	if key := info.KerberosKeys[0]; key.KeyType != 18 || !bytes.Equal(key.Value, aes256) || key.Salt != salt || key.IterationCount != 4096 || key.Category != KerberosKeyCurrent {
		t.Errorf("AES256 key = %+v", key)
	}
	if key := info.KerberosKeys[1]; key.KeyType != 17 || !bytes.Equal(key.Value, aes128) || key.Category != KerberosKeyOld {
		t.Errorf("AES128 old key = %+v", key)
	}
	if info.CleartextPassword != "P@ssw0rd!" || !bytes.Equal(info.CleartextPasswordRaw, cleartext) {
		t.Errorf("cleartext = %q (%x)", info.CleartextPassword, info.CleartextPasswordRaw)
	}
	if len(info.WDigestHashes) != 29 || info.WDigestHashes[28][0] != 29 {
		t.Errorf("WDigest hashes = %d", len(info.WDigestHashes))
	}
}

func TestParseSupplementalCredentialsRejectsMalformedValues(t *testing.T) {
	newer, _, _, _ := kerberosNewValue()
	binary.LittleEndian.PutUint32(newer[24+20:24+24], uint32(len(newer)+1))
	tests := []struct {
		name string
		blob []byte
	}{
		{name: "short header", blob: []byte{1, 2, 3}},
		{name: "bad signature", blob: make([]byte, 111)},
		{name: "key beyond property", blob: supplementalBlob(supplementalProperty("Primary:Kerberos-Newer-Keys", newer))},
		{name: "odd cleartext", blob: supplementalBlob(supplementalProperty("Primary:CLEARTEXT", []byte{1}))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseSupplementalCredentials(tt.blob); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
