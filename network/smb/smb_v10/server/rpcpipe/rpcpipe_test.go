package rpcpipe

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	srvsvcfunctions "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/functions"
	wkssvcfunctions "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// maxPipeAnswer is the ceiling the SMB server passes to Transact. It is not
// exported by the server, so it is repeated here; a test that used a different
// value would not be exercising the path a real client takes.
const maxPipeAnswer = 64 * 1024

// pipeInvoker drives a Handler the way a client does: it marshals a call's
// parameters, frames a request PDU, hands it to Transact, reassembles the reply
// fragments and decodes the response stub.
//
// It satisfies ndr.Invoker, so the interfaces' own client stubs run against it
// unchanged. That is the point of testing through it: the client declares the
// parameter sets privately, and a server that disagreed with them about the wire
// would fail here rather than only against a real client.
type pipeInvoker struct {
	t       *testing.T
	handler *Handler
	pipe    string
	callID  uint32
}

// bind performs the association's bind and returns the bind_ack.
func (p *pipeInvoker) bind(abstract syntax.SyntaxID) *pdu.BindAck {
	p.t.Helper()

	p.callID++
	bind := &pdu.Bind{
		Header:      pdu.NewHeader(pdu.PacketTypeBind, pdu.PFCFirstFrag|pdu.PFCLastFrag, p.callID),
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		ContextList: []pdu.ContextElement{{
			ContextID:        0,
			AbstractSyntax:   abstract,
			TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
		}},
	}
	request, err := bind.Marshal()
	if err != nil {
		p.t.Fatalf("failed to marshal the bind: %v", err)
	}

	reply, more, err := p.handler.Transact(p.pipe, request, maxPipeAnswer)
	if err != nil {
		p.t.Fatalf("Transact refused the bind on %q: %v", p.pipe, err)
	}
	if more {
		p.t.Fatalf("a bind_ack of %d bytes was reported as incomplete", len(reply))
	}

	header, err := pdu.PeekHeader(reply)
	if err != nil {
		p.t.Fatalf("the bind reply is not a PDU: %v", err)
	}
	if header.PacketType != pdu.PacketTypeBindAck {
		p.t.Fatalf("the bind on %q was answered with %s, want bind_ack", p.pipe, header.PacketType)
	}

	ack := &pdu.BindAck{}
	if _, err := ack.Unmarshal(reply); err != nil {
		p.t.Fatalf("failed to decode the bind_ack: %v", err)
	}
	return ack
}

// Invoke runs one call over the pipe.
func (p *pipeInvoker) Invoke(in ndr.Call, out any) error {
	stub, err := ndr.Request(in)
	if err != nil {
		return fmt.Errorf("marshal the request stub: %w", err)
	}

	p.callID++
	request := &pdu.Request{
		Header: pdu.NewHeader(pdu.PacketTypeRequest, pdu.PFCFirstFrag|pdu.PFCLastFrag, p.callID),
		Opnum:  in.Opnum(),
		Stub:   stub,
	}
	framed, err := request.Marshal()
	if err != nil {
		return fmt.Errorf("marshal the request PDU: %w", err)
	}

	reply, more, err := p.handler.Transact(p.pipe, framed, maxPipeAnswer)
	if err != nil {
		return fmt.Errorf("transact on %q: %w", p.pipe, err)
	}
	if more {
		return fmt.Errorf("the answer was cut at %d bytes, which this test's ceiling should not do", len(reply))
	}

	answer, err := reassemble(reply)
	if err != nil {
		return err
	}
	if err := ndr.Unmarshal(answer, out); err != nil {
		return fmt.Errorf("unmarshal the response stub: %w", err)
	}
	return nil
}

