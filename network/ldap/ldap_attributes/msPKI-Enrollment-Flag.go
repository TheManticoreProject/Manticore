package ldap_attributes

import "fmt"

type MSPKIEnrollmentFlag uint32

// msPKI-Enrollment-Flag Attribute
// Src: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-crtd/ec71fd43-61c2-407b-83c9-b52272dec8a1
const (
	// CT_FLAG_INCLUDE_SYMMETRIC_ALGORITHMS
	// This flag instructs the client and server to include a Secure/Multipurpose Internet Mail Extensions (S/MIME)
	// certificate extension, as specified in [RFC4262], in the request and in the issued certificate.
	MSPKI_ENROLLMENT_FLAG_INCLUDE_SYMMETRIC_ALGORITHMS = 0x00000001

	// CT_FLAG_PEND_ALL_REQUESTS
	// This flag instructs the CA to put all requests in a pending state.
	MSPKI_ENROLLMENT_FLAG_PEND_ALL_REQUESTS = 0x00000002

	// CT_FLAG_PUBLISH_TO_KRA_CONTAINER
	// This flag instructs the CA to publish the issued certificate to the key recovery agent (KRA) container
	// in Active Directory, as specified in [MS-ADTS].
	MSPKI_ENROLLMENT_FLAG_PUBLISH_TO_KRA_CONTAINER = 0x00000004

	// CT_FLAG_PUBLISH_TO_DS
	// This flag instructs CA servers to append the issued certificate to the userCertificate attribute,
	// as specified in [RFC4523], on the user object in Active Directory. The server processing rules for
	// this flag are specified in [MS-WCCE] section 3.2.2.6.2.1.4.5.6.
	MSPKI_ENROLLMENT_FLAG_PUBLISH_TO_DS = 0x00000008

	// CT_FLAG_AUTO_ENROLLMENT_CHECK_USER_DS_CERTIFICATE
	// This flag instructs clients not to do autoenrollment for a certificate based on this template if the
	// user's userCertificate attribute (specified in [RFC4523]) in Active Directory has a valid certificate
	// based on the same template.
	MSPKI_ENROLLMENT_FLAG_AUTO_ENROLLMENT_CHECK_USER_DS_CERTIFICATE = 0x00000010

	// CT_FLAG_AUTO_ENROLLMENT
	// This flag instructs clients to perform autoenrollment for the specified template.
	MSPKI_ENROLLMENT_FLAG_AUTO_ENROLLMENT = 0x00000020

	// CT_FLAG_PREVIOUS_APPROVAL_VALIDATE_REENROLLMENT
	// This flag instructs clients to sign the renewal request using the private key of the existing certificate.
	// For more information, see [MS-WCCE] section 3.2.2.6.2.1.4.5.6. This flag also instructs the CA to process
	// the renewal requests as specified in [MS-WCCE] section 3.2.2.6.2.1.4.5.6.
	MSPKI_ENROLLMENT_FLAG_PREVIOUS_APPROVAL_VALIDATE_REENROLLMENT = 0x00000040

	// CT_FLAG_USER_INTERACTION_REQUIRED
	// This flag instructs the client to obtain user consent before attempting to enroll for a certificate
	// that is based on the specified template.
	MSPKI_ENROLLMENT_FLAG_USER_INTERACTION_REQUIRED = 0x00000100

	// CT_FLAG_REMOVE_INVALID_CERTIFICATE_FROM_PERSONAL_STORE
	// This flag instructs the autoenrollment client to delete any certificates that are no longer needed
	// based on the specific template from the local certificate storage. For information about autoenrollment
	// and the local certificate storage, see [MS-CERSOD] section 2.1.2.2.2.
	MSPKI_ENROLLMENT_FLAG_REMOVE_INVALID_CERTIFICATE_FROM_PERSONAL_STORE = 0x00000400

	// CT_FLAG_ALLOW_ENROLL_ON_BEHALF_OF
	// This flag instructs the server to allow enroll on behalf of (EOBO) functionality.
	MSPKI_ENROLLMENT_FLAG_ALLOW_ENROLL_ON_BEHALF_OF = 0x00000800

	// CT_FLAG_ADD_OCSP_NOCHECK
	// This flag instructs the server to not include revocation information and add the id-pkix-ocsp-nocheck
	// extension, as specified in [RFC2560] section 4.2.2.2.1, to the certificate that is issued.
	MSPKI_ENROLLMENT_FLAG_ADD_OCSP_NOCHECK = 0x00001000

	// CT_FLAG_ENABLE_KEY_REUSE_ON_NT_TOKEN_KEYSET_STORAGE_FULL
	// This flag instructs the client to reuse the private key for a smart card–based certificate renewal
	// if it is unable to create a new private key on the card.
	MSPKI_ENROLLMENT_FLAG_ENABLE_KEY_REUSE_ON_NT_TOKEN_KEYSET_STORAGE_FULL = 0x00002000

	// CT_FLAG_NOREVOCATIONINFOINISSUEDCERTS
	// This flag instructs the server to not include revocation information in the issued certificate.
	MSPKI_ENROLLMENT_FLAG_NOREVOCATIONINFOINISSUEDCERTS = 0x00004000

	// CT_FLAG_INCLUDE_BASIC_CONSTRAINTS_FOR_EE_CERTS
	// This flag instructs the server to include Basic Constraints extension (specified in [RFC3280]
	// section 4.2.1.10) in the end entity certificates.
	MSPKI_ENROLLMENT_FLAG_INCLUDE_BASIC_CONSTRAINTS_FOR_EE_CERTS = 0x00008000

	// CT_FLAG_ALLOW_PREVIOUS_APPROVAL_KEYBASEDRENEWAL_VALIDATE_REENROLLMENT
	// This flag instructs the CA to ignore the requirement for Enroll permissions on the template when
	// processing renewal requests as specified in [MS-WCCE] section 3.2.2.6.2.1.4.5.6.
	MSPKI_ENROLLMENT_FLAG_ALLOW_PREVIOUS_APPROVAL_KEYBASEDRENEWAL_VALIDATE_REENROLLMENT = 0x00010000

	// CT_FLAG_ISSUANCE_POLICIES_FROM_REQUEST
	// This flag indicates that the certificate issuance policies to be included in the issued certificate
	// come from the request rather than from the template. The template contains a list of all of the issuance
	// policies that the request is allowed to specify; if the request contains policies that are not listed
	// in the template, then the request is rejected. For the processing rules of this flag, see [MS-WCCE]
	// section 3.2.2.6.2.1.4.5.8.
	MSPKI_ENROLLMENT_FLAG_ISSUANCE_POLICIES_FROM_REQUEST = 0x00020000

	// CT_FLAG_SKIP_AUTO_RENEWAL
	// This flag indicates that the certificate should not be auto-renewed, although it has a valid template.
	MSPKI_ENROLLMENT_FLAG_SKIP_AUTO_RENEWAL = 0x00040000

	// CT_FLAG_NO_SECURITY_EXTENSION
	// This flag instructs the CA to not include the security extension szOID_NTDS_CA_SECURITY_EXT
	// (OID:1.3.6.1.4.1.311.25.2), as specified in [MS-WCCE] sections 2.2.2.7.7.4 and 3.2.2.6.2.1.4.5.9,
	// in the issued certificate.
	MSPKI_ENROLLMENT_FLAG_NO_SECURITY_EXTENSION = 0x00080000
)

