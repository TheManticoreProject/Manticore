package structures

// LSAPR_TRUSTED_DOMAIN_INFO is the discriminated union of trusted-domain information
// classes ([MS-LSAD] 2.2.7.9). The discriminant Class is a TRUSTED_INFORMATION_CLASS;
// the wire form is the discriminant followed by the single selected arm ([C706] section
// 14.3.8). Case values follow the TRUSTED_INFORMATION_CLASS enum (1..13).
//
// The TrustedDomainInformationBasic arm (case 5) is a LSAPR_TRUST_INFORMATION, since
// LSAPR_TRUSTED_DOMAIN_INFORMATION_BASIC is an IDL typedef of that existing type.
type LSAPR_TRUSTED_DOMAIN_INFO struct {
	Class                   TRUSTED_INFORMATION_CLASS                      `ndr:"switch"`
	TrustedDomainNameInfo   LSAPR_TRUSTED_DOMAIN_NAME_INFO                 `ndr:"case=1"`
	TrustedControllersInfo  LSAPR_TRUSTED_CONTROLLERS_INFO                 `ndr:"case=2"`
	TrustedPosixOffsetInfo  TRUSTED_POSIX_OFFSET_INFO                      `ndr:"case=3"`
	TrustedPasswordInfo     LSAPR_TRUSTED_PASSWORD_INFO                    `ndr:"case=4"`
	TrustedDomainInfoBasic  LSAPR_TRUST_INFORMATION                        `ndr:"case=5"`
	TrustedDomainInfoEx     LSAPR_TRUSTED_DOMAIN_INFORMATION_EX            `ndr:"case=6"`
	TrustedAuthInfo         LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION          `ndr:"case=7"`
	TrustedFullInfo         LSAPR_TRUSTED_DOMAIN_FULL_INFORMATION          `ndr:"case=8"`
	TrustedAuthInfoInternal LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL `ndr:"case=9"`
	TrustedFullInfoInternal LSAPR_TRUSTED_DOMAIN_FULL_INFORMATION_INTERNAL `ndr:"case=10"`
	TrustedDomainInfoEx2    LSAPR_TRUSTED_DOMAIN_INFORMATION_EX2           `ndr:"case=11"`
	TrustedFullInfo2        LSAPR_TRUSTED_DOMAIN_FULL_INFORMATION2         `ndr:"case=12"`
	TrustedDomainSETs       TRUSTED_DOMAIN_SUPPORTED_ENCRYPTION_TYPES      `ndr:"case=13"`
}
