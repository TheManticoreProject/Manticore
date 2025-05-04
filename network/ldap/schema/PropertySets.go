package schema

const (
	PROPERTY_SET_ACCOUNT_RESTRICTIONS                   = "4c164200-20c0-11d0-a768-00aa006e0529"
	PROPERTY_SET_DNS_HOST_NAME_ATTRIBUTES               = "72e39547-7b18-11d1-adef-00c04fd8d5cd"
	PROPERTY_SET_DOMAIN_PASSWORD_AND_LOCKOUT_POLICIES   = "c7407360-20bf-11d0-a768-00aa006e0529"
	PROPERTY_SET_GENERAL_INFORMATION                    = "59ba2f42-79a2-11d0-9020-00c04fc2d3cf"
	PROPERTY_SET_GROUP_MEMBERSHIP                       = "bc0ac240-79a9-11d0-9020-00c04fc2d4cf"
	PROPERTY_SET_LOGON_INFORMATION                      = "5f202010-79a5-11d0-9020-00c04fc2d4cf"
	PROPERTY_SET_MS_TS_GATEWAYACCESS                    = "ffa6f046-ca4b-4feb-b40d-04dfee722543"
	PROPERTY_SET_OTHER_DOMAIN_PARAMETERS_FOR_USE_BY_SAM = "b8119fd0-04f6-4762-ab7a-4986c76b3f9a"
	PROPERTY_SET_PERSONAL_INFORMATION                   = "77b5b886-944a-11d1-aebd-0000f80367c1"
	PROPERTY_SET_PHONE_AND_MAIL_OPTIONS                 = "e45795b2-9455-11d1-aebd-0000f80367c1"
	PROPERTY_SET_PRIVATE_INFORMATION                    = "91e647de-d96f-4b70-9557-d63ff4f3ccd8"
	PROPERTY_SET_PUBLIC_INFORMATION                     = "e48d0154-bcf8-11d1-8702-00c04fb96050"
	PROPERTY_SET_REMOTE_ACCESS_INFORMATION              = "037088f8-0ae1-11d2-b422-00a0c968f939"
	PROPERTY_SET_TERMINAL_SERVER_LICENSE_SERVER         = "5805bc62-bdc9-4428-a5e2-856a0f4c185e"
	PROPERTY_SET_WEB_INFORMATION                        = "e45795b3-9455-11d1-aebd-0000f80367c1"
)

var PropertySetToGUID = map[string]string{
	"PROPERTY_SET_ACCOUNT_RESTRICTIONS":                   PROPERTY_SET_ACCOUNT_RESTRICTIONS,
	"PROPERTY_SET_DNS_HOST_NAME_ATTRIBUTES":               PROPERTY_SET_DNS_HOST_NAME_ATTRIBUTES,
	"PROPERTY_SET_DOMAIN_PASSWORD_AND_LOCKOUT_POLICIES":   PROPERTY_SET_DOMAIN_PASSWORD_AND_LOCKOUT_POLICIES,
	"PROPERTY_SET_GENERAL_INFORMATION":                    PROPERTY_SET_GENERAL_INFORMATION,
	"PROPERTY_SET_GROUP_MEMBERSHIP":                       PROPERTY_SET_GROUP_MEMBERSHIP,
	"PROPERTY_SET_LOGON_INFORMATION":                      PROPERTY_SET_LOGON_INFORMATION,
	"PROPERTY_SET_MS_TS_GATEWAYACCESS":                    PROPERTY_SET_MS_TS_GATEWAYACCESS,
	"PROPERTY_SET_OTHER_DOMAIN_PARAMETERS_FOR_USE_BY_SAM": PROPERTY_SET_OTHER_DOMAIN_PARAMETERS_FOR_USE_BY_SAM,
	"PROPERTY_SET_PERSONAL_INFORMATION":                   PROPERTY_SET_PERSONAL_INFORMATION,
	"PROPERTY_SET_PHONE_AND_MAIL_OPTIONS":                 PROPERTY_SET_PHONE_AND_MAIL_OPTIONS,
	"PROPERTY_SET_PRIVATE_INFORMATION":                    PROPERTY_SET_PRIVATE_INFORMATION,
	"PROPERTY_SET_PUBLIC_INFORMATION":                     PROPERTY_SET_PUBLIC_INFORMATION,
	"PROPERTY_SET_REMOTE_ACCESS_INFORMATION":              PROPERTY_SET_REMOTE_ACCESS_INFORMATION,
	"PROPERTY_SET_TERMINAL_SERVER_LICENSE_SERVER":         PROPERTY_SET_TERMINAL_SERVER_LICENSE_SERVER,
	"PROPERTY_SET_WEB_INFORMATION":                        PROPERTY_SET_WEB_INFORMATION,
}

