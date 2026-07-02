package mslsad

// POLICY_LSA_SERVER_ROLE enumerates the role of an LSA server ([MS-LSAD] 2.2.4.4). As an
// NDR enum it is transmitted as a 16-bit unsigned value ([C706] section 14.3.6).
type POLICY_LSA_SERVER_ROLE uint16

const (
	PolicyServerRoleBackup  POLICY_LSA_SERVER_ROLE = 2
	PolicyServerRolePrimary POLICY_LSA_SERVER_ROLE = 3
)
