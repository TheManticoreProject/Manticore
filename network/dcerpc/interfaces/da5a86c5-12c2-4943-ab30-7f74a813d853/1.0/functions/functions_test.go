package functions_test

// IDL source: [MS-PCQ] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-pcq/dcee10e3-0512-495e-9566-26e56cc21c5c
// A fetched copy is kept at ms-pcq.idl in the interface directory.

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/da5a86c5-12c2-4943-ab30-7f74a813d853/1.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspcq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-pcq"
)

// responder is an ndr.Invoker that replies with a canned response stub.
type responder struct {
	resp []byte
}

func (r *responder) Invoke(in ndr.Call, out any) error {
	if _, err := ndr.Request(in); err != nil {
		return err
	}
	if r.resp == nil {
		return nil
	}
	return ndr.Response(r.resp, out)
}

// queryCounterInfoRespMirror reproduces the wire shape of the PerflibV2QueryCounterInfo
// response so a canned server stub can be produced by marshalling it.
type queryCounterInfoRespMirror struct {
	PdwOutSize ndr.DWORD
	PdwRtnSize ndr.DWORD
	LpData     []uint8   `ndr:"ref,varying"`
	Status     ndr.DWORD `ndr:"retval"`
}

func queryCounterInfoStub(t *testing.T, status ndr.DWORD, rtnSize ndr.DWORD) []byte {
	t.Helper()
	stub, err := ndr.Marshal(&queryCounterInfoRespMirror{
		PdwOutSize: 0,
		PdwRtnSize: rtnSize,
		LpData:     []uint8{},
		Status:     status,
	})
	if err != nil {
		t.Fatalf("marshal canned response: %v", err)
	}
	return stub
}

// TestPerflibV2QueryCounterInfo_SizeProbeTolerated pins the size-probe tolerance of the
// query methods: [MS-PCQ] 3.1.4.6 has the server answer a dwInSize of 0 with
// ERROR_NOT_ENOUGH_MEMORY (0x00000008) and the required buffer size in pdwRtnSize, and
// the stub must report that as a successful probe rather than a failure. This is the term
// that a rewrite of the status condition could drop while still compiling.
func TestPerflibV2QueryCounterInfo_SizeProbeTolerated(t *testing.T) {
	stub := queryCounterInfoStub(t, 0x00000008, 4096)
	_, rtnSize, _, err := functions.PerflibV2QueryCounterInfo(&responder{resp: stub}, mspcq.RPC_HQUERY{}, 0)
	if err != nil {
		t.Fatalf("ERROR_NOT_ENOUGH_MEMORY must be tolerated as the size probe, got error: %v", err)
	}
	if rtnSize != 4096 {
		t.Errorf("pdwRtnSize = %d, want 4096", rtnSize)
	}
}

// TestPerflibV2QueryCounterInfo_Success pins that a zero return value is not an error.
func TestPerflibV2QueryCounterInfo_Success(t *testing.T) {
	stub := queryCounterInfoStub(t, 0x00000000, 0)
	if _, _, _, err := functions.PerflibV2QueryCounterInfo(&responder{resp: stub}, mspcq.RPC_HQUERY{}, 0); err != nil {
		t.Fatalf("ERROR_SUCCESS must not be an error, got: %v", err)
	}
}

// TestPerflibV2QueryCounterInfo_Error pins that a code outside the tolerated pair still
// fails, and that the failure names the code through the shared [MS-ERREF] 2.2 table.
func TestPerflibV2QueryCounterInfo_Error(t *testing.T) {
	stub := queryCounterInfoStub(t, 0x00000005, 0)
	_, _, _, err := functions.PerflibV2QueryCounterInfo(&responder{resp: stub}, mspcq.RPC_HQUERY{}, 0)
	if err == nil {
		t.Fatal("expected an error for ERROR_ACCESS_DENIED")
	}
	if !strings.Contains(err.Error(), "ERROR_ACCESS_DENIED") {
		t.Errorf("error = %q, want it to name ERROR_ACCESS_DENIED", err.Error())
	}
}

// validateCountersRespMirror reproduces the wire shape of the PerflibV2ValidateCounters
// response, whose lpData is a conformant (non-varying) in/out array.
type validateCountersRespMirror struct {
	LpData []uint8   `ndr:"ref,conformant"`
	Status ndr.DWORD `ndr:"retval"`
}

// TestPerflibV2ValidateCounters_NoSizeProbeTolerance pins that the methods which never
// had the size-probe tolerance did not acquire one: PerflibV2ValidateCounters takes a
// caller-sized in/out buffer, so ERROR_NOT_ENOUGH_MEMORY is a plain failure here and must
// surface as an error naming that code.
func TestPerflibV2ValidateCounters_NoSizeProbeTolerance(t *testing.T) {
	stub, err := ndr.Marshal(&validateCountersRespMirror{
		LpData: []uint8{},
		Status: 0x00000008,
	})
	if err != nil {
		t.Fatalf("marshal canned response: %v", err)
	}
	_, err = functions.PerflibV2ValidateCounters(&responder{resp: stub}, mspcq.RPC_HQUERY{}, 0, []uint8{}, 0)
	if err == nil {
		t.Fatal("expected an error: ValidateCounters has no size-probe form")
	}
	if !strings.Contains(err.Error(), "ERROR_NOT_ENOUGH_MEMORY") {
		t.Errorf("error = %q, want it to name ERROR_NOT_ENOUGH_MEMORY", err.Error())
	}
}
