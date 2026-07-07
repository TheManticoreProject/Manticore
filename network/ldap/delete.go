package ldap

import (
	goldapv3 "github.com/go-ldap/ldap/v3"
)

// Delete performs an LDAP delete operation, removing the entry with the specified distinguished name.
//
// Parameters:
//   - distinguishedName: A string representing the distinguished name (DN) of the LDAP entry to be deleted.
//
// Returns:
//   - An error object if the delete operation fails, otherwise nil.
//
// The function creates a new LDAP delete request for the provided distinguished name and performs the
// delete operation using the established LDAP connection.
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
//	err = session.Delete("cn=John Doe,dc=example,dc=com")
//	if err != nil {
//		logger.Error(fmt.Sprintf("Failed to delete LDAP entry: %s", err))
//	} else {
//		logger.Info("Successfully deleted LDAP entry")
//	}
//
// Note:
//   - A standard delete operation only removes leaf entries. To delete an entry together with its
//     descendants, use DeleteWithControls with the subtree-delete control.
//   - Ensure that the LDAP connection is properly established before calling this function.
func (ldapSession *Session) Delete(distinguishedName string) error {
	return ldapSession.connection.Del(goldapv3.NewDelRequest(distinguishedName, nil))
}

// DeleteWithControls performs an LDAP delete operation with the provided controls.
//
// Parameters:
//   - distinguishedName: A string representing the distinguished name (DN) of the LDAP entry to be deleted.
//   - controls: A slice of goldapv3.Control to send with the delete request.
//
// Returns:
//   - An error object if the delete operation fails, otherwise nil.
//
// This is the controlled form of Delete. It is useful, for example, to attach the subtree-delete control
// (OID 1.2.840.113556.1.4.805) so that an entry is removed together with all of its descendants.
//
// Example usage:
//
//	controls := NewControlsWithOIDs([]string{"1.2.840.113556.1.4.805"}, true)
//	err := session.DeleteWithControls("ou=Stale,dc=example,dc=com", controls)
//	if err != nil {
//		logger.Error(fmt.Sprintf("Failed to delete LDAP subtree: %s", err))
//	}
func (ldapSession *Session) DeleteWithControls(distinguishedName string, controls []goldapv3.Control) error {
	return ldapSession.connection.Del(goldapv3.NewDelRequest(distinguishedName, controls))
}
