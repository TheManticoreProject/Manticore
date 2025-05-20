package ldap_attributes

// General Enrollment Flags
// Src: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-crtd/6cc7eb79-3e84-477a-b398-b0ff2b68a6c0
const (
	// Reserved. All protocols MUST ignore this flag.
	ENROLLMENT_FLAG_ADD_EMAIL = 0x00000002
	// Reserved. All protocols MUST ignore this flag.
	ENROLLMENT_FLAG_PUBLISH_TO_DS = 0x00000008
	// Reserved. All protocols MUST ignore this flag.
	ENROLLMENT_FLAG_EXPORTABLE_KEY = 0x00000010
	// This flag is the same as CT_FLAG_AUTO_ENROLLMENT. Indicates that auto-enrollment is enabled for this certificate template.
	ENROLLMENT_FLAG_AUTO_ENROLLMENT = 0x00000020
	// This flag indicates that this certificate template is for an end entity that represents a machine.
	ENROLLMENT_FLAG_MACHINE_TYPE = 0x00000040
	// This flag indicates a certificate request for a CA certificate.
	ENROLLMENT_FLAG_IS_CA = 0x00000080
	// This flag indicates that a certificate based on this section needs to include a template name certificate extension.
	ENROLLMENT_FLAG_ADD_TEMPLATE_NAME = 0x00000200
	// This flag indicates a certificate request for cross-certifying a certificate. Processing rules are specified in [MS-WCCE].
	ENROLLMENT_FLAG_IS_CROSS_CA = 0x00000800
	// This flag indicates that the record of a certificate request for a certificate that is issued need not be persisted by the CA.
	ENROLLMENT_FLAG_DONOTPERSISTINDB = 0x00001000
	// This flag indicates that the template SHOULD not be modified in any way; it is not used by the client or server in the Windows Client Certificate Enrollment Protocol.
	ENROLLMENT_FLAG_IS_DEFAULT = 0x00010000
	// This flag indicates that the template MAY be modified if required; it is not used by the client or server in the Windows Client Certificate Enrollment Protocol.
	ENROLLMENT_FLAG_IS_MODIFIED = 0x00020000
)
