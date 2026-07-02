package mslsad

// TRUSTED_INFORMATION_CLASS enumerates the trusted-domain information classes that
// select the arm of LSAPR_TRUSTED_DOMAIN_INFO ([MS-LSAD] 2.2.7.28). As an NDR enum it is
// transmitted as a 16-bit unsigned value ([C706] section 14.3.6).
type TRUSTED_INFORMATION_CLASS uint16

const (
	TrustedDomainNameInformation          TRUSTED_INFORMATION_CLASS = 1
	TrustedControllersInformation         TRUSTED_INFORMATION_CLASS = 2
	TrustedPosixOffsetInformation         TRUSTED_INFORMATION_CLASS = 3
	TrustedPasswordInformation            TRUSTED_INFORMATION_CLASS = 4
	TrustedDomainInformationBasic         TRUSTED_INFORMATION_CLASS = 5
	TrustedDomainInformationEx            TRUSTED_INFORMATION_CLASS = 6
	TrustedDomainAuthInformation          TRUSTED_INFORMATION_CLASS = 7
	TrustedDomainFullInformation          TRUSTED_INFORMATION_CLASS = 8
	TrustedDomainAuthInformationInternal  TRUSTED_INFORMATION_CLASS = 9
	TrustedDomainFullInformationInternal  TRUSTED_INFORMATION_CLASS = 10
	TrustedDomainInformationEx2Internal   TRUSTED_INFORMATION_CLASS = 11
	TrustedDomainFullInformation2Internal TRUSTED_INFORMATION_CLASS = 12
	TrustedDomainSupportedEncryptionTypes TRUSTED_INFORMATION_CLASS = 13
)
