package functions

// IDL source: [MS-MQDS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqds/7907bc25-e4e6-40ef-b990-9172d1808e94
// A fetched copy is kept at ms-mqds.idl in the interface directory.

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// s_DSSetPropsRequest carries the [in] parameters of S_DSSetProps.
type s_DSSetPropsRequest struct {
	DwObjectType ndr.DWORD
	PwcsPathName ndr.WSTR
	Cp           ndr.DWORD
	AProp        []ndr.DWORD          `ndr:"ref,size_is=Cp"`
	ApVar        []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
}

func (*s_DSSetPropsRequest) Opnum() uint16 { return dscomm.OpnumS_DSSetProps }

// s_DSSetPropsResponse carries the [out] parameters and return value of S_DSSetProps.
type s_DSSetPropsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// S_DSSetProps calls S_DSSetProps (opnum 3) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSSetProps(rpc ndr.Invoker, dwObjectType ndr.DWORD, pwcsPathName ndr.WSTR, cp ndr.DWORD, aProp []ndr.DWORD, apVar []msmqmq.PROPVARIANT) (err error) {
	req := &s_DSSetPropsRequest{
		DwObjectType: dwObjectType,
		PwcsPathName: pwcsPathName,
		Cp:           cp,
		AProp:        aProp,
		ApVar:        apVar,
	}
	var resp s_DSSetPropsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSSetProps: %w", err)
		return
	}
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSSetProps failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}
