package kerbcrypto

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// unhex decodes a whitespace-separated hex string (as printed in the RFCs).
func unhex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(strings.ReplaceAll(strings.Join(strings.Fields(s), ""), " ", ""))
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}

// withConfounder runs fn with randRead patched to emit conf, restoring it after.
func withConfounder(conf []byte, fn func()) {
	orig := randRead
	randRead = func(buf []byte) (int, error) {
		copy(buf, conf)
		return len(buf), nil
	}
	defer func() { randRead = orig }()
	fn()
}

// RFC 8009 Appendix A base-keys for key usage 2.
var (
	rfc8009Base128 = "37 05 D9 60 80 C1 77 28 A0 E8 00 EA B6 E0 D2 3C"
	rfc8009Base256 = "6D 40 4D 37 FA F7 9F 9D F0 D3 35 68 D3 20 66 98 00 EB 48 36 47 2E A8 A0 26 D1 6B 71 82 46 0C 52"
)

// TestKDFHMACSHA2Vectors validates KDF-HMAC-SHA2 against RFC 8009 Appendix A
// (Kc/Ke/Ki derivation for key usage 2).
func TestKDFHMACSHA2Vectors(t *testing.T) {
	p128, _ := aes2ParamsFor(iana.ETypeAES128CTSHMACSHA256)
	p256, _ := aes2ParamsFor(iana.ETypeAES256CTSHMACSHA384)
	base128 := unhex(t, rfc8009Base128)
	base256 := unhex(t, rfc8009Base256)

	cases := []struct {
		name    string
		base    []byte
		p       aes2Params
		purpose byte
		outBits int
		want    string
	}{
		{"aes128 Kc", base128, p128, 0x99, 128, "B3 1A 01 8A 48 F5 47 76 F4 03 E9 A3 96 32 5D C3"},
		{"aes128 Ke", base128, p128, 0xAA, 128, "9B 19 7D D1 E8 C5 60 9D 6E 67 C3 E3 7C 62 C7 2E"},
		{"aes128 Ki", base128, p128, 0x55, 128, "9F DA 0E 56 AB 2D 85 E1 56 9A 68 86 96 C2 6A 6C"},
		{"aes256 Kc", base256, p256, 0x99, 192, "EF 57 18 BE 86 CC 84 96 3D 8B BB 50 31 E9 F5 C4 BA 41 F2 8F AF 69 E7 3D"},
		{"aes256 Ke", base256, p256, 0xAA, 256, "56 AB 22 BE E6 3D 82 D7 BC 52 27 F6 77 3F 8E A7 A5 EB 1C 82 51 60 C3 83 12 98 0C 44 2E 5C 7E 49"},
		{"aes256 Ki", base256, p256, 0x55, 192, "69 B1 65 14 E3 CD 8E 56 B8 20 10 D5 C7 30 12 B6 22 C4 D0 0F FC 23 ED 1F"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := kdfHMACSHA2(c.p.newHash, c.base, usageConstant(2, c.purpose), c.outBits)
			if want := unhex(t, c.want); !bytes.Equal(got, want) {
				t.Errorf("got %X, want %X", got, want)
			}
		})
	}
}

// TestAES2StringToKeyVectors validates the RFC 8009 Appendix A string-to-key
// base-key derivation. The RFC's saltp embeds a 16-byte sequence before the
// principal salt "ATHENA.MIT.EDUraeburn".
func TestAES2StringToKeyVectors(t *testing.T) {
	rnd := unhex(t, "10 DF 9D D7 83 E5 BC 8A CE A1 73 0E 74 35 5F 61")
	salt := string(rnd) + "ATHENA.MIT.EDUraeburn"

	p128, _ := aes2ParamsFor(iana.ETypeAES128CTSHMACSHA256)
	got128, err := aes2StringToKey("password", salt, 32768, p128)
	if err != nil {
		t.Fatal(err)
	}
	if want := unhex(t, "08 9B CA 48 B1 05 EA 6E A7 7C A5 D2 F3 9D C5 E7"); !bytes.Equal(got128, want) {
		t.Errorf("aes128 base-key: got %X, want %X", got128, want)
	}

	p256, _ := aes2ParamsFor(iana.ETypeAES256CTSHMACSHA384)
	got256, err := aes2StringToKey("password", salt, 32768, p256)
	if err != nil {
		t.Fatal(err)
	}
	if want := unhex(t, "45 BD 80 6D BF 6A 83 3A 9C FF C1 C9 45 89 A2 22 36 7A 79 BC 21 C4 13 71 89 06 E9 F5 78 A7 84 67"); !bytes.Equal(got256, want) {
		t.Errorf("aes256 base-key: got %X, want %X", got256, want)
	}
}

