package server_test

import (
	"encoding/binary"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
)

func TestDFSReferralRequestStatus(t *testing.T) {
	_, port := startServer(t, server.Config{})
	c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.TreeConnect("IPC$"); err != nil {
		t.Fatalf("TreeConnect IPC$: %v", err)
	}

	fileID := types.SMB2_FILEID{Persistent: ^uint64(0), Volatile: ^uint64(0)}
	path := `\FILESRV\manticore-does-not-exist`
	for _, tc := range []struct {
		name    string
		ctlCode uint32
		input   []byte
	}{
		{"FSCTL_DFS_GET_REFERRALS", filesystem.FSCTL_DFS_GET_REFERRALS, dfsReferralRequest(path)},
		{"FSCTL_DFS_GET_REFERRALS_EX", filesystem.FSCTL_DFS_GET_REFERRALS_EX, dfsReferralRequestEx(path, "Default-First-Site-Name")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertIOCTLStatus(t, c, fileID, tc.ctlCode, tc.input, true, "NT_STATUS_NOT_FOUND (0xc0000225)")
		})
	}

	t.Run("malformed classic request", func(t *testing.T) {
		assertIOCTLStatus(t, c, fileID, filesystem.FSCTL_DFS_GET_REFERRALS, []byte{4, 0}, true, "NT_STATUS_INVALID_PARAMETER (0xc000000d)")
	})
	t.Run("malformed extended request", func(t *testing.T) {
		assertIOCTLStatus(t, c, fileID, filesystem.FSCTL_DFS_GET_REFERRALS_EX, dfsReferralRequest(path), true, "NT_STATUS_INVALID_PARAMETER (0xc000000d)")
	})
	t.Run("domain referral on non-DC", func(t *testing.T) {
		assertIOCTLStatus(t, c, fileID, filesystem.FSCTL_DFS_GET_REFERRALS, dfsReferralRequest(""), true, "NT_STATUS_INVALID_PARAMETER (0xc000000d)")
	})
	t.Run("DC referral on non-DC", func(t *testing.T) {
		assertIOCTLStatus(t, c, fileID, filesystem.FSCTL_DFS_GET_REFERRALS, dfsReferralRequest(`\LAB`), true, "NT_STATUS_INVALID_PARAMETER (0xc000000d)")
	})
	t.Run("SYSVOL referral on non-DC", func(t *testing.T) {
		assertIOCTLStatus(t, c, fileID, filesystem.FSCTL_DFS_GET_REFERRALS, dfsReferralRequest(`\LAB\SYSVOL`), true, "NT_STATUS_NOT_FOUND (0xc0000225)")
	})
	t.Run("unknown domain-based namespace", func(t *testing.T) {
		assertIOCTLStatus(t, c, fileID, filesystem.FSCTL_DFS_GET_REFERRALS, dfsReferralRequest(`\LAB\missing`), true, "NT_STATUS_DFS_UNAVAILABLE (0xc000026d)")
	})
	t.Run("missing FSCTL flag", func(t *testing.T) {
		assertIOCTLStatus(t, c, fileID, filesystem.FSCTL_DFS_GET_REFERRALS, dfsReferralRequest(path), false, "NT_STATUS_NOT_SUPPORTED (0xc00000bb)")
	})
	t.Run("non-sentinel FileId", func(t *testing.T) {
		assertIOCTLStatus(t, c, types.SMB2_FILEID{}, filesystem.FSCTL_DFS_GET_REFERRALS, dfsReferralRequest(path), true, "NT_STATUS_INVALID_PARAMETER (0xc000000d)")
	})
}

type ioctlClient interface {
	Ioctl(types.SMB2_FILEID, uint32, []byte, bool, uint32) ([]byte, error)
}

func assertIOCTLStatus(t *testing.T, c ioctlClient, fileID types.SMB2_FILEID, ctlCode uint32, input []byte, isFSCTL bool, want string) {
	t.Helper()
	_, err := c.Ioctl(fileID, ctlCode, input, isFSCTL, 65536)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("IOCTL returned %v, want %s", err, want)
	}
}

func dfsReferralRequest(path string) []byte {
	return append(binary.LittleEndian.AppendUint16(nil, 4), utf16Bytes(path)...)
}

func dfsReferralRequestEx(path, site string) []byte {
	pathBytes := utf16Bytes(path)
	siteBytes := utf16Bytes(site)
	requestData := binary.LittleEndian.AppendUint16(nil, uint16(len(pathBytes)))
	requestData = append(requestData, pathBytes...)
	requestData = binary.LittleEndian.AppendUint16(requestData, uint16(len(siteBytes)))
	requestData = append(requestData, siteBytes...)

	request := binary.LittleEndian.AppendUint16(nil, 4)
	request = binary.LittleEndian.AppendUint16(request, 1) // SiteName is present.
	request = binary.LittleEndian.AppendUint32(request, uint32(len(requestData)))
	return append(request, requestData...)
}

func utf16Bytes(value string) []byte {
	codeUnits := utf16.Encode([]rune(value + "\x00"))
	encoded := make([]byte, 0, len(codeUnits)*2)
	for _, codeUnit := range codeUnits {
		encoded = binary.LittleEndian.AppendUint16(encoded, codeUnit)
	}
	return encoded
}