var MSPKIEnrollmentFlagMap = map[MSPKIEnrollmentFlag]string{
	MSPKI_ENROLLMENT_FLAG_INCLUDE_SYMMETRIC_ALGORITHMS:                                  "Include Symmetric Algorithms",
	MSPKI_ENROLLMENT_FLAG_PEND_ALL_REQUESTS:                                             "Pending All Requests",
	MSPKI_ENROLLMENT_FLAG_PUBLISH_TO_KRA_CONTAINER:                                      "Publish to KRA Container",
	MSPKI_ENROLLMENT_FLAG_PUBLISH_TO_DS:                                                 "Publish to DS",
	MSPKI_ENROLLMENT_FLAG_AUTO_ENROLLMENT_CHECK_USER_DS_CERTIFICATE:                     "Auto Enrollment Check User DS Certificate",
	MSPKI_ENROLLMENT_FLAG_AUTO_ENROLLMENT:                                               "Auto Enrollment",
	MSPKI_ENROLLMENT_FLAG_PREVIOUS_APPROVAL_VALIDATE_REENROLLMENT:                       "Previous Approval Validate Reenrollment",
	MSPKI_ENROLLMENT_FLAG_USER_INTERACTION_REQUIRED:                                     "User Interaction Required",
	MSPKI_ENROLLMENT_FLAG_REMOVE_INVALID_CERTIFICATE_FROM_PERSONAL_STORE:                "Remove Invalid Certificate From Personal Store",
	MSPKI_ENROLLMENT_FLAG_ALLOW_ENROLL_ON_BEHALF_OF:                                     "Allow Enroll On Behalf Of",
	MSPKI_ENROLLMENT_FLAG_ADD_OCSP_NOCHECK:                                              "Add OCSP No Check",
	MSPKI_ENROLLMENT_FLAG_ENABLE_KEY_REUSE_ON_NT_TOKEN_KEYSET_STORAGE_FULL:              "Enable Key Reuse On NT Token Keyset Storage Full",
	MSPKI_ENROLLMENT_FLAG_NOREVOCATIONINFOINISSUEDCERTS:                                 "No Revocation Info In Issued Certs",
	MSPKI_ENROLLMENT_FLAG_INCLUDE_BASIC_CONSTRAINTS_FOR_EE_CERTS:                        "Include Basic Constraints For EE Certs",
	MSPKI_ENROLLMENT_FLAG_ALLOW_PREVIOUS_APPROVAL_KEYBASEDRENEWAL_VALIDATE_REENROLLMENT: "Allow Previous Approval Key Based Renewal Validate Reenrollment",
	MSPKI_ENROLLMENT_FLAG_ISSUANCE_POLICIES_FROM_REQUEST:                                "Issuance Policies From Request",
	MSPKI_ENROLLMENT_FLAG_SKIP_AUTO_RENEWAL:                                             "Skip Auto Renewal",
	MSPKI_ENROLLMENT_FLAG_NO_SECURITY_EXTENSION:                                         "No Security Extension",
}

// String returns the string representation of the enrollment flag
func (flag MSPKIEnrollmentFlag) String() string {
	if val, ok := MSPKIEnrollmentFlagMap[flag]; ok {
		return val
	}
	return fmt.Sprintf("UnknownEnrollmentFlag(%d)", flag)
}
