package functions_test

import (
	"bytes"
	"testing"

	inetinfo "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82ad4280-036b-11cf-972c-00aa006887b0/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82ad4280-036b-11cf-972c-00aa006887b0/2.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msirp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-irp"
)

// responder is an ndr.Invoker that records the marshalled request stub (so the
// on-the-wire NDR layout can be asserted) and replies with a canned response stub.
type responder struct {
	stub  []byte
	opnum uint16
	resp  []byte
}

func (r *responder) Invoke(in ndr.Call, out any) error {
	b, err := ndr.Request(in)
	if err != nil {
		return err
	}
	r.stub = b
	r.opnum = in.Opnum()
	if r.resp == nil {
		return nil
	}
	return ndr.Response(r.resp, out)
}

// TestR_InetInfoGetVersion exercises the simplest method: a [unique,string] server
// handle plus a reserved DWORD in, a DWORD version out, and the trailing return code.
func TestR_InetInfoGetVersion(t *testing.T) {
	// Canned response: pdwVersion = 0x00050000, ERROR_SUCCESS.
	resp := []byte{
		0x00, 0x00, 0x05, 0x00, // pdwVersion
		0x00, 0x00, 0x00, 0x00, // return code (ERROR_SUCCESS)
	}
	r := &responder{resp: resp}
	version, err := functions.R_InetInfoGetVersion(r, nil, 7)
	if err != nil {
		t.Fatalf("R_InetInfoGetVersion: %v", err)
	}
	if version != 0x00050000 {
		t.Errorf("version = %#x, want 0x00050000", version)
	}
	if r.opnum != inetinfo.OpnumR_InetInfoGetVersion {
		t.Errorf("opnum = %d, want %d", r.opnum, inetinfo.OpnumR_InetInfoGetVersion)
	}
	// Request stub: a NULL [unique] pszServer is a zero referent id, then dwReserved.
	wantStub := []byte{
		0x00, 0x00, 0x00, 0x00, // pszServer NULL referent
		0x07, 0x00, 0x00, 0x00, // dwReserved
	}
	if !bytes.Equal(r.stub, wantStub) {
		t.Errorf("request stub:\n got %x\nwant %x", r.stub, wantStub)
	}
}

// TestR_InetInfoGetVersion_Error verifies a nonzero return code surfaces as an error
// carrying the hex code (ERROR_ACCESS_DENIED, 0x00000005).
func TestR_InetInfoGetVersion_Error(t *testing.T) {
	resp := []byte{
		0x00, 0x00, 0x00, 0x00, // pdwVersion
		0x05, 0x00, 0x00, 0x00, // return code (ERROR_ACCESS_DENIED)
	}
	_, err := functions.R_InetInfoGetVersion(&responder{resp: resp}, nil, 0)
	if err == nil {
		t.Fatal("expected an error for a nonzero return code")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("0x00000005")) {
		t.Errorf("error = %q, want it to contain 0x00000005", err.Error())
	}
}

// queryStatsRespMirror reproduces the wire shape of the R_InetInfoQueryStatistics
// response: the inline [out, switch_is] union (a top-level [ref] pointer contributes
// no referent id) followed by the return code. Marshalling it yields a canned server
// stub whose decode by the wrapper is the property under test.
type queryStatsRespMirror struct {
	StatsInfo msirp.INET_INFO_STATISTICS_INFO
	Status    uint32 `ndr:"retval"`
}

// TestR_InetInfoQueryStatistics_Union verifies the wrapper decodes the [out] statistics
// union (case 0 → *INET_INFO_STATISTICS_0) selected by the inline discriminant.
func TestR_InetInfoQueryStatistics_Union(t *testing.T) {
	stub, err := ndr.Marshal(&queryStatsRespMirror{
		StatsInfo: msirp.INET_INFO_STATISTICS_INFO{
			Tag: 0,
			InetStats0: &msirp.INET_INFO_STATISTICS_0{
				CacheCtrs:    msirp.INETA_CACHE_STATISTICS{FilesCached: 12, CurrentFileCacheSize: 0x1_2345_6789},
				AtqCtrs:      msirp.INETA_ATQ_STATISTICS{TotalAllowedRequests: 99},
				NAuxCounters: 1,
				RgCounters:   [20]ndr.DWORD{7},
			},
		},
		Status: 0,
	})
	if err != nil {
		t.Fatalf("marshal canned response: %v", err)
	}

	stats, err := functions.R_InetInfoQueryStatistics(&responder{resp: stub}, nil, 0, 0)
	if err != nil {
		t.Fatalf("R_InetInfoQueryStatistics: %v", err)
	}
	if stats.Tag != 0 {
		t.Fatalf("union Tag = %d, want 0", stats.Tag)
	}
	if stats.InetStats0 == nil {
		t.Fatal("InetStats0 arm decoded as nil")
	}
	if stats.InetStats0.CacheCtrs.FilesCached != 12 ||
		stats.InetStats0.CacheCtrs.CurrentFileCacheSize != 0x1_2345_6789 ||
		stats.InetStats0.AtqCtrs.TotalAllowedRequests != 99 {
		t.Errorf("decoded stats mismatch: %+v", stats.InetStats0)
	}
}

// TestR_InetInfoGetAdminInformation_Null verifies the double-pointer [out] result
// (LPINET_INFO_CONFIG_INFO *ppConfig) decodes a NULL inner [unique] pointer as a nil
// return without error.
func TestR_InetInfoGetAdminInformation_Null(t *testing.T) {
	resp := []byte{
		0x00, 0x00, 0x00, 0x00, // ppConfig NULL referent
		0x00, 0x00, 0x00, 0x00, // return code (ERROR_SUCCESS)
	}
	cfg, err := functions.R_InetInfoGetAdminInformation(&responder{resp: resp}, nil, 0)
	if err != nil {
		t.Fatalf("R_InetInfoGetAdminInformation: %v", err)
	}
	if cfg != nil {
		t.Errorf("ppConfig = %+v, want nil", cfg)
	}
}
