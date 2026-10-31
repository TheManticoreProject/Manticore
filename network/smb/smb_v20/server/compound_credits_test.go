package server_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
	"github.com/TheManticoreProject/Manticore/windows/filesystem/infoclass"
)

// TestServerRelatedCompoundLargeTransactSize checks that a CREATE+QUERY_INFO+CLOSE
// chain stays within the credits the client holds when the server advertises an
// 8 MiB MaxTransactSize: the QUERY_INFO alone would otherwise charge 128 credits,
// and the chain as a whole more than the window, so the server drops the
// connection.
func TestServerRelatedCompoundLargeTransactSize(t *testing.T) {
	for _, config := range []server.Config{
		{MaxReadSize: 8 << 20, MaxWriteSize: 8 << 20},
		{MaxReadSize: 8 << 20, MaxWriteSize: 8 << 20, RequireSigning: true},
	} {
		_, port := startServer(t, config)
		c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 3; i++ {
			info, err := c.CreateQueryInfoClose("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, 0,
				1, uint8(infoclass.FileStandardInformation), 0)
			if err != nil {
				t.Fatalf("CREATE+QUERY_INFO+CLOSE #%d (signing %v): %v", i, config.RequireSigning, err)
			}
			var standard filesystem.FileStandardInformation
			if err := standard.Unmarshal(info); err != nil || standard.EndOfFile != 11 {
				t.Fatalf("FileStandardInformation = %+v, %v", standard, err)
			}
		}
	}
}
