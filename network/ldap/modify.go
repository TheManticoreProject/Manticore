package ldap

import (
	"fmt"

	goldapv3 "github.com/go-ldap/ldap/v3"
)

type ModifyRequest struct {
	// Distinguished names
	DistinguishedName string
	// Actions to perform on attributes
	Attributes []*Action
	// LDAP controls
	Controls []goldapv3.Control
}

type Action struct {
	Attribute string
	// Values
	AddValues       []string
	DelValues       []string
	ReplaceValues   []string
	IncrementValues []string
	OverwriteValues []string
}

// NewModifyRequest creates a new instance of ModifyRequest with the specified distinguished name.
//
// Parameters:
//   - distinguishedName: A string representing the distinguished name (DN) of the LDAP entry to be modified.
//
// Returns:
//   - A pointer to a newly created ModifyRequest instance with the specified distinguished name and an empty list of attributes.
//
// The function initializes the ModifyRequest struct with the provided distinguished name and an empty slice of attributes.
// This struct can then be used to add, delete, increment, or replace attributes for the specified LDAP entry.
//
// Example usage:
//
//	modifyRequest := NewModifyRequest("cn=John Doe,dc=example,dc=com")
//	modifyRequest.Add("description", []string{"New description"})
//	modifyRequest.Delete("telephoneNumber", []string{"123-456-7890"})
//	modifyRequest.Replace("mail", []string{"john.doe@example.com"})
//	err := ldapSession.Modify("cn=John Doe,dc=example,dc=com", modifyRequest)
//	if err != nil {
//		log.Fatalf("Failed to modify LDAP entry: %s", err)
//	}
func NewModifyRequest(distinguishedName string) *ModifyRequest {
	return &ModifyRequest{
		DistinguishedName: distinguishedName,
		Attributes:        make([]*Action, 0),
	}
}

// AddControl adds a control to the ModifyRequest.
//
// Parameters:
//   - control: A pointer to a goldapv3.Control struct representing the control to be added.
func (req *ModifyRequest) AddControl(control goldapv3.Control) {
	if req.Controls == nil {
		req.Controls = make([]goldapv3.Control, 0)
	}
	req.Controls = append(req.Controls, control)
}

// Add adds a new attribute and its values to the ModifyRequest.
//
// Parameters:
//   - attrType: A string representing the type of the attribute to be added.
//   - attrVals: A slice of strings representing the values of the attribute to be added.
//
// The function appends a new Attribute struct to the Attributes slice of the ModifyRequest. The Attribute struct
// contains the provided attribute type and values. This function can be used to add new attributes to an LDAP entry
// as part of a modify operation.
//
// Example usage:
//
//	modifyRequest := NewModifyRequest("cn=John Doe,dc=example,dc=com")
//	modifyRequest.Add("description", []string{"New description"})
//	modifyRequest.Add("telephoneNumber", []string{"123-456-7890"})
//
// In this example, the modifyRequest will contain two attributes to be added to the LDAP entry with the distinguished
// name "cn=John Doe,dc=example,dc=com". The first attribute is "description" with the value "New description", and the
// second attribute is "telephoneNumber" with the value "123-456-7890".
func (req *ModifyRequest) Add(attrType string, attrVals []string) {
	req.Attributes = append(req.Attributes, &Action{
		Attribute: attrType,
		AddValues: attrVals,
	})
}

// Delete removes an attribute and its values from the ModifyRequest.
//
// Parameters:
//   - attrType: A string representing the type of the attribute to be removed.
//   - attrVals: A slice of strings representing the values of the attribute to be removed.
//
// The function appends a new Attribute struct to the Attributes slice of the ModifyRequest. The Attribute struct
// contains the provided attribute type and values. This function can be used to remove attributes from an LDAP entry
// as part of a modify operation.
//
// Example usage:
//
//	modifyRequest := NewModifyRequest("cn=John Doe,dc=example,dc=com")
//	modifyRequest.Delete("description", []string{"Old description"})
//	modifyRequest.Delete("telephoneNumber", []string{"123-456-7890"})
//
// In this example, the modifyRequest will contain two attributes to be removed from the LDAP entry with the distinguished
// name "cn=John Doe,dc=example,dc=com". The first attribute is "description" with the value "Old description", and the
// second attribute is "telephoneNumber" with the value "123-456-7890".
//
// Note:
//   - If the specified attribute and values do not exist in the LDAP entry, the delete operation will have no effect.
func (req *ModifyRequest) Delete(attrType string, attrVals []string) {
	req.Attributes = append(req.Attributes, &Action{
		Attribute: attrType,
		DelValues: attrVals,
	})
}

