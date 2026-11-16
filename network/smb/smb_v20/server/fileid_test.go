package server_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// TestCreateFileIdHalvesAreDistinct checks that a CREATE response's FileId has
// two unrelated halves rather than the equal, counting-from-one halves a plain
// per-connection counter produces. Windows returns two independent numbers here;
// equal halves below 0x10000 were a server fingerprint (M-075 / #2044).
func TestCreateFileIdHalvesAreDistinct(t *testing.T) {
	_, port := startServer(t, server.Config{})
	c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}

	// Open several handles on one connection so a counter would reveal itself as
	// consecutive, equal-halved ids.
	var ids []types.SMB2_FILEID
	for _, name := range []string{"a.txt", "b.txt", "c.txt", "d.txt"} {
		id, err := c.CreateFile(name, fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
			fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN_IF, 0)
		if err != nil {
			t.Fatalf("create %q: %v", name, err)
		}
		ids = append(ids, id)
	}

	seenVolatile := map[uint64]bool{}
	for i, id := range ids {
		if id.Persistent == id.Volatile {
			t.Errorf("handle %d: FileId Persistent == Volatile (0x%016x); Windows returns two unrelated values", i, id.Volatile)
		}
		if id.Persistent == 0 || id.Volatile == 0 {
			t.Errorf("handle %d: FileId half is zero (persistent=0x%016x volatile=0x%016x)", i, id.Persistent, id.Volatile)
		}
		// The published detection rule flags a FileId whose halves are equal and
		// below 0x10000. Neither condition must hold.
		if id.Persistent == id.Volatile && id.Volatile < 0x10000 {
			t.Errorf("handle %d: FileId matches the small-equal-halves detection rule", i)
		}
		if seenVolatile[id.Volatile] {
			t.Errorf("handle %d: Volatile 0x%016x repeats a live handle's", i, id.Volatile)
		}
		seenVolatile[id.Volatile] = true
	}

	// The real, distinct-halved FileId still resolves to its handle.
	if _, err := c.WriteFile(ids[0], 0, []byte("hi")); err != nil {
		t.Fatalf("write with the real FileId: %v", err)
	}

	// A FileId carrying the right Volatile but a forged Persistent (here the old
	// equal-halves guess) is refused: the server validates both halves, not just
	// the Volatile lookup key.
	forged := types.SMB2_FILEID{Persistent: ids[0].Volatile, Volatile: ids[0].Volatile}
	if ids[0].Persistent == forged.Persistent {
		t.Fatal("test precondition failed: the real Persistent already equals the Volatile")
	}
	_, err = c.WriteFile(forged, 0, []byte("x"))
	if err == nil {
		t.Fatal("write with a forged Persistent was accepted, want STATUS_FILE_CLOSED")
	}
	wantHex := fmt.Sprintf("0x%08x", uint32(nt_status.NT_STATUS_FILE_CLOSED))
	if !strings.Contains(err.Error(), wantHex) {
		t.Errorf("forged write error = %q, want it to report STATUS_FILE_CLOSED (%s)", err, wantHex)
	}
}
