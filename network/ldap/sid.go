package ldap

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// ParseSIDFromBytes parses raw bytes representing an SID and returns the SID string
//
// This function takes raw bytes representing an SID and parses them to extract the revision level,
// identifier authority, sub-authorities, and relative identifier. It then constructs and returns the SID string
// in the format "S-<revisionLevel>-<identifierAuthority>-<subAuthorities>-<relativeIdentifier>".
//
// Parameters:
//   - sidBytes ([]byte): The raw bytes representing the SID.
//
// Returns:
//   - string: The parsed SID string. If the input is not a valid SID or if an error occurs, the function returns an empty string.
func ParseSIDFromBytes(sidBytes []byte) string {
	debug := false

	// Ensure the SID has a valid format
	if len(sidBytes) < 8 || sidBytes[0] != 1 {
		return ""
	}

	// Extract revisionLevel
	revisionLevel := int(sidBytes[0])
	if debug {
		fmt.Printf("revisionLevel       = 0x%02x\n", revisionLevel)
		fmt.Println(sidBytes[1:])
	}
	// Extract subAuthorityCount
	subAuthorityCount := int(sidBytes[1])
	if debug {
		fmt.Printf("subAuthorityCount   = 0x%02x\n", subAuthorityCount)
		fmt.Println(sidBytes[2:])
	}

	// Extract identifierAuthority
	identifierAuthority := uint64(sidBytes[2+0]) << 40
	identifierAuthority |= uint64(sidBytes[2+1]) << 32
	identifierAuthority |= uint64(sidBytes[2+2]) << 24
	identifierAuthority |= uint64(sidBytes[2+3]) << 16
	identifierAuthority |= uint64(sidBytes[2+4]) << 8
	identifierAuthority |= uint64(sidBytes[2+5])
	if debug {
		fmt.Printf("identifierAuthority = 0x%08x\n", identifierAuthority)
		fmt.Println(sidBytes[8:])
	}

	// Extract subAuthorities
	subAuthorities := make([]string, 0)
	for k := 0; k < subAuthorityCount-1; k++ {
		subAuthority := binary.LittleEndian.Uint32(sidBytes[8+(4*k):])
		subAuthorities = append(subAuthorities, fmt.Sprintf("%d", subAuthority))
		if debug {
			fmt.Printf("subAuthority        = 0x%08x\n", subAuthority)
			fmt.Println(sidBytes[8+(4*k):])
		}
	}

	// Parse the relativeIdentifier
	relativeIdentifier := binary.LittleEndian.Uint32(sidBytes[8+((subAuthorityCount-1)*4):])
	if debug {
		fmt.Printf("relativeIdentifier  = 0x%08x\n", relativeIdentifier)
		fmt.Println(sidBytes[8+((subAuthorityCount-1)*4):])
	}

	// Construct the parsed SID
	parsedSID := fmt.Sprintf("S-%d-%d-%s-%d", revisionLevel, identifierAuthority, strings.Join(subAuthorities, "-"), relativeIdentifier)

	return parsedSID
}

// LookupSID retrieves the name of the object associated with the given SID from the LDAP directory.
//
// This function performs an LDAP search to find the object with the specified SID within all naming contexts
// of the LDAP directory. It retrieves the "name" attribute of the object and returns it.
//
// Parameters:
//   - SID (string): The Security Identifier (SID) to search for.
//
// Returns:
//   - string: The name of the object associated with the given SID if found, otherwise an empty string.
//
// Example usage:
//
//	ldapSession := &Session{}
//	SID := "S-1-5-21-3623811015-3361044348-30300820-1013"
//	name := ldapSession.LookupSID(SID)
//	if name != "" {
//	    fmt.Printf("The name associated with SID %s is %s\n", SID, name)
//	} else {
//	    fmt.Printf("No object found for SID %s\n", SID)
//	}
//
// Note:
//   - This function assumes that the Session struct has a valid connection object and that the GetAllNamingContexts
//     and QueryWholeSubtree methods are implemented correctly.
func (ldapSession *Session) LookupSID(SID string) (string, error) {
	query := fmt.Sprintf("(objectSid=%s)", SID)
	attributes := []string{"name"}

	namingContexts, err := ldapSession.GetAllNamingContexts()
	if err != nil {
		return "", fmt.Errorf("error fetching naming contexts: %w", err)
	}

	for _, namingContext := range namingContexts {
		lookupIdentityResults, err := ldapSession.QueryWholeSubtree(namingContext, query, attributes)
		if err != nil {
			return "", fmt.Errorf("error querying LDAP: %w", err)
		}
		if len(lookupIdentityResults) != 0 {
			return lookupIdentityResults[0].GetAttributeValue("name"), nil
		}
	}

	return "", fmt.Errorf("no object found for SID %s", SID)
}