// reassemble joins the response fragments of a reply, and reports a fault as an
// error naming its status.
func reassemble(stream []byte) ([]byte, error) {
	stub := []byte{}

	for offset := 0; offset < len(stream); {
		header, err := pdu.PeekHeader(stream[offset:])
		if err != nil {
			return nil, fmt.Errorf("the PDU at offset %d has no readable header: %w", offset, err)
		}
		length := int(header.FragLength)
		if length < pdu.HeaderSize || offset+length > len(stream) {
			return nil, fmt.Errorf("the PDU at offset %d claims %d bytes, past the %d remaining",
				offset, length, len(stream)-offset)
		}
		fragment := stream[offset : offset+length]
		offset += length

		switch header.PacketType {
		case pdu.PacketTypeResponse:
			response := &pdu.Response{}
			if _, err := response.Unmarshal(fragment); err != nil {
				return nil, fmt.Errorf("failed to decode a response fragment: %w", err)
			}
			stub = append(stub, response.Stub...)

		case pdu.PacketTypeFault:
			fault := &pdu.Fault{}
			if _, err := fault.Unmarshal(fragment); err != nil {
				return nil, fmt.Errorf("failed to decode a fault: %w", err)
			}
			return nil, fmt.Errorf("the call faulted with %s", pdu.FaultStatus(fault.Status))

		default:
			return nil, fmt.Errorf("the reply carries a %s PDU", header.PacketType)
		}
	}

	return stub, nil
}

// testHandler builds a Handler over a fixed share list.
func testHandler(shares ...ShareEntry) *Handler {
	return New(Options{
		ServerName: "MANTICORE",
		DomainName: "WORKGROUP",
		Shares:     func() []ShareEntry { return shares },
	})
}

// sampleShares is a share list with the shapes that matter: an ordinary disk
// share with a remark, an administrative share, a printer, and one with no
// remark at all.
func sampleShares() []ShareEntry {
	return []ShareEntry{
		{Name: "PUBLIC", Comment: "Public files", Type: STYPE_DISKTREE},
		{Name: "IPC$", Comment: "Remote IPC", Type: STYPE_IPC | STYPE_SPECIAL},
		{Name: "PRINTER", Comment: "The printer", Type: STYPE_PRINTQ},
		{Name: "SCRATCH", Type: STYPE_DISKTREE},
	}
}

func TestHandlerSatisfiesThePipeContract(t *testing.T) {
	// The compile-time assertion in the package covers the method set; this
	// checks the behaviour the server relies on: a pipe it serves opens, one it
	// does not is refused rather than silently accepted.
	handler := testHandler()

	for _, name := range []string{"srvsvc", "wkssvc", `\PIPE\srvsvc`, "PIPE/wkssvc", "SrvSvc"} {
		if err := handler.OpenPipe(name); err != nil {
			t.Errorf("OpenPipe(%q) failed: %v", name, err)
			continue
		}
		if err := handler.ClosePipe(name); err != nil {
			t.Errorf("ClosePipe(%q) failed: %v", name, err)
		}
	}

	for _, name := range []string{"lsarpc", "samr", "", "srvsvc2"} {
		if err := handler.OpenPipe(name); err == nil {
			t.Errorf("OpenPipe(%q) succeeded for a pipe this handler does not serve", name)
		}
	}
}

func TestPipesServed(t *testing.T) {
	got := testHandler().Pipes()
	want := []string{"srvsvc", "wkssvc"}

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Pipes reports %v, want %v", got, want)
	}
}

func TestTransactOnAnUnservedPipeFails(t *testing.T) {
	handler := testHandler()

	if _, _, err := handler.Transact("lsarpc", []byte{0}, maxPipeAnswer); err == nil {
		t.Error("Transact answered on a pipe the handler does not serve")
	}
}

func TestBindOnEachPipeNamesItsOwnInterface(t *testing.T) {
	handler := testHandler(sampleShares()...)

	// The srvsvc pipe accepts srvsvc and refuses wkssvc, and the other way
	// round. That is the whole of what one-interface-per-pipe means, and it is
	// what a client relies on to discover it opened the wrong pipe.
	cases := []struct {
		pipe     string
		abstract syntax.SyntaxID
		accepted bool
	}{
		{"srvsvc", srvsvcSyntax(), true},
		{"srvsvc", wkssvcSyntax(), false},
		{"wkssvc", wkssvcSyntax(), true},
		{"wkssvc", srvsvcSyntax(), false},
	}

	for _, test := range cases {
		name := fmt.Sprintf("%s pipe binding %s", test.pipe, test.abstract.UUID.ToFormatD())
		t.Run(name, func(t *testing.T) {
			bind := &pdu.Bind{
				Header:      pdu.NewHeader(pdu.PacketTypeBind, pdu.PFCFirstFrag|pdu.PFCLastFrag, 1),
				MaxXmitFrag: 4280,
				MaxRecvFrag: 4280,
				ContextList: []pdu.ContextElement{{
					AbstractSyntax:   test.abstract,
					TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
				}},
			}
			request, err := bind.Marshal()
			if err != nil {
				t.Fatalf("failed to marshal the bind: %v", err)
			}

			reply, _, err := handler.Transact(test.pipe, request, maxPipeAnswer)
			if err != nil {
				t.Fatalf("Transact failed: %v", err)
			}
			header, err := pdu.PeekHeader(reply)
			if err != nil {
				t.Fatalf("the reply is not a PDU: %v", err)
			}

			want := pdu.PacketTypeBindNak
			if test.accepted {
				want = pdu.PacketTypeBindAck
			}
			if header.PacketType != want {
				t.Errorf("the bind was answered with %s, want %s", header.PacketType, want)
			}
		})
	}
}

