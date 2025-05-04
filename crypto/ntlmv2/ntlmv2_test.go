package ntlmv2_test

// import (
// 	"testing"

// 	"github.com/TheManticoreProject/Manticore/crypto/ntlmv2"
// )

// func TestNTLMv2HashToHashcatString(t *testing.T) {
// 	testCases := []struct {
// 		testName        string
// 		domain          string
// 		username        string
// 		password        string
// 		serverChallenge [8]byte
// 		clientChallenge [8]byte
// 		expectedHash    string
// 	}{
// 		{
// 			testName:        "Domain username and password to NTLMv2 hashcat string",
// 			domain:          "LAB",
// 			username:        "Podalirius",
// 			password:        "Admin123!",
// 			serverChallenge: [8]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88},
// 			clientChallenge: [8]byte{0x14, 0x78, 0x18, 0x65, 0x59, 0x40, 0x07, 0x4B, 0xBB, 0x99, 0xDA, 0x22, 0x21, 0x04, 0xF1, 0x82},
// 			expectedHash:    "Podalirius::LAB:1122334455667788:147818655940074BBB99DA222104F182:0101000000000000EDC8BDBFD3AADB014B40D4784E921B03000000000200080034004B003200490001001E00570049004E002D00360050004C0059004E003500350045004400420032000400140034004B00320049002E004C004F00430041004C0003003400570049004E002D00360050004C0059004E003500350045004400420032002E0034004B00320049002E004C004F00430041004C000500140034004B00320049002E004C004F00430041004C000800300030000000000000000100000000200000512F465482D41399998EA0D3D7E64F2D0C26B6CF64F7E9CDDA71B01B0574A47F0A001000000000000000000000000000000000000900280048005400540050002F00770069006E002D00360070006C0079006E003500350065006400620032000000000000000000", // Replace with the actual expected hash
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc // capture range variable
// 		t.Run(tc.testName, func(t *testing.T) {
// 			t.Parallel()
// 			ntlm, err := ntlmv2.NewNTLMv2(tc.domain, tc.username, tc.password, tc.serverChallenge, tc.clientChallenge)
// 			if err != nil {
// 				t.Fatalf("Expected no error, got %v", err)
// 			}

// 			hashcatString, err := ntlm.ToHashcatString()
// 			if err != nil {
// 				t.Fatalf("Expected no error, got %v", err)
// 			}

// 			if hashcatString != tc.expectedHash {
// 				t.Errorf("For domain %s, username %s: expected hash %s, got %s", tc.domain, tc.username, tc.expectedHash, hashcatString)
// 			}
// 		})
// 	}
// }
