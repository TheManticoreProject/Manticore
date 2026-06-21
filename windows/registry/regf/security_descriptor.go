package regf

import (
	"fmt"

	"github.com/TheManticoreProject/winacl/securitydescriptor"
)

// SecurityDescriptorParsed decodes this key's SK record into a winacl
// NtSecurityDescriptor (owner, group, DACL, SACL). It returns nil (no error) when the key
// references no SK record. Use the raw SecurityDescriptor accessor instead when only the
// undecoded bytes are needed.
func (k *KeyNode) SecurityDescriptorParsed() (*securitydescriptor.NtSecurityDescriptor, error) {
	raw, err := k.SecurityDescriptor()
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	sd := securitydescriptor.NewSecurityDescriptor()
	if _, err := sd.Unmarshal(raw); err != nil {
		return nil, fmt.Errorf("regf: parsing security descriptor: %w", err)
	}
	return sd, nil
}

// GetSecurityDescriptor decodes the SECURITY_DESCRIPTOR of the key at the given path into
// a winacl NtSecurityDescriptor.
//
// Parameters:
//   - keyPath (string): key path relative to the root key.
//
// Returns:
//   - The parsed security descriptor, or nil if the key has no SK record.
//   - An error if the key is not found or the descriptor cannot be parsed.
func (h *Hive) GetSecurityDescriptor(keyPath string) (*securitydescriptor.NtSecurityDescriptor, error) {
	nk, err := h.FindKey(keyPath)
	if err != nil {
		return nil, err
	}
	return nk.SecurityDescriptorParsed()
}