func srvsvcSyntax() syntax.SyntaxID { return srvsvc.SyntaxID() }

func wkssvcSyntax() syntax.SyntaxID {
	return (&wkssvcService{}).AbstractSyntax()
}

func TestNetrShareEnumLevel1ReportsEveryShare(t *testing.T) {
	shares := sampleShares()
	handler := testHandler(shares...)

	rpc := &pipeInvoker{t: t, handler: handler, pipe: "srvsvc"}
	rpc.bind(srvsvcSyntax())

	info, total, resume, err := srvsvcfunctions.NetrShareEnum(rpc, "",
		mssrvs.SHARE_ENUM_STRUCT{Level: 1, ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: 1}},
		maxPreferredLength, nil)
	if err != nil {
		t.Fatalf("NetrShareEnum at level 1 failed: %v", err)
	}

	if int(total) != len(shares) {
		t.Errorf("TotalEntries is %d, want the %d shares registered", total, len(shares))
	}
	if resume != nil {
		t.Errorf("a resume handle of %d came back for a call that sent none", *resume)
	}
	if info.Level != 1 || info.ShareInfo.Tag != 1 {
		t.Errorf("the answer reports level %d and union tag %d, want 1 and 1", info.Level, info.ShareInfo.Tag)
	}

	container := info.ShareInfo.Level1
	if container == nil {
		t.Fatal("the level 1 container is a null pointer, so no shares came back")
	}
	if int(container.EntriesRead) != len(shares) || len(container.Buffer) != len(shares) {
		t.Fatalf("the container reports %d entries and carries %d, want %d of each",
			container.EntriesRead, len(container.Buffer), len(shares))
	}

	for i, want := range shares {
		got := container.Buffer[i]
		if string(got.Shi1Netname) != want.Name {
			t.Errorf("entry %d is named %q, want %q", i, got.Shi1Netname, want.Name)
		}
		if uint32(got.Shi1Type) != want.Type {
			t.Errorf("entry %d (%s) has type 0x%08X, want 0x%08X", i, want.Name, got.Shi1Type, want.Type)
		}
		if string(got.Shi1Remark) != want.Comment {
			t.Errorf("entry %d (%s) has remark %q, want %q", i, want.Name, got.Shi1Remark, want.Comment)
		}
	}
}

func TestNetrShareEnumLevel0ReportsNamesOnly(t *testing.T) {
	shares := sampleShares()
	rpc := &pipeInvoker{t: t, handler: testHandler(shares...), pipe: "srvsvc"}
	rpc.bind(srvsvcSyntax())

	info, total, _, err := srvsvcfunctions.NetrShareEnum(rpc, "",
		mssrvs.SHARE_ENUM_STRUCT{Level: 0, ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: 0}},
		maxPreferredLength, nil)
	if err != nil {
		t.Fatalf("NetrShareEnum at level 0 failed: %v", err)
	}
	if int(total) != len(shares) {
		t.Errorf("TotalEntries is %d, want %d", total, len(shares))
	}

	container := info.ShareInfo.Level0
	if container == nil {
		t.Fatal("the level 0 container is a null pointer")
	}
	if info.ShareInfo.Level1 != nil {
		t.Error("the level 1 container is also set, so the union carries two arms")
	}
	if len(container.Buffer) != len(shares) {
		t.Fatalf("the container carries %d entries, want %d", len(container.Buffer), len(shares))
	}
	for i, want := range shares {
		if string(container.Buffer[i].Shi0Netname) != want.Name {
			t.Errorf("entry %d is named %q, want %q", i, container.Buffer[i].Shi0Netname, want.Name)
		}
	}
}

