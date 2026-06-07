package client_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/informationlevels"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// trans2DataReply builds a marshalled SMB_COM_TRANSACTION2 reply whose Trans2_Data
// block carries data. With no setup words the data block begins 55 bytes from the
// start of the SMB header (header 32 + WordCount 1 + 10 words + ByteCount 2).
func trans2DataReply(t *testing.T, data []byte) []byte {
	t.Helper()
	const dataOffset = 55

	resp := commands.NewTransaction2Response()
	resp.TotalDataCount = types.USHORT(len(data))
	resp.DataCount = types.USHORT(len(data))
	resp.DataOffset = types.USHORT(dataOffset)
	resp.Trans2_Data = []types.UCHAR(data)

	return marshalResponse(t, resp)
}

// sentTransaction2 decodes the captured request as a Transaction2Request.
func sentTransaction2(t *testing.T, raw []byte) *commands.Transaction2Request {
	t.Helper()
	msg := message.NewMessage()
	if err := msg.Unmarshal(raw); err != nil {
		t.Fatalf("failed to decode sent request: %v", err)
	}
	req, ok := msg.Command.(*commands.Transaction2Request)
	if !ok {
		t.Fatalf("sent command is %T, want *Transaction2Request", msg.Command)
	}
	return req
}

func TestGetFileStandardInfoDecodesResponse(t *testing.T) {
	want := &informationlevels.SMB_QUERY_FILE_STANDARD_INFO{Numberoflinks: 1, Directory: 1}
	want.Allocationsize.QuadPart = 0x4000
	want.Endoffile.QuadPart = 0x3210
	infoBytes, err := want.Marshal()
	if err != nil {
		t.Fatalf("marshal info: %v", err)
	}

	tr := &capturingTransport{response: trans2DataReply(t, infoBytes)}
	c := newSessionClient(tr)

	got, err := c.GetFileStandardInfo("\\dir\\file.txt")
	if err != nil {
		t.Fatalf("GetFileStandardInfo: %v", err)
	}
	if got.Allocationsize.QuadPart != want.Allocationsize.QuadPart ||
		got.Endoffile.QuadPart != want.Endoffile.QuadPart ||
		got.Numberoflinks != want.Numberoflinks || got.Directory != want.Directory {
		t.Errorf("decoded info mismatch:\n got %+v\n want %+v", got, want)
	}

	// The request must be TRANS2_QUERY_PATH_INFORMATION at the STANDARD level.
	req := sentTransaction2(t, tr.sent)
	if len(req.Setup) != 1 || uint16(req.Setup[0]) != 0x0005 {
		t.Errorf("expected setup [0x0005] (QUERY_PATH_INFORMATION), got %v", req.Setup)
	}
	if len(req.Trans2_Parameters) < 2 {
		t.Fatalf("expected information level in parameters, got %d bytes", len(req.Trans2_Parameters))
	}
	gotLevel := uint16(req.Trans2_Parameters[0]) | uint16(req.Trans2_Parameters[1])<<8
	if gotLevel != client.InfoLevelQueryFileStandard {
		t.Errorf("expected information level 0x%04x, got 0x%04x", client.InfoLevelQueryFileStandard, gotLevel)
	}
}

func TestGetFsSizeInfoDecodesResponse(t *testing.T) {
	want := &informationlevels.SMB_QUERY_FS_SIZE_INFO{Sectorsperallocationunit: 8, Bytespersector: 512}
	want.Totalallocationunits.QuadPart = 0x100000
	want.Totalfreeallocationunits.QuadPart = 0x80000
	infoBytes, _ := want.Marshal()

	tr := &capturingTransport{response: trans2DataReply(t, infoBytes)}
	c := newSessionClient(tr)

	got, err := c.GetFsSizeInfo()
	if err != nil {
		t.Fatalf("GetFsSizeInfo: %v", err)
	}
	if *got != *want {
		t.Errorf("decoded fs size mismatch:\n got %+v\n want %+v", got, want)
	}

	req := sentTransaction2(t, tr.sent)
	if len(req.Setup) != 1 || uint16(req.Setup[0]) != 0x0003 {
		t.Errorf("expected setup [0x0003] (QUERY_FS_INFORMATION), got %v", req.Setup)
	}
}

func TestSetFileEndOfFileRequestShape(t *testing.T) {
	// SET responses carry no data; an empty TRANS2 reply with status 0 suffices.
	tr := &capturingTransport{response: trans2DataReply(t, nil)}
	c := newSessionClient(tr)

	if err := c.SetFileEndOfFile(7, 0x1122334455); err != nil {
		t.Fatalf("SetFileEndOfFile: %v", err)
	}

	req := sentTransaction2(t, tr.sent)
	if len(req.Setup) != 1 || uint16(req.Setup[0]) != 0x0008 {
		t.Errorf("expected setup [0x0008] (SET_FILE_INFORMATION), got %v", req.Setup)
	}
	// Parameters: FID(2) + InformationLevel(2) + Reserved(2).
	if len(req.Trans2_Parameters) < 6 {
		t.Fatalf("expected >=6 parameter bytes, got %d", len(req.Trans2_Parameters))
	}
	gotFID := uint16(req.Trans2_Parameters[0]) | uint16(req.Trans2_Parameters[1])<<8
	gotLevel := uint16(req.Trans2_Parameters[2]) | uint16(req.Trans2_Parameters[3])<<8
	if gotFID != 7 || gotLevel != client.InfoLevelSetFileEndOfFile {
		t.Errorf("expected FID 7 level 0x%04x, got FID %d level 0x%04x", client.InfoLevelSetFileEndOfFile, gotFID, gotLevel)
	}
	// Data: the 8-byte end-of-file value.
	if len(req.Trans2_Data) != 8 {
		t.Errorf("expected 8 data bytes (EndOfFile), got %d", len(req.Trans2_Data))
	}
}

func TestQueryInfoWithoutSession(t *testing.T) {
	c := &client.Client{Connection: &client.Connection{Server: &client.Server{}}}
	if _, err := c.GetFileBasicInfo("\\x"); err == nil {
		t.Error("expected error without a session")
	}
	if err := c.SetFileEndOfFile(1, 0); err == nil {
		t.Error("expected error without a session")
	}
}
