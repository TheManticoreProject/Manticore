package server

import (
	"strconv"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// emptyFileMaps counts the files the table still keeps a grant map for although
// no grant is left on them.
func emptyFileMaps(t *oplockTable) int {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	empty := 0
	for _, grants := range t.grants {
		if len(grants) == 0 {
			empty++
		}
	}
	return empty
}

// A break that withdraws the last grant on a file drops the file from the
// table, as a release does, so the table holds no file without a grant.
func TestOplockTableDropsFilesWithoutGrants(t *testing.T) {
	table := newOplockTable()
	tree := &Tree{Share: &Share{Name: "files"}}
	conn := &Connection{transport: discardTransport{}}
	session := &Session{}
	request := &commands.CreateRequest{RequestedOplockLevel: commands.SMB2_OPLOCK_LEVEL_II}

	var actors, holders []*Open
	for i := 0; i < 50; i++ {
		path := "file" + strconv.Itoa(i) + ".txt"
		holder := &Open{session: session, tree: tree, path: path}
		if level, _ := table.grant(conn, holder, request); level != commands.SMB2_OPLOCK_LEVEL_II {
			t.Fatalf("grant = level %d, want level II", level)
		}
		// The writer holds no oplock of its own, so the break empties the file.
		actor := &Open{session: session, tree: tree, path: path}
		table.breakOthers(actor)
		actors, holders = append(actors, actor), append(holders, holder)
	}
	if n := emptyFileMaps(table); n != 0 {
		t.Fatalf("%d files kept in the oplock table with no grant left on them", n)
	}

	for i := range actors {
		table.release(actors[i])
		table.release(holders[i])
	}
	if n := len(table.grants); n != 0 {
		t.Fatalf("oplock table holds %d files after every open was released", n)
	}
}
