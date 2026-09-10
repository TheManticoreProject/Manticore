package rpcinterface_b97db8b24c6311cfbff608002be23f2f_3_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the clusapi interface
// (b97db8b2-4c63-11cf-bff6-08002be23f2f v3.0, [MS-CMRP]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "b97db8b2-4c63-11cf-bff6-08002be23f2f" {
		t.Errorf("UUID = %s, want b97db8b2-4c63-11cf-bff6-08002be23f2f", got)
	}
	if s.MajorVersion != 3 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 3.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// that a couple of anchor opnums resolve to the expected method names.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) differ in size",
			len(OpnumToName), len(NameToOpnum))
	}
	if OpnumToName[OpnumApiOpenCluster] != "ApiOpenCluster" {
		t.Errorf("opnum 0 = %q, want ApiOpenCluster", OpnumToName[OpnumApiOpenCluster])
	}
	if OpnumToName[OpnumApiCloseCluster] != "ApiCloseCluster" {
		t.Errorf("opnum 1 = %q, want ApiCloseCluster", OpnumToName[OpnumApiCloseCluster])
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestOpnumGaps checks that the "not used on the wire" opnums (e.g. 80, 264) are not
// present in the opnum map, while their neighbors are.
func TestOpnumGaps(t *testing.T) {
	for _, gap := range []uint16{80, 264} {
		if name, ok := OpnumToName[gap]; ok {
			t.Errorf("opnum %d should be absent (not used on the wire), got %q", gap, name)
		}
	}
	if _, ok := OpnumToName[79]; !ok {
		t.Errorf("opnum 79 (ApiNodeControl) should be present")
	}
}

// TestDocumentedStatusCodesResolveThroughWin32 pins that the Win32 codes [MS-CMRP]
// documents for the ClusAPI methods resolve through the shared [MS-ERREF] 2.2 table
// under their specification names, that cluster codes this interface never enumerated
// now render by name rather than as undecoded hex, and that a value the specification
// does not define still renders as hex.
func TestDocumentedStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000001: "ERROR_INVALID_FUNCTION",
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000003: "ERROR_PATH_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x0000000D: "ERROR_INVALID_DATA",
		0x00000022: "ERROR_WRONG_DISK",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000034: "ERROR_DUP_NAME",
		0x00000046: "ERROR_SHARING_PAUSED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x00000072: "ERROR_INVALID_TARGET_HANDLE",
		0x00000078: "ERROR_CALL_NOT_IMPLEMENTED",
		0x0000007B: "ERROR_INVALID_NAME",
		0x00000091: "ERROR_DIR_NOT_EMPTY",
		0x000000B7: "ERROR_ALREADY_EXISTS",
		0x000000EA: "ERROR_MORE_DATA",
		0x00000103: "ERROR_NO_MORE_ITEMS",
		0x0000029C: "ERROR_ASSERTION_FAILURE",
		0x000003E5: "ERROR_IO_PENDING",
		0x00000423: "ERROR_CIRCULAR_DEPENDENCY",
		0x00000428: "ERROR_EXCEPTION_IN_SERVICE",
		0x0000045B: "ERROR_SHUTDOWN_IN_PROGRESS",
		0x000004E7: "ERROR_SERVER_SHUTDOWN_IN_PROGRESS",
		0x00000522: "ERROR_PRIVILEGE_NOT_HELD",
		0x0000052D: "ERROR_PASSWORD_RESTRICTION",
		0x0000053A: "ERROR_INVALID_SECURITY_DESCR",
		0x0000055C: "ERROR_SPECIAL_GROUP",
		0x00001389: "ERROR_DEPENDENT_RESOURCE_EXISTS",
		0x0000138A: "ERROR_DEPENDENCY_NOT_FOUND",
		0x0000138B: "ERROR_DEPENDENCY_ALREADY_EXISTS",
		0x0000138C: "ERROR_RESOURCE_NOT_ONLINE",
		0x0000138D: "ERROR_HOST_NODE_NOT_AVAILABLE",
		0x0000138E: "ERROR_RESOURCE_NOT_AVAILABLE",
		0x0000138F: "ERROR_RESOURCE_NOT_FOUND",
		0x00001390: "ERROR_SHUTDOWN_CLUSTER",
		0x00001392: "ERROR_OBJECT_ALREADY_EXISTS",
		0x00001394: "ERROR_GROUP_NOT_AVAILABLE",
		0x00001395: "ERROR_GROUP_NOT_FOUND",
		0x00001397: "ERROR_HOST_NODE_NOT_RESOURCE_OWNER",
		0x00001398: "ERROR_HOST_NODE_NOT_GROUP_OWNER",
		0x0000139A: "ERROR_RESMON_ONLINE_FAILED",
		0x0000139B: "ERROR_RESOURCE_ONLINE",
		0x0000139D: "ERROR_NOT_QUORUM_CAPABLE",
		0x0000139F: "ERROR_INVALID_STATE",
		0x000013A0: "ERROR_RESOURCE_PROPERTIES_STORED",
		0x000013A1: "ERROR_NOT_QUORUM_CLASS",
		0x000013A2: "ERROR_CORE_RESOURCE",
		0x000013AB: "ERROR_NETWORK_NOT_AVAILABLE",
		0x000013AC: "ERROR_NODE_NOT_AVAILABLE",
		0x000013AD: "ERROR_ALL_NODES_NOT_AVAILABLE",
		0x000013AE: "ERROR_RESOURCE_FAILED",
		0x000013B2: "ERROR_CLUSTER_NODE_NOT_FOUND",
		0x000013B5: "ERROR_CLUSTER_NETWORK_NOT_FOUND",
		0x000013B7: "ERROR_CLUSTER_NETINTERFACE_NOT_FOUND",
		0x000013B8: "ERROR_CLUSTER_INVALID_REQUEST",
		0x000013BA: "ERROR_CLUSTER_NODE_DOWN",
		0x000013C2: "ERROR_CLUSTER_NODE_NOT_PAUSED",
		0x000013CD: "ERROR_DEPENDENCY_NOT_ALLOWED",
		0x000013CF: "ERROR_NODE_CANT_HOST_RESOURCE",
		0x000013D0: "ERROR_CLUSTER_NODE_NOT_READY",
		0x000013D1: "ERROR_CLUSTER_NODE_SHUTTING_DOWN",
		0x000013D6: "ERROR_CLUSTER_RESOURCE_TYPE_NOT_FOUND",
		0x000013D7: "ERROR_CLUSTER_RESTYPE_NOT_SUPPORTED",
		0x000013DC: "ERROR_RESMON_INVALID_STATE",
		0x00001714: "ERROR_CLUSTER_GROUP_MOVING",
		0x00001728: "ERROR_QUORUM_NOT_ALLOWED_IN_THIS_GROUP",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// ERROR_CLUSTER_NODE_EXISTS and ERROR_CLUSTER_JOIN_ABORTED sit in the same
	// [MS-ERREF] 2.2 cluster range as the codes above but were outside the subset this
	// interface used to declare, so they printed as 0x000013b0 and 0x000013d2. They
	// resolve by name now.
	for code, name := range map[uint32]string{
		0x000013B0: "ERROR_CLUSTER_NODE_EXISTS",
		0x000013D2: "ERROR_CLUSTER_JOIN_ABORTED",
	} {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want 0xdeadbeef", got)
	}
}

// TestPipeName documents that this interface has no named pipe: MS-CMRP v3.0 uses
// ncacn_ip_tcp with a dynamic endpoint, so PipeName is intentionally empty.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}
