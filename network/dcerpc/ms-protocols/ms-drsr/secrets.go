package msdrsr

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"

	drsrtypes "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// AccountSecrets holds the credential material decrypted from one replicated object.
// NTHash/LMHash are the current 16-byte hashes (valid only when HasNT/HasLM is set);
// the history slices are previous hashes as the DC returns them. SupplementalCredentials
// is the transport-decrypted supplementalCredentials blob (Kerberos keys, cleartext, …)
// left unparsed — decoding it is future work.
type AccountSecrets struct {
	DN             string
	SAMAccountName string
	RID            uint32
	NTHash         [16]byte
	LMHash         [16]byte
	HasNT          bool
	HasLM          bool
	IsDeleted      bool
	NTHistory      [][16]byte
	LMHistory      [][16]byte

	SupplementalCredentials []byte
}

// DecryptSecrets decrypts the secret attributes of every replicated object that carries
// an objectSid (i.e. a security principal), using the connection's NTLM session key. It
// must be called on the same Client whose ReplicateSingleObject (or full-NC call)
// produced the result, because decryption is keyed by that connection's session key.
func (c *Client) DecryptSecrets(res *ReplicationResult) ([]*AccountSecrets, error) {
	sessionKey := c.SessionKey()
	if len(sessionKey) == 0 {
		return nil, fmt.Errorf("msdrsr: no session key (not connected, or anonymous bind)")
	}
	var out []*AccountSecrets
	for _, obj := range res.Objects {
		sid := firstValue(findAttr(obj, res.PrefixTable, oidObjectSid))
		if sid == nil {
			continue // not a security principal (no objectSid) — nothing to decrypt
		}
		rid, err := ridFromSID(sid)
		if err != nil {
			return nil, fmt.Errorf("msdrsr: %s: %w", obj.DN, err)
		}
		sec, err := decryptObjectSecrets(obj, res.PrefixTable, sessionKey, rid)
		if err != nil {
			return nil, fmt.Errorf("msdrsr: %s: %w", obj.DN, err)
		}
		out = append(out, sec)
	}
	return out, nil
}

// decryptObjectSecrets extracts and decrypts the secret attributes of a single object.
func decryptObjectSecrets(obj ReplicatedObject, prefixTable drsrtypes.SCHEMA_PREFIX_TABLE, sessionKey []byte, rid uint32) (*AccountSecrets, error) {
	sec := &AccountSecrets{DN: obj.DN, RID: rid}

	// isDeleted is a 4-byte LE boolean: TRUE = 0x01000000
	if v := firstValue(findAttr(obj, prefixTable, oidIsDeleted)); v != nil && len(v) >= 4 && v[0] != 0 {
		sec.IsDeleted = true
	}

	if v := firstValue(findAttr(obj, prefixTable, oidSAMAccountName)); v != nil {
		sec.SAMAccountName = utf16leToString(v)
	}

	if v := firstValue(findAttr(obj, prefixTable, oidUnicodePwd)); v != nil {
		h, err := decryptHash(sessionKey, v, rid)
		if err != nil {
			return nil, fmt.Errorf("unicodePwd: %w", err)
		}
		sec.NTHash, sec.HasNT = h, true
	}
	if v := firstValue(findAttr(obj, prefixTable, oidDBCSPwd)); v != nil {
		h, err := decryptHash(sessionKey, v, rid)
		if err != nil {
			return nil, fmt.Errorf("dBCSPwd: %w", err)
		}
		sec.LMHash, sec.HasLM = h, true
	}
	if v := firstValue(findAttr(obj, prefixTable, oidNTPwdHistory)); v != nil {
		hist, err := decryptHashHistory(sessionKey, v, rid)
		if err != nil {
			return nil, fmt.Errorf("ntPwdHistory: %w", err)
		}
		sec.NTHistory = hist
	}
	if v := firstValue(findAttr(obj, prefixTable, oidLMPwdHistory)); v != nil {
		hist, err := decryptHashHistory(sessionKey, v, rid)
		if err != nil {
			return nil, fmt.Errorf("lmPwdHistory: %w", err)
		}
		sec.LMHistory = hist
	}
	if v := firstValue(findAttr(obj, prefixTable, oidSupplementalCredentials)); v != nil {
		blob, err := decryptReplicatedValue(sessionKey, v)
		if err != nil {
			return nil, fmt.Errorf("supplementalCredentials: %w", err)
		}
		sec.SupplementalCredentials = blob
	}
	return sec, nil
}

// firstValue returns the first value of a multi-valued attribute, or nil.
func firstValue(vals [][]byte) []byte {
	if len(vals) == 0 {
		return nil
	}
	return vals[0]
}

// ridFromSID extracts the RID (last sub-authority) from a binary SID ([MS-DTYP] 2.4.2.2:
// revision, sub-authority count, 6-byte identifier authority, then count 32-bit LE
// sub-authorities). The RID is the final sub-authority.
func ridFromSID(sid []byte) (uint32, error) {
	if len(sid) < 8 {
		return 0, fmt.Errorf("SID too short: %d bytes", len(sid))
	}
	count := int(sid[1])
	end := 8 + 4*count
	if count == 0 || end > len(sid) {
		return 0, fmt.Errorf("malformed SID: %d sub-authorities, %d bytes", count, len(sid))
	}
	return binary.LittleEndian.Uint32(sid[end-4 : end]), nil
}

// utf16leToString decodes UTF-16LE bytes (an AD unicode string attribute value) to a Go
// string, dropping a trailing NUL if present.
func utf16leToString(b []byte) string {
	units := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		units = append(units, binary.LittleEndian.Uint16(b[i:i+2]))
	}
	for len(units) > 0 && units[len(units)-1] == 0 {
		units = units[:len(units)-1]
	}
	return string(utf16.Decode(units))
}