var GUIDToPropertySet = map[string]string{
	PROPERTY_SET_ACCOUNT_RESTRICTIONS:                   "PROPERTY_SET_ACCOUNT_RESTRICTIONS",
	PROPERTY_SET_DNS_HOST_NAME_ATTRIBUTES:               "PROPERTY_SET_DNS_HOST_NAME_ATTRIBUTES",
	PROPERTY_SET_DOMAIN_PASSWORD_AND_LOCKOUT_POLICIES:   "PROPERTY_SET_DOMAIN_PASSWORD_AND_LOCKOUT_POLICIES",
	PROPERTY_SET_GENERAL_INFORMATION:                    "PROPERTY_SET_GENERAL_INFORMATION",
	PROPERTY_SET_GROUP_MEMBERSHIP:                       "PROPERTY_SET_GROUP_MEMBERSHIP",
	PROPERTY_SET_LOGON_INFORMATION:                      "PROPERTY_SET_LOGON_INFORMATION",
	PROPERTY_SET_MS_TS_GATEWAYACCESS:                    "PROPERTY_SET_MS_TS_GATEWAYACCESS",
	PROPERTY_SET_OTHER_DOMAIN_PARAMETERS_FOR_USE_BY_SAM: "PROPERTY_SET_OTHER_DOMAIN_PARAMETERS_FOR_USE_BY_SAM",
	PROPERTY_SET_PERSONAL_INFORMATION:                   "PROPERTY_SET_PERSONAL_INFORMATION",
	PROPERTY_SET_PHONE_AND_MAIL_OPTIONS:                 "PROPERTY_SET_PHONE_AND_MAIL_OPTIONS",
	PROPERTY_SET_PRIVATE_INFORMATION:                    "PROPERTY_SET_PRIVATE_INFORMATION",
	PROPERTY_SET_PUBLIC_INFORMATION:                     "PROPERTY_SET_PUBLIC_INFORMATION",
	PROPERTY_SET_REMOTE_ACCESS_INFORMATION:              "PROPERTY_SET_REMOTE_ACCESS_INFORMATION",
	PROPERTY_SET_TERMINAL_SERVER_LICENSE_SERVER:         "PROPERTY_SET_TERMINAL_SERVER_LICENSE_SERVER",
	PROPERTY_SET_WEB_INFORMATION:                        "PROPERTY_SET_WEB_INFORMATION",
}

