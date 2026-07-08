package client

// White-box tests for the security-context seam. They use a fake securityContext to pin the
// client's PDU framing (auth_length/frag_length, byte order, pad stripping) independent of
// any provider's crypto, and confirm the client asks the provider for the token length and
// hands it the signed region and stub. NTLM crypto itself is covered in the security package.

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// fakeSecCtx is a deterministic securityContext: it "seals" by XORing the stub with 0xFF
// (reversible and observable on the wire) and emits a token of tokenLen 0xAB bytes. It
// records what it was handed so tests can assert the client passed the right slices.
type fakeSecCtx struct {
	tokenLen   int
	shortToken bool // return a wrong-length token to exercise the client's length check

	gotSignedRegion []byte
	gotStub         []byte
	gotSeal         bool
	gotAuthValue    []byte
	unprotectCalled bool
}

func (f *fakeSecCtx) AuthValueLen(bool) int { return f.tokenLen }

func (f *fakeSecCtx) ProtectRequest(signedRegion, stub []byte, seal bool) ([]byte, []byte, error) {
	f.gotSignedRegion = append([]byte(nil), signedRegion...)
	f.gotStub = append([]byte(nil), stub...)
	f.gotSeal = seal
	onWire := append([]byte(nil), stub...)
	if seal {
		for i := range onWire {
			onWire[i] ^= 0xff
		}
	}
	n := f.tokenLen
	if f.shortToken {
		n--
	}
	return onWire, bytes.Repeat([]byte{0xAB}, n), nil
}

func (f *fakeSecCtx) UnprotectResponse(signedRegion, stub, authValue []byte, seal bool) ([]byte, error) {
	f.unprotectCalled = true
	f.gotSeal = seal
	f.gotAuthValue = append([]byte(nil), authValue...)
	out := append([]byte(nil), stub...)
	if seal {
		for i := range out {
			out[i] ^= 0xff
		}
	}
	return out, nil
}

func TestMarshalProtectedRequestFraming(t *testing.T) {
	for _, tc := range []struct {
		name  string
		level uint8
		seal  bool
	}{
		{"privacy seals stub", pdu.AuthLevelPktPrivacy, true},
		{"integrity leaves stub", pdu.AuthLevelPktIntegrity, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeSecCtx{tokenLen: 16}
			c := &Client{authType: pdu.AuthTypeNTLMSSP, authLevel: tc.level, sec: fake}

			stub := []byte("hello-rpc-stub!") // 15 bytes -> 1 pad byte -> 16
			req := &pdu.Request{ContextID: 0, Opnum: 5, Stub: stub}
			req.Header = pdu.NewHeader(pdu.PacketTypeRequest, pdu.PFCFirstFrag|pdu.PFCLastFrag, 1)

			raw, err := c.marshalProtectedRequest(req)
			if err != nil {
				t.Fatalf("marshalProtectedRequest: %v", err)
			}

			hdr, err := pdu.PeekHeader(raw)
			if err != nil {
				t.Fatalf("PeekHeader: %v", err)
			}
			if hdr.AuthLength != 16 {
				t.Errorf("auth_length = %d, want 16", hdr.AuthLength)
			}
			if int(hdr.FragLength) != len(raw) {
				t.Errorf("frag_length = %d, want %d", hdr.FragLength, len(raw))
			}

			// Layout: header(16) | bodyHdr(8) | onWireStub(16) | sec_trailer(8) | token(16).
			const stubStart = pdu.HeaderSize + 8
			stubPad := make([]byte, 16)
			copy(stubPad, stub)
			onWire := raw[stubStart : stubStart+16]
			want := append([]byte(nil), stubPad...)
			if tc.seal {
				for i := range want {
					want[i] ^= 0xff
				}
			}
			if !bytes.Equal(onWire, want) {
				t.Errorf("on-wire stub = %x, want %x", onWire, want)
			}
			token := raw[len(raw)-16:]
			if !bytes.Equal(token, bytes.Repeat([]byte{0xAB}, 16)) {
				t.Errorf("token = %x, want 16x0xAB", token)
			}
			// sec_trailer auth_pad_length is the third byte of the trailer.
			if got := raw[stubStart+16+2]; got != 1 {
				t.Errorf("auth_pad_length = %d, want 1", got)
			}
			if fake.gotSeal != tc.seal {
				t.Errorf("provider seal flag = %v, want %v", fake.gotSeal, tc.seal)
			}
			if len(fake.gotSignedRegion) != len(raw)-16 {
				t.Errorf("signed region len = %d, want %d", len(fake.gotSignedRegion), len(raw)-16)
			}
			if !bytes.Equal(fake.gotStub, stubPad) {
				t.Errorf("provider stub = %x, want %x", fake.gotStub, stubPad)
			}
		})
	}
}

