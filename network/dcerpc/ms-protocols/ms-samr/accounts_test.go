package ms_samr

import (
	"errors"
	"testing"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// fakeTransport is an in-memory transport.Transport for driving the DCE/RPC client
// without a network.
type fakeTransport struct {
	sent      [][]byte
	recvQueue [][]byte
}

func (f *fakeTransport) Connect() error { return nil }
func (f *fakeTransport) Send(p []byte) error {
	f.sent = append(f.sent, append([]byte(nil), p...))
	return nil
}
func (f *fakeTransport) Recv() ([]byte, error) {
	if len(f.recvQueue) == 0 {
		return nil, errors.New("recv queue empty")
	}
	c := f.recvQueue[0]
	f.recvQueue = f.recvQueue[1:]
	return c, nil
}
func (f *fakeTransport) Close() error        { return nil }
func (f *fakeTransport) MaxXmitFrag() uint16 { return 4280 }
func (f *fakeTransport) MaxRecvFrag() uint16 { return 4280 }
func (f *fakeTransport) queue(b []byte)      { f.recvQueue = append(f.recvQueue, b) }

func bindAck(t *testing.T) []byte {
	t.Helper()
	ack := &pdu.BindAck{
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		Results:     []pdu.PresentationResult{{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDRTransferSyntax()}},
	}
	b, err := ack.Marshal()
	if err != nil {
		t.Fatalf("bind_ack marshal: %v", err)
	}
	return b
}

func responsePDU(t *testing.T, callID uint32, stubBytes []byte) []byte {
	t.Helper()
	resp := &pdu.Response{Stub: stubBytes}
	resp.Header = pdu.NewHeader(pdu.PacketTypeResponse, pdu.PFCFirstFrag|pdu.PFCLastFrag, callID)
	b, err := resp.Marshal()
	if err != nil {
		t.Fatalf("response marshal: %v", err)
	}
	return b
}

func stub(t *testing.T, v any) []byte {
	t.Helper()
	b, err := ndr.Marshal(v)
	if err != nil {
		t.Fatalf("marshal response stub: %v", err)
	}
	return b
}

// connectedClient returns an ms_samr.Client whose samr association is already bound over
// ft, bypassing the pipe binder so the in-memory transport can drive it.
func connectedClient(t *testing.T, ft *fakeTransport) *Client {
	t.Helper()
	ft.queue(bindAck(t))
	rpc := dcerpcclient.NewClient(ft)
	if err := rpc.Bind(samr.SyntaxID()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	return &Client{rpc: rpc, connected: true}
}

func opnumOf(t *testing.T, requestPDU []byte) uint16 {
	t.Helper()
	var req pdu.Request
	if _, err := req.Unmarshal(requestPDU); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	return req.Opnum
}

// Mirrors of the samr [out] response structures, for marshalling canned replies.
type connect5Resp struct {
	OutVersion      ndr.DWORD
	OutRevisionInfo mssamr.SAMPR_REVISION_INFO
	ServerHandle    mssamr.SAMPR_HANDLE
	Status          ndr.DWORD `ndr:"retval"`
}

type openHandleResp struct {
	Handle mssamr.SAMPR_HANDLE
	Status ndr.DWORD `ndr:"retval"`
}

type membersResp struct {
	Members mssamr.SAMPR_PSID_ARRAY_OUT
	Status  ndr.DWORD `ndr:"retval"`
}

func sentinelHandle() mssamr.SAMPR_HANDLE {
	var h mssamr.SAMPR_HANDLE
	for i := range h {
		h[i] = byte(i + 1)
	}
	return h
}

func sidInfo(t *testing.T, s string) mssamr.SAMPR_SID_INFORMATION {
	t.Helper()
	sid, err := msdtyp.ParseSID(s)
	if err != nil {
		t.Fatalf("ParseSID(%q): %v", s, err)
	}
	return mssamr.SAMPR_SID_INFORMATION{SidPointer: &sid}
}

// TestLocalAliasMembers drives the full enum-local-admins chain (SamrConnect5 ->
// SamrOpenDomain(Builtin) -> SamrOpenAlias(544) -> SamrGetMembersInAlias) and checks the
// member SIDs come back and the members call went out.
func TestLocalAliasMembers(t *testing.T) {
	ft := &fakeTransport{}
	c := connectedClient(t, ft)

	h := sentinelHandle()
	wantSIDs := []string{"S-1-5-21-1111-2222-3333-500", "S-1-5-21-1111-2222-3333-512"}
	members := mssamr.SAMPR_PSID_ARRAY_OUT{
		Count: 2,
		Sids:  []mssamr.SAMPR_SID_INFORMATION{sidInfo(t, wantSIDs[0]), sidInfo(t, wantSIDs[1])},
	}

	ft.queue(responsePDU(t, 2, stub(t, &connect5Resp{ServerHandle: h}))) // OpenServer (SamrConnect5)
	ft.queue(responsePDU(t, 3, stub(t, &openHandleResp{Handle: h})))     // OpenDomain
	ft.queue(responsePDU(t, 4, stub(t, &openHandleResp{Handle: h})))     // OpenAlias
	ft.queue(responsePDU(t, 5, stub(t, &membersResp{Members: members}))) // GetMembersInAlias
	ft.queue(responsePDU(t, 6, stub(t, &openHandleResp{})))              // CloseHandle(alias)
	ft.queue(responsePDU(t, 7, stub(t, &openHandleResp{})))              // CloseHandle(domain)
	ft.queue(responsePDU(t, 8, stub(t, &openHandleResp{})))              // CloseHandle(server)

	got, err := c.LocalAliasMembers(AliasRIDAdministrators)
	if err != nil {
		t.Fatalf("LocalAliasMembers() error = %v", err)
	}
	if len(got) != len(wantSIDs) {
		t.Fatalf("got %d members %v, want %d %v", len(got), got, len(wantSIDs), wantSIDs)
	}
	for i := range wantSIDs {
		if got[i] != wantSIDs[i] {
			t.Errorf("member[%d] = %q, want %q", i, got[i], wantSIDs[i])
		}
	}

	// The members read must be a SamrGetMembersInAlias call (sent[0] is the bind).
	if op := opnumOf(t, ft.sent[4]); op != samr.OpnumSamrGetMembersInAlias {
		t.Fatalf("members call opnum = %d, want SamrGetMembersInAlias (%d)", op, samr.OpnumSamrGetMembersInAlias)
	}
}

// TestBuiltinDomainSID checks the well-known Builtin SID is what SamrOpenDomain is handed.
func TestBuiltinDomainSID(t *testing.T) {
	sid, err := BuiltinDomainSID()
	if err != nil {
		t.Fatalf("BuiltinDomainSID() error = %v", err)
	}
	if sid.String() != "S-1-5-32" {
		t.Fatalf("BuiltinDomainSID() = %q, want S-1-5-32", sid.String())
	}
}
