package ldap_attributes

// We need to check if they are real
const (
	// Microsoft EKU OIDs
	EKU_CLIENT_AUTHENTICATION     = "1.3.6.1.5.5.7.3.2"
	EKU_SERVER_AUTHENTICATION     = "1.3.6.1.5.5.7.3.1"
	EKU_CODE_SIGNING              = "1.3.6.1.5.5.7.3.3"
	EKU_EMAIL_PROTECTION          = "1.3.6.1.5.5.7.3.4"
	EKU_TIME_STAMPING             = "1.3.6.1.5.5.7.3.8"
	EKU_OCSP_SIGNING              = "1.3.6.1.5.5.7.3.9"
	EKU_IPSEC_END_SYSTEM          = "1.3.6.1.5.5.7.3.5"
	EKU_IPSEC_TUNNEL              = "1.3.6.1.5.5.7.3.6"
	EKU_IPSEC_USER                = "1.3.6.1.5.5.7.3.7"
	EKU_ANY                       = "2.5.29.37.0"
	EKU_CERTIFICATE_REQUEST_AGENT = "1.3.6.1.4.1.311.20.2.1"
	EKU_SMART_CARD_LOGON          = "1.3.6.1.4.1.311.20.2.2"
	EKU_DS_EMAIL_REPLICATION      = "1.3.6.1.4.1.311.21.19"
	EKU_KDC_AUTHENTICATION        = "1.3.6.1.5.2.3.5"
	EKU_FILE_RECOVERY             = "1.3.6.1.4.1.311.10.3.4"
	EKU_QUALIFIED_SUBORDINATION   = "1.3.6.1.4.1.311.10.3.10"
	EKU_KEY_RECOVERY_AGENT        = "1.3.6.1.4.1.311.21.6"
	EKU_CA_EXCHANGE               = "1.3.6.1.4.1.311.21.5"
	EKU_LIFETIME_SIGNING          = "1.3.6.1.4.1.311.10.3.13"
	EKU_DOCUMENT_SIGNING          = "1.3.6.1.4.1.311.10.3.12"
	EKU_KEY_PACK_LICENSES         = "1.3.6.1.4.1.311.10.6.2"
	EKU_KEY_PACK_SILENT_USER      = "1.3.6.1.4.1.311.10.6.1"
)
