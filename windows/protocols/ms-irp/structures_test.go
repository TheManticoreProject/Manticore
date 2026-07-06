package msirp

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts
// the result is deeply equal to in. This is the wire-shape acceptance gate for the
// [MS-IRP] (inetinfo) NDR structures in the absence of a live IIS host.
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}
	var out T
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("%s: round trip mismatch:\n in:  %+v\n out: %+v", name, in, out)
	}
}

func wstr(s string) *ndr.WSTR { w := ndr.WSTR(s); return &w }

// TestINET_INFO_CAP_FLAGS exercises the simplest fixed-size structure (two DWORDs).
func TestINET_INFO_CAP_FLAGS(t *testing.T) {
	roundTrip(t, "INET_INFO_CAP_FLAGS", INET_INFO_CAP_FLAGS{Flag: 0x0000000f, Mask: 0xffffffff})
}

// TestINET_INFO_CAPABILITIES_STRUCT exercises a [unique] pointer to a conformant array
// (CapFlags, sized by NumCapFlags) of INET_INFO_CAP_FLAGS — the shape returned by
// R_InetInfoGetServerCapabilities.
func TestINET_INFO_CAPABILITIES_STRUCT(t *testing.T) {
	roundTrip(t, "INET_INFO_CAPABILITIES_STRUCT", INET_INFO_CAPABILITIES_STRUCT{
		CapVersion:   1,
		ProductType:  2,
		MajorVersion: 10,
		MinorVersion: 0,
		BuildNumber:  17763,
		NumCapFlags:  2,
		CapFlags: []INET_INFO_CAP_FLAGS{
			{Flag: 0x1, Mask: 0x1},
			{Flag: 0x2, Mask: 0x3},
		},
	})
}

// TestINET_INFO_IP_SEC_LIST exercises an embedded (flexible-array-member) conformant
// array whose maximum_count is hoisted to the front of the structure — no referent id.
func TestINET_INFO_IP_SEC_LIST(t *testing.T) {
	roundTrip(t, "INET_INFO_IP_SEC_LIST", INET_INFO_IP_SEC_LIST{
		CEntries: 2,
		AIPSecEntry: []INET_INFO_IP_SEC_ENTRY{
			{DwMask: 0xffffff00, DwNetwork: 0x0a000000},
			{DwMask: 0xffff0000, DwNetwork: 0xac100000},
		},
	})
}

// TestINET_INFO_SITE_LIST exercises an embedded conformant array of entries that each
// carry a [unique] string pointer (deferred referents follow the whole array body).
func TestINET_INFO_SITE_LIST(t *testing.T) {
	roundTrip(t, "INET_INFO_SITE_LIST", INET_INFO_SITE_LIST{
		CEntries: 2,
		ASiteEntry: []INET_INFO_SITE_ENTRY{
			{PszComment: wstr("Default Web Site"), DwInstance: 1},
			{PszComment: wstr("FTP Site"), DwInstance: 2},
		},
	})
}

// TestINET_INFO_VIRTUAL_ROOT_ENTRY exercises a mix of [unique] string pointers and a
// fixed [257]WCHAR password buffer.
func TestINET_INFO_VIRTUAL_ROOT_ENTRY(t *testing.T) {
	var pw [257]uint16
	for i, r := range "s3cr3t" {
		pw[i] = uint16(r)
	}
	roundTrip(t, "INET_INFO_VIRTUAL_ROOT_ENTRY", INET_INFO_VIRTUAL_ROOT_ENTRY{
		PszRoot:         wstr("/"),
		PszAddress:      wstr("10.0.0.1"),
		PszDirectory:    wstr(`C:\inetpub\wwwroot`),
		DwMask:          0x00000007,
		PszAccountName:  wstr("IUSR"),
		AccountPassword: pw,
		DwError:         0,
	})
}

