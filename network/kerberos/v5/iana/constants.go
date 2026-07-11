// Package iana holds the IANA/RFC-registered numeric constants of the Kerberos
// v5 protocol (RFC 4120 and the crypto/checksum registries): message types,
// principal name types, encryption and checksum type IDs, pre-authentication
// data types, error codes, key-usage numbers, and flag bit positions.
//
// It is a dependency-free leaf package: both the wire-message layer
// (network/kerberos/v5/messages) and the crypto engine
// (network/kerberos/v5/crypto) import it, so neither has to depend on the
// other merely to share a constant. Keeping every protocol constant in one
// place also gives a single source of truth as MS-KILE / RFC 8009 etypes and
// additional PA-DATA types are added.
package iana

// Kerberos protocol version number (RFC 4120).
const KerberosV5 = 5

// Message type constants (RFC 4120 Section 7.5.7).
const (
	MsgTypeASReq   = 10 // AS-REQ
	MsgTypeASRep   = 11 // AS-REP
	MsgTypeTGSReq  = 12 // TGS-REQ
	MsgTypeTGSRep  = 13 // TGS-REP
	MsgTypeAPReq   = 14 // AP-REQ
	MsgTypeAPRep   = 15 // AP-REP
	MsgTypeKRBCred = 22 // KRB-CRED
	MsgTypeError   = 30 // KRB-ERROR
)

// Principal name type constants (RFC 4120 Section 6.2 / 7.5.8).
const (
	NameTypeUnknown    = 0  // NT-UNKNOWN
	NameTypePrincipal  = 1  // NT-PRINCIPAL
	NameTypeSRVInst    = 2  // NT-SRV-INST (krbtgt)
	NameTypeSRVHST     = 3  // NT-SRV-HST
	NameTypeSRVXHST    = 4  // NT-SRV-XHST
	NameTypeUID        = 5  // NT-UID
	NameTypeX500       = 6  // NT-X500-PRINCIPAL
	NameTypeSMTP       = 7  // NT-SMTP-NAME
	NameTypeEnterprise = 10 // NT-ENTERPRISE
)

// Encryption type constants (RFC 3961, RFC 3962, RFC 8009, RFC 4757).
const (
	ETypeDESCBCCRC           = 1  // des-cbc-crc (legacy)
	ETypeDESCBCMD5           = 3  // des-cbc-md5 (legacy)
	ETypeDES3CBCSHA1KD       = 16 // des3-cbc-sha1-kd (legacy)
	ETypeAES128CTSHMACSHA196 = 17 // aes128-cts-hmac-sha1-96 (RFC 3962)
	ETypeAES256CTSHMACSHA196 = 18 // aes256-cts-hmac-sha1-96 (RFC 3962)
	ETypeAES128CTSHMACSHA256 = 19 // aes128-cts-hmac-sha256-128 (RFC 8009)
	ETypeAES256CTSHMACSHA384 = 20 // aes256-cts-hmac-sha384-192 (RFC 8009)
	ETypeRC4HMAC             = 23 // rc4-hmac / arcfour-hmac (RFC 4757)
	ETypeRC4HMACExp          = 24 // rc4-hmac-exp (RFC 4757)
)

// Checksum type constants (RFC 3961/3962/8009/4757 registry).
const (
	CksumTypeCRC32               = 1    // crc32
	CksumTypeRSAMD5              = 7    // rsa-md5
	CksumTypeHMACSHA1DES3KD      = 12   // hmac-sha1-des3-kd
	CksumTypeHMACSHA196AES128    = 15   // hmac-sha1-96-aes128
	CksumTypeHMACSHA196AES256    = 16   // hmac-sha1-96-aes256
	CksumTypeHMACSHA256128AES128 = 19   // hmac-sha256-128-aes128 (RFC 8009)
	CksumTypeHMACSHA384192AES256 = 20   // hmac-sha384-192-aes256 (RFC 8009)
	CksumTypeHMACMD5             = -138 // hmac-md5 (rc4), aka 0xffffff76
)