func TestNetrShareEnumOnAnEmptyServer(t *testing.T) {
	// A server with no shares answers an empty list, not a failure: that is how
	// a client learns there is nothing to browse.
	rpc := &pipeInvoker{t: t, handler: testHandler(), pipe: "srvsvc"}
	rpc.bind(srvsvcSyntax())

	info, total, _, err := srvsvcfunctions.NetrShareEnum(rpc, "",
		mssrvs.SHARE_ENUM_STRUCT{Level: 1, ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: 1}},
		maxPreferredLength, nil)
	if err != nil {
		t.Fatalf("NetrShareEnum on a server with no shares failed: %v", err)
	}
	if total != 0 {
		t.Errorf("TotalEntries is %d, want 0", total)
	}
	if info.ShareInfo.Level1 == nil {
		t.Fatal("the level 1 container is a null pointer, want an empty one")
	}
	if len(info.ShareInfo.Level1.Buffer) != 0 {
		t.Errorf("the container carries %d entries, want none", len(info.ShareInfo.Level1.Buffer))
	}
}

func TestNetrShareEnumWithNoShareSourceReportsNone(t *testing.T) {
	// Options.Shares left nil. Reporting nothing is the honest answer for an
	// endpoint with no share source, and a nil call would panic on the
	// connection's goroutine.
	rpc := &pipeInvoker{t: t, handler: New(Options{}), pipe: "srvsvc"}
	rpc.bind(srvsvcSyntax())

	_, total, _, err := srvsvcfunctions.NetrShareEnum(rpc, "",
		mssrvs.SHARE_ENUM_STRUCT{Level: 1, ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: 1}},
		maxPreferredLength, nil)
	if err != nil {
		t.Fatalf("NetrShareEnum with no share source failed: %v", err)
	}
	if total != 0 {
		t.Errorf("TotalEntries is %d, want 0", total)
	}
}

func TestNetrShareEnumRefusesAnUnservedLevel(t *testing.T) {
	rpc := &pipeInvoker{t: t, handler: testHandler(sampleShares()...), pipe: "srvsvc"}
	rpc.bind(srvsvcSyntax())

	// Level 2 has a union arm this does not fill, and level 99 has none at all.
	// Both are a successful call reporting ERROR_INVALID_LEVEL, which is what
	// makes a client fall back rather than give up.
	for _, level := range []ndr.DWORD{2, 502, 99} {
		t.Run(fmt.Sprintf("level %d", level), func(t *testing.T) {
			_, _, _, err := srvsvcfunctions.NetrShareEnum(rpc, "",
				mssrvs.SHARE_ENUM_STRUCT{Level: level, ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: level}},
				maxPreferredLength, nil)
			if err == nil {
				t.Fatalf("level %d was answered as if it were served", level)
			}
			if !strings.Contains(err.Error(), "ERROR_INVALID_LEVEL") {
				t.Errorf("level %d failed with %v, want ERROR_INVALID_LEVEL", level, err)
			}
		})
	}
}

func TestNetrShareEnumUnknownOpnumFaults(t *testing.T) {
	handler := testHandler(sampleShares()...)

	// NetrShareAdd, opnum 14, which this does not implement.
	request := &pdu.Request{
		Header: pdu.NewHeader(pdu.PacketTypeRequest, pdu.PFCFirstFrag|pdu.PFCLastFrag, 1),
		Opnum:  srvsvc.OpnumNetrShareAdd,
		Stub:   make([]byte, 16),
	}
	framed, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the request: %v", err)
	}

	reply, _, err := handler.Transact("srvsvc", framed, maxPipeAnswer)
	if err != nil {
		t.Fatalf("Transact failed: %v", err)
	}
	if _, err := reassemble(reply); err == nil {
		t.Fatal("an opnum the interface does not implement was answered with a response")
	} else if !strings.Contains(err.Error(), "nca_s_op_rng_error") {
		t.Errorf("the call failed with %v, want a fault of nca_s_op_rng_error", err)
	}
}

