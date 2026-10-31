package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// sentRequestTrailer checks one sent authenticated request fragment against the
// sealing rules: it fits in maxXmit, the sec_trailer is 16-byte aligned from the start
// of the stub ([MS-RPCE] 2.2.2.11), and auth_pad_length accounts for every byte between
// the request's stub chunk and the sec_trailer.
func sentRequestTrailer(frag []byte, maxXmit int) (stubLen int, err error) {
	hdr, err := pdu.PeekHeader(frag)
	if err != nil {
		return 0, err
	}
	if int(hdr.FragLength) != len(frag) {
		return 0, fmt.Errorf("frag_length %d, fragment is %d bytes", hdr.FragLength, len(frag))
	}
	if len(frag) > maxXmit {
		return 0, fmt.Errorf("frag_length %d exceeds max_xmit_frag %d", len(frag), maxXmit)
	}
	trailerAt := len(frag) - int(hdr.AuthLength) - pdu.SecTrailerSize
	paddedStub := trailerAt - requestHeaderOverhead
	if paddedStub%16 != 0 {
		return 0, fmt.Errorf("sec_trailer at stub offset %d is not 16-byte aligned", paddedStub)
	}
	pad := int(frag[trailerAt+2])
	if pad > 15 {
		return 0, fmt.Errorf("auth_pad_length %d exceeds 15", pad)
	}
	return paddedStub - pad, nil
}

func TestSendRequestSecTrailerAlignment(t *testing.T) {
	kerberos := func(etype int, key []byte) SecurityContext {
		return &kerberosSecurityContext{ctx: &gssapi.SecContext{SessionKey: key, SessionEType: etype}, spnego: true}
	}
	for _, tc := range []struct {
		name     string
		authType uint8
		sec      SecurityContext
	}{
		// A fixed 16-byte token that leaves the stub length unchanged, as NTLM does.
		{"fixed token", pdu.AuthTypeNTLMSSP, &fakeSecCtx{tokenLen: 16}},
		{"kerberos rc4", pdu.AuthTypeGSSNegotiate, kerberos(iana.ETypeRC4HMAC, bytes.Repeat([]byte{0x11}, 16))},
		{"kerberos aes256", pdu.AuthTypeGSSNegotiate, kerberos(iana.ETypeAES256CTSHMACSHA196, bytes.Repeat([]byte{0x22}, 32))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const maxXmit = 1001
			ft := &fakeTransport{maxXmit: maxXmit, maxRecv: maxXmit}
			c := NewClient(ft)
			c.authType = tc.authType
			c.authLevel = pdu.AuthLevelPktPrivacy
			c.sec = tc.sec
			c.sendFragMax = maxXmit
			c.callID = 2

			stub := make([]byte, 2000)
			for i := range stub {
				stub[i] = byte(i)
			}
			if err := c.sendRequest(1, stub); err != nil {
				t.Fatalf("sendRequest: %v", err)
			}
			if len(ft.sent) < 2 {
				t.Fatalf("a %d-byte stub was sent in %d fragment(s), want several", len(stub), len(ft.sent))
			}

			total := 0
			for i, frag := range ft.sent {
				n, err := sentRequestTrailer(frag, maxXmit)
				if err != nil {
					t.Fatalf("fragment %d: %v", i, err)
				}
				// alloc_hint is the stub bytes remaining from this fragment on.
				if hint := int(binary.LittleEndian.Uint32(frag[pdu.HeaderSize : pdu.HeaderSize+4])); hint != len(stub)-total {
					t.Errorf("fragment %d: alloc_hint %d, want %d", i, hint, len(stub)-total)
				}
				total += n
			}
			if total != len(stub) {
				t.Errorf("fragments carry %d stub bytes once auth padding is removed, want %d", total, len(stub))
			}
		})
	}
}
