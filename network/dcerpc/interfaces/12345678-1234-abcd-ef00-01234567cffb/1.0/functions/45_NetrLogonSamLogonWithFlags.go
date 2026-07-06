package functions

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// netrLogonSamLogonWithFlagsRequest carries the [in] parameters of NetrLogonSamLogonWithFlags.
type netrLogonSamLogonWithFlagsRequest struct {
	LogonServer         *ndr.WSTR                      `ndr:"unique"`
	ComputerName        *ndr.WSTR                      `ndr:"unique"`
	Authenticator       *msnrpc.NETLOGON_AUTHENTICATOR `ndr:"unique"`
	ReturnAuthenticator *msnrpc.NETLOGON_AUTHENTICATOR `ndr:"unique"`
	LogonLevel          msnrpc.NETLOGON_LOGON_INFO_CLASS
	LogonInformation    msnrpc.NETLOGON_LEVEL
	ValidationLevel     msnrpc.NETLOGON_VALIDATION_INFO_CLASS
	ExtraFlags          ndr.DWORD
}

func (*netrLogonSamLogonWithFlagsRequest) Opnum() uint16 {
	return logon.OpnumNetrLogonSamLogonWithFlags
}

// netrLogonSamLogonWithFlagsResponse carries the [out] parameters and return value of NetrLogonSamLogonWithFlags.
type netrLogonSamLogonWithFlagsResponse struct {
	ReturnAuthenticator   *msnrpc.NETLOGON_AUTHENTICATOR `ndr:"unique"`
	ValidationInformation msnrpc.NETLOGON_VALIDATION
	Authoritative         uint8
	ExtraFlags            ndr.DWORD
	Status                ndr.DWORD `ndr:"retval"`
}

// NetrLogonSamLogonWithFlags calls NetrLogonSamLogonWithFlags (opnum 45) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonSamLogonWithFlags(rpc ndr.Invoker, logonServer *ndr.WSTR, computerName *ndr.WSTR, authenticator *msnrpc.NETLOGON_AUTHENTICATOR, returnAuthenticator *msnrpc.NETLOGON_AUTHENTICATOR, logonLevel msnrpc.NETLOGON_LOGON_INFO_CLASS, logonInformation msnrpc.NETLOGON_LEVEL, validationLevel msnrpc.NETLOGON_VALIDATION_INFO_CLASS, extraFlags ndr.DWORD) (ReturnAuthenticator *msnrpc.NETLOGON_AUTHENTICATOR, ValidationInformation msnrpc.NETLOGON_VALIDATION, Authoritative uint8, ExtraFlags ndr.DWORD, err error) {
	req := &netrLogonSamLogonWithFlagsRequest{
		LogonServer:         logonServer,
		ComputerName:        computerName,
		Authenticator:       authenticator,
		ReturnAuthenticator: returnAuthenticator,
		LogonLevel:          logonLevel,
		LogonInformation:    logonInformation,
		ValidationLevel:     validationLevel,
		ExtraFlags:          extraFlags,
	}
	var resp netrLogonSamLogonWithFlagsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonSamLogonWithFlags: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	ValidationInformation = resp.ValidationInformation
	Authoritative = resp.Authoritative
	ExtraFlags = resp.ExtraFlags
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonSamLogonWithFlags failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
