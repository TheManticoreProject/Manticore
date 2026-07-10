package ldap

import (
	"fmt"
)

// msDS-KeyCredentialLink write primitives (the "key trust" / Windows Hello for
// Business key-based credential attribute, MS-ADTS / MS-KCL).
//
// msDS-KeyCredentialLink holds one or more DN-with-binary values, each binding a
// raw public key (a KEYCREDENTIALLINK_BLOB) to the object. Writing an
// attacker-controlled key here and then authenticating with PKINIT using the
// matching private key is the Shadow Credentials technique. These helpers are the
// LDAP read/add/remove path only: composing and parsing the blob itself lives in
// windows/keycredentiallink, which is deliberately NOT imported here (it depends
// on this package's DNWithBinary type, so importing it back would create a cycle).
//
// Callers therefore pass the fully-formed DN-with-binary value string to add, and
// a predicate over the raw binary blob to decide what to keep on removal.

// KeyCredentialLinkAttribute is the LDAP attribute holding key-trust credentials.
const KeyCredentialLinkAttribute = "msDS-KeyCredentialLink"

// AddKeyCredentialLink appends a DN-with-binary value to targetDN's
// msDS-KeyCredentialLink, preserving any existing key credentials (an LDAP
// modify-add, not a replace). value must already be in DN-with-binary string form
// ("B:<hex-digit-count>:<hex>:<DN>"), e.g. produced from a
// windows/keycredentiallink.KeyCredentialLink DNWithBinary via its String().
func AddKeyCredentialLink(sess *Session, targetDN string, value string) error {
	req := NewModifyRequest(targetDN)
	req.Add(KeyCredentialLinkAttribute, []string{value})
	if err := sess.Modify(req); err != nil {
		return fmt.Errorf("keycredentiallink: add %s on %q: %w", KeyCredentialLinkAttribute, targetDN, err)
	}
	return nil
}

// AddKeyCredentialLinkDNBinary is AddKeyCredentialLink taking a *DNWithBinary
// instead of its string form; it marshals the value and appends it.
func AddKeyCredentialLinkDNBinary(sess *Session, targetDN string, value *DNWithBinary) error {
	if value == nil {
		return fmt.Errorf("keycredentiallink: nil DNWithBinary value")
	}
	return AddKeyCredentialLink(sess, targetDN, value.String())
}

// GetKeyCredentialLinks reads targetDN's current msDS-KeyCredentialLink values and
// returns them in their DN-with-binary string form. The slice is empty (nil) when
// the attribute is unset.
func GetKeyCredentialLinks(sess *Session, targetDN string) ([]string, error) {
	entries, err := sess.QueryBaseObject(targetDN, "(objectClass=*)", []string{KeyCredentialLinkAttribute})
	if err != nil {
		return nil, fmt.Errorf("keycredentiallink: read %s on %q: %w", KeyCredentialLinkAttribute, targetDN, err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("keycredentiallink: no entry returned for %q", targetDN)
	}
	return entries[0].GetAttributeValues(KeyCredentialLinkAttribute), nil
}

// RemoveKeyCredentialLink reads targetDN's current msDS-KeyCredentialLink values,
// parses the binary blob out of each DN-with-binary value, and rewrites the
// attribute keeping only the values for which keep(binaryBlob) returns true (or
// flushes the attribute if none remain). Values that are not well-formed
// DN-with-binary are conservatively kept.
//
// The caller MUST match on the stable KeyID inside the blob, not on the exact
// value string. Active Directory rewrites parts of the stored
// KEYCREDENTIALLINK_BLOB (e.g. the key's last-logon timestamp) and normalizes the
// DN portion of the DN-with-binary value, so the bytes read back never match the
// bytes written and an exact-value delete is unreliable. The KeyID (entry type
// 0x01) is assigned at creation and does not change, so decode it from binaryBlob
// (with windows/keycredentiallink) and compare that. keep returns true to retain a
// value, false to drop it.
func RemoveKeyCredentialLink(sess *Session, targetDN string, keep func(binaryBlob []byte) bool) error {
	entries, err := sess.QueryBaseObject(targetDN, "(objectClass=*)", []string{KeyCredentialLinkAttribute})
	if err != nil {
		return fmt.Errorf("keycredentiallink: read %s on %q: %w", KeyCredentialLinkAttribute, targetDN, err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("keycredentiallink: no entry returned for %q", targetDN)
	}

	values := entries[0].GetAttributeValues(KeyCredentialLinkAttribute)
	survivors, removed := FilterKeyCredentialLinks(values, keep)
	if !removed {
		return fmt.Errorf("keycredentiallink: no matching value to remove on %q", targetDN)
	}

	if len(survivors) == 0 {
		return sess.FlushAttributeValues(targetDN, KeyCredentialLinkAttribute)
	}
	req := NewModifyRequest(targetDN)
	req.Replace(KeyCredentialLinkAttribute, survivors)
	if err := sess.Modify(req); err != nil {
		return fmt.Errorf("keycredentiallink: rewrite %s on %q: %w", KeyCredentialLinkAttribute, targetDN, err)
	}
	return nil
}

// FilterKeyCredentialLinks applies keep to the binary blob of each DN-with-binary
// value and returns the values to retain plus whether any value was dropped. A
// value that does not parse as DN-with-binary is kept unconditionally. This is the
// pure, server-free core of RemoveKeyCredentialLink (and is unit-tested directly).
func FilterKeyCredentialLinks(values []string, keep func(binaryBlob []byte) bool) (survivors []string, removed bool) {
	for _, v := range values {
		var dnb DNWithBinary
		if _, err := dnb.Unmarshal([]byte(v)); err != nil {
			// Not a DN-with-binary value: keep it untouched.
			survivors = append(survivors, v)
			continue
		}
		if keep(dnb.BinaryData) {
			survivors = append(survivors, v)
			continue
		}
		removed = true
	}
	return survivors, removed
}
