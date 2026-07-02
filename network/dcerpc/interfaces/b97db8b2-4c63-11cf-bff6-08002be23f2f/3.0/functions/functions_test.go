package functions_test

import (
	"bytes"
	"testing"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// captureInvoker records the marshalled request stub and opnum without any network I/O,
// so the on-the-wire NDR layout of the ClusAPI requests can be asserted. Invoke leaves
// the response zero, so status-returning stubs see ERROR_SUCCESS and return no error.
type captureInvoker struct {
	stub  []byte
	opnum uint16
}

func (c *captureInvoker) Invoke(in ndr.Call, out any) error {
	b, err := ndr.Request(in)
	if err != nil {
		return err
	}
	c.stub = b
	c.opnum = in.Opnum()
	return nil
}

// utf16 returns the UTF-16LE bytes of an ASCII string, to search for it inside a stub.
func utf16(s string) []byte {
	b := make([]byte, 0, len(s)*2)
	for i := 0; i < len(s); i++ {
		b = append(b, s[i], 0)
	}
	return b
}

// TestApiOpenClusterMarshal checks the parameterless open (opnum 0) whose return value
// is the HCLUSTER_RPC context handle (retval), decoded from a zero response as the
// zero handle with no error.
func TestApiOpenClusterMarshal(t *testing.T) {
	cap := &captureInvoker{}
	h, status, err := functions.ApiOpenCluster(cap)
	if err != nil {
		t.Fatalf("ApiOpenCluster: %v", err)
	}
	if cap.opnum != clusapi.OpnumApiOpenCluster {
		t.Fatalf("opnum = %d, want %d", cap.opnum, clusapi.OpnumApiOpenCluster)
	}
	if h != (mscmrp.HCLUSTER_RPC{}) || status != clusapi.StatusSuccess {
		t.Fatalf("zero response: handle=%v status=%#x", h, status)
	}
	if len(cap.stub) != 0 {
		t.Fatalf("ApiOpenCluster takes no [in] parameters, stub = %x", cap.stub)
	}
}

// TestApiSetClusterNameMarshal checks opnum 2 and that the [in,string] cluster name is
// carried as a UTF-16LE NDR string.
func TestApiSetClusterNameMarshal(t *testing.T) {
	cap := &captureInvoker{}
	if _, err := functions.ApiSetClusterName(cap, ndr.WSTR("CLUSTER1")); err != nil {
		t.Fatalf("ApiSetClusterName: %v", err)
	}
	if cap.opnum != clusapi.OpnumApiSetClusterName {
		t.Fatalf("opnum = %d, want %d", cap.opnum, clusapi.OpnumApiSetClusterName)
	}
	if !bytes.Contains(cap.stub, utf16("CLUSTER1")) {
		t.Fatalf("UTF-16LE %q not found in stub %x", "CLUSTER1", cap.stub)
	}
}

// TestApiCreateResourceMarshal checks opnum 9 and that the leading HGROUP_RPC context
// handle plus the resource name/type strings are marshalled.
func TestApiCreateResourceMarshal(t *testing.T) {
	cap := &captureInvoker{}
	_, _, _, err := functions.ApiCreateResource(cap,
		mscmrp.HGROUP_RPC{}, ndr.WSTR("Disk1"), ndr.WSTR("Physical Disk"), 0)
	if err != nil {
		t.Fatalf("ApiCreateResource: %v", err)
	}
	if cap.opnum != clusapi.OpnumApiCreateResource {
		t.Fatalf("opnum = %d, want %d", cap.opnum, clusapi.OpnumApiCreateResource)
	}
	if !bytes.Contains(cap.stub, utf16("Disk1")) {
		t.Fatalf("UTF-16LE %q not found in stub %x", "Disk1", cap.stub)
	}
	if !bytes.Contains(cap.stub, utf16("Physical Disk")) {
		t.Fatalf("UTF-16LE %q not found in stub %x", "Physical Disk", cap.stub)
	}
}
