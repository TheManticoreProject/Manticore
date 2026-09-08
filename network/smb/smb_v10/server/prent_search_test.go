package server

import (
	"encoding/binary"
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// searchServer serves a directory of short-named files plus one whose name cannot
// fit the 8.3 field.
func searchServer(t *testing.T, names ...string) (*MemoryFileSystem, *smb1client.Client) {
	t.Helper()

	fs := NewMemoryFileSystem("FILES")
	for _, name := range names {
		if err := fs.AddFile(name, []byte("x")); err != nil {
			t.Fatalf("AddFile(%q) error = %v", name, err)
		}
	}
	_, client := fileServer(t, fs, false)
	return fs, client
}

// searchEntriesOf decodes the packed entries out of an SMB_COM_SEARCH reply.
func searchEntriesOf(t *testing.T, body []byte) []types.SMB_DIRECTORY_INFORMATION {
	t.Helper()

	response := commands.NewSearchResponse()
	if _, err := response.Unmarshal(body); err != nil {
		t.Fatalf("the reply did not decode: %v", err)
	}

	packed := []byte(response.SMB_Directory_Information.Buffer)
	if len(packed)%types.SMB_DIRECTORY_INFORMATION_SIZE != 0 {
		t.Fatalf("the entry block is %d bytes, not a whole number of %d-byte entries",
			len(packed), types.SMB_DIRECTORY_INFORMATION_SIZE)
	}

	entries := []types.SMB_DIRECTORY_INFORMATION{}
	for at := 0; at+types.SMB_DIRECTORY_INFORMATION_SIZE <= len(packed); at += types.SMB_DIRECTORY_INFORMATION_SIZE {
		entry := types.NewSMB_DIRECTORY_INFORMATION()
		if _, err := entry.Unmarshal(packed[at:]); err != nil {
			t.Fatalf("entry at %d did not decode: %v", at, err)
		}
		entries = append(entries, *entry)
	}

	if int(response.Count) != len(entries) {
		t.Errorf("the reply reports %d entries and carries %d", response.Count, len(entries))
	}
	return entries
}

// sendSearch sends an SMB_COM_SEARCH, optionally continuing from a resume key.
func sendSearch(
	t *testing.T,
	client *smb1client.Client,
	pattern string,
	maxCount int,
	resumeKey *types.SMB_RESUME_KEY,
) (uint32, []byte) {
	t.Helper()

	request := commands.NewSearchRequest()
	request.MaxCount = types.USHORT(maxCount)
	if err := request.FileName.SetString(pattern); err != nil {
		t.Fatalf("SetString() error = %v", err)
	}
	if resumeKey != nil {
		request.ResumeKey = *resumeKey
		request.ResumeKeyPresent = true
	}

	return sendLegacy(t, client, codes.SMB_COM_SEARCH, request)
}

// TestSearchEnumeratesADirectory asserts the core-set search returns the entries
// of a directory, packed at the fixed stride.
func TestSearchEnumeratesADirectory(t *testing.T) {
	_, client := searchServer(t, "ONE.TXT", "TWO.TXT", "THREE.TXT")

	status, body := sendSearch(t, client, `\*`, 10, nil)
	if status != 0 {
		t.Fatalf("the search reported 0x%08X, want success", status)
	}

	entries := searchEntriesOf(t, body)
	if len(entries) != 3 {
		t.Fatalf("the search returned %d entries, want 3", len(entries))
	}

	found := map[string]bool{}
	for _, entry := range entries {
		found[entry.FileName.GetString()] = true
	}
	for _, want := range []string{"ONE.TXT", "TWO.TXT", "THREE.TXT"} {
		if !found[want] {
			t.Errorf("the listing does not contain %q; it has %v", want, found)
		}
	}
}

// TestSearchResumesFromItsKey asserts a search continues where the previous
// response left off, and finishes.
//
// The position travels in the resume key rather than in a server-side table, so
// this is the assertion that the key is both issued and honoured — a server that
// ignored it would restart the listing every call and the client would loop.
func TestSearchResumesFromItsKey(t *testing.T) {
	_, client := searchServer(t, "ONE.TXT", "TWO.TXT", "THREE.TXT", "FOUR.TXT")

	// One entry at a time, so every step needs the key.
	seen := []string{}
	var key *types.SMB_RESUME_KEY

	for round := 0; round < 8; round++ {
		status, body := sendSearch(t, client, `\*`, 1, key)
		if status == uint32(nt_status.NT_STATUS_NO_MORE_FILES) {
			break
		}
		if status != 0 {
			t.Fatalf("round %d reported 0x%08X", round, status)
		}

		entries := searchEntriesOf(t, body)
		if len(entries) != 1 {
			t.Fatalf("round %d returned %d entries, want 1", round, len(entries))
		}
		seen = append(seen, entries[0].FileName.GetString())

		last := entries[len(entries)-1].ResumeKey
		key = &last
	}

	if len(seen) != 4 {
		t.Fatalf("the walk saw %d entries in total, want 4: %v", len(seen), seen)
	}

	// And no entry was seen twice, which is what a resume key that did not
	// advance would produce.
	unique := map[string]bool{}
	for _, name := range seen {
		if unique[name] {
			t.Errorf("%q was returned twice; the resume key is not advancing", name)
		}
		unique[name] = true
	}
}

// TestSearchRefusesAnInventedResumeKey asserts a key the server did not issue is
// refused rather than read as a position.
//
// Resuming from a position derived from arbitrary bytes would return an arbitrary
// part of the directory and present it as a continuation.
func TestSearchRefusesAnInventedResumeKey(t *testing.T) {
	_, client := searchServer(t, "ONE.TXT")

	invented := types.SMB_RESUME_KEY{Reserved: 0xFF}
	binary.LittleEndian.PutUint32(invented.ServerState[0:4], 0x41414141)

	status, _ := sendSearch(t, client, `\*`, 10, &invented)
	if status != uint32(nt_status.NT_STATUS_INVALID_PARAMETER) {
		t.Errorf("an invented resume key reported 0x%08X, want STATUS_INVALID_PARAMETER (0x%08X)",
			status, uint32(nt_status.NT_STATUS_INVALID_PARAMETER))
	}
}

// TestSearchEchoesTheClientState asserts the client's half of the resume key comes
// back untouched.
//
// [MS-CIFS] section 2.2.4.58.1: "The value provided by the client MUST be returned
// in each ResumeKey provided in the response." It is the client's state, not the
// server's, and a client may be keeping its own position in it.
func TestSearchEchoesTheClientState(t *testing.T) {
	_, client := searchServer(t, "ONE.TXT", "TWO.TXT")

	// A first call to obtain a key this server issued, then continue with the
	// client's own bytes written into it.
	status, body := sendSearch(t, client, `\*`, 1, nil)
	if status != 0 {
		t.Fatalf("the first search reported 0x%08X", status)
	}
	key := searchEntriesOf(t, body)[0].ResumeKey
	key.ClientState = [4]types.UCHAR{0xDE, 0xAD, 0xBE, 0xEF}

	status, body = sendSearch(t, client, `\*`, 1, &key)
	if status != 0 {
		t.Fatalf("the continued search reported 0x%08X", status)
	}

	got := searchEntriesOf(t, body)[0].ResumeKey.ClientState
	if got != [4]types.UCHAR{0xDE, 0xAD, 0xBE, 0xEF} {
		t.Errorf("ClientState came back as %v, want it echoed unchanged", got)
	}
}

// TestSearchOmitsNamesThatDoNotFitTheField asserts a name too long for the 8.3
// field is left out rather than truncated.
//
// The entry's name field is thirteen bytes and this server has no 8.3 alias to
// substitute. Truncating would name a different file, and inventing an alias would
// hand the client a name no subsequent open could resolve — so the entry is
// omitted, and a client that needs it has to use TRANS2_FIND_FIRST2.
func TestSearchOmitsNamesThatDoNotFitTheField(t *testing.T) {
	_, client := searchServer(t, "SHORT.TXT", "a-very-long-file-name.txt")

	status, body := sendSearch(t, client, `\*`, 10, nil)
	if status != 0 {
		t.Fatalf("the search reported 0x%08X, want success", status)
	}

	entries := searchEntriesOf(t, body)
	if len(entries) != 1 {
		t.Fatalf("the search returned %d entries, want only the short-named one", len(entries))
	}
	if got := entries[0].FileName.GetString(); got != "SHORT.TXT" {
		t.Errorf("the listing names %q, want %q", got, "SHORT.TXT")
	}

	// The long-named file is still there, and still listable at an NT level.
	nt, err := client.ListDirectory(`\`)
	if err != nil {
		t.Fatalf("listing at an NT level failed: %v", err)
	}
	if !containsName(namesOf(nt), "a-very-long-file-name.txt") {
		t.Error("the long-named file is missing from an NT listing too, so it was not merely omitted")
	}
}

// TestSearchOfAnExhaustedDirectoryReportsNoMoreFiles asserts the end of a walk is
// reported rather than answered with an empty success.
//
// A client given an empty successful answer asks again forever.
func TestSearchOfAnExhaustedDirectoryReportsNoMoreFiles(t *testing.T) {
	_, client := searchServer(t, "ONE.TXT")

	status, body := sendSearch(t, client, `\*`, 10, nil)
	if status != 0 {
		t.Fatalf("the first search reported 0x%08X", status)
	}
	key := searchEntriesOf(t, body)[0].ResumeKey

	status, _ = sendSearch(t, client, `\*`, 10, &key)
	if status != uint32(nt_status.NT_STATUS_NO_MORE_FILES) {
		t.Errorf("an exhausted search reported 0x%08X, want STATUS_NO_MORE_FILES (0x%08X)",
			status, uint32(nt_status.NT_STATUS_NO_MORE_FILES))
	}
}

// TestFindAndFindUniqueAreServed asserts the other two search commands answer.
func TestFindAndFindUniqueAreServed(t *testing.T) {
	_, client := searchServer(t, "ONE.TXT", "TWO.TXT")

	t.Run("find", func(t *testing.T) {
		request := commands.NewFindRequest()
		request.MaxCount = types.USHORT(10)
		if err := request.FileName.SetString(`\*`); err != nil {
			t.Fatalf("SetString() error = %v", err)
		}

		status, body := sendLegacy(t, client, codes.SMB_COM_FIND, request)
		if status != 0 {
			t.Fatalf("SMB_COM_FIND reported 0x%08X, want success", status)
		}

		response := commands.NewFindResponse()
		if _, err := response.Unmarshal(body); err != nil {
			t.Fatalf("the reply did not decode: %v", err)
		}
		if int(response.Count) != 2 || len(response.DirectoryInformationData) != 2 {
			t.Fatalf("SMB_COM_FIND returned Count=%d with %d entries, want 2",
				response.Count, len(response.DirectoryInformationData))
		}
	})

	t.Run("find unique", func(t *testing.T) {
		request := commands.NewFindUniqueRequest()
		request.MaxCount = types.USHORT(10)
		if err := request.FileName.SetString(`\ONE.TXT`); err != nil {
			t.Fatalf("SetString() error = %v", err)
		}

		status, body := sendLegacy(t, client, codes.SMB_COM_FIND_UNIQUE, request)
		if status != 0 {
			t.Fatalf("SMB_COM_FIND_UNIQUE reported 0x%08X, want success", status)
		}

		response := commands.NewFindUniqueResponse()
		if _, err := response.Unmarshal(body); err != nil {
			t.Fatalf("the reply did not decode: %v", err)
		}
		if int(response.Count) != 1 {
			t.Fatalf("SMB_COM_FIND_UNIQUE returned %d entries for one name", response.Count)
		}
		if got := response.DirectoryInformationData[0].FileName.GetString(); got != "ONE.TXT" {
			t.Errorf("it names %q, want %q", got, "ONE.TXT")
		}
	})

	t.Run("find unique on a missing name", func(t *testing.T) {
		request := commands.NewFindUniqueRequest()
		request.MaxCount = types.USHORT(10)
		if err := request.FileName.SetString(`\NOSUCH.TXT`); err != nil {
			t.Fatalf("SetString() error = %v", err)
		}

		status, _ := sendLegacy(t, client, codes.SMB_COM_FIND_UNIQUE, request)
		if status == 0 {
			t.Error("SMB_COM_FIND_UNIQUE succeeded for a name that does not exist")
		}
	})

	t.Run("find close", func(t *testing.T) {
		request := commands.NewFindCloseRequest()
		status, _ := sendLegacy(t, client, codes.SMB_COM_FIND_CLOSE, request)
		if status != 0 {
			t.Fatalf("SMB_COM_FIND_CLOSE reported 0x%08X, want success", status)
		}
	})
}