// Key usage numbers (RFC 4120 Section 7.5.1). Key usage 0 is not permitted.
const (
	KeyUsageASReqPAEncTimestamp    = 1  // PA-ENC-TIMESTAMP, client key
	KeyUsageKDCRepTicket           = 2  // Ticket enc-part, service key
	KeyUsageASRepEncPart           = 3  // AS-REP enc-part, client key
	KeyUsageTGSReqKDCReqBodyAD     = 4  // TGS-REQ enc-authz-data, TGS session key
	KeyUsageTGSReqKDCReqBodyADSub  = 5  // TGS-REQ enc-authz-data, auth subkey
	KeyUsageTGSReqAuthCksum        = 6  // TGS-REQ authenticator cksum, TGS session key
	KeyUsageTGSReqPAAPReqAuthen    = 7  // TGS-REQ PA-TGS-REQ authenticator, TGS session key
	KeyUsageTGSRepEncSessionKey    = 8  // TGS-REP enc-part, TGS session key
	KeyUsageTGSRepEncSubSessionKey = 9  // TGS-REP enc-part, auth subkey
	KeyUsageAPReqAuthCksum         = 10 // AP-REQ authenticator cksum
	KeyUsageAPReqAuthen            = 11 // AP-REQ authenticator, app session key
	KeyUsageAPRepEncPart           = 12 // AP-REP enc-part, app session key
	KeyUsageKRBPrivEncPart         = 13 // KRB-PRIV enc-part
	KeyUsageKRBCredEncPart         = 14 // KRB-CRED enc-part
	KeyUsageKRBSafeCksum           = 15 // KRB-SAFE cksum
	KeyUsageKerbNonKerbSalt        = 16 // KERB_NON_KERB_SALT (MS-PAC PAC_CREDENTIAL_INFO)
	KeyUsageKerbNonKerbCksumSalt   = 17 // KERB_NON_KERB_CKSUM_SALT (MS-PAC signatures)
	KeyUsageADKDCIssuedCksum       = 19 // AD-KDC-ISSUED checksum
)

// Pre-authentication data type constants (RFC 4120 Section 7.5.2, MS-KILE, RFC 6113).
const (
	PATGSReq             = 1   // PA-TGS-REQ (DER AP-REQ)
	PAEncTimestamp       = 2   // PA-ENC-TIMESTAMP
	PAPWSalt             = 3   // PA-PW-SALT (raw salt, not ASN.1)
	PAETypeInfo          = 11  // PA-ETYPE-INFO
	PAPKASReqOld         = 14  // PA-PK-AS-REQ_OLD (PKINIT)
	PAPKASRepOld         = 15  // PA-PK-AS-REP_OLD (PKINIT)
	PAPKASReq            = 16  // PA-PK-AS-REQ (PKINIT, MS-PKCA)
	PAPKASRep            = 17  // PA-PK-AS-REP (PKINIT, MS-PKCA)
	PAETypeInfo2         = 19  // PA-ETYPE-INFO2 (replaces PA-ETYPE-INFO)
	PAForUser            = 129 // PA-FOR-USER (MS-SFU S4U2Self)
	PASvrReferralInfo    = 20  // PA-SVR-REFERRAL-INFO (RFC 6806)
	PAPACRequest         = 128 // KERB-PA-PAC-REQUEST (MS-KILE)
	PAFXCookie           = 133 // PA-FX-COOKIE (RFC 6113 FAST)
	PAFXFast             = 136 // PA-FX-FAST (RFC 6113 FAST)
	PAFXError            = 137 // PA-FX-ERROR (RFC 6113 FAST)
	PAEncryptedChallenge = 138 // PA-ENCRYPTED-CHALLENGE (RFC 6113 FAST)
	PAKeyListReq         = 161 // KERB-KEY-LIST-REQ (MS-KILE)
	PAKeyListRep         = 162 // KERB-KEY-LIST-REP (MS-KILE)
	PASupportedEnctypes  = 165 // PA-SUPPORTED-ENCTYPES (MS-KILE)
	PAPACOptions         = 167 // PA-PAC-OPTIONS (MS-KILE)
)