// Increment adds an increment operation for an attribute to the ModifyRequest.
//
// Parameters:
//   - attrType: A string representing the type of the attribute to be incremented.
//   - attrVal: A string representing the value to increment the attribute by.
//
// The function appends a new Attribute struct to the Attributes slice of the ModifyRequest. The Attribute struct
// contains the provided attribute type and the increment value. This function can be used to increment the value
// of an attribute in an LDAP entry as part of a modify operation.
//
// Example usage:
//
//	modifyRequest := NewModifyRequest("cn=John Doe,dc=example,dc=com")
//	modifyRequest.Increment("loginCount", "1")
//
// In this example, the modifyRequest will contain an increment operation for the "loginCount" attribute with the
// increment value "1". This means that the "loginCount" attribute of the LDAP entry with the distinguished name
// "cn=John Doe,dc=example,dc=com" will be incremented by 1.
//
// Note:
//   - The attribute type must support increment operations for this function to have an effect.
//   - The increment value should be a valid string representation of the increment amount.
func (req *ModifyRequest) Increment(attrType string, attrVal string) {
	req.Attributes = append(req.Attributes, &Action{
		Attribute:       attrType,
		IncrementValues: []string{attrVal},
	})
}

// Replace sets a replace operation for an attribute in the ModifyRequest.
//
// Parameters:
//   - attrType: A string representing the type of the attribute to be replaced.
//   - attrVals: A slice of strings representing the new values for the attribute.
//
// The function appends a new Attribute struct to the Attributes slice of the ModifyRequest. The Attribute struct
// contains the provided attribute type and the new values. This function can be used to replace the values of an
// attribute in an LDAP entry as part of a modify operation.
//
// Example usage:
//
//	modifyRequest := NewModifyRequest("cn=John Doe,dc=example,dc=com")
//	modifyRequest.Replace("description", []string{"New description"})
//
// In this example, the modifyRequest will contain a replace operation for the "description" attribute with the
// new value "New description". This means that the "description" attribute of the LDAP entry with the distinguished
// name "cn=John Doe,dc=example,dc=com" will be replaced with the new value.
//
// Note:
//   - If the specified attribute does not exist in the LDAP entry, the replace operation will add the attribute
//     with the provided values.
//   - The replace operation will remove any existing values for the specified attribute and set the new values.
func (req *ModifyRequest) Replace(attrType string, attrVals []string) {
	req.Attributes = append(req.Attributes, &Action{
		Attribute:     attrType,
		ReplaceValues: attrVals,
	})
}

// Overwrite sets a overwrite operation for an attribute in the ModifyRequest.
//
// Parameters:
//   - attrType: A string representing the type of the attribute to be overwritten.
//   - attrVals: A slice of strings representing the new values for the attribute.
func (ldapSession *Session) Overwrite(distinguishedName string, attrType string, attrVals []string) error {
	// Overwrite the content of the attribute with the new values
	m1 := goldapv3.NewModifyRequest(distinguishedName, nil)
	m1.Delete(attrType, nil)

	// Execute the modify request
	err := ldapSession.connection.Modify(m1)
	if err != nil {
		// Check if the error is of type NO_ATTRIBUTE_OR_VAL (code 16)
		if ldapErr, ok := err.(*goldapv3.Error); ok && ldapErr.ResultCode == 16 {
			// Ignore NO_ATTRIBUTE_OR_VAL error as it just means the attribute doesn't exist yet
			// LDAP Result Code 16 "No Such Attribute": 00002076: AtrErr: DSID-030F1BC2, #1:
			// 0: 00002076: DSID-030F1BC2, problem 1001 (NO_ATTRIBUTE_OR_VAL), data 0, Att d (description)
			// This is expected behavior, so we can ignore it
		} else {
			// Return other types of errors
			return fmt.Errorf("error flushing attribute: %s", err)
		}
	}

	m2 := goldapv3.NewModifyRequest(distinguishedName, nil)
	m2.Add(attrType, attrVals)

	// Execute the modify request
	err = ldapSession.connection.Modify(m2)
	if err != nil {
		return fmt.Errorf("error replacing attribute: %s", err)
	}

	return nil
}

