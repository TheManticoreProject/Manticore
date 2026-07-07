package ldap

import (
	goldapv3 "github.com/go-ldap/ldap/v3"
)

// AddRequest represents an LDAP add operation that creates a new entry.
//
// Fields:
//   - DistinguishedName: the distinguished name (DN) of the entry to be created.
//   - Attributes: the ordered list of attributes for the new entry. By convention
//     the objectClass attribute is added first.
//   - Controls: optional LDAP controls to send with the request.
type AddRequest struct {
	// Distinguished name
	DistinguishedName string
	// Attributes of the new entry
	Attributes []*AddAttribute
	// LDAP controls
	Controls []goldapv3.Control
}

// AddAttribute is a single attribute (type and values) of an entry being created.
//
// Binary attribute values are supported: pass each value as string(bytes). The
// underlying LDAP encoder writes attribute values as raw octet strings, so the
// bytes are transmitted verbatim.
type AddAttribute struct {
	Type string
	Vals []string
}

// NewAddRequest creates a new AddRequest for the specified distinguished name.
//
// Parameters:
//   - distinguishedName: A string representing the distinguished name (DN) of the LDAP entry to be created.
//
// Returns:
//   - A pointer to a newly created AddRequest instance with the specified distinguished name and an empty
//     list of attributes.
//
// Example usage:
//
//	addRequest := NewAddRequest("cn=John Doe,dc=example,dc=com")
//	addRequest.Attribute("objectClass", []string{"top", "person"})
//	addRequest.Attribute("cn", []string{"John Doe"})
//	addRequest.Attribute("sn", []string{"Doe"})
//	err := ldapSession.Add(addRequest)
//	if err != nil {
//		log.Fatalf("Failed to add LDAP entry: %s", err)
//	}
func NewAddRequest(distinguishedName string) *AddRequest {
	return &AddRequest{
		DistinguishedName: distinguishedName,
		Attributes:        make([]*AddAttribute, 0),
	}
}

// AddControl adds a control to the AddRequest.
//
// Parameters:
//   - control: A goldapv3.Control interface representing the control to be added.
func (req *AddRequest) AddControl(control goldapv3.Control) {
	if req.Controls == nil {
		req.Controls = make([]goldapv3.Control, 0)
	}
	req.Controls = append(req.Controls, control)
}

// Attribute appends an attribute and its values to the AddRequest.
//
// Parameters:
//   - attrType: A string representing the type of the attribute to be added.
//   - attrVals: A slice of strings representing the values of the attribute.
//
// The function appends a new AddAttribute to the Attributes slice of the AddRequest. Attributes are kept
// in the order in which they are appended. Binary values are supported by passing string(bytes).
//
// Example usage:
//
//	addRequest := NewAddRequest("cn=John Doe,dc=example,dc=com")
//	addRequest.Attribute("objectClass", []string{"top", "person"})
//	addRequest.Attribute("cn", []string{"John Doe"})
func (req *AddRequest) Attribute(attrType string, attrVals []string) {
	req.Attributes = append(req.Attributes, &AddAttribute{
		Type: attrType,
		Vals: attrVals,
	})
}

// Add performs an LDAP add operation using the provided AddRequest, creating a new entry on the server.
//
// Parameters:
//   - addRequest: A pointer to an AddRequest struct describing the entry to create.
//
// Returns:
//   - An error object if the add operation fails, otherwise nil.
//
// The function creates a new LDAP add request using the distinguished name, attributes, and controls from
// the provided AddRequest, then performs the add operation using the established LDAP connection.
//
// Example usage:
//
//	session, err := NewSession("ldap.example.com", 389, credentials, false, false)
//	if err != nil {
//		logger.Error(fmt.Sprintf("Failed to create session: %s", err))
//		return
//	}
//	success, err := session.Connect()
//	if !success {
//		logger.Warn(fmt.Sprintf("Failed to connect to LDAP server: %s", err))
//		return
//	}
//
//	addRequest := NewAddRequest("cn=John Doe,dc=example,dc=com")
//	addRequest.Attribute("objectClass", []string{"top", "person"})
//	addRequest.Attribute("cn", []string{"John Doe"})
//	addRequest.Attribute("sn", []string{"Doe"})
//	err = session.Add(addRequest)
//	if err != nil {
//		logger.Error(fmt.Sprintf("Failed to add LDAP entry: %s", err))
//	} else {
//		logger.Info("Successfully added LDAP entry")
//	}
//
// Note:
//   - Ensure that the LDAP connection is properly established before calling this function.
func (ldapSession *Session) Add(addRequest *AddRequest) error {
	// Create a new add request
	a := goldapv3.NewAddRequest(addRequest.DistinguishedName, addRequest.Controls)

	// Add the attributes to the add request, preserving their order
	for _, attribute := range addRequest.Attributes {
		a.Attribute(attribute.Type, attribute.Vals)
	}

	return ldapSession.connection.Add(a)
}