func TestNetrShareEnumUndecodableStubFaults(t *testing.T) {
	handler := testHandler(sampleShares()...)

	// Three bytes where a SHARE_ENUM_STRUCT and three more parameters belong.
	request := &pdu.Request{
		Header: pdu.NewHeader(pdu.PacketTypeRequest, pdu.PFCFirstFrag|pdu.PFCLastFrag, 1),
		Opnum:  srvsvc.OpnumNetrShareEnum,
		Stub:   []byte{1, 2, 3},
	}
	framed, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the request: %v", err)
	}

	reply, _, err := handler.Transact("srvsvc", framed, maxPipeAnswer)
	if err != nil {
		t.Fatalf("Transact failed: %v", err)
	}
	if _, err := reassemble(reply); err == nil {
		t.Fatal("a stub that cannot be decoded was answered with a response")
	} else if !strings.Contains(err.Error(), "nca_s_fault_ndr") {
		t.Errorf("the call failed with %v, want a fault of nca_s_fault_ndr", err)
	}
}

func TestNetrShareEnumResumesWhereItStopped(t *testing.T) {
	// A budget too small for the whole list makes the enumeration report part of
	// it with ERROR_MORE_DATA and a resume handle, which the client sends back to
	// continue. Walking it to the end has to yield every share exactly once.
	shares := sampleShares()
	rpc := &pipeInvoker{t: t, handler: testHandler(shares...), pipe: "srvsvc"}
	rpc.bind(srvsvcSyntax())

	handle := ndr.DWORD(0)
	seen := []string{}

	for round := 0; round < len(shares)+2; round++ {
		info, total, resume, err := srvsvcfunctions.NetrShareEnum(rpc, "",
			mssrvs.SHARE_ENUM_STRUCT{Level: 1, ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: 1}},
			48, &handle)
		if err != nil {
			t.Fatalf("round %d failed: %v", round, err)
		}
		if resume == nil {
			t.Fatalf("round %d came back with no resume handle for a call that sent one", round)
		}

		container := info.ShareInfo.Level1
		if container == nil {
			t.Fatalf("round %d came back with a null container", round)
		}
		if len(container.Buffer) == 0 {
			t.Fatalf("round %d reported no entries, so the enumeration cannot advance", round)
		}
		for _, entry := range container.Buffer {
			seen = append(seen, string(entry.Shi1Netname))
		}

		if *resume == 0 {
			// The enumeration finished: the last answer carried the remainder.
			if int(total) != len(container.Buffer) {
				t.Errorf("the final round reports %d entries remaining but carried %d",
					total, len(container.Buffer))
			}
			break
		}
		handle = *resume
	}

	if want := []string{"PUBLIC", "IPC$", "PRINTER", "SCRATCH"}; strings.Join(seen, ",") != strings.Join(want, ",") {
		t.Errorf("walking the enumeration saw %v, want %v exactly once each", seen, want)
	}
}

func TestNetrShareEnumResumeHandlePastTheEndFinishes(t *testing.T) {
	// A share removed between two calls can leave a handle past the end. The
	// enumeration is over, and saying so is what lets the client stop.
	rpc := &pipeInvoker{t: t, handler: testHandler(sampleShares()...), pipe: "srvsvc"}
	rpc.bind(srvsvcSyntax())

	handle := ndr.DWORD(9999)
	info, total, resume, err := srvsvcfunctions.NetrShareEnum(rpc, "",
		mssrvs.SHARE_ENUM_STRUCT{Level: 1, ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: 1}},
		maxPreferredLength, &handle)
	if err != nil {
		t.Fatalf("a resume handle past the end failed: %v", err)
	}
	if total != 0 {
		t.Errorf("TotalEntries is %d, want 0 from a handle past the end", total)
	}
	if resume == nil || *resume != 0 {
		t.Errorf("the resume handle came back as %v, want 0 to end the enumeration", resume)
	}
	if info.ShareInfo.Level1 == nil || len(info.ShareInfo.Level1.Buffer) != 0 {
		t.Error("the answer carries entries for a handle past the end")
	}
}

