package ldap

import (
	"fmt"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"
)

// Security Information constants for NT Security Descriptor flags, as carried in the
// LDAP_SERVER_SD_FLAGS_OID control value.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/3888c2b7-35b9-45b7-afeb-b772aa932dd0
//
// SACL_SECURITY_INFORMATION is the one to think about before requesting: reading a
// SACL needs the SE_SECURITY_NAME privilege, and a domain controller answers a client
// that lacks it with the nTSecurityDescriptor attribute absent from the entry rather
// than with an error. A caller that only needs the owner, the group and the DACL
// should ask for exactly those, which is what SECURITY_INFORMATION_DEFAULT is.
const (
	OWNER_SECURITY_INFORMATION = 0x1 // Owner identifier of the object
	GROUP_SECURITY_INFORMATION = 0x2 // Primary group identifier
	DACL_SECURITY_INFORMATION  = 0x4 // Discretionary access control list (DACL) of the object
	SACL_SECURITY_INFORMATION  = 0x8 // System access control list (SACL) of the object

	// SECURITY_INFORMATION_DEFAULT is the everything-but-the-SACL set, which is what
	// an unprivileged read of a security descriptor has to ask for.
	SECURITY_INFORMATION_DEFAULT = OWNER_SECURITY_INFORMATION | GROUP_SECURITY_INFORMATION | DACL_SECURITY_INFORMATION
)

type ControlMicrosoftSDFlags struct {
	Criticality  bool
	ControlValue int32
}

func (c *ControlMicrosoftSDFlags) GetControlType() string {
	// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/3888c2b7-35b9-45b7-afeb-b772aa932dd0
	return "1.2.840.113556.1.4.801"
}

func (c *ControlMicrosoftSDFlags) Encode() *ber.Packet {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "1.2.840.113556.1.4.801", "Control Type"))
	if c.Criticality {
		packet.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, c.Criticality, "Criticality"))
	}
	p2 := ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, nil, "Control Value(SDFlags)")
	seq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "SDFlags")
	seq.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, c.ControlValue, "Flags"))
	p2.AppendChild(seq)
	packet.AppendChild(p2)
	return packet
}

func (c *ControlMicrosoftSDFlags) String() string {
	return fmt.Sprintf(
		"Control Type: %s (%q)",
		"SD Flags - Microsoft",
		"1.2.840.113556.1.4.801")
}

// NewControlMicrosoftSDFlags returns a ControlMicrosoftSDFlags control
func NewControlMicrosoftSDFlags() *ControlMicrosoftSDFlags {
	return &ControlMicrosoftSDFlags{}
}

// GetNtSecurityDescriptorOf retrieves the NT Security Descriptor of an LDAP entry.
//
// Parameters:
//   - dn: A string representing the distinguished name (DN) of the LDAP entry to retrieve the NT Security Descriptor from.
//
// Returns:
//   - A string representing the NT Security Descriptor of the LDAP entry.
//   - An error if the search operation fails.
func (s *Session) GetNtSecurityDescriptorOf(distinguishedName string) (string, error) {
	// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/3888c2b7-35b9-45b7-afeb-b772aa932dd0

	control := &ControlMicrosoftSDFlags{
		Criticality:  false,
		ControlValue: int32(SECURITY_INFORMATION_DEFAULT),
	}

	searchRequest := ldap.NewSearchRequest(
		distinguishedName,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(distinguishedName=%s)", distinguishedName),
		[]string{"nTSecurityDescriptor"},
		[]ldap.Control{control},
	)

	searchResult, err := s.connection.SearchWithPaging(searchRequest, 1000)
	if err != nil {
		return "", fmt.Errorf("error searching for nTSecurityDescriptor: %w", err)
	}

	if len(searchResult.Entries) == 0 {
		return "", fmt.Errorf("no entry returned for %q; access may be denied", distinguishedName)
	}

	ntsd := searchResult.Entries[0].GetEqualFoldRawAttributeValue("nTSecurityDescriptor")

	return string(ntsd), nil
}
