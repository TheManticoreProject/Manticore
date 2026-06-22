package ntds

import (
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
	"fmt"
)

const (
	// pekListHeaderLen is the size of the encrypted pekList header (8) plus its key
	// material (16) before the encrypted body.
	pekListHeaderLen = 24
	// pekPlainHeaderLen is the fixed header at the start of the decrypted pekList,
	// before the PEK entries.
	pekPlainHeaderLen = 32
	// pekEntryLen is the size of one decrypted PEK entry: a 4-byte prefix (index, or
	// header+padding) followed by the 16-byte key.
	pekEntryLen = 20
	// rc4Rounds is the number of key-material rounds folded into the MD5 derivation of
	// the RC4 key in the legacy pekList format.
	rc4Rounds = 1000
)

// DecryptPEKList decrypts the pekList attribute with the SYSTEM boot key and returns the
// contained PEKs (usually one). Both on-disk formats are supported, selected by the
// first header byte:
//
//   - 0x02 (up to Windows Server 2012 R2): RC4 keyed by MD5(bootKey + keyMaterial*1000).
//   - 0x03 (Windows Server 2016+): AES-CBC keyed by bootKey, IV = keyMaterial.
//
// bootKey is the 16-byte boot key derived from the SYSTEM hive (the caller supplies it;
// boot-key derivation is out of scope here).
func DecryptPEKList(bootKey, pekList []byte) ([]PEK, error) {
	if len(bootKey) < 16 {
		return nil, fmt.Errorf("ntds: boot key too short: %d bytes", len(bootKey))
	}
	if len(pekList) < pekListHeaderLen {
		return nil, fmt.Errorf("ntds: pekList too short: %d bytes", len(pekList))
	}
	header := pekList[0:8]
	keyMaterial := pekList[8:24]
	encrypted := pekList[24:]

	switch header[0] {
	case 0x02:
		md := md5.New()
		md.Write(bootKey[:16])
		for i := 0; i < rc4Rounds; i++ {
			md.Write(keyMaterial)
		}
		c, err := rc4.NewCipher(md.Sum(nil))
		if err != nil {
			return nil, fmt.Errorf("ntds: pekList RC4 key: %w", err)
		}
		plain := make([]byte, len(encrypted))
		c.XORKeyStream(plain, encrypted)
		return parsePEKListRC4(plain), nil

	case 0x03:
		plain, err := decryptAESCBC(bootKey[:16], encrypted, keyMaterial)
		if err != nil {
			return nil, err
		}
		return parsePEKListAES(plain), nil

	default:
		return nil, fmt.Errorf("ntds: unknown pekList format 0x%02X", header[0])
	}
}

// parsePEKListRC4 extracts the PEK keys from a decrypted legacy (RC4) pekList: a 32-byte
// header followed by fixed 20-byte entries whose last 16 bytes are the key.
func parsePEKListRC4(plain []byte) []PEK {
	if len(plain) < pekPlainHeaderLen {
		return nil
	}
	body := plain[pekPlainHeaderLen:]
	var peks []PEK
	for off := 0; off+pekEntryLen <= len(body); off += pekEntryLen {
		peks = append(peks, PEK(append([]byte(nil), body[off+4:off+pekEntryLen]...)))
	}
	return peks
}

// parsePEKListAES extracts the PEK keys from a decrypted AES pekList: a 32-byte header
// followed by entries of a 4-byte little-endian index and a 16-byte key, in ascending
// index order; parsing stops at the first non-sequential index (the list terminator).
func parsePEKListAES(plain []byte) []PEK {
	if len(plain) < pekPlainHeaderLen {
		return nil
	}
	body := plain[pekPlainHeaderLen:]
	var peks []PEK
	want := uint32(0)
	for off := 0; off+pekEntryLen <= len(body); off += pekEntryLen {
		if binary.LittleEndian.Uint32(body[off:off+4]) != want {
			break
		}
		peks = append(peks, PEK(append([]byte(nil), body[off+4:off+pekEntryLen]...)))
		want++
	}
	return peks
}
