package spnego

import (
	"encoding/asn1"
	"fmt"
)

// OIDs for various authentication mechanisms
var (
	// SPNEGO OID: 1.3.6.1.5.5.2
	// iso.org.dod.internet.security.mechanism.snego
	// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-spng/94ccc4f8-d224-495f-8d31-4f58d1af598e
	SpnegoOID = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 2}

	// SPNegoEx OID: 1.3.6.1.4.1.311.2.2.30
	// iso.org.dod.internet.private.enterprise.Microsoft.security.mechanisms.SPNegoEx
	// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-spng/f5edf48c-57cc-4c61-bff9-ee19b9cd059e
	SPNegoExOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 2, 30}

	// NTLM OID: 1.3.6.1.4.1.311.2.2.10
	// iso.org.dod.internet.private.enterprise.Microsoft.security.mechanisms.NTLM
	// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/e21c0b07-8662-41b7-8853-2b9184eab0db
	NtlmOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 2, 10}

	// Kerberos OID: 1.2.840.113554.1.2.2
	// iso.org.dod.internet.private.enterprise.Microsoft.security.mechanisms.Kerberos
	// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-spng/211417c4-11ef-46c0-a8fb-f178a51c2088
	KerberosOID = asn1.ObjectIdentifier{1, 2, 840, 113554, 1, 2, 2}

	// MSKerberosOID: 1.2.840.48018.1.2.2 — Microsoft's legacy "MS KRB5" mechanism
	// OID. Windows advertises this in the SPNEGO mechTypes for Kerberos, and it is
	// what Windows RPC clients send, so it is used for the RPC Negotiate bind.
	MSKerberosOID = asn1.ObjectIdentifier{1, 2, 840, 48018, 1, 2, 2}
)

// OIDtoString converts an OID to a human-readable string representation
func OIDtoString(oid asn1.ObjectIdentifier) string {
	oidString := oid.String()

	switch {
	case oid.Equal(SpnegoOID):
		return fmt.Sprintf("%s (SPNego)", oidString)
	case oid.Equal(SPNegoExOID):
		return fmt.Sprintf("%s (SPNegoEx)", oidString)
	case oid.Equal(NtlmOID):
		return fmt.Sprintf("%s (NTLMSSP - Microsoft NTLM Security Support Provider)", oidString)
	case oid.Equal(KerberosOID):
		return fmt.Sprintf("%s (Kerberos)", oidString)
	default:
		return oidString
	}
}
