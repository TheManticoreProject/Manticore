package server

import (
	"strings"
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/filesystem/infoclass"
)

// TestAlternateNameHasNoShortName asserts both SMB1 query paths refuse the 8.3
// alternate name with STATUS_OBJECT_NAME_NOT_FOUND instead of returning the long
// name. The volume generates no short names, which is how Windows answers with
// short-name generation disabled; returning the long name was a fingerprint
// (the SMB1 counterpart of #2059, tracked as #2151).
func TestAlternateNameHasNoShortName(t *testing.T) {
	const objectNameNotFound = "0xc0000034" // STATUS_OBJECT_NAME_NOT_FOUND

	client, fid := passthroughServer(t, "a long file name.txt", "contents")

	// Pass-through FileAlternateNameInformation (class 21), over an open FID.
	if _, err := client.QueryFileInformation(fid, passthroughLevel(infoclass.FileAlternateNameInformation)); err == nil {
		t.Error("pass-through FileAlternateNameInformation was answered; want STATUS_OBJECT_NAME_NOT_FOUND")
	} else if !strings.Contains(err.Error(), objectNameNotFound) {
		t.Errorf("pass-through alt-name error = %q, want status %s", err, objectNameNotFound)
	}

	// Legacy SMB_QUERY_FILE_ALT_NAME_INFO (0x0108), over a path.
	if _, err := client.QueryPathInformation("a long file name.txt", smb1client.InfoLevelQueryFileAltName); err == nil {
		t.Error("legacy SMB_QUERY_FILE_ALT_NAME_INFO was answered; want STATUS_OBJECT_NAME_NOT_FOUND")
	} else if !strings.Contains(err.Error(), objectNameNotFound) {
		t.Errorf("legacy alt-name error = %q, want status %s", err, objectNameNotFound)
	}

	// The long-name class is unchanged: it still answers.
	if _, err := client.QueryPathInformation("a long file name.txt", smb1client.InfoLevelQueryFileName); err != nil {
		t.Errorf("SMB_QUERY_FILE_NAME_INFO failed: %v", err)
	}
	if _, err := client.QueryFileInformation(fid, passthroughLevel(infoclass.FileNameInformation)); err != nil {
		t.Errorf("pass-through FileNameInformation failed: %v", err)
	}
}
