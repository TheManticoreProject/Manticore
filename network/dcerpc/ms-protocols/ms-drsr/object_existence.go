package msdrsr

import (
	"bytes"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	drsrtypes "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// ObjectExistenceResult reports whether the server's digest matches the supplied digest.
// When it does not match, GUIDs contains the server's object sequence for the requested
// range after applying the supplied up-to-date vector.
type ObjectExistenceResult struct {
	DigestMatches bool
	GUIDs         []guid.GUID
}

// GetObjectExistence calls IDL_DRSGetObjectExistence (opnum 23) for a range in a naming
// context. start must identify the first object in the client's sequence and count is the
// maximum number of objects in that sequence. upToDateVector is the merged client/server
// replication state used to exclude objects that have not reached both replicas. The
// digest is the MD5 digest of the client's GUID sequence; a mismatch makes the server
// return its sequence for reconciliation.
func (c *Client) GetObjectExistence(ncDN string, start guid.GUID, count uint32, upToDateVector *drsrtypes.UPTODATE_VECTOR_V1_EXT, digest [16]byte) (*ObjectExistenceResult, error) {
	if !c.bound {
		return nil, fmt.Errorf("msdrsr: not connected")
	}
	if ncDN == "" {
		return nil, fmt.Errorf("msdrsr: GetObjectExistence: naming context is empty")
	}
	startUUID := drsrtypes.UUIDFromGUID(start)
	if startUUID.IsZero() {
		return nil, fmt.Errorf("msdrsr: GetObjectExistence: start GUID is zero")
	}
	if count == 0 {
		return nil, fmt.Errorf("msdrsr: GetObjectExistence: count is zero")
	}
	if upToDateVector == nil {
		return nil, fmt.Errorf("msdrsr: GetObjectExistence: up-to-date vector is nil")
	}
	if upToDateVector.DwVersion != 1 {
		return nil, fmt.Errorf("msdrsr: GetObjectExistence: up-to-date vector version is %d, want 1", upToDateVector.DwVersion)
	}
	if int(upToDateVector.CNumCursors) != len(upToDateVector.RgCursors) {
		return nil, fmt.Errorf("msdrsr: GetObjectExistence: up-to-date vector count is %d, have %d cursors", upToDateVector.CNumCursors, len(upToDateVector.RgCursors))
	}
	for i := 1; i < len(upToDateVector.RgCursors); i++ {
		if bytes.Compare(upToDateVector.RgCursors[i-1].UuidDsa.Octets[:], upToDateVector.RgCursors[i].UuidDsa.Octets[:]) > 0 {
			return nil, fmt.Errorf("msdrsr: GetObjectExistence: up-to-date vector cursors are not sorted by invocation GUID")
		}
	}

	nc := drsrtypes.NewDSNameFromDN(ncDN)
	msgIn := drsrtypes.DRS_MSG_EXISTREQ{
		Tag: 1,
		V1: drsrtypes.DRS_MSG_EXISTREQ_V1{
			GuidStart:            startUUID,
			CGuids:               ndr.DWORD(count),
			PNC:                  &nc,
			PUpToDateVecCommonV1: upToDateVector,
			Md5Digest:            digest,
		},
	}
	_, msgOut, err := functions.IDL_DRSGetObjectExistence(c.rpc, c.handle, 1, msgIn)
	if err != nil {
		return nil, fmt.Errorf("msdrsr: GetObjectExistence: %w", err)
	}

	result := &ObjectExistenceResult{DigestMatches: msgOut.V1.DwStatusFlags&1 != 0}
	result.GUIDs = make([]guid.GUID, 0, len(msgOut.V1.RgGuids))
	for _, objectGUID := range msgOut.V1.RgGuids {
		result.GUIDs = append(result.GUIDs, objectGUID.GUID())
	}
	return result, nil
}