func TestNetrShareEnumFragmentsALongList(t *testing.T) {
	// A share list well past one fragment, so the reply is several response PDUs
	// that the client reassembles. Reporting the whole list in one exchange is
	// what MAX_PREFERRED_LENGTH asks for, so the fragmentation is the server's
	// problem and not the client's.
	shares := make([]ShareEntry, 0, 200)
	for i := 0; i < 200; i++ {
		shares = append(shares, ShareEntry{
			Name:    fmt.Sprintf("SHARE%03d", i),
			Comment: fmt.Sprintf("share number %d, with a remark long enough to matter", i),
			Type:    STYPE_DISKTREE,
		})
	}

	handler := testHandler(shares...)
	rpc := &pipeInvoker{t: t, handler: handler, pipe: "srvsvc"}
	rpc.bind(srvsvcSyntax())

	info, total, _, err := srvsvcfunctions.NetrShareEnum(rpc, "",
		mssrvs.SHARE_ENUM_STRUCT{Level: 1, ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: 1}},
		maxPreferredLength, nil)
	if err != nil {
		t.Fatalf("NetrShareEnum over a long list failed: %v", err)
	}
	if int(total) != len(shares) {
		t.Errorf("TotalEntries is %d, want %d", total, len(shares))
	}

	container := info.ShareInfo.Level1
	if container == nil {
		t.Fatal("the container is a null pointer")
	}
	if len(container.Buffer) != len(shares) {
		t.Fatalf("the container carries %d entries, want %d", len(container.Buffer), len(shares))
	}
	for i, want := range shares {
		if string(container.Buffer[i].Shi1Netname) != want.Name {
			t.Fatalf("entry %d is named %q, want %q — the fragments did not reassemble in order",
				i, container.Buffer[i].Shi1Netname, want.Name)
		}
	}
}

func TestNetrWkstaGetInfoLevel100(t *testing.T) {
	rpc := &pipeInvoker{t: t, handler: testHandler(), pipe: "wkssvc"}
	rpc.bind(wkssvcSyntax())

	info, err := wkssvcfunctions.NetrWkstaGetInfo(rpc, nil, 100)
	if err != nil {
		t.Fatalf("NetrWkstaGetInfo at level 100 failed: %v", err)
	}
	if info.Tag != 100 {
		t.Errorf("the union tag is %d, want 100", info.Tag)
	}

	level := info.WkstaInfo100
	if level == nil {
		t.Fatal("the level 100 arm is a null pointer")
	}
	if level.Wki100_platform_id != platformIDNT {
		t.Errorf("the platform id is %d, want PLATFORM_ID_NT (%d)", level.Wki100_platform_id, platformIDNT)
	}
	if level.Wki100_computername == nil || string(*level.Wki100_computername) != "MANTICORE" {
		t.Errorf("the computer name is %v, want MANTICORE", level.Wki100_computername)
	}
	if level.Wki100_langroup == nil || string(*level.Wki100_langroup) != "WORKGROUP" {
		t.Errorf("the workgroup is %v, want WORKGROUP", level.Wki100_langroup)
	}
	if level.Wki100_ver_major != defaultVersionMajor || level.Wki100_ver_minor != defaultVersionMinor {
		t.Errorf("the version is %d.%d, want the default %d.%d",
			level.Wki100_ver_major, level.Wki100_ver_minor, defaultVersionMajor, defaultVersionMinor)
	}
}

func TestNetrWkstaGetInfoLevel101(t *testing.T) {
	handler := New(Options{ServerName: "HOST", DomainName: "LAB", VersionMajor: 10, VersionMinor: 0})
	rpc := &pipeInvoker{t: t, handler: handler, pipe: "wkssvc"}
	rpc.bind(wkssvcSyntax())

	info, err := wkssvcfunctions.NetrWkstaGetInfo(rpc, nil, 101)
	if err != nil {
		t.Fatalf("NetrWkstaGetInfo at level 101 failed: %v", err)
	}

	level := info.WkstaInfo101
	if level == nil {
		t.Fatal("the level 101 arm is a null pointer")
	}
	if info.WkstaInfo100 != nil {
		t.Error("the level 100 arm is also set, so the union carries two arms")
	}
	if level.Wki101_computername == nil || string(*level.Wki101_computername) != "HOST" {
		t.Errorf("the computer name is %v, want HOST", level.Wki101_computername)
	}
	if level.Wki101_ver_major != 10 || level.Wki101_ver_minor != 0 {
		t.Errorf("the version is %d.%d, want the configured 10.0", level.Wki101_ver_major, level.Wki101_ver_minor)
	}
	// The LAN root is a Windows redirector path, which this host has none of.
	// A null pointer says so; an empty string would claim it has one that is
	// nameless.
	if level.Wki101_lanroot != nil {
		t.Errorf("the LAN root is %q, want a null pointer", *level.Wki101_lanroot)
	}
}