// aes2EncCase is one RFC 8009 Appendix A sample encryption.
type aes2EncCase struct {
	name       string
	conf       string
	plaintext  string
	ciphertext string
}

// TestAES2EncryptDecryptVectors validates encryption (with the RFC's fixed
// confounders) and decryption round-trips against RFC 8009 Appendix A.
func TestAES2EncryptDecryptVectors(t *testing.T) {
	tests := []struct {
		etype int
		base  string
		cases []aes2EncCase
	}{
		{
			etype: iana.ETypeAES128CTSHMACSHA256,
			base:  rfc8009Base128,
			cases: []aes2EncCase{
				{"empty", "7E 58 95 EA F2 67 24 35 BA D8 17 F5 45 A3 71 48", "",
					"EF 85 FB 89 0B B8 47 2F 4D AB 20 39 4D CA 78 1D AD 87 7E DA 39 D5 0C 87 0C 0D 5A 0A 8E 48 C7 18"},
				{"lt-block", "7B CA 28 5E 2F D4 13 0F B5 5B 1A 5C 83 BC 5B 24", "00 01 02 03 04 05",
					"84 D7 F3 07 54 ED 98 7B AB 0B F3 50 6B EB 09 CF B5 54 02 CE F7 E6 87 7C E9 9E 24 7E 52 D1 6E D4 42 1D FD F8 97 6C"},
				{"eq-block", "56 AB 21 71 3F F6 2C 0A 14 57 20 0F 6F A9 94 8F", "00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F",
					"35 17 D6 40 F5 0D DC 8A D3 62 87 22 B3 56 9D 2A E0 74 93 FA 82 63 25 40 80 EA 65 C1 00 8E 8F C2 95 FB 48 52 E7 D8 3E 1E 7C 48 C3 7E EB E6 B0 D3"},
				{"gt-block", "A7 A4 E2 9A 47 28 CE 10 66 4F B6 4E 49 AD 3F AC", "00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F 10 11 12 13 14",
					"72 0F 73 B1 8D 98 59 CD 6C CB 43 46 11 5C D3 36 C7 0F 58 ED C0 C4 43 7C 55 73 54 4C 31 C8 13 BC E1 E6 D0 72 C1 86 B3 9A 41 3C 2F 92 CA 9B 83 34 A2 87 FF CB FC"},
			},
		},
		{
			etype: iana.ETypeAES256CTSHMACSHA384,
			base:  rfc8009Base256,
			cases: []aes2EncCase{
				{"empty", "F7 64 E9 FA 15 C2 76 47 8B 2C 7D 0C 4E 5F 58 E4", "",
					"41 F5 3F A5 BF E7 02 6D 91 FA F9 BE 95 91 95 A0 58 70 72 73 A9 6A 40 F0 A0 19 60 62 1A C6 12 74 8B 9B BF BE 7E B4 CE 3C"},
				{"lt-block", "B8 0D 32 51 C1 F6 47 14 94 25 6F FE 71 2D 0B 9A", "00 01 02 03 04 05",
					"4E D7 B3 7C 2B CA C8 F7 4F 23 C1 CF 07 E6 2B C7 B7 5F B3 F6 37 B9 F5 59 C7 F6 64 F6 9E AB 7B 60 92 23 75 26 EA 0D 1F 61 CB 20 D6 9D 10 F2"},
				{"eq-block", "53 BF 8A 0D 10 52 65 D4 E2 76 42 86 24 CE 5E 63", "00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F",
					"BC 47 FF EC 79 98 EB 91 E8 11 5C F8 D1 9D AC 4B BB E2 E1 63 E8 7D D3 7F 49 BE CA 92 02 77 64 F6 8C F5 1F 14 D7 98 C2 27 3F 35 DF 57 4D 1F 93 2E 40 C4 FF 25 5B 36 A2 66"},
				{"gt-block", "76 3E 65 36 7E 86 4F 02 F5 51 53 C7 E3 B5 8A F1", "00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F 10 11 12 13 14",
					"40 01 3E 2D F5 8E 87 51 95 7D 28 78 BC D2 D6 FE 10 1C CF D5 56 CB 1E AE 79 DB 3C 3E E8 64 29 F2 B2 A6 02 AC 86 FE F6 EC B6 47 D6 29 5F AE 07 7A 1F EB 51 75 08 D2 C1 6B 41 92 E0 1F 62"},
			},
		},
	}

	for _, tt := range tests {
		base := unhex(t, tt.base)
		for _, c := range tt.cases {
			conf := unhex(t, c.conf)
			pt := unhex(t, c.plaintext)
			wantCT := unhex(t, c.ciphertext)

			t.Run(nameFor(tt.etype, c.name, "enc"), func(t *testing.T) {
				var got []byte
				var err error
				withConfounder(conf, func() { got, err = Encrypt(tt.etype, base, 2, pt) })
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, wantCT) {
					t.Errorf("ciphertext:\n got  %X\n want %X", got, wantCT)
				}
			})

			t.Run(nameFor(tt.etype, c.name, "dec"), func(t *testing.T) {
				got, err := Decrypt(tt.etype, base, 2, wantCT)
				if err != nil {
					t.Fatalf("decrypt: %v", err)
				}
				if !bytes.Equal(got, pt) {
					t.Errorf("plaintext: got %X, want %X", got, pt)
				}
			})
		}
	}
}