// Modify performs an LDAP modify operation on the specified distinguished name (DN) using the provided ModifyRequest.
//
// Parameters:
//   - distinguishedName: A string representing the distinguished name (DN) of the LDAP entry to be modified.
//   - modifyRequest: A pointer to a ModifyRequest struct containing the modifications to be applied.
//
// Returns:
//   - An error object if the modify operation fails, otherwise nil.
//
// The function creates a new LDAP modify request using the provided distinguished name and the attributes from the
// ModifyRequest. It then performs the modify operation using the established LDAP connection in the Session struct.
//
// Example usage:
//
//	session := &Session{}
//	err := session.InitSession("ldap.example.com", 389, false, true, "EXAMPLE", "user", "password", false)
//	if err != nil {
//		logger.Error(fmt.Sprintf("Failed to initialize session: %s", err))
//		return
//	}
//	success, err := session.Connect()
//	if !success {
//		logger.Warn(fmt.Sprintf("Failed to connect to LDAP server: %s", err))
//		return
//	}
//
//	modifyRequest := NewModifyRequest("cn=John Doe,dc=example,dc=com")
//	modifyRequest.Replace("description", []string{"New description"})
//	err = session.Modify("cn=John Doe,dc=example,dc=com", modifyRequest)
//	if err != nil {
//		logger.Error(fmt.Sprintf("Failed to modify LDAP entry: %s", err))
//	} else {
//		logger.Info("Successfully modified LDAP entry")
//	}
//
// Note:
//   - The ModifyRequest struct should contain the desired modifications, such as add, delete, or replace operations
//     for specific attributes.
//   - Ensure that the LDAP connection is properly established before calling this function.
func (ldapSession *Session) Modify(modifyRequest *ModifyRequest) error {
	m := goldapv3.NewModifyRequest(modifyRequest.DistinguishedName, modifyRequest.Controls)
	for _, attribute := range modifyRequest.Attributes {
		if len(attribute.AddValues) > 0 {
			// Add the new values to the attribute
			m.Add(attribute.Attribute, attribute.AddValues)
		} else if len(attribute.DelValues) > 0 {
			// Delete the values from the attribute
			m.Delete(attribute.Attribute, attribute.DelValues)
		} else if len(attribute.ReplaceValues) > 0 {
			// Replace the content of the attribute with the new values
			m.Replace(attribute.Attribute, attribute.ReplaceValues)
		} else if len(attribute.IncrementValues) > 0 {
			// Increment the value of the attribute
			m.Increment(attribute.Attribute, attribute.IncrementValues[0])
		}
	}
	return ldapSession.connection.Modify(m)
}

// AddStringToAttributeList adds a string to an attribute list
//
// Parameters:
//   - dn: A string representing the distinguished name (DN) of the LDAP entry to be modified.
//   - attributeName: A string representing the name of the attribute to be modified.
//   - valueToAdd: A string representing the value to be added to the attribute.
//
// Returns:
//   - An error object if the add operation fails, otherwise nil.
func (ldapSession *Session) AddStringToAttributeList(dn string, attributeName string, valueToAdd string) error {
	// Create a modify request
	modifyRequest := goldapv3.NewModifyRequest(dn, nil)
	modifyRequest.Add(attributeName, []string{valueToAdd})

	// Execute the modify request
	err := ldapSession.connection.Modify(modifyRequest)
	if err != nil {
		return fmt.Errorf("error adding value to attribute: %s", err)
	}

	return nil
}

// FlushAttribute flushes the attribute by deleting it
//
// Parameters:
//   - dn: A string representing the distinguished name (DN) of the LDAP entry to be modified.
//   - attributeName: A string representing the name of the attribute to be flushed.
//
// Returns:
//   - An error object if the flush operation fails, otherwise nil.
func (ldapSession *Session) FlushAttribute(dn string, attributeName string) error {
	// Create a modify request
	modifyRequest := goldapv3.NewModifyRequest(dn, nil)
	modifyRequest.Delete(attributeName, nil)

	// Execute the modify request
	err := ldapSession.connection.Modify(modifyRequest)
	if err != nil {
		return fmt.Errorf("error flushing attribute: %s", err)
	}

	return nil
}
