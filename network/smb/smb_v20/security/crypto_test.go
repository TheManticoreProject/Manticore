package security

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

func TestGMACNonce(t *testing.T) {
	m := message.NewMessage()
	m.Header.MessageId = 2
	m.SetCommand(commands.NewTreeDisconnectRequest())
	request, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	for _, tc := range []struct {
		name string
		wire []byte
		want string
	}{
		{"client request", request, "020000000000000000000000"},
		{"server response", func() []byte {
			wire := bytes.Clone(request)
			wire[signFlagsOffset] |= 1 // SMB2_FLAGS_SERVER_TO_REDIR
			return wire
		}(), "020000000000000001000000"},
		{"cancel request", func() []byte {
			wire := bytes.Clone(request)
			wire[12] = 0x0c // SMB2_CANCEL
			return wire
		}(), "020000000000000002000000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := gmacNonce(tc.wire), mustHex(t, tc.want); !bytes.Equal(got, want) {
				t.Errorf("gmacNonce = %X, want %X", got, want)
			}
		})
	}
}
func TestSMB3CMACSignAndVerify(t *testing.T) {
	key := mustHex(t, "0B7E9C5CAC36C0F6EA9AB275298CEDCE")

	m := message.NewMessage()
	m.Header.MessageId = 7
	m.SetCommand(commands.NewTreeDisconnectRequest())
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	Sign(dialects.SMB2_DIALECT_3_1_1, -1, key, wire)

	decoded := message.NewMessage()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !decoded.Header.Flags.IsSigned() {
		t.Errorf("SMB2_FLAGS_SIGNED not set after CMAC signing")
	}

	if !Verify(dialects.SMB2_DIALECT_3_1_1, -1, key, wire) {
		t.Errorf("CMAC signature failed to verify with the correct key")
	}
	if Verify(dialects.SMB2_DIALECT_3_1_1, -1, mustHex(t, "00000000000000000000000000000000"), wire) {
		t.Errorf("CMAC signature verified with a wrong key")
	}
	wire[len(wire)-1] ^= 0xFF
	if Verify(dialects.SMB2_DIALECT_3_1_1, -1, key, wire) {
		t.Errorf("CMAC signature verified for a tampered message")
	}
}
func TestGMACSignAndVerify(t *testing.T) {
	key := mustHex(t, "0B7E9C5CAC36C0F6EA9AB275298CEDCE")

	m := message.NewMessage()
	m.Header.MessageId = 42
	m.SetCommand(commands.NewTreeDisconnectRequest())
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	Sign(dialects.SMB2_DIALECT_3_1_1, commands.SMB2_SIGNING_ALG_AES_GMAC, key, wire)

	decoded := message.NewMessage()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !decoded.Header.Flags.IsSigned() {
		t.Errorf("SMB2_FLAGS_SIGNED not set after GMAC signing")
	}

	if !Verify(dialects.SMB2_DIALECT_3_1_1, commands.SMB2_SIGNING_ALG_AES_GMAC, key, wire) {
		t.Errorf("GMAC signature failed to verify with the correct key")
	}
	wrongKey := make([]byte, 16)
	if Verify(dialects.SMB2_DIALECT_3_1_1, commands.SMB2_SIGNING_ALG_AES_GMAC, wrongKey, wire) {
		t.Errorf("GMAC signature verified with a wrong key")
	}
	wire[len(wire)-1] ^= 0xFF
	if Verify(dialects.SMB2_DIALECT_3_1_1, commands.SMB2_SIGNING_ALG_AES_GMAC, key, wire) {
		t.Errorf("GMAC signature verified for a tampered message")
	}
}
func TestGMACSignatureNotCMAC(t *testing.T) {
	key16 := mustHex(t, "0B7E9C5CAC36C0F6EA9AB275298CEDCE")

	buildMessage := func() []byte {
		m := message.NewMessage()
		m.Header.MessageId = 7
		m.SetCommand(commands.NewTreeDisconnectRequest())
		wire, _ := m.Marshal()
		return wire
	}

	cmacWire := buildMessage()
	signMessageCMAC(key16, cmacWire)
	cmacSig := make([]byte, 16)
	copy(cmacSig, cmacWire[signSignatureOffset:signSignatureOffset+signSignatureLength])

	gmacWire := buildMessage()
	signMessageGMAC(key16, gmacWire)
	gmacSig := make([]byte, 16)
	copy(gmacSig, gmacWire[signSignatureOffset:signSignatureOffset+signSignatureLength])

	if bytes.Equal(cmacSig, gmacSig) {
		t.Errorf("GMAC and CMAC produced the same signature")
	}
}
func TestCCMKnownAnswerRFC3610(t *testing.T) {
	key := mustHex(t, "C0C1C2C3C4C5C6C7C8C9CACBCCCDCECF")
	nonce := mustHex(t, "00000003020100A0A1A2A3A4A5") // 13 bytes -> L=2
	aad := mustHex(t, "0001020304050607")
	plaintext := mustHex(t, "08090A0B0C0D0E0F101112131415161718191A1B1C1D1E")
	// Expected ciphertext || 8-byte MAC from RFC 3610 packet vector #1.
	wantCT := mustHex(t, "588C979A61C663D2F066D0C2C0F989806D5F6B61DAC384")
	wantMAC := mustHex(t, "17E8D12CFDF926E0")

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("aes: %v", err)
	}

	// The RFC 3610 vector uses an 8-byte tag; exercise the generic core through a
	// dedicated tag length rather than the SMB-fixed 16.
	tag := ccmTag(block, nonce, plaintext, aad, 15-len(nonce), len(wantMAC))
	var counter0 [16]byte
	ccmCounterBlock(counter0[:], nonce, 15-len(nonce), 0)
	var s0 [16]byte
	block.Encrypt(s0[:], counter0[:])
	gotMAC := make([]byte, 8)
	for i := range gotMAC {
		gotMAC[i] = tag[i] ^ s0[i]
	}
	if !bytes.Equal(gotMAC, wantMAC) {
		t.Errorf("CCM MAC = %X, want %X", gotMAC, wantMAC)
	}

	ct := make([]byte, len(plaintext))
	ccmCTR(block, nonce, 15-len(nonce), plaintext, ct)
	if !bytes.Equal(ct, wantCT) {
		t.Errorf("CCM ciphertext = %X, want %X", ct, wantCT)
	}
}
func TestCCMSealOpenRoundTrip(t *testing.T) {
	key := mustHex(t, "00112233445566778899AABBCCDDEEFF")
	block, _ := aes.NewCipher(key)
	nonce := mustHex(t, "0102030405060708090A0B") // 11 bytes
	aad := []byte("transform-header-aad-32bytes!!!!")
	plaintext := []byte("the quick brown fox jumps over the lazy SMB dog")

	sealed, err := ccmSeal(block, nonce, plaintext, aad)
	if err != nil {
		t.Fatalf("ccmSeal: %v", err)
	}
	if len(sealed) != len(plaintext)+ccmTagSize {
		t.Fatalf("sealed length = %d, want %d", len(sealed), len(plaintext)+ccmTagSize)
	}
	got, err := ccmOpen(block, nonce, sealed, aad)
	if err != nil {
		t.Fatalf("ccmOpen: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("ccmOpen = %q, want %q", got, plaintext)
	}
	sealed[0] ^= 0xFF
	if _, err := ccmOpen(block, nonce, sealed, aad); err == nil {
		t.Errorf("ccmOpen accepted a tampered ciphertext")
	}
}
func TestPreauthHashMatchesPublishedVector(t *testing.T) {
	negReq := mustHex(t, ""+
		"FE534D4240000100000000000000800000000000000000000100000000000000FFFE000000000000"+
		"00000000000000000000000000000000000000000000000024000500000000003F000000ECD86F32"+
		"6276024F9F7752B89BB33F3A70000000020000000202100200030203110300000100260000000000"+
		"010020000100FA49E6578F1F3A9F4CD3E9CC14A67AA884B3D05844E0E5A118225C15887F32FF0000"+
		"0200060000000000020002000100")
	want := mustHex(t, ""+
		"DD94EFC5321BB618A2E208BA8920D2F422992526947A409B5037DE1E0FE8C736"+
		"2B8C47122594CDE0CE26AA9DFC8BCDBDE0621957672623351A7540F1E54A0426")

	got := PreauthUpdate(make([]byte, PreauthHashSize), negReq)
	if !bytes.Equal(got, want) {
		t.Errorf("pre-auth hash mismatch:\n got %s\nwant %s", hex.EncodeToString(got), hex.EncodeToString(want))
	}
}

// TestEncryptDecryptRoundTrip checks that a TRANSFORM_HEADER frame decrypts back
// to the original for every cipher and that tampering is detected.
func TestEncryptDecryptRoundTrip(t *testing.T) {
	key128 := mustHex(t, "629BCBC54422A0F572B97F45989B6073")
	key256 := mustHex(t, "629BCBC54422A0F572B97F45989B6073A1B2C3D4E5F60718293A4B5C6D7E8F90")
	plaintext := []byte("an SMB2 message long enough to span more than one AES block")

	for _, tc := range []struct {
		name   string
		cipher uint16
		key    []byte
	}{
		{"AES-128-CCM", commands.SMB2_ENCRYPTION_AES128_CCM, key128},
		{"AES-128-GCM", commands.SMB2_ENCRYPTION_AES128_GCM, key128},
		{"AES-256-CCM", commands.SMB2_ENCRYPTION_AES256_CCM, key256},
		{"AES-256-GCM", commands.SMB2_ENCRYPTION_AES256_GCM, key256},
	} {
		t.Run(tc.name, func(t *testing.T) {
			frame, err := Encrypt(tc.cipher, tc.key, 0x1234, 7, plaintext)
			if err != nil {
				t.Fatalf("Encrypt: %v", err)
			}
			if !IsTransformHeader(frame) || TransformSessionID(frame) != 0x1234 {
				t.Fatalf("frame does not carry the expected TRANSFORM_HEADER")
			}
			got, err := Decrypt(tc.cipher, tc.key, frame)
			if err != nil || !bytes.Equal(got, plaintext) {
				t.Fatalf("Decrypt = %q, %v", got, err)
			}
			frame[len(frame)-1] ^= 0xFF
			if _, err := Decrypt(tc.cipher, tc.key, frame); err == nil {
				t.Error("Decrypt accepted a tampered frame")
			}
		})
	}
}
