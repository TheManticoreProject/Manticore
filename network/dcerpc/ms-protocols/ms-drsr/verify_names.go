package msdrsr

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	drsrtypes "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// DRS_VERIFY_DSNAMES is the IDL_DRSVerifyNames flag selecting verification by DSNAME
// (the names are distinguished names) ([MS-DRSR] 4.1.26.2).
const DRS_VERIFY_DSNAMES = 0

// VerifiedName is the result of verifying one object name: the resolved GUID (zero if the
// object was not found) and its returned distinguished name.
type VerifiedName struct {
	Input string
	DN    string
	GUID  guid.GUID
	Found bool
}

// VerifyNames checks whether the given distinguished names resolve to objects on the DC
// via IDL_DRSVerifyNames (opnum 8), returning one result per input in order. It requests
// no attributes (existence only). This is read-only.
func (c *Client) VerifyNames(dns []string) ([]VerifiedName, error) {
	if !c.bound {
		return nil, fmt.Errorf("msdrsr: not connected")
	}
	rp := make([]*drsrtypes.DSNAME, len(dns))
	for i, dn := range dns {
		d := drsrtypes.NewDSNameFromDN(dn)
		rp[i] = &d
	}
	msgIn := drsrtypes.DRS_MSG_VERIFYREQ{
		Tag: 1,
		V1: drsrtypes.DRS_MSG_VERIFYREQ_V1{
			DwFlags: DRS_VERIFY_DSNAMES,
			CNames:  ndr.DWORD(len(dns)),
			RpNames: rp,
		},
	}
	_, msgOut, err := functions.IDL_DRSVerifyNames(c.rpc, c.handle, 1, msgIn)
	if err != nil {
		return nil, fmt.Errorf("msdrsr: VerifyNames: %w", err)
	}
	out := make([]VerifiedName, 0, len(dns))
	for i, e := range msgOut.V1.RpEntInf {
		vn := VerifiedName{}
		if i < len(dns) {
			vn.Input = dns[i]
		}
		if e.PName != nil {
			vn.GUID = e.PName.Guid.GUID()
			vn.DN = decodeWChars(e.PName.StringName)
			vn.Found = !e.PName.Guid.IsZero()
		}
		out = append(out, vn)
	}
	return out, nil
}
