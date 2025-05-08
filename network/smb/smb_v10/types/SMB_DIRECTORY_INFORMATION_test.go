package types_test

import (
	"bytes"
	"encoding/hex"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func TestSMB_DIRECTORY_INFORMATION_Marshal(t *testing.T) {
	testCases := []struct {
		name           string
		input          *types.SMB_DIRECTORY_INFORMATION
		expectedOutput string
		expectError    bool
	}{
		{
			name: "Basic file entry",
			input: &types.SMB_DIRECTORY_INFORMATION{
				ResumeKey:      types.SMB_RESUME_KEY{},
				FileAttributes: types.UCHAR(types.ATTR_ARCHIVE),
				LastWriteTime:  *types.NewSMB_TIMEFromTime(time.Date(2021, time.December, 3, 0, 0, 0, 0, time.UTC)),
				LastWriteDate:  *types.NewSMB_DATEFromDate(2021, 12, 3),
				FileSize:       types.ULONG(1024),
				FileName:       *types.NewOEM_STRINGFromString("TEST.TXT"),
			},
			expectedOutput: "05150000000000000000000000000000000000000000000020008001b7d8e7d70183530004000004544553542e5458542020202000",
			expectError:    false,
		},
		{
			name: "Directory entry",
			input: &types.SMB_DIRECTORY_INFORMATION{
				ResumeKey:      types.SMB_RESUME_KEY{},
				FileAttributes: types.UCHAR(types.ATTR_DIRECTORY),
				LastWriteTime:  *types.NewSMB_TIMEFromTime(time.Date(2021, time.December, 3, 0, 0, 0, 0, time.UTC)),
				LastWriteDate:  *types.NewSMB_DATEFromDate(2021, 12, 3),
				FileSize:       types.ULONG(0),
				FileName:       *types.NewOEM_STRINGFromString("FOLDER"),
			},
			expectedOutput: "05150000000000000000000000000000000000000000000010008001b7d8e7d70183530000000004464f4c44455220202020202000",
			expectError:    false,
		},
		{
			name: "With Resume Key",
			input: &types.SMB_DIRECTORY_INFORMATION{
				ResumeKey: types.SMB_RESUME_KEY{
					SMB_STRING: types.SMB_STRING{
						BufferFormat: types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK,
						Length:       0,
						Buffer:       []byte{},
					},
					Reserved:    0,
					ServerState: [16]types.UCHAR{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
					ClientState: [4]types.UCHAR{1, 2, 3, 4},
				},
				FileAttributes: types.UCHAR(types.ATTR_DIRECTORY),
				LastWriteTime:  *types.NewSMB_TIMEFromTime(time.Date(2021, time.December, 3, 0, 0, 0, 0, time.UTC)),
				LastWriteDate:  *types.NewSMB_DATEFromDate(2021, 12, 3),
				FileSize:       types.ULONG(0),
				FileName:       *types.NewOEM_STRINGFromString("PODA.LIRIUS"),
			},
			expectedOutput: "051500000102030405060708090a0b0c0d0e0f100102030410008001b7d8e7d70183530000000004504f44412e4c49524955532000",
			expectError:    false,
		},
		{
			name: "Filename with single character",
			input: &types.SMB_DIRECTORY_INFORMATION{
				ResumeKey:      types.SMB_RESUME_KEY{},
				FileAttributes: types.UCHAR(types.ATTR_ARCHIVE),
				LastWriteTime:  *types.NewSMB_TIMEFromTime(time.Date(2021, time.December, 3, 0, 0, 0, 0, time.UTC)),
				LastWriteDate:  *types.NewSMB_DATEFromDate(2021, 12, 3),
				FileSize:       types.ULONG(512),
				FileName:       *types.NewOEM_STRINGFromString("A"),
			},
			expectedOutput: "05150000000000000000000000000000000000000000000020008001b7d8e7d7018353000200000441202020202020202020202000",
			expectError:    false,
		},
		{
			name: "Filename with exactly 12 characters",
			input: &types.SMB_DIRECTORY_INFORMATION{
				ResumeKey:      types.SMB_RESUME_KEY{},
				FileAttributes: types.UCHAR(types.ATTR_ARCHIVE),
				LastWriteTime:  *types.NewSMB_TIMEFromTime(time.Date(2021, time.December, 3, 0, 0, 0, 0, time.UTC)),
				LastWriteDate:  *types.NewSMB_DATEFromDate(2021, 12, 3),
				FileSize:       types.ULONG(2048),
				FileName:       *types.NewOEM_STRINGFromString("ABCDEFGHIJKL"),
			},
			expectedOutput: "05150000000000000000000000000000000000000000000020008001b7d8e7d701835300080000044142434445464748494a4b4c00",
			expectError:    false,
		},
		{
			name: "Filename too long (13+ characters)",
			input: &types.SMB_DIRECTORY_INFORMATION{
				ResumeKey:      types.SMB_RESUME_KEY{},
				FileAttributes: types.UCHAR(types.ATTR_ARCHIVE),
				LastWriteTime:  *types.NewSMB_TIMEFromTime(time.Date(2021, time.December, 3, 0, 0, 0, 0, time.UTC)),
				LastWriteDate:  *types.NewSMB_DATEFromDate(2021, 12, 3),
				FileSize:       types.ULONG(4096),
				FileName:       *types.NewOEM_STRINGFromString("ABCDEFGHIJKLM"),
			},
			expectedOutput: "",
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.input.Marshal()

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			expectedBytes, err := hex.DecodeString(tc.expectedOutput)
			if err != nil {
				t.Fatalf("Failed to decode expected output: %v", err)
			}

			if !bytes.Equal(result, expectedBytes) {
				t.Errorf("Marshal output mismatch\nExpected: %s\nGot:      %s", tc.expectedOutput, hex.EncodeToString(result))
				if len(result) != len(tc.expectedOutput) {
					t.Errorf("Length mismatch: expected %d, got %d", len(tc.expectedOutput), len(result))
				} else {
					for i := 0; i < len(result); i++ {
						if result[i] != tc.expectedOutput[i] {
							t.Errorf("Mismatch at index %d: expected 0x%02x, got 0x%02x", i, tc.expectedOutput[i], result[i])
						}
					}
				}
			}
		})
	}
}

func TestNewSMB_DIRECTORY_INFORMATION(t *testing.T) {
	dirInfo := types.NewSMB_DIRECTORY_INFORMATION()

	if dirInfo == nil {
		t.Fatal("NewSMB_DIRECTORY_INFORMATION returned nil")
	}

	// Verify default values
	if dirInfo.FileAttributes != 0 {
		t.Errorf("Expected FileAttributes to be 0, got %d", dirInfo.FileAttributes)
	}

	if dirInfo.FileSize != 0 {
		t.Errorf("Expected FileSize to be 0, got %d", dirInfo.FileSize)
	}
}