// TestINET_INFO_VIRTUAL_ROOT_LIST exercises the embedded conformant array of the above
// entries.
func TestINET_INFO_VIRTUAL_ROOT_LIST(t *testing.T) {
	roundTrip(t, "INET_INFO_VIRTUAL_ROOT_LIST", INET_INFO_VIRTUAL_ROOT_LIST{
		CEntries: 1,
		AVirtRootEntry: []INET_INFO_VIRTUAL_ROOT_ENTRY{
			{PszRoot: wstr("/"), PszDirectory: wstr(`C:\web`), DwMask: 0x4},
		},
	})
}

// TestINET_LOG_CONFIGURATION exercises a structure that is entirely fixed-size
// (scalars and fixed WCHAR arrays), with no pointers or conformant arrays.
func TestINET_LOG_CONFIGURATION(t *testing.T) {
	var dir, ds [260]uint16
	var name30 [30]uint16
	var user, pass [257]uint16
	copyUTF16(dir[:], `C:\LogFiles`)
	copyUTF16(name30[:], "InternetLog")
	roundTrip(t, "INET_LOG_CONFIGURATION", INET_LOG_CONFIGURATION{
		InetLogType:          1,
		IlPeriod:             2,
		RgchLogFileDirectory: dir,
		CbSizeForTruncation:  0x100000,
		RgchDataSource:       ds,
		RgchTableName:        name30,
		RgchUserName:         user,
		RgchPassword:         pass,
	})
}

func copyUTF16(dst []uint16, s string) {
	i := 0
	for _, r := range s {
		if i >= len(dst) {
			break
		}
		dst[i] = uint16(r)
		i++
	}
}

// TestINET_INFO_CONFIG_INFO exercises the largest administration structure: several
// [unique] string and [unique] structure pointers, a WORD, an LCID, fixed BYTE/WCHAR
// arrays, a 16-bit short, and three nested [unique] list pointers.
func TestINET_INFO_CONFIG_INFO(t *testing.T) {
	var product [64]uint8
	copy(product[:], []byte("IIS"))
	var anonPw [257]uint16
	copyUTF16(anonPw[:], "anon")
	roundTrip(t, "INET_INFO_CONFIG_INFO", INET_INFO_CONFIG_INFO{
		FieldControl:        0x0000ffff,
		DwConnectionTimeout: 300,
		DwMaxConnections:    1000,
		LpszAdminName:       wstr("Administrator"),
		LpszAdminEmail:      wstr("admin@example.test"),
		LpszServerComment:   wstr("Default"),
		LpLogConfig:         &INET_LOG_CONFIGURATION{InetLogType: 1, IlPeriod: 3},
		LangId:              0x0409,
		LocalId:             0x00000409,
		ProductId:           product,
		FLogAnonymous:       1,
		FLogNonAnonymous:    0,
		LpszAnonUserName:    wstr("IUSR"),
		SzAnonPassword:      anonPw,
		DwAuthentication:    0x00000001,
		SPort:               80,
		DenyIPList: &INET_INFO_IP_SEC_LIST{
			CEntries:    1,
			AIPSecEntry: []INET_INFO_IP_SEC_ENTRY{{DwMask: 0xffffffff, DwNetwork: 0x7f000001}},
		},
		GrantIPList:  nil,
		VirtualRoots: nil,
	})
}

// TestINET_INFO_STATISTICS_INFO exercises the [switch_type(unsigned long)] union with
// its case(0) arm selected — a [unique] pointer to INET_INFO_STATISTICS_0. The inline
// discriminant precedes the selected arm ([C706] 14.3.8).
func TestINET_INFO_STATISTICS_INFO(t *testing.T) {
	roundTrip(t, "INET_INFO_STATISTICS_INFO/case0", INET_INFO_STATISTICS_INFO{
		Tag: 0,
		InetStats0: &INET_INFO_STATISTICS_0{
			CacheCtrs:    INETA_CACHE_STATISTICS{FilesCached: 5, CurrentFileCacheSize: 0x1_0000_0000, MaximumFileCacheSize: 0x2_0000_0000},
			AtqCtrs:      INETA_ATQ_STATISTICS{TotalAllowedRequests: 42},
			NAuxCounters: 2,
			RgCounters:   [20]ndr.DWORD{1, 2},
		},
	})
	// The [default] arm carries no body: an unknown discriminant is a bare tag.
	roundTrip(t, "INET_INFO_STATISTICS_INFO/default", INET_INFO_STATISTICS_INFO{Tag: 9})
}

