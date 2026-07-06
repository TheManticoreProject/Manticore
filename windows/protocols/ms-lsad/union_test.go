package mslsad

import (
	"testing"

	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// ustr builds an RPC_UNICODE_STRING from a Go string for test fixtures.
func ustr(_ *testing.T, s string) msdtyp.RPC_UNICODE_STRING {
	return msdtyp.NewUnicodeString(s)
}

// TestLSAPR_POLICY_INFORMATION_RoundTrip exercises the policy-information union with two
// different arms selected: case 6 (POLICY_LSA_SERVER_ROLE_INFO) and case 3
// (LSAPR_POLICY_PRIMARY_DOM_INFO with a [unique] SID pointer).
func TestLSAPR_POLICY_INFORMATION_RoundTrip(t *testing.T) {
	roundTrip(t, "PolicyServerRoleInfo(case6)", LSAPR_POLICY_INFORMATION{
		Class:                PolicyLsaServerRoleInformation,
		PolicyServerRoleInfo: POLICY_LSA_SERVER_ROLE_INFO{LsaServerRole: PolicyServerRolePrimary},
	})

	roundTrip(t, "PolicyPrimaryDomainInfo(case3)", LSAPR_POLICY_INFORMATION{
		Class: PolicyPrimaryDomainInformation,
		PolicyPrimaryDomainInfo: LSAPR_POLICY_PRIMARY_DOM_INFO{
			Name: ustr(t, "CONTOSO"),
			Sid:  mustSID(t, "S-1-5-21-1004336348-1177238915-682003330"),
		},
	})
}

// TestLSAPR_TRUSTED_DOMAIN_INFO_RoundTrip exercises the trusted-domain union with the
// TrustedPosixOffsetInformation arm (case 3) selected.
func TestLSAPR_TRUSTED_DOMAIN_INFO_RoundTrip(t *testing.T) {
	roundTrip(t, "TrustedPosixOffsetInfo(case3)", LSAPR_TRUSTED_DOMAIN_INFO{
		Class:                  TrustedPosixOffsetInformation,
		TrustedPosixOffsetInfo: TRUSTED_POSIX_OFFSET_INFO{Offset: 0x12345678},
	})
}

// TestLSA_FOREST_TRUST_INFORMATION_RoundTrip exercises the array-of-pointers-to-records
// shape where each record embeds the ForestTrustData union with a TopLevelName arm
// (case 0).
func TestLSA_FOREST_TRUST_INFORMATION_RoundTrip(t *testing.T) {
	in := LSA_FOREST_TRUST_INFORMATION{
		RecordCount: 2,
		Entries: []*LSA_FOREST_TRUST_RECORD{
			{
				Flags:           0x1,
				ForestTrustType: ForestTrustTopLevelName,
				Time:            msdtyp.LARGE_INTEGER(0x5566778811223344),
				ForestTrustData: LSA_FOREST_TRUST_DATA{
					ForestTrustType: ForestTrustTopLevelName,
					TopLevelName:    ustr(t, "contoso.com"),
				},
			},
			{
				Flags:           0x2,
				ForestTrustType: ForestTrustTopLevelName,
				ForestTrustData: LSA_FOREST_TRUST_DATA{
					ForestTrustType: ForestTrustTopLevelName,
					TopLevelName:    ustr(t, "fabrikam.com"),
				},
			},
		},
	}
	roundTrip(t, "LSA_FOREST_TRUST_INFORMATION", in)
}
