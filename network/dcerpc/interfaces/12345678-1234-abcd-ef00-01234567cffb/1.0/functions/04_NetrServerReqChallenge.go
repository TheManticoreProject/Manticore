package functions

import (
	"fmt"

	netlogon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

type netrServerReqChallengeRequest struct {
	PrimaryName     *ndr.WSTR `ndr:"unique"`
	ComputerName    ndr.WSTR
	ClientChallenge structures.NETLOGON_CREDENTIAL
}

func (*netrServerReqChallengeRequest) Opnum() uint16 { return netlogon.OpnumNetrServerReqChallenge }

type netrServerReqChallengeResponse struct {
	ServerChallenge structures.NETLOGON_CREDENTIAL
	Status          ndr.DWORD `ndr:"retval"`
}

// NetrServerReqChallenge calls NetrServerReqChallenge ([MS-NRPC] 3.5.4.4.1, opnum 4): the
// client sends a challenge and receives the server's challenge, the first step of secure
// channel negotiation.
//
// Parameters:
//   - rpc: A DCE/RPC client bound to the Netlogon interface.
//   - primaryName: The DC name (empty string sends a NULL unique pointer).
//   - computerName: The NetBIOS name of the client computer.
//   - clientChallenge: The client challenge.
//
// Returns:
//   - The server challenge, the NTSTATUS return value, and a transport error.
func NetrServerReqChallenge(rpc ndr.Invoker, primaryName, computerName string, clientChallenge structures.NETLOGON_CREDENTIAL) (structures.NETLOGON_CREDENTIAL, uint32, error) {
	req := &netrServerReqChallengeRequest{
		ComputerName:    ndr.WSTR(computerName),
		ClientChallenge: clientChallenge,
	}
	if primaryName != "" {
		pn := ndr.WSTR(primaryName)
		req.PrimaryName = &pn
	}

	var resp netrServerReqChallengeResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.NETLOGON_CREDENTIAL{}, 0, fmt.Errorf("NetrServerReqChallenge: %w", err)
	}
	return resp.ServerChallenge, uint32(resp.Status), nil
}
