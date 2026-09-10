package wirediff

// Offsets within the fixed 64-byte SMB2 header ([MS-SMB2] 2.2.1.2).
const (
	smb2HeaderSize      = 64
	smb2MessageIdOffset = 24
	smb2TreeIdOffset    = 36
	smb2SessionIdOffset = 40
	smb2SignatureOffset = 48
)

// SMB2HeaderMask is the set of header fields that legitimately differ between
// two recordings of the same exchange.
//
// Nothing else in the header is excused. CreditCharge, Credit, Flags and
// NextCommand are all deterministic for a given request, and a change in any of
// them is exactly what this comparison exists to catch.
func SMB2HeaderMask() Mask {
	return Mask{
		// MessageId advances per request and per credit charged, so it depends
		// on how much traffic preceded this message.
		{Name: "SMB2 header MessageId", Start: smb2MessageIdOffset, Len: 8},
		// TreeId and SessionId are assigned by the server.
		{Name: "SMB2 header TreeId", Start: smb2TreeIdOffset, Len: 4},
		{Name: "SMB2 header SessionId", Start: smb2SessionIdOffset, Len: 8},
		// The signature is keyed on a session key derived from fresh challenges.
		{Name: "SMB2 header Signature", Start: smb2SignatureOffset, Len: 16},
	}
}

// SMB2NegotiateRequestMask masks the header fields above plus the two parts of a
// NEGOTIATE request that carry fresh randomness.
//
// It deliberately does *not* mask Capabilities, SecurityMode, ClientStartTime or
// the dialect list. Those are the fields a passive observer uses to tell one
// implementation from another, so they are the ones worth pinning.
//
// contextsOffset is where the negotiate context list begins, which varies with
// the number of dialects offered; pass 0 to mask only the ClientGuid.
func SMB2NegotiateRequestMask(contextsOffset int) Mask {
	mask := SMB2HeaderMask()

	// ClientGuid: identifies the client machine, and a conforming client
	// generates it per machine rather than per connection.
	mask = append(mask, Span{Name: "NEGOTIATE ClientGuid", Start: smb2HeaderSize + 12, Len: 16})

	if contextsOffset > 0 {
		// The pre-auth integrity context carries a fresh 32-byte salt. Its
		// position depends on the contexts emitted before it, so the caller
		// supplies where the list starts and everything from there is excused.
		mask = append(mask, Span{Name: "NEGOTIATE contexts (pre-auth salt)", Start: contextsOffset, Len: 1 << 16})
	}
	return mask
}

// SMB2SessionSetupRequestMask masks the header plus the whole security buffer.
//
// The buffer is a GSS-API token whose every interesting part is fresh per
// session — the client challenge, the encrypted session key, the MIC, the
// mechListMIC and a timestamp — so comparing it byte for byte across runs would
// only ever report noise. What this leaves pinned is the SESSION_SETUP structure
// around it: the flags, the security mode, the capabilities and the channel.
//
// bufferOffset is where the security buffer begins in the message.
func SMB2SessionSetupRequestMask(bufferOffset int) Mask {
	mask := SMB2HeaderMask()
	mask = append(mask, Span{Name: "SESSION_SETUP security buffer", Start: bufferOffset, Len: 1 << 16})
	return mask
}
