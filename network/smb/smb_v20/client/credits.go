package client

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
)

// creditGranularity is the payload size, in bytes, that a single credit covers.
// [MS-SMB2] 3.1.5.2 counts a request against the credits it holds in 64 KiB
// units, so a one-credit request may carry at most this many bytes of payload.
const creditGranularity = 65536

// targetCredits is the credit window the client tries to hold. The server grants
// only one credit alongside the NEGOTIATE response, so without asking for more
// the client could never issue a request larger than one credit's worth. 128
// credits corresponds to the 8 MiB maximum a Windows server advertises, which is
// the largest single request the negotiated limits allow.
const targetCredits = 128

// creditsFor returns the number of credits a payload of n bytes consumes:
// (n - 1) / 65536 + 1, per [MS-SMB2] 3.1.5.2. An empty payload still charges one
// credit, because every request costs at least one.
func creditsFor(n uint32) uint16 {
	if n <= creditGranularity {
		return 1
	}
	charge := (n-1)/creditGranularity + 1
	// A charge cannot exceed the 16-bit field; the payload clamp in
	// maxPayloadForRequest keeps callers well below this in practice.
	if charge > 0xFFFF {
		return 0xFFFF
	}
	return uint16(charge)
}

// requestCreditCharge returns the CreditCharge a request must carry. The charge
// covers the larger of the payload sent and the payload expected back
// ([MS-SMB2] 3.1.5.2), so a read charges for the data it asks for even though
// the request itself is small.
//
// Commands whose payload is bounded by the fixed part of the structure — every
// command not listed here — charge a single credit.
func requestCreditCharge(command command_interface.CommandInterface) uint16 {
	var payload uint32

	switch cmd := command.(type) {
	case *commands.ReadRequest:
		// The response carries Length bytes back.
		payload = uint32(cmd.Length)
	case *commands.WriteRequest:
		payload = uint32(len(cmd.Data))
	case *commands.QueryDirectoryRequest:
		payload = uint32(cmd.OutputBufferLength)
	case *commands.QueryInfoRequest:
		payload = maxU32(uint32(cmd.OutputBufferLength), uint32(len(cmd.Input)))
	case *commands.SetInfoRequest:
		payload = uint32(len(cmd.Buffer))
	case *commands.IoctlRequest:
		payload = maxU32(uint32(cmd.MaxOutputResponse), uint32(len(cmd.Input)))
	case *commands.ChangeNotifyRequest:
		payload = uint32(cmd.OutputBufferLength)
	}

	return creditsFor(payload)
}

// maxPayloadForRequest returns the largest payload the client may currently ask
// for in a single request: the smaller of the server's advertised ceiling and
// what the credits on hand can pay for.
//
// Sizing a request from the server's ceiling alone is what makes an 8 MiB
// request go out against a one-credit grant, which the server rejects with
// STATUS_INVALID_PARAMETER ([MS-SMB2] 3.3.5.2.5).
func (c *Client) maxPayloadForRequest(serverMax uint32) uint32 {
	// Before the first response arrives the client holds the single credit the
	// server grants with NEGOTIATE.
	held := c.Connection.Credits
	if held == 0 {
		held = 1
	}

	// The SMB 2.0.2 dialect has no credit model: CreditCharge MUST be 0 and a
	// request may not exceed one credit's worth of payload.
	if c.Connection.Dialect < dialects.SMB2_DIALECT_2_1_0 {
		held = 1
	}

	budget := uint32(held) * creditGranularity
	if serverMax > 0 && serverMax < budget {
		return serverMax
	}
	return budget
}

// creditRequest returns the CreditRequest to place on an outgoing request: enough
// to replace the credits this request spends, plus enough to grow the window
// toward targetCredits. The server grants what it chooses; asking is what lets
// the window grow past the single credit NEGOTIATE leaves the client with.
//
// Before a dialect that supports multi-credit is negotiated there is no window to
// grow — the SMB 2.0.2 dialect has no credit model at all — so the request asks
// only for what it spends.
func (c *Client) creditRequest(charge uint16) uint16 {
	want := charge
	if c.Connection.Dialect >= dialects.SMB2_DIALECT_2_1_0 {
		if held := c.Connection.Credits; held < targetCredits {
			if shortfall := targetCredits - held; shortfall > want {
				want = shortfall
			}
		}
	}
	if want == 0 {
		want = 1
	}
	return want
}

// spendCredits deducts what a request charges from the credits on hand. The
// server replenishes them through the Credit field of each response.
func (c *Client) spendCredits(charge uint16) {
	if c.Connection.Credits >= charge {
		c.Connection.Credits -= charge
		return
	}
	c.Connection.Credits = 0
}

// grantCredits adds the credits a response carries to the credits on hand.
func (c *Client) grantCredits(granted uint16) {
	total := uint32(c.Connection.Credits) + uint32(granted)
	if total > 0xFFFF {
		total = 0xFFFF
	}
	c.Connection.Credits = uint16(total)
}

func maxU32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}
