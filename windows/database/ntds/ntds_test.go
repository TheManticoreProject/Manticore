package ntds

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// The vectors below were produced independently in Python (impacket's transformKey +
// pycryptodome AES/DES/RC4/MD5), so a matching Go result cross-validates the whole
// decryption chain against impacket's NTDSHashes behaviour. bootKey, PEK, RID and the
// expected hashes are fixed; see the generator in the PR description.
const (
	vecBootKey    = "000102030405060708090a0b0c0d0e0f"
	vecPEK        = "11111111111111111111111111111111"
	vecRID        = 500
	vecNT         = "31d6cfe0d16ae931b73c59d7e0c089c0" // empty password
	vecNT2        = "8846f7eaee8fb117ad06bdd830b7586c" // "password"
	vecPekListRC4 = "0200000000000000aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa01eb46cbb54c926d566dc36e44e20d3946a3ad276548def414bf2828cd1559a6f5c526b2917ec2e28bc891bb52d9c16b467b20ae"
	vecPekListAES = "0300000000000000bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbc39afd91d9c08639d2b9ba5164af295fc0c882f4b49f03f76c7a17779d6da072c4e8260f08ecbe541e01bccd19663ef79b246a11b831e98ec6d292f1172b837d"
	vecHashRC4    = "0000000000000000cccccccccccccccccccccccccccccccc79844a39148e7b88768caac1f96caba6"
	vecHashAES    = "1300000000000000dddddddddddddddddddddddddddddddd00000000d311e3bdbf0024e051a53759f80e2b71"
	vecHistRC4    = "0000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee21ce6082c5db0c5d0a2d5ec09d7b8478be753ac8ee36b9c66a741f872ca6891e"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}

func TestDecryptPEKListRC4(t *testing.T) {
	peks, err := DecryptPEKList(mustHex(t, vecBootKey), mustHex(t, vecPekListRC4))
	if err != nil {
		t.Fatalf("DecryptPEKList: %v", err)
	}
	if len(peks) != 1 {
		t.Fatalf("got %d PEKs, want 1", len(peks))
	}
	if hex.EncodeToString(peks[0]) != vecPEK {
		t.Errorf("PEK = %x, want %s", peks[0], vecPEK)
	}
}

func TestDecryptPEKListAES(t *testing.T) {
	peks, err := DecryptPEKList(mustHex(t, vecBootKey), mustHex(t, vecPekListAES))
	if err != nil {
		t.Fatalf("DecryptPEKList: %v", err)
	}
	if len(peks) != 1 || hex.EncodeToString(peks[0]) != vecPEK {
		t.Fatalf("PEKs = %x, want one PEK %s", peks, vecPEK)
	}
}

func TestDecryptHashRC4AndAES(t *testing.T) {
	peks := []PEK{mustHex(t, vecPEK)}
	for _, tc := range []struct {
		name, blob string
	}{
		{"RC4", vecHashRC4},
		{"AES", vecHashAES},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecryptHash(peks, vecRID, mustHex(t, tc.blob))
			if err != nil {
				t.Fatalf("DecryptHash: %v", err)
			}
			if hex.EncodeToString(got) != vecNT {
				t.Errorf("hash = %x, want %s", got, vecNT)
			}
		})
	}
}

func TestDecryptHashHistory(t *testing.T) {
	peks := []PEK{mustHex(t, vecPEK)}
	hist, err := DecryptHashHistory(peks, vecRID, mustHex(t, vecHistRC4))
	if err != nil {
		t.Fatalf("DecryptHashHistory: %v", err)
	}
	if len(hist) != 2 {
		t.Fatalf("got %d history entries, want 2", len(hist))
	}
	if hex.EncodeToString(hist[0]) != vecNT || hex.EncodeToString(hist[1]) != vecNT2 {
		t.Errorf("history = [%x %x], want [%s %s]", hist[0], hist[1], vecNT, vecNT2)
	}
}

func TestDeriveDESKeysVector(t *testing.T) {
	// Cross-checked against impacket transformKey for RID 500.
	k1, k2 := deriveDESKeys(500)
	if !bytes.Equal(k1, mustHex(t, "f40040000ea00400")) || !bytes.Equal(k2, mustHex(t, "007a00200006d002")) {
		t.Errorf("RID 500 DES keys = %x / %x, want f40040000ea00400 / 007a00200006d002", k1, k2)
	}
}

func TestDecryptErrors(t *testing.T) {
	if _, err := DecryptPEKList(make([]byte, 8), make([]byte, 64)); err == nil {
		t.Error("short boot key accepted")
	}
	if _, err := DecryptPEKList(make([]byte, 16), []byte{0x99, 0, 0, 0}); err == nil {
		t.Error("unknown pekList format / short list accepted")
	}
	if _, err := DecryptHash([]PEK{mustHex(t, vecPEK)}, 1, make([]byte, 4)); err == nil {
		t.Error("short blob accepted")
	}
}
