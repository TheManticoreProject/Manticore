package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrLogonComputeServerDigestRequest carries the [in] parameters of NetrLogonComputeServerDigest.
type netrLogonComputeServerDigestRequest struct {
	ServerName  *ndr.WSTR `ndr:"unique"`
	Rid         ndr.DWORD
	Message     []uint8 `ndr:"ref,size_is=MessageSize"`
	MessageSize ndr.DWORD
}

func (*netrLogonComputeServerDigestRequest) Opnum() uint16 {
	return logon.OpnumNetrLogonComputeServerDigest
}

// netrLogonComputeServerDigestResponse carries the [out] parameters and return value of NetrLogonComputeServerDigest.
type netrLogonComputeServerDigestResponse struct {
	NewMessageDigest [16]int8
	OldMessageDigest [16]int8
	Status           ndr.DWORD `ndr:"retval"`
}

// NetrLogonComputeServerDigest calls NetrLogonComputeServerDigest (opnum 24) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrLogonComputeServerDigest(rpc ndr.Invoker, serverName *ndr.WSTR, rid ndr.DWORD, message []uint8, messageSize ndr.DWORD) (NewMessageDigest [16]int8, OldMessageDigest [16]int8, err error) {
	req := &netrLogonComputeServerDigestRequest{
		ServerName:  serverName,
		Rid:         rid,
		Message:     message,
		MessageSize: messageSize,
	}
	var resp netrLogonComputeServerDigestResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrLogonComputeServerDigest: %w", err)
		return
	}
	NewMessageDigest = resp.NewMessageDigest
	OldMessageDigest = resp.OldMessageDigest
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrLogonComputeServerDigest failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