// TestAES2DecryptTamper confirms a flipped MAC byte is rejected.
func TestAES2DecryptTamper(t *testing.T) {
	base := unhex(t, rfc8009Base128)
	ct := unhex(t, "EF 85 FB 89 0B B8 47 2F 4D AB 20 39 4D CA 78 1D AD 87 7E DA 39 D5 0C 87 0C 0D 5A 0A 8E 48 C7 18")
	ct[len(ct)-1] ^= 0x01
	if _, err := Decrypt(iana.ETypeAES128CTSHMACSHA256, base, 2, ct); err != ErrIntegrityCheckFailed {
		t.Fatalf("want ErrIntegrityCheckFailed, got %v", err)
	}
}

// TestAES2ChecksumVectors validates get_mic against RFC 8009 Appendix A.
func TestAES2ChecksumVectors(t *testing.T) {
	msg := unhex(t, "00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F 10 11 12 13 14")

	got128, err := GetChecksum(iana.CksumTypeHMACSHA256128AES128, unhex(t, rfc8009Base128), 2, msg)
	if err != nil {
		t.Fatal(err)
	}
	if want := unhex(t, "D7 83 67 18 66 43 D6 7B 41 1C BA 91 39 FC 1D EE"); !bytes.Equal(got128, want) {
		t.Errorf("hmac-sha256-128-aes128: got %X, want %X", got128, want)
	}

	got256, err := GetChecksum(iana.CksumTypeHMACSHA384192AES256, unhex(t, rfc8009Base256), 2, msg)
	if err != nil {
		t.Fatal(err)
	}
	if want := unhex(t, "45 EE 79 15 67 EE FC A3 7F 4A C1 E0 22 2D E8 0D 43 C3 BF A0 66 99 67 2A"); !bytes.Equal(got256, want) {
		t.Errorf("hmac-sha384-192-aes256: got %X, want %X", got256, want)
	}

	if !VerifyChecksum(iana.CksumTypeHMACSHA256128AES128, unhex(t, rfc8009Base128), 2, msg, got128) {
		t.Error("VerifyChecksum returned false for a valid checksum")
	}
}

func nameFor(etype int, c, op string) string {
	e := "aes128sha256"
	if etype == iana.ETypeAES256CTSHMACSHA384 {
		e = "aes256sha384"
	}
	return e + "/" + c + "/" + op
}