// TestW3_STATISTICS_STRUCT exercises the W3 statistics union (case 0 → *W3_STATISTICS_1),
// which embeds two LARGE_INTEGER counters.
func TestW3_STATISTICS_STRUCT(t *testing.T) {
	roundTrip(t, "W3_STATISTICS_STRUCT/case0", W3_STATISTICS_STRUCT{
		Tag: 0,
		StatInfo1: &W3_STATISTICS_1{
			TotalBytesSent:     msdtyp.LARGE_INTEGER(1 << 40),
			TotalBytesReceived: msdtyp.LARGE_INTEGER(1 << 20),
			TotalGets:          123,
			CurrentConnections: 7,
			RgCounters:         [20]ndr.DWORD{9, 8, 7},
		},
	})
}

// TestFTP_STATISTICS_STRUCT exercises the FTP statistics union (case 0 → *FTP_STATISTICS_0).
func TestFTP_STATISTICS_STRUCT(t *testing.T) {
	roundTrip(t, "FTP_STATISTICS_STRUCT/case0", FTP_STATISTICS_STRUCT{
		Tag: 0,
		StatInfo0: &FTP_STATISTICS_0{
			TotalBytesSent:       msdtyp.LARGE_INTEGER(4096),
			CurrentConnections:   3,
			TotalAllowedRequests: 100,
		},
	})
}

// TestIIS_USER_INFO_1_CONTAINER exercises a [unique] pointer to a conformant array
// (Buffer, sized by EntriesRead) of user records that each carry a [unique] string.
func TestIIS_USER_INFO_1_CONTAINER(t *testing.T) {
	roundTrip(t, "IIS_USER_INFO_1_CONTAINER", IIS_USER_INFO_1_CONTAINER{
		EntriesRead: 2,
		Buffer: []IIS_USER_INFO_1{
			{IdUser: 1, PszUser: wstr("alice"), FAnonymous: 0, InetHost: 0x0a000001, TConnect: 100},
			{IdUser: 2, PszUser: wstr("bob"), FAnonymous: 1, InetHost: 0x0a000002, TConnect: 200},
		},
	})
}

// TestIIS_USER_ENUM_STRUCT exercises the nested union: a Level discriminant beside a
// [switch_is(Level)] union whose case(1) arm is a [unique] pointer to the container.
func TestIIS_USER_ENUM_STRUCT(t *testing.T) {
	roundTrip(t, "IIS_USER_ENUM_STRUCT/level1", IIS_USER_ENUM_STRUCT{
		Level: 1,
		ConfigInfo: USER_ENUM_UNION{
			Tag: 1,
			Level1: &IIS_USER_INFO_1_CONTAINER{
				EntriesRead: 1,
				Buffer:      []IIS_USER_INFO_1{{IdUser: 7, PszUser: wstr("carol")}},
			},
		},
	})
	// The [default] arm is empty; an [in,out] enum with no data is a bare discriminant.
	roundTrip(t, "IIS_USER_ENUM_STRUCT/default", IIS_USER_ENUM_STRUCT{Level: 0, ConfigInfo: USER_ENUM_UNION{Tag: 0}})
}

// TestDWORDLONG_Alias asserts DWORDLONG is the 8-byte unsigned alias used by the cache
// counters, so a full 64-bit value survives a round trip inside a structure.
func TestDWORDLONG_Alias(t *testing.T) {
	roundTrip(t, "INETA_CACHE_STATISTICS/dwordlong", INETA_CACHE_STATISTICS{
		CurrentFileCacheSize: DWORDLONG(0xdead_beef_0000_0001),
		MaximumFileCacheSize: DWORDLONG(0xffff_ffff_ffff_ffff),
	})
}
