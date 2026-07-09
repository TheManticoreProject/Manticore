package ldap

import (
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
)

// tlsServerEndPointPrefix is the RFC 5929 §4 channel-binding type prefix that
// precedes the certificate hash in the GSS-API application data.
const tlsServerEndPointPrefix = "tls-server-end-point:"

// tlsServerEndPointCBT builds the RFC 5929 "tls-server-end-point" channel-binding
// token for an LDAPS server certificate: the ASCII prefix "tls-server-end-point:"
// followed by the hash of the server's DER-encoded certificate. This is the
// application-data component of the GSS-API channel bindings that Active Directory
// validates for LDAP channel binding (a.k.a. EPA / "Extended Protection for
// Authentication").
func tlsServerEndPointCBT(cert *x509.Certificate) []byte {
	return append([]byte(tlsServerEndPointPrefix), certificateHash(cert)...)
}

// certificateHash hashes a certificate's raw DER using the algorithm RFC 5929 §4.1
// prescribes: the certificate's own signature hash, except that MD5 and SHA-1 (and
// any unknown/unhashed signature) are upgraded to SHA-256.
func certificateHash(cert *x509.Certificate) []byte {
	switch cert.SignatureAlgorithm {
	case x509.SHA384WithRSA, x509.ECDSAWithSHA384, x509.SHA384WithRSAPSS:
		sum := sha512.Sum384(cert.Raw)
		return sum[:]
	case x509.SHA512WithRSA, x509.ECDSAWithSHA512, x509.SHA512WithRSAPSS:
		sum := sha512.Sum512(cert.Raw)
		return sum[:]
	default:
		// SHA-256 and, per RFC 5929, the MD5/SHA-1 and undefined cases.
		sum := sha256.Sum256(cert.Raw)
		return sum[:]
	}
}
