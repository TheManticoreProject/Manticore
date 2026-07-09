package pac

import (
	"fmt"
	"time"
	"unicode/utf16"

	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// This file is the read side of the KERB_VALIDATION_INFO logon-info buffer: it
// resolves a parsed PAC's ulType 0x00000001 buffer to the typed structure decoded
// by UnmarshalKerbValidationInfo (build.go) and adds convenience accessors that
// turn the raw NDR fields into Go strings, SIDs, and times. The decode itself is
// symmetric with the encoder in build.go (both use the same structs and the NDR
// "type serialization version 1" framing, [MS-RPCE] 2.2.6).

// filetimeNever is the FILETIME sentinel (0x7FFFFFFFFFFFFFFF) the KDC writes for
// "never" times such as LogoffTime, KickOffTime, and PasswordMustChange.
const filetimeNever = uint64(0x7FFFFFFFFFFFFFFF)

// LogonInfo locates the logon-info buffer (PAC_INFO_BUFFER ulType 0x00000001) and
// NDR-decodes it into a KERB_VALIDATION_INFO ([MS-PAC] 2.5). It returns an error
// if the PAC has no logon-info buffer or the buffer does not decode.
func (p *PAC) LogonInfo() (*KERB_VALIDATION_INFO, error) {
	b, ok := p.Buffer(BufferLogonInfo)
	if !ok {
		return nil, fmt.Errorf("pac: no logon-info buffer (ulType 0x%08x)", BufferLogonInfo)
	}
	return UnmarshalKerbValidationInfo(b.Data)
}

// Uint64 returns the FILETIME as a single 64-bit count of 100-nanosecond intervals
// since 1601-01-01 (the low half is the least significant).
func (f FILETIME) Uint64() uint64 {
	return uint64(f.HighDateTime)<<32 | uint64(f.LowDateTime)
}

// IsNever reports whether the FILETIME is the "never" sentinel
// (0x7FFFFFFFFFFFFFFF), used for LogoffTime/KickOffTime/PasswordMustChange when the
// corresponding limit does not apply.
func (f FILETIME) IsNever() bool { return f.Uint64() == filetimeNever }

// Time converts the FILETIME to a Go time in UTC. A zero FILETIME (password never
// set) and the "never" sentinel both map to the zero time.Time, so callers can
// test the result with time.Time.IsZero. It is the inverse of FileTimeFromTime.
func (f FILETIME) Time() time.Time {
	v := f.Uint64()
	if v == 0 || v == filetimeNever {
		return time.Time{}
	}
	return time.Unix(0, int64(v-filetimeEpochDelta)*100).UTC()
}

// String decodes the RPC_UNICODE_STRING to a Go string. Only the first Length/2
// UTF-16 code units are significant ([MS-DTYP] 2.3.10 counts bytes, and the buffer
// carries no NUL terminator); any extra allocated units are dropped.
func (u RPC_UNICODE_STRING) String() string {
	units := u.Buffer
	if n := int(u.Length / 2); n >= 0 && n < len(units) {
		units = units[:n]
	}
	return string(utf16.Decode(units))
}

// UserName returns the account's samAccountName (EffectiveName) as a Go string.
func (k *KERB_VALIDATION_INFO) UserName() string { return k.EffectiveName.String() }

// FullNameString returns the account's display name (FullName).
func (k *KERB_VALIDATION_INFO) FullNameString() string { return k.FullName.String() }

// LogonServerName returns the NetBIOS name of the KDC that issued the ticket.
func (k *KERB_VALIDATION_INFO) LogonServerName() string { return k.LogonServer.String() }

// LogonDomainNameString returns the NetBIOS name of the account's domain.
func (k *KERB_VALIDATION_INFO) LogonDomainNameString() string { return k.LogonDomainName.String() }

// UserRID returns the account's relative identifier (UserId).
func (k *KERB_VALIDATION_INFO) UserRID() uint32 { return k.UserId }

// PrimaryGroupRID returns the RID of the account's primary group.
func (k *KERB_VALIDATION_INFO) PrimaryGroupRID() uint32 { return k.PrimaryGroupId }

// AccountControl returns the userAccountControl flags carried in the PAC.
func (k *KERB_VALIDATION_INFO) AccountControl() uint32 { return k.UserAccountControl }

// DomainSID returns the account domain's SID (LogonDomainId), or nil if absent.
func (k *KERB_VALIDATION_INFO) DomainSID() *msdtyp.RPC_SID { return k.LogonDomainId }

// sidWithRID appends a relative identifier to a domain SID and renders the result
// in "S-1-5-21-…-<rid>" textual form. It returns "" if domain is nil.
func sidWithRID(domain *msdtyp.RPC_SID, rid uint32) string {
	if domain == nil {
		return ""
	}
	subs := make([]uint32, 0, len(domain.SubAuthority)+1)
	subs = append(subs, domain.SubAuthority...)
	subs = append(subs, rid)
	full := msdtyp.RPC_SID{
		Revision:            domain.Revision,
		SubAuthorityCount:   uint8(len(subs)),
		IdentifierAuthority: domain.IdentifierAuthority,
		SubAuthority:        subs,
	}
	return full.String()
}

// UserSID returns the account's SID as the domain SID with the UserId RID
// appended, or "" if the PAC has no LogonDomainId.
func (k *KERB_VALIDATION_INFO) UserSID() string {
	return sidWithRID(k.LogonDomainId, k.UserId)
}

// PrimaryGroupSID returns the primary group's SID (domain SID + PrimaryGroupId),
// or "" if the PAC has no LogonDomainId.
func (k *KERB_VALIDATION_INFO) PrimaryGroupSID() string {
	return sidWithRID(k.LogonDomainId, k.PrimaryGroupId)
}

// GroupRIDs returns the RIDs of the account's groups in its own domain (GroupIds).
func (k *KERB_VALIDATION_INFO) GroupRIDs() []uint32 {
	rids := make([]uint32, 0, len(k.GroupIds))
	for _, g := range k.GroupIds {
		rids = append(rids, g.RelativeId)
	}
	return rids
}

// GroupSIDs returns the account's account-domain group memberships (GroupIds) as
// full SID strings (domain SID + each group RID). It returns nil if the PAC has no
// LogonDomainId.
func (k *KERB_VALIDATION_INFO) GroupSIDs() []string {
	if k.LogonDomainId == nil {
		return nil
	}
	sids := make([]string, 0, len(k.GroupIds))
	for _, g := range k.GroupIds {
		sids = append(sids, sidWithRID(k.LogonDomainId, g.RelativeId))
	}
	return sids
}

// ExtraSIDs returns the ExtraSids member (SIDs from domains other than the account
// domain, e.g. universal groups and injected SIDs) as textual SID strings.
func (k *KERB_VALIDATION_INFO) ExtraSIDs() []string {
	sids := make([]string, 0, len(k.ExtraSids))
	for _, s := range k.ExtraSids {
		if s.Sid == nil {
			continue
		}
		sids = append(sids, s.Sid.String())
	}
	return sids
}