func TestMarshalProtectedRequestTokenLengthMismatch(t *testing.T) {
	fake := &fakeSecCtx{tokenLen: 16, shortToken: true}
	c := &Client{authType: pdu.AuthTypeNTLMSSP, authLevel: pdu.AuthLevelPktIntegrity, sec: fake}
	req := &pdu.Request{Opnum: 1, Stub: []byte("x")}
	req.Header = pdu.NewHeader(pdu.PacketTypeRequest, pdu.PFCFirstFrag|pdu.PFCLastFrag, 1)
	if _, err := c.marshalProtectedRequest(req); err == nil {
		t.Fatal("expected error when provider returns a token of the wrong length")
	}
}

func TestUnprotectResponseStubPlumbing(t *testing.T) {
	fake := &fakeSecCtx{tokenLen: 16}
	c := &Client{authType: pdu.AuthTypeNTLMSSP, authLevel: pdu.AuthLevelPktIntegrity, sec: fake}

	stub := []byte("resp-stub-data") // 14 bytes -> 2 pad bytes
	frag := buildResponseFrag(t, stub, pdu.AuthLevelPktIntegrity, bytes.Repeat([]byte{0xCD}, 16))

	got, err := c.unprotectResponseStub(frag)
	if err != nil {
		t.Fatalf("unprotectResponseStub: %v", err)
	}
	if !bytes.Equal(got, stub) {
		t.Errorf("recovered stub = %q, want %q (pad must be stripped)", got, stub)
	}
	if !fake.unprotectCalled {
		t.Error("provider UnprotectResponse was not called")
	}
	if !bytes.Equal(fake.gotAuthValue, bytes.Repeat([]byte{0xCD}, 16)) {
		t.Errorf("provider auth_value = %x, want 16x0xCD", fake.gotAuthValue)
	}
}

func TestSetAuthProviderNetlogonRequiresBindToken(t *testing.T) {
	// Netlogon without a bind token is rejected.
	if err := (&Client{}).SetAuthProvider(pdu.AuthTypeNetlogon, pdu.AuthLevelPktPrivacy, &fakeSecCtx{tokenLen: 56}, nil); err == nil {
		t.Fatal("SetAuthProvider accepted netlogon with a nil bind token")
	}
	// Netlogon with a bind token is accepted.
	if err := (&Client{}).SetAuthProvider(pdu.AuthTypeNetlogon, pdu.AuthLevelPktPrivacy, &fakeSecCtx{tokenLen: 56}, []byte{1, 2, 3, 4}); err != nil {
		t.Fatalf("SetAuthProvider(netlogon, bindToken) = %v, want nil", err)
	}
	// A non-netlogon single-leg provider may still omit the bind token.
	if err := (&Client{}).SetAuthProvider(pdu.AuthTypeGSSKerberos, pdu.AuthLevelPktPrivacy, &fakeSecCtx{tokenLen: 16}, nil); err != nil {
		t.Fatalf("SetAuthProvider(non-netlogon, nil bindToken) = %v, want nil", err)
	}
}

func TestAuthVerifierOverheadUsesProviderTokenLen(t *testing.T) {
	c := &Client{authLevel: pdu.AuthLevelPktPrivacy, sec: &fakeSecCtx{tokenLen: 56}}
	// 3 worst-case pad + 8-byte sec_trailer + 56-byte token.
	if got := c.authVerifierOverhead(); got != 3+pdu.SecTrailerSize+56 {
		t.Errorf("authVerifierOverhead = %d, want %d", got, 3+pdu.SecTrailerSize+56)
	}
}

// buildResponseFrag assembles a minimal authenticated response fragment: header, 8-byte
// response body header, padded stub, sec_trailer, and token.
func buildResponseFrag(t *testing.T, stub []byte, authLevel uint8, token []byte) []byte {
	t.Helper()
	pad := (4 - len(stub)%4) % 4
	stubPad := make([]byte, len(stub)+pad)
	copy(stubPad, stub)
	st := pdu.SecTrailer{AuthType: pdu.AuthTypeNTLMSSP, AuthLevel: authLevel, AuthPadLength: uint8(pad)}

	body := make([]byte, 8) // response body header (alloc_hint, ctx_id, cancel_count, reserved)
	h := pdu.NewHeader(pdu.PacketTypeResponse, pdu.PFCFirstFrag|pdu.PFCLastFrag, 1)
	fragLen := pdu.HeaderSize + len(body) + len(stubPad) + pdu.SecTrailerSize + len(token)
	h.AuthLength = uint16(len(token))
	h.FragLength = uint16(fragLen)
	hb, err := h.Marshal()
	if err != nil {
		t.Fatalf("header marshal: %v", err)
	}

	frag := make([]byte, 0, fragLen)
	frag = append(frag, hb...)
	frag = append(frag, body...)
	frag = append(frag, stubPad...)
	frag = append(frag, st.Marshal()...)
	frag = append(frag, token...)
	return frag
}
