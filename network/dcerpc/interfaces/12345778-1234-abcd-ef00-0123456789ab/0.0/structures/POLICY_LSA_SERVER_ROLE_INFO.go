package structures

// POLICY_LSA_SERVER_ROLE_INFO contains the role of an LSA server ([MS-LSAD] 2.2.4.8).
type POLICY_LSA_SERVER_ROLE_INFO struct {
	LsaServerRole POLICY_LSA_SERVER_ROLE `ndr:"enum"`
}
