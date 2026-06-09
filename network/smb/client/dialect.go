package client

import (
	"github.com/TheManticoreProject/Manticore/network/smb"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
)

// SMB1 dialect strings (MS-CIFS) and SMB2 markers used in the dialect list of a
// multi-protocol SMB_COM_NEGOTIATE request.
const (
	// SMB1DialectNTLM012 is the modern SMB1 dialect ("NT LM 0.12").
	SMB1DialectNTLM012 = "NT LM 0.12"
	// SMB2DialectString2002 pins SMB 2.0.2 when offered in an SMB1 negotiate.
	SMB2DialectString2002 = "SMB 2.002"
	// SMB2DialectStringWildcard offers "any SMB2"; the server replies with the
	// wildcard revision 0x02FF and the client follows up with a native SMB2
	// negotiate to pin the exact dialect.
	SMB2DialectStringWildcard = "SMB 2.???"
)

// smb2DialectFor maps an abstract protocol version to its concrete SMB2 wire
// dialect revision. ok is false for SMB1 (which has no 16-bit revision) and for
// values that are not SMB2 dialects. The abstract SMB_VERSION_2_0 family marker
// maps to the lowest real SMB2 dialect, SMB 2.0.2.
func smb2DialectFor(v smb.SMBProtocolVersion) (dialects.Dialect, bool) {
	switch v {
	case smb.SMB_VERSION_2_0, smb.SMB_VERSION_2_0_2:
		return dialects.SMB2_DIALECT_2_0_2, true
	case smb.SMB_VERSION_2_1:
		return dialects.SMB2_DIALECT_2_1_0, true
	case smb.SMB_VERSION_3_0:
		return dialects.SMB2_DIALECT_3_0_0, true
	case smb.SMB_VERSION_3_0_2:
		return dialects.SMB2_DIALECT_3_0_2, true
	case smb.SMB_VERSION_3_1_1:
		return dialects.SMB2_DIALECT_3_1_1, true
	}
	return 0, false
}

// versionForSMB2Dialect is the reverse mapping, used to interpret the
// DialectRevision returned in an SMB2 NEGOTIATE response. ok is false for the
// wildcard revision and for any value that is not a real SMB2 dialect (notably
// 0x0200, which is the abstract family marker, not a wire dialect).
func versionForSMB2Dialect(d dialects.Dialect) (smb.SMBProtocolVersion, bool) {
	switch d {
	case dialects.SMB2_DIALECT_2_0_2:
		return smb.SMB_VERSION_2_0_2, true
	case dialects.SMB2_DIALECT_2_1_0:
		return smb.SMB_VERSION_2_1, true
	case dialects.SMB2_DIALECT_3_0_0:
		return smb.SMB_VERSION_3_0, true
	case dialects.SMB2_DIALECT_3_0_2:
		return smb.SMB_VERSION_3_0_2, true
	case dialects.SMB2_DIALECT_3_1_1:
		return smb.SMB_VERSION_3_1_1, true
	}
	return 0, false
}

// defaultPreference is the preference list used when Options.Preferred is empty:
// every supported version, best first.
func defaultPreference() []smb.SMBProtocolVersion {
	return []smb.SMBProtocolVersion{
		smb.SMB_VERSION_3_1_1,
		smb.SMB_VERSION_3_0_2,
		smb.SMB_VERSION_3_0,
		smb.SMB_VERSION_2_1,
		smb.SMB_VERSION_2_0_2,
		smb.SMB_VERSION_1_0,
	}
}