// KDC error code constants (RFC 4120 Section 7.5.9).
const (
	ErrNone               = 0  // KDC_ERR_NONE
	ErrNameExp            = 1  // KDC_ERR_NAME_EXP
	ErrServiceExp         = 2  // KDC_ERR_SERVICE_EXP
	ErrBadPVNO            = 3  // KDC_ERR_BAD_PVNO
	ErrCPrincipalUnknown  = 6  // KDC_ERR_C_PRINCIPAL_UNKNOWN
	ErrSPrincipalUnknown  = 7  // KDC_ERR_S_PRINCIPAL_UNKNOWN
	ErrPrincipalNotUnique = 8  // KDC_ERR_PRINCIPAL_NOT_UNIQUE
	ErrNullKey            = 9  // KDC_ERR_NULL_KEY
	ErrPolicy             = 12 // KDC_ERR_POLICY
	ErrBadOption          = 13 // KDC_ERR_BADOPTION
	ErrETypeNoSupp        = 14 // KDC_ERR_ETYPE_NOSUPP
	ErrSumTypeNoSupp      = 15 // KDC_ERR_SUMTYPE_NOSUPP
	ErrPADataTypeNoSupp   = 16 // KDC_ERR_PADATA_TYPE_NOSUPP
	ErrClientRevoked      = 18 // KDC_ERR_CLIENT_REVOKED
	ErrClientNotYet       = 21 // KDC_ERR_CLIENT_NOTYET
	ErrKeyExpired         = 23 // KDC_ERR_KEY_EXPIRED
	ErrPreauthFailed      = 24 // KDC_ERR_PREAUTH_FAILED
	ErrPreauthRequired    = 25 // KDC_ERR_PREAUTH_REQUIRED
	ErrServerNoMatch      = 26 // KDC_ERR_SERVER_NOMATCH
	ErrMustUseUser2User   = 27 // KDC_ERR_MUST_USE_USER2USER
	ErrBadIntegrity       = 31 // KRB_AP_ERR_BAD_INTEGRITY
	ErrTktExpired         = 32 // KRB_AP_ERR_TKT_EXPIRED
	ErrTktNYV             = 33 // KRB_AP_ERR_TKT_NYV
	ErrRepeat             = 34 // KRB_AP_ERR_REPEAT
	ErrNotUs              = 35 // KRB_AP_ERR_NOT_US
	ErrBadMatch           = 36 // KRB_AP_ERR_BADMATCH
	ErrSkew               = 37 // KRB_AP_ERR_SKEW
	ErrBadAddr            = 38 // KRB_AP_ERR_BADADDR
	ErrBadVersion         = 39 // KRB_AP_ERR_BADVERSION
	ErrMsgType            = 40 // KRB_AP_ERR_MSG_TYPE
	ErrModified           = 41 // KRB_AP_ERR_MODIFIED
	ErrGeneric            = 60 // KRB_ERR_GENERIC
	ErrFieldTooLong       = 61 // KRB_ERR_FIELD_TOOLONG
	ErrResponseTooBig     = 52 // KRB_ERR_RESPONSE_TOO_BIG
	ErrWrongRealm         = 68 // KDC_ERR_WRONG_REALM
	ErrUserToUserRequired = 69 // KRB_AP_ERR_USER_TO_USER_REQUIRED
)

// AP options bit positions (RFC 4120 Section 5.5.1). Bit 0 = MSB.
const (
	APOptionUseSessionKey = 1 // use-session-key
	APOptionMutualAuth    = 2 // mutual-required
)

// Ticket flag bit positions (RFC 4120 Section 5.3 / 2.x). Bit 0 = MSB.
const (
	TicketFlagForwardable  = 1  // forwardable
	TicketFlagForwarded    = 2  // forwarded
	TicketFlagProxiable    = 3  // proxiable
	TicketFlagProxy        = 4  // proxy
	TicketFlagMayPostdate  = 5  // may-postdate
	TicketFlagPostdated    = 6  // postdated
	TicketFlagInvalid      = 7  // invalid
	TicketFlagRenewable    = 8  // renewable
	TicketFlagInitial      = 9  // initial
	TicketFlagPreAuthent   = 10 // pre-authent
	TicketFlagHWAuthent    = 11 // hw-authent
	TicketFlagTransitCheck = 12 // transited-policy-checked
	TicketFlagOKAsDelegate = 13 // ok-as-delegate
)

// KDC options bit positions (RFC 4120 Section 5.4.1, RFC 6806, U2U). Bit 0 = MSB.
// Note: KDCOptions and TicketFlags differ — do not copy one into the other.
const (
	KDCOptionForwardable       = 1  // forwardable
	KDCOptionForwarded         = 2  // forwarded
	KDCOptionProxiable         = 3  // proxiable
	KDCOptionProxy             = 4  // proxy
	KDCOptionAllowPostdate     = 5  // allow-postdate
	KDCOptionPostdated         = 6  // postdated
	KDCOptionRenewable         = 8  // renewable
	KDCOptionCNameInAddlTkt    = 14 // cname-in-additional-ticket (MS-SFU S4U2Proxy; RFC 4120 leaves bit 14 unused)
	KDCOptionCanonicalize      = 15 // canonicalize (RFC 6806)
	KDCOptionDisableTransitChk = 26 // disable-transited-check
	KDCOptionRenewableOK       = 27 // renewable-ok
	KDCOptionEncTktInSKey      = 28 // enc-tkt-in-skey (user-to-user)
	KDCOptionRenew             = 30 // renew
	KDCOptionValidate          = 31 // validate
)
