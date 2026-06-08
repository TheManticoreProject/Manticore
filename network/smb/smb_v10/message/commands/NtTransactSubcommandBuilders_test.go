package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestNtTransactRequestSetupRoundTrip verifies the new Setup-words field on
// NtTransactRequest marshals after Function and round-trips (SetupCount derived).
func TestNtTransactRequestSetupRoundTrip(t *testing.T) {
	r := commands.NewNtTransactRequest()
	r.Function = types.USHORT(subcommands.NT_TRANSACT_IOCTL)
	r.Setup = []types.USHORT{0x0078, 0x0014, 0x4002, 0x0001}

	raw, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out commands.NtTransactRequest
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Function != r.Function {
		t.Errorf("Function: got 0x%04x want 0x%04x", out.Function, r.Function)
	}
	if out.SetupCount != 4 {
		t.Errorf("SetupCount: got %d want 4", out.SetupCount)
	}
	if len(out.Setup) != 4 || out.Setup[0] != 0x0078 || out.Setup[3] != 0x0001 {
		t.Errorf("Setup round trip: got %v", out.Setup)
	}
}

func TestNtTransactIoctlSetup(t *testing.T) {
	setup := commands.NewNtTransactIoctlSetup(subcommands.FSCTL_SRV_COPYCHUNK, 0x4003, true, 0)
	if len(setup) != 4 {
		t.Fatalf("setup words: got %d want 4", len(setup))
	}
	// Reassemble the 8 setup octets and check FunctionCode/FID/IsFsctl/IsFlags.
	b := make([]byte, 8)
	for i, w := range setup {
		binary.LittleEndian.PutUint16(b[i*2:], uint16(w))
	}
	if got := binary.LittleEndian.Uint32(b[0:4]); got != subcommands.FSCTL_SRV_COPYCHUNK {
		t.Errorf("FunctionCode: got 0x%08x want 0x%08x", got, subcommands.FSCTL_SRV_COPYCHUNK)
	}
	if got := binary.LittleEndian.Uint16(b[4:6]); got != 0x4003 {
		t.Errorf("FID: got 0x%04x want 0x4003", got)
	}
	if b[6] != 1 || b[7] != 0 {
		t.Errorf("IsFsctl/IsFlags: got %d/%d want 1/0", b[6], b[7])
	}
}

func TestNotifyChangeBuilderAndResponse(t *testing.T) {
	setup := subcommands.NtTransactNotifyChangeSetup{
		CompletionFilter: subcommands.FILE_NOTIFY_CHANGE_NAME,
		FID:              0x4002,
		WatchTree:        true,
	}
	req, err := commands.NewNtTransactNotifyChangeRequest(setup, 0x1000)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	// Full request round-trips (setup-only, no NT_Trans_Parameters/Data).
	raw, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out commands.NtTransactRequest
	if _, err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Function != types.USHORT(subcommands.NT_TRANSACT_NOTIFY_CHANGE) || out.SetupCount != 4 {
		t.Fatalf("decoded Function=0x%04x SetupCount=%d", out.Function, out.SetupCount)
	}
	// Setup words decode back to the notify-change setup.
	setupBytes := make([]byte, len(out.Setup)*2)
	for i, w := range out.Setup {
		binary.LittleEndian.PutUint16(setupBytes[i*2:], uint16(w))
	}
	var decoded subcommands.NtTransactNotifyChangeSetup
	if _, err := decoded.Unmarshal(setupBytes); err != nil {
		t.Fatalf("decode setup: %v", err)
	}
	if decoded != setup {
		t.Errorf("setup round trip: got %+v want %+v", decoded, setup)
	}

	// Response: a FILE_NOTIFY_INFORMATION list in NT_Trans_Parameters.
	listBytes, _ := subcommands.MarshalFileNotifyInformationList([]subcommands.FileNotifyInformation{
		{Action: subcommands.FILE_ACTION_ADDED, FileName: "a.txt"},
		{Action: subcommands.FILE_ACTION_REMOVED, FileName: "b.log"},
	})
	resp := commands.NewNtTransactResponse()
	resp.Parameters = listBytes
	items, err := resp.NotifyChangeInformation()
	if err != nil {
		t.Fatalf("NotifyChangeInformation: %v", err)
	}
	if len(items) != 2 || items[0].FileName != "a.txt" || items[1].Action != subcommands.FILE_ACTION_REMOVED {
		t.Errorf("notify list: %+v", items)
	}
}

