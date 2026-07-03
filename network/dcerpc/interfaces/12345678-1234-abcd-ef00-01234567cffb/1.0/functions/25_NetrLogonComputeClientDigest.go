package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrLogonComputeClientDigestRequest carries the [in] parameters of NetrLogonComputeClientDigest.
type netrLogonComputeClientDigestRequest struct {
	ServerName  *ndr.WSTR `ndr:"unique"`
	DomainName  *ndr.WSTR `ndr:"unique"`
	Message     []uint8   `ndr:"ref,size_is=MessageSize"`
	MessageSize ndr.DWORD
}

func (*netrLogonComputeClientDigestRequest) Opnum() uint16 {
	return logon.OpnumNetrLogonComputeClientDigest
}

// netrLogonComputeClientDigestResponse carries the [out] parameters and return value of NetrLogonComputeClientDigest.
type netrLogonComputeClientDigestResponse struct {
	NewMessageDigest [16]int8
	OldMessageDigest [16]int8
	Status           ndr.DWORD `ndr:"retval"`
}

// NetrLogonComputeClientDigest calls NetrLogonComputeClientDigest (opnum 25) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonComputeClientDigest(rpc ndr.Invoker, serverName *ndr.WSTR, domainName *ndr.WSTR, message []uint8, messageSize ndr.DWORD) (NewMessageDigest [16]int8, OldMessageDigest [16]int8, err error) {
	req := &netrLogonComputeClientDigestRequest{
		ServerName:  serverName,
		DomainName:  domainName,
		Message:     message,
		MessageSize: messageSize,
	}
	var resp netrLogonComputeClientDigestResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonComputeClientDigest: %w", err)
		return
	}
	NewMessageDigest = resp.NewMessageDigest
	OldMessageDigest = resp.OldMessageDigest
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonComputeClientDigest failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
