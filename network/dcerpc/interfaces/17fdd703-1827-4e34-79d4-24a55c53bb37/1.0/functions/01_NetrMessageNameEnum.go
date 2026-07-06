package functions

// IDL source: [MS-MSRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-msrp/181965ff-fab4-4ad4-a8d7-16b444cc4e66
// A fetched copy is kept at ms-msrp.idl in the interface directory.

import (
	"fmt"

	msgsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/17fdd703-1827-4e34-79d4-24a55c53bb37/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmsrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-msrp"
)

// netrMessageNameEnumRequest carries the [in] parameters of NetrMessageNameEnum.
type netrMessageNameEnumRequest struct {
	ServerName   *ndr.WSTR `ndr:"unique"`
	InfoStruct   msmsrp.MSG_ENUM_STRUCT
	PrefMaxLen   ndr.DWORD
	ResumeHandle *ndr.DWORD `ndr:"unique"`
}

func (*netrMessageNameEnumRequest) Opnum() uint16 { return msgsvc.OpnumNetrMessageNameEnum }

// netrMessageNameEnumResponse carries the [out] parameters and return value of NetrMessageNameEnum.
type netrMessageNameEnumResponse struct {
	InfoStruct   msmsrp.MSG_ENUM_STRUCT
	TotalEntries ndr.DWORD
	ResumeHandle *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// NetrMessageNameEnum calls NetrMessageNameEnum (opnum 1) ([MS-MSRP] — verify the parameter
// modeling and status handling).
func NetrMessageNameEnum(rpc ndr.Invoker, serverName *ndr.WSTR, infoStruct msmsrp.MSG_ENUM_STRUCT, prefMaxLen ndr.DWORD, resumeHandle *ndr.DWORD) (InfoStruct msmsrp.MSG_ENUM_STRUCT, TotalEntries ndr.DWORD, ResumeHandle *ndr.DWORD, err error) {
	req := &netrMessageNameEnumRequest{
		ServerName:   serverName,
		InfoStruct:   infoStruct,
		PrefMaxLen:   prefMaxLen,
		ResumeHandle: resumeHandle,
	}
	var resp netrMessageNameEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrMessageNameEnum: %w", err)
		return
	}
	InfoStruct = resp.InfoStruct
	TotalEntries = resp.TotalEntries
	ResumeHandle = resp.ResumeHandle
	if uint32(resp.Status) != msgsvc.StatusSuccess {
		err = fmt.Errorf("NetrMessageNameEnum failed: %s", msgsvc.StatusString(uint32(resp.Status)))
	}
	return
}
