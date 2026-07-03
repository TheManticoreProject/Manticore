package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrLogonSamLogonExRequest carries the [in] parameters of NetrLogonSamLogonEx.
type netrLogonSamLogonExRequest struct {
	LogonServer      *ndr.WSTR `ndr:"unique"`
	ComputerName     *ndr.WSTR `ndr:"unique"`
	LogonLevel       msnrpc.NETLOGON_LOGON_INFO_CLASS
	LogonInformation msnrpc.NETLOGON_LEVEL
	ValidationLevel  msnrpc.NETLOGON_VALIDATION_INFO_CLASS
	ExtraFlags       ndr.DWORD
}

func (*netrLogonSamLogonExRequest) Opnum() uint16 { return logon.OpnumNetrLogonSamLogonEx }

// netrLogonSamLogonExResponse carries the [out] parameters and return value of NetrLogonSamLogonEx.
type netrLogonSamLogonExResponse struct {
	ValidationInformation msnrpc.NETLOGON_VALIDATION
	Authoritative         uint8
	ExtraFlags            ndr.DWORD
	Status                ndr.DWORD `ndr:"retval"`
}

// NetrLogonSamLogonEx calls NetrLogonSamLogonEx (opnum 39) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonSamLogonEx(rpc ndr.Invoker, logonServer *ndr.WSTR, computerName *ndr.WSTR, logonLevel msnrpc.NETLOGON_LOGON_INFO_CLASS, logonInformation msnrpc.NETLOGON_LEVEL, validationLevel msnrpc.NETLOGON_VALIDATION_INFO_CLASS, extraFlags ndr.DWORD) (ValidationInformation msnrpc.NETLOGON_VALIDATION, Authoritative uint8, ExtraFlags ndr.DWORD, err error) {
	req := &netrLogonSamLogonExRequest{
		LogonServer:      logonServer,
		ComputerName:     computerName,
		LogonLevel:       logonLevel,
		LogonInformation: logonInformation,
		ValidationLevel:  validationLevel,
		ExtraFlags:       extraFlags,
	}
	var resp netrLogonSamLogonExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonSamLogonEx: %w", err)
		return
	}
	ValidationInformation = resp.ValidationInformation
	Authoritative = resp.Authoritative
	ExtraFlags = resp.ExtraFlags
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonSamLogonEx failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