var PropertySetToAttributeDisplayNames = map[string][]string{
	PROPERTY_SET_REMOTE_ACCESS_INFORMATION:              {"ms-ds-token-group-names", "ms-ds-token-group-names-global-and-universal", "ms-ds-token-group-names-no-gc-acceptable", "msnpallowdialin", "msnpcallingstationid", "msradiuscallbacknumber", "msradiusframedipaddress", "msradiusframedroute", "msradiusservicetype", "token-groups", "token-groups-global-and-universal", "token-groups-no-gc-acceptable"},
	PROPERTY_SET_ACCOUNT_RESTRICTIONS:                   {"account-expires", "ms-ds-allowed-to-act-on-behalf-of-other-identity", "ms-ds-user-account-control-computed", "ms-ds-user-password-expiry-time-computed", "pwd-last-set", "user-account-control", "user-parameters"},
	PROPERTY_SET_TERMINAL_SERVER_LICENSE_SERVER:         {"ms-ts-expiredate", "ms-ts-expiredate2", "ms-ts-expiredate3", "ms-ts-expiredate4", "ms-ts-licenseversion", "ms-ts-licenseversion2", "ms-ts-licenseversion3", "ms-ts-licenseversion4", "ms-ts-managingls", "ms-ts-managingls2", "ms-ts-managingls3", "ms-ts-managingls4", "terminal-server"},
	PROPERTY_SET_GENERAL_INFORMATION:                    {"admin-description", "code-page", "country-code", "display-name", "object-sid", "primary-group-id", "sam-account-name", "sam-account-type", "sd-rights-effective", "show-in-advanced-view-only", "sid-history", "uid", "user-comment"},
	PROPERTY_SET_LOGON_INFORMATION:                      {"bad-pwd-count", "home-directory", "home-drive", "last-logoff", "last-logon", "last-logon-timestamp", "logon-count", "logon-hours", "logon-workstation", "profile-path", "script-path", "user-workstations"},
	PROPERTY_SET_DNS_HOST_NAME_ATTRIBUTES:               {"dns-host-name", "ms-ds-additional-dns-host-name"},
	PROPERTY_SET_PERSONAL_INFORMATION:                   {"address", "address-home", "assistant", "comment", "country-name", "facsimile-telephone-number", "international-isdn-number", "locality-name", "ms-ds-cloudextensionattribute1", "ms-ds-cloudextensionattribute10", "ms-ds-cloudextensionattribute11", "ms-ds-cloudextensionattribute12", "ms-ds-cloudextensionattribute13", "ms-ds-cloudextensionattribute14", "ms-ds-cloudextensionattribute15", "ms-ds-cloudextensionattribute16", "ms-ds-cloudextensionattribute17", "ms-ds-cloudextensionattribute18", "ms-ds-cloudextensionattribute19", "ms-ds-cloudextensionattribute2", "ms-ds-cloudextensionattribute20", "ms-ds-cloudextensionattribute3", "ms-ds-cloudextensionattribute4", "ms-ds-cloudextensionattribute5", "ms-ds-cloudextensionattribute6", "ms-ds-cloudextensionattribute7", "ms-ds-cloudextensionattribute8", "ms-ds-cloudextensionattribute9", "ms-ds-external-directory-object-id", "ms-ds-failed-interactive-logon-count", "ms-ds-failed-interactive-logon-count-at-last-successful-logon", "ms-ds-geocoordinates-altitude", "ms-ds-geocoordinates-latitude", "ms-ds-geocoordinates-longitude", "ms-ds-host-service-account", "ms-ds-last-failed-interactive-logon-time", "ms-ds-last-successful-interactive-logon-time", "ms-ds-supported-encryption-types", "msmq-digests", "msmq-sign-certificates", "personal-title", "phone-fax-other", "phone-home-other", "phone-home-primary", "phone-ip-other", "phone-ip-primary", "phone-isdn-primary", "phone-mobile-other", "phone-mobile-primary", "phone-office-other", "phone-pager-other", "phone-pager-primary", "physical-delivery-office-name", "picture", "post-office-box", "postal-address", "postal-code", "preferred-delivery-method", "registered-address", "state-or-province-name", "street-address", "telephone-number", "teletex-terminal-identifier", "telex-number", "telex-primary", "user-cert", "user-shared-folder", "user-shared-folder-other", "user-smime-certificate", "x121-address", "x509-cert"},
	PROPERTY_SET_PRIVATE_INFORMATION:                    {"ms-pki-accountcredentials", "ms-pki-credential-roaming-tokens", "ms-pki-dpapimasterkeys", "ms-pki-roamingtimestamp"},
	"9b026da6-0d3c-465c-8bee-5199d7165cba":              {"ms-ds-key-credential-link"},
	"a29b89fd-c7e8-11d0-9bae-00c04fd92ef5":              {"domain-wide-policy", "efspolicy", "public-key-policy"},
	"a29b89fe-c7e8-11d0-9bae-00c04fd92ef5":              {"domain-policy-reference", "machine-password-change-interval"},
	"a29b8a01-c7e8-11d0-9bae-00c04fd92ef5":              {"local-policy-reference", "machine-wide-policy", "quality-of-service"},
	PROPERTY_SET_OTHER_DOMAIN_PARAMETERS_FOR_USE_BY_SAM: {"domain-replica", "force-logoff", "modified-count", "oem-information", "server-role", "server-state", "uas-compat"},
	PROPERTY_SET_GROUP_MEMBERSHIP:                       {"is-member-of-dl", "member"},
	PROPERTY_SET_DOMAIN_PASSWORD_AND_LOCKOUT_POLICIES:   {"lock-out-observation-window", "lockout-duration", "lockout-threshold", "max-pwd-age", "min-pwd-age", "min-pwd-length", "pwd-history-length", "pwd-properties"},
	PROPERTY_SET_WEB_INFORMATION:                        {"www-home-page", "www-page-other"},
	PROPERTY_SET_PUBLIC_INFORMATION:                     {"additional-information", "allowed-attributes", "allowed-attributes-effective", "allowed-child-classes", "allowed-child-classes-effective", "alt-security-identities", "common-name", "company", "department", "description", "display-name-printable", "division", "e-mail-addresses", "given-name", "initials", "legacy-exchange-dn", "manager", "ms-ds-allowed-to-delegate-to", "ms-ds-approx-immed-subordinates", "ms-ds-auxiliary-classes", "ms-ds-hab-seniority-index", "ms-ds-phonetic-company-name", "ms-ds-phonetic-department", "ms-ds-phonetic-display-name", "ms-ds-phonetic-first-name", "ms-ds-phonetic-last-name", "ms-ds-source-object-dn", "obj-dist-name", "object-category", "object-class", "object-guid", "organization-name", "organizational-unit-name", "other-mailbox", "proxy-addresses", "rdn", "reports", "service-principal-name", "show-in-address-book", "surname", "system-flags", "text-country", "title", "user-principal-name"},
}