func TestNetrWkstaGetInfoWithNoNamesConfigured(t *testing.T) {
	rpc := &pipeInvoker{t: t, handler: New(Options{}), pipe: "wkssvc"}
	rpc.bind(wkssvcSyntax())

	info, err := wkssvcfunctions.NetrWkstaGetInfo(rpc, nil, 100)
	if err != nil {
		t.Fatalf("NetrWkstaGetInfo failed: %v", err)
	}
	if info.WkstaInfo100 == nil {
		t.Fatal("the level 100 arm is a null pointer")
	}
	if info.WkstaInfo100.Wki100_computername != nil {
		t.Errorf("the computer name is %q, want a null pointer when none is configured",
			*info.WkstaInfo100.Wki100_computername)
	}
}

func TestNetrWkstaGetInfoRefusesAnUnservedLevel(t *testing.T) {
	rpc := &pipeInvoker{t: t, handler: testHandler(), pipe: "wkssvc"}
	rpc.bind(wkssvcSyntax())

	for _, level := range []ndr.DWORD{102, 502, 1013, 0} {
		t.Run(fmt.Sprintf("level %d", level), func(t *testing.T) {
			_, err := wkssvcfunctions.NetrWkstaGetInfo(rpc, nil, level)
			if err == nil {
				t.Fatalf("level %d was answered as if it were served", level)
			}
			if !strings.Contains(err.Error(), "ERROR_INVALID_LEVEL") {
				t.Errorf("level %d failed with %v, want ERROR_INVALID_LEVEL", level, err)
			}
		})
	}
}

func TestForServerReportsTheServersOwnShares(t *testing.T) {
	srv, err := server.NewServer(server.Config{})
	if err != nil {
		t.Fatalf("failed to build a server: %v", err)
	}

	pipes := ForServer(srv, Options{ServerName: "MANTICORE"})
	if err := srv.AddShare(&server.Share{
		Name:  "IPC$",
		Type:  server.ShareTypeNamedPipe,
		Pipes: pipes,
	}); err != nil {
		t.Fatalf("failed to add IPC$: %v", err)
	}
	if err := srv.AddShare(&server.Share{
		Name:    "PUBLIC",
		Type:    server.ShareTypeDisk,
		Comment: "Public files",
		FS:      server.NewMemoryFileSystem(""),
	}); err != nil {
		t.Fatalf("failed to add PUBLIC: %v", err)
	}

	rpc := &pipeInvoker{t: t, handler: pipes, pipe: "srvsvc"}
	rpc.bind(srvsvcSyntax())

	info, _, _, err := srvsvcfunctions.NetrShareEnum(rpc, "",
		mssrvs.SHARE_ENUM_STRUCT{Level: 1, ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: 1}},
		maxPreferredLength, nil)
	if err != nil {
		t.Fatalf("NetrShareEnum failed: %v", err)
	}
	container := info.ShareInfo.Level1
	if container == nil {
		t.Fatal("the container is a null pointer")
	}

	// Shares come out of a map, so the order is not fixed; what matters is that
	// both are there with the right type.
	types := map[string]uint32{}
	remarks := map[string]string{}
	for _, entry := range container.Buffer {
		types[string(entry.Shi1Netname)] = uint32(entry.Shi1Type)
		remarks[string(entry.Shi1Netname)] = string(entry.Shi1Remark)
	}

	if got, want := types["IPC$"], STYPE_IPC|STYPE_SPECIAL; got != want {
		t.Errorf("IPC$ is reported as type 0x%08X, want 0x%08X", got, want)
	}
	if got, want := types["PUBLIC"], STYPE_DISKTREE; got != want {
		t.Errorf("PUBLIC is reported as type 0x%08X, want 0x%08X", got, want)
	}
	if got := remarks["PUBLIC"]; got != "Public files" {
		t.Errorf("PUBLIC's remark is %q, want %q", got, "Public files")
	}

	// A share added after the handler was built appears in the next enumeration,
	// because the share list is read on each call.
	if err := srv.AddShare(&server.Share{
		Name: "LATER",
		Type: server.ShareTypeDisk,
		FS:   server.NewMemoryFileSystem(""),
	}); err != nil {
		t.Fatalf("failed to add LATER: %v", err)
	}

	info, _, _, err = srvsvcfunctions.NetrShareEnum(rpc, "",
		mssrvs.SHARE_ENUM_STRUCT{Level: 1, ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: 1}},
		maxPreferredLength, nil)
	if err != nil {
		t.Fatalf("the second NetrShareEnum failed: %v", err)
	}
	found := false
	for _, entry := range info.ShareInfo.Level1.Buffer {
		if string(entry.Shi1Netname) == "LATER" {
			found = true
		}
	}
	if !found {
		t.Error("a share added after the handler was built does not appear in an enumeration")
	}
}

