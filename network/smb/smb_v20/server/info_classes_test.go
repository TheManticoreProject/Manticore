package server_test

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
	"github.com/TheManticoreProject/Manticore/windows/filesystem/infoclass"
)

// TestServerAnswersAddedInformationClasses lists the share in the
// FileDirectoryInformation and FileIdExtdDirectoryInformation classes and
// queries a file's streams and the volume's sector size, decoding each answer
// with the windows/filesystem structures.
func TestServerAnswersAddedInformationClasses(t *testing.T) {
	_, port := startServer(t, server.Config{})
	c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	root, err := c.CreateFile("", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, fileflags.FILE_DIRECTORY_FILE)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseFile(root)

	const restart = 0x01 // SMB2_RESTART_SCANS
	raw, err := c.QueryDirectory(root, uint8(infoclass.FileDirectoryInformation), "hello.txt", restart)
	if err != nil {
		t.Fatalf("QueryDirectory(FileDirectoryInformation): %v", err)
	}
	dir := filesystem.ParseFileDirectoryInformationList(raw)
	if len(dir) != 1 || dir[0].FileName != "hello.txt" || dir[0].EndOfFile != 11 {
		t.Errorf("FileDirectoryInformation listing = %+v", dir)
	}

	raw, err = c.QueryDirectory(root, uint8(infoclass.FileIdExtdDirectoryInformation), "hello.txt", restart)
	if err != nil {
		t.Fatalf("QueryDirectory(FileIdExtdDirectoryInformation): %v", err)
	}
	extd := filesystem.ParseFileIdExtdDirectoryInformationList(raw)
	if len(extd) != 1 || extd[0].FileName != "hello.txt" || extd[0].EndOfFile != 11 {
		t.Fatalf("FileIdExtdDirectoryInformation listing = %+v", extd)
	}

	file, err := c.CreateFile("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseFile(file)

	// The listing's 128-bit FileId carries the same index FileInternalInformation reports.
	internal, err := c.QueryInfo(file, 0x01, uint8(infoclass.FileInternalInformation), 0)
	if err != nil || len(internal) != 8 {
		t.Fatalf("FileInternalInformation = %x, %v", internal, err)
	}
	if got := binary.LittleEndian.Uint64(extd[0].FileId[:8]); got != binary.LittleEndian.Uint64(internal) {
		t.Errorf("FileId low half %#x, FileInternalInformation %#x", got, binary.LittleEndian.Uint64(internal))
	}

	streams, err := c.QueryFileStreamInfo(file)
	if err != nil {
		t.Fatalf("QueryFileStreamInfo: %v", err)
	}
	if len(streams) != 1 || streams[0].StreamName != "::$DATA" || streams[0].StreamSize != 11 {
		t.Errorf("streams = %+v", streams)
	}

	sectors, err := c.QueryFsSectorSizeInfo(file)
	if err != nil {
		t.Fatalf("QueryFsSectorSizeInfo: %v", err)
	}
	if sectors.LogicalBytesPerSector == 0 || sectors.PhysicalBytesPerSectorForAtomicity != sectors.LogicalBytesPerSector {
		t.Errorf("sector size = %+v", sectors)
	}
}