func TestSecurityDescBuilders(t *testing.T) {
	params := subcommands.NtTransactSecurityDescParameters{
		FID:                 0x4002,
		SecurityInformation: subcommands.OWNER_SECURITY_INFORMATION | subcommands.DACL_SECURITY_INFORMATION,
	}
	descriptor := []byte{0x01, 0x00, 0x04, 0x80, 0xDE, 0xAD, 0xBE, 0xEF} // opaque self-relative SD bytes

	// Query request: params in NT_Trans_Parameters, MaxParameterCount = 4.
	q, err := commands.NewNtTransactQuerySecurityDescRequest(params, 0x400)
	if err != nil {
		t.Fatalf("query build: %v", err)
	}
	if q.Function != types.USHORT(subcommands.NT_TRANSACT_QUERY_SECURITY_DESC) || q.MaxParameterCount != 4 {
		t.Errorf("query: Function=0x%04x MaxParameterCount=%d", q.Function, q.MaxParameterCount)
	}
	var qp subcommands.NtTransactSecurityDescParameters
	if _, err := qp.Unmarshal(q.NT_Trans_Parameters); err != nil || qp != params {
		t.Errorf("query params: %+v err=%v", qp, err)
	}

	// Set request: params + descriptor in NT_Trans_Data.
	s, err := commands.NewNtTransactSetSecurityDescRequest(params, descriptor)
	if err != nil {
		t.Fatalf("set build: %v", err)
	}
	if s.Function != types.USHORT(subcommands.NT_TRANSACT_SET_SECURITY_DESC) {
		t.Errorf("set Function=0x%04x", s.Function)
	}
	if !bytes.Equal(s.NT_Trans_Data, descriptor) || s.DataCount != types.ULONG(len(descriptor)) {
		t.Errorf("set data mismatch: % x (count %d)", s.NT_Trans_Data, s.DataCount)
	}

	// Response: LengthNeeded parameter + descriptor data.
	respParams := subcommands.NtTransactQuerySecurityDescResponseParameters{LengthNeeded: uint32(len(descriptor))}
	pb, _ := respParams.Marshal()
	resp := commands.NewNtTransactResponse()
	resp.Parameters = pb
	resp.Data = descriptor
	gotParams, gotSD, err := resp.QuerySecurityDescriptor()
	if err != nil {
		t.Fatalf("QuerySecurityDescriptor: %v", err)
	}
	if gotParams.LengthNeeded != uint32(len(descriptor)) || !bytes.Equal(gotSD, descriptor) {
		t.Errorf("query response: %+v / % x", gotParams, gotSD)
	}
}

func TestQuotaBuilderAndResponse(t *testing.T) {
	params := subcommands.NtTransQueryQuotaRequestParameters{FID: 0x4002, ReturnSingleEntry: true}
	req, err := commands.NewNtTransactQueryQuotaRequest(params, nil, 0x2000)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if req.Function != types.USHORT(subcommands.NT_TRANSACT_QUERY_QUOTA) {
		t.Errorf("Function=0x%04x", req.Function)
	}
	var qp subcommands.NtTransQueryQuotaRequestParameters
	if _, err := qp.Unmarshal(req.NT_Trans_Parameters); err != nil || qp != params {
		t.Errorf("quota params: %+v err=%v", qp, err)
	}

	// Response: DataLength parameter + one FILE_QUOTA_INFORMATION record.
	rec := subcommands.FileQuotaInformation{QuotaUsed: 4096, QuotaThreshold: -1, QuotaLimit: -1, Sid: []byte{0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}}
	recBytes, _ := rec.Marshal()
	respParams := subcommands.NtTransQuotaResponseParameters{DataLength: uint32(len(recBytes))}
	pb, _ := respParams.Marshal()
	resp := commands.NewNtTransactResponse()
	resp.Parameters = pb
	resp.Data = recBytes
	gotParams, list, err := resp.QuotaInformation()
	if err != nil {
		t.Fatalf("QuotaInformation: %v", err)
	}
	if gotParams.DataLength != uint32(len(recBytes)) || len(list) != 1 || list[0].QuotaUsed != 4096 {
		t.Errorf("quota response: %+v list=%+v", gotParams, list)
	}
}