func TestForServerWithoutAServerReportsNoShares(t *testing.T) {
	rpc := &pipeInvoker{t: t, handler: ForServer(nil, Options{}), pipe: "srvsvc"}
	rpc.bind(srvsvcSyntax())

	_, total, _, err := srvsvcfunctions.NetrShareEnum(rpc, "",
		mssrvs.SHARE_ENUM_STRUCT{Level: 1, ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: 1}},
		maxPreferredLength, nil)
	if err != nil {
		t.Fatalf("NetrShareEnum failed: %v", err)
	}
	if total != 0 {
		t.Errorf("TotalEntries is %d, want 0", total)
	}
}

func TestShareTypeOf(t *testing.T) {
	cases := []struct {
		name      string
		shareType server.ShareType
		want      uint32
	}{
		{"PUBLIC", server.ShareTypeDisk, STYPE_DISKTREE},
		{"C$", server.ShareTypeDisk, STYPE_DISKTREE | STYPE_SPECIAL},
		{"IPC$", server.ShareTypeNamedPipe, STYPE_IPC | STYPE_SPECIAL},
		{"PRINT", server.ShareTypePrinter, STYPE_PRINTQ},
		{"PRINT$", server.ShareTypePrinter, STYPE_PRINTQ | STYPE_SPECIAL},
		{"ODD", server.ShareTypeAny, STYPE_DISKTREE},
	}

	for _, test := range cases {
		if got := ShareTypeOf(test.name, test.shareType); got != test.want {
			t.Errorf("ShareTypeOf(%q, %q) is 0x%08X, want 0x%08X", test.name, test.shareType, got, test.want)
		}
		if got := ShareTypeOf(test.name, test.shareType) & STYPE_MASK; got > STYPE_IPC {
			t.Errorf("ShareTypeOf(%q, %q) masks to kind %d, which is not a share kind",
				test.name, test.shareType, got)
		}
	}
}

func TestConcurrentClientsOfOnePipe(t *testing.T) {
	// The SMB server calls a pipe handler on the goroutine of whichever
	// connection is asking, and one handler serves every connection. Under -race
	// this is what proves the handler can be shared.
	handler := testHandler(sampleShares()...)

	var waiting sync.WaitGroup
	for client := 0; client < 8; client++ {
		waiting.Add(1)
		go func(client int) {
			defer waiting.Done()

			pipe := "srvsvc"
			if client%2 == 1 {
				pipe = "wkssvc"
			}
			if err := handler.OpenPipe(pipe); err != nil {
				t.Errorf("client %d: OpenPipe failed: %v", client, err)
				return
			}
			defer func() {
				if err := handler.ClosePipe(pipe); err != nil {
					t.Errorf("client %d: ClosePipe failed: %v", client, err)
				}
			}()

			rpc := &pipeInvoker{t: t, handler: handler, pipe: pipe, callID: uint32(client) * 1000}
			if pipe == "srvsvc" {
				rpc.bind(srvsvcSyntax())
			} else {
				rpc.bind(wkssvcSyntax())
			}

			for round := 0; round < 10; round++ {
				if pipe == "srvsvc" {
					if _, _, _, err := srvsvcfunctions.NetrShareEnum(rpc, "",
						mssrvs.SHARE_ENUM_STRUCT{Level: 1, ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: 1}},
						maxPreferredLength, nil); err != nil {
						t.Errorf("client %d round %d: NetrShareEnum failed: %v", client, round, err)
						return
					}
					continue
				}
				if _, err := wkssvcfunctions.NetrWkstaGetInfo(rpc, nil, 100); err != nil {
					t.Errorf("client %d round %d: NetrWkstaGetInfo failed: %v", client, round, err)
					return
				}
			}
		}(client)
	}
	waiting.Wait()
}
