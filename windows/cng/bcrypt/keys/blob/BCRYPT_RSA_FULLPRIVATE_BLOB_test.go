package blob_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"math/big"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/blob"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/headers"
)

func TestBCRYPT_RSA_FULLPRIVATE_BLOB_MarshalUnmarshal(t *testing.T) {
	// Each field is filled with a distinct byte so a misplaced offset is detected.
	fill := func(n int, v byte) []byte { return bytes.Repeat([]byte{v}, n) }

	header := headers.BCRYPT_RSA_KEY_BLOB{
		BitLength:   64,
		CbPublicExp: 3,
		CbModulus:   8,
		CbPrime1:    4,
		CbPrime2:    4,
	}
	input := blob.BCRYPT_RSA_FULLPRIVATE_BLOB{
		PublicExponent:  []byte{0x01, 0x00, 0x01},
		Modulus:         fill(8, 0x11),
		Prime1:          fill(4, 0x22),
		Prime2:          fill(4, 0x33),
		Exponent1:       fill(4, 0x44),
		Exponent2:       fill(4, 0x55),
		Coefficient:     fill(4, 0x66),
		PrivateExponent: fill(8, 0x77),
	}

	data, err := input.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if len(data) != 39 {
		t.Fatalf("Marshal() length = %d, want 39", len(data))
	}

	rk := &blob.BCRYPT_RSA_FULLPRIVATE_BLOB{}
	n, err := rk.Unmarshal(header, data)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if n != len(data) {
		t.Errorf("Unmarshal() bytesRead = %d, want %d", n, len(data))
	}
	if !rk.Equal(&input) {
		t.Errorf("Unmarshal() = %+v, want %+v", rk, input)
	}

	// Truncating the private exponent must be reported as an error.
	if _, err := (&blob.BCRYPT_RSA_FULLPRIVATE_BLOB{}).Unmarshal(header, data[:len(data)-1]); err == nil {
		t.Errorf("Unmarshal() on truncated blob: expected error, got nil")
	}
}

func TestBCRYPT_RSA_FULLPRIVATE_BLOB_FromRSAPrivateKey(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() error = %v", err)
	}
	key.Precompute()

	cbModulus := (key.N.BitLen() + 7) / 8
	cbPrime1 := (key.Primes[0].BitLen() + 7) / 8
	cbPrime2 := (key.Primes[1].BitLen() + 7) / 8
	pad := func(v *big.Int, n int) []byte { return v.FillBytes(make([]byte, n)) }
	pubExp := big.NewInt(int64(key.E)).Bytes()

	header := headers.BCRYPT_RSA_KEY_BLOB{
		BitLength:   uint32(key.N.BitLen()),
		CbPublicExp: uint32(len(pubExp)),
		CbModulus:   uint32(cbModulus),
		CbPrime1:    uint32(cbPrime1),
		CbPrime2:    uint32(cbPrime2),
	}

	// Raw body laid out as BCryptExportKey(BCRYPT_RSAFULLPRIVATE_BLOB) does after the header.
	var raw []byte
	raw = append(raw, pubExp...)
	raw = append(raw, pad(key.N, cbModulus)...)
	raw = append(raw, pad(key.Primes[0], cbPrime1)...)
	raw = append(raw, pad(key.Primes[1], cbPrime2)...)
	raw = append(raw, pad(key.Precomputed.Dp, cbPrime1)...)
	raw = append(raw, pad(key.Precomputed.Dq, cbPrime2)...)
	raw = append(raw, pad(key.Precomputed.Qinv, cbPrime1)...)
	raw = append(raw, pad(key.D, cbModulus)...)

	rk := &blob.BCRYPT_RSA_FULLPRIVATE_BLOB{}
	n, err := rk.Unmarshal(header, raw)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if n != len(raw) {
		t.Errorf("Unmarshal() bytesRead = %d, want %d", n, len(raw))
	}

	checks := []struct {
		name string
		got  []byte
		want *big.Int
	}{
		{"Modulus", rk.Modulus, key.N},
		{"Prime1", rk.Prime1, key.Primes[0]},
		{"Prime2", rk.Prime2, key.Primes[1]},
		{"Exponent1", rk.Exponent1, key.Precomputed.Dp},
		{"Exponent2", rk.Exponent2, key.Precomputed.Dq},
		{"Coefficient", rk.Coefficient, key.Precomputed.Qinv},
		{"PrivateExponent", rk.PrivateExponent, key.D},
	}
	for _, c := range checks {
		if new(big.Int).SetBytes(c.got).Cmp(c.want) != 0 {
			t.Errorf("%s mismatch after unmarshal", c.name)
		}
	}

	out, err := rk.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if !bytes.Equal(out, raw) {
		t.Errorf("Marshal() did not reproduce the original blob")
	}
}
