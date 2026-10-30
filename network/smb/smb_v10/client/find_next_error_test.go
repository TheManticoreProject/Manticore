package client

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// findFirst2MoreToCome returns a FIND_FIRST2 response with one entry and
// EndOfSearch = 0, so the client goes on to FIND_NEXT2.
func findFirst2MoreToCome(t *testing.T) []byte {
	t.Helper()
	params := make([]byte, 10)
	binary.LittleEndian.PutUint16(params[0:], 0x1234) // SID
	binary.LittleEndian.PutUint16(params[2:], 1)      // SearchCount
	binary.LittleEndian.PutUint16(params[4:], 0)      // EndOfSearch
	entry := bothDirInfoEntry("a.txt", true, 0)
	r := commands.NewTransaction2Response()
	r.TotalParameterCount = types.USHORT(len(params))
	r.ParameterCount = types.USHORT(len(params))
	r.ParameterOffset = 56
	r.Pad1 = []types.UCHAR{0}
	r.Trans2_Parameters = params
	r.TotalDataCount = types.USHORT(len(entry))
	r.DataCount = types.USHORT(len(entry))
	r.DataOffset = types.USHORT(56 + len(params))
	r.Trans2_Data = entry
	m := message.NewMessage()
	m.Header.SetFlags(flags.FLAGS_REPLY)
	m.AddCommand(r)
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func findClose2Reply(t *testing.T) []byte {
	t.Helper()
	m := message.NewMessage()
	m.Header.SetFlags(flags.FLAGS_REPLY)
	m.AddCommand(commands.NewFindClose2Response())
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// A FIND_NEXT2 failure other than STATUS_NO_MORE_FILES must be reported with the
// entries read so far, not hidden behind a nil error.
func TestListEntriesReportsFindNextFailure(t *testing.T) {
	tr := &scriptedTransport{replies: [][]byte{
		findFirst2MoreToCome(t),
		interimTrans2Response(t, 0xC0000022), // STATUS_ACCESS_DENIED
		findClose2Reply(t),
	}}
	c := &Client{Transport: tr, Connection: &Connection{Server: &Server{}}, Session: &Session{}}
	entries, err := c.ListEntries(`\*`)
	if len(entries) != 1 {
		t.Fatalf("ListEntries returned %d entries, want the 1 entry of the first page", len(entries))
	}
	if err == nil {
		t.Fatal("ListEntries returned a nil error after FIND_NEXT2 failed with STATUS_ACCESS_DENIED")
	}
}

// STATUS_NO_MORE_FILES on FIND_NEXT2 is a normal end of search.
func TestListEntriesTreatsNoMoreFilesAsEnd(t *testing.T) {
	tr := &scriptedTransport{replies: [][]byte{
		findFirst2MoreToCome(t),
		interimTrans2Response(t, 0x80000006), // STATUS_NO_MORE_FILES
		findClose2Reply(t),
	}}
	c := &Client{Transport: tr, Connection: &Connection{Server: &Server{}}, Session: &Session{}}
	entries, err := c.ListEntries(`\*`)
	if err != nil {
		t.Fatalf("ListEntries error = %v, want nil for STATUS_NO_MORE_FILES", err)
	}
	if len(entries) != 1 {
		t.Fatalf("ListEntries returned %d entries, want 1", len(entries))
	}
}