func TestCopychunkBuilderAndResponse(t *testing.T) {
	cc := subcommands.SrvCopychunkCopy{
		CopychunkResumeKey: [24]byte{0x2D, 0x0B},
		Chunks:             []subcommands.SrvCopychunk{{SourceOffset: 0, DestinationOffset: 0, Length: 0x063C}},
	}
	req, err := commands.NewNtTransactCopychunkRequest(0x4003, cc)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if req.Function != types.USHORT(subcommands.NT_TRANSACT_IOCTL) || len(req.Setup) != 4 {
		t.Fatalf("Function=0x%04x setup=%d", req.Function, len(req.Setup))
	}
	// Setup FSCTL code.
	b := make([]byte, 8)
	for i, w := range req.Setup {
		binary.LittleEndian.PutUint16(b[i*2:], uint16(w))
	}
	if binary.LittleEndian.Uint32(b[0:4]) != subcommands.FSCTL_SRV_COPYCHUNK {
		t.Errorf("setup FSCTL code mismatch")
	}
	// NT_Trans_Data round-trips to the copy struct.
	var gotCopy subcommands.SrvCopychunkCopy
	if _, err := gotCopy.Unmarshal(req.NT_Trans_Data); err != nil {
		t.Fatalf("copy unmarshal: %v", err)
	}
	if len(gotCopy.Chunks) != 1 || gotCopy.Chunks[0].Length != 0x063C {
		t.Errorf("copy data: %+v", gotCopy)
	}

	// Response.
	rb, _ := (&subcommands.SrvCopychunkResponse{ChunksWritten: 1, TotalBytesWritten: 0x063C}).Marshal()
	resp := commands.NewNtTransactResponse()
	resp.Data = rb
	got, err := resp.CopychunkResponse()
	if err != nil {
		t.Fatalf("CopychunkResponse: %v", err)
	}
	if got.ChunksWritten != 1 || got.TotalBytesWritten != 0x063C {
		t.Errorf("copychunk response: %+v", got)
	}
}

func TestSnapshotAndResumeKeyResponses(t *testing.T) {
	// Resume-key request build.
	rk := commands.NewNtTransactRequestResumeKeyRequest(0x4002, 0x400)
	b := make([]byte, 8)
	for i, w := range rk.Setup {
		binary.LittleEndian.PutUint16(b[i*2:], uint16(w))
	}
	if binary.LittleEndian.Uint32(b[0:4]) != subcommands.FSCTL_SRV_REQUEST_RESUME_KEY {
		t.Errorf("resume-key setup FSCTL mismatch")
	}

	// Resume-key response parse.
	rkResp, _ := (&subcommands.SrvRequestResumeKeyResponse{CopychunkResumeKey: [24]byte{0xAA, 0xBB}}).Marshal()
	resp := commands.NewNtTransactResponse()
	resp.Data = rkResp
	gotRK, err := resp.RequestResumeKey()
	if err != nil || gotRK.CopychunkResumeKey[0] != 0xAA {
		t.Fatalf("RequestResumeKey: %+v err=%v", gotRK, err)
	}

	// Snapshot enumerate request build + response parse.
	snReq := commands.NewNtTransactEnumerateSnapshotsRequest(0x4002, 0x2000)
	if snReq.Function != types.USHORT(subcommands.NT_TRANSACT_IOCTL) {
		t.Errorf("snapshot Function=0x%04x", snReq.Function)
	}
	snBytes, _ := (&subcommands.SrvSnapshotArray{NumberOfSnapShots: 1, Snapshots: []string{"@GMT-2020.01.02-03.04.05"}}).Marshal()
	resp2 := commands.NewNtTransactResponse()
	resp2.Data = snBytes
	got, err := resp2.SnapshotArray()
	if err != nil {
		t.Fatalf("SnapshotArray: %v", err)
	}
	if len(got.Snapshots) != 1 || got.Snapshots[0] != "@GMT-2020.01.02-03.04.05" {
		t.Errorf("snapshot array: %+v", got)
	}
}
