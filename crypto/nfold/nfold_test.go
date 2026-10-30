package nfold

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// TestNFoldRFC3961Vectors tests all vectors from RFC 3961 and known reference implementations.
// Vectors verified against jcmturner/gokrb5 and jfjallid/gokrb5 (both canonical RFC 3961 impls).
func TestNFoldRFC3961Vectors(t *testing.T) {
	tests := []struct {
		input    []byte
		n        int
		expected string
	}{
		// nfold(64, "kerberos") = "kerberos" — identity case (lcm(64,64)=64, one block)
		{[]byte("kerberos"), 64, "6b65726265726f73"},
		// nfold(128, "kerberos") — RFC 3961 reference vector
		{[]byte("kerberos"), 128, "6b65726265726f737b9b5b2b93132b93"},
		// nfold(168, "kerberos") — RFC 3961 reference vector
		{[]byte("kerberos"), 168, "8372c236344e5f1550cd0747e15d62ca7a5a3bcea4"},
		// nfold(56, "password") — RFC 3961 reference vector
		{[]byte("password"), 56, "78a07b6caf85fa"},
		// nfold(168, "password") — RFC 3961 reference vector
		{[]byte("password"), 168, "59e4a8ca7c0385c3c37b3f6d2000247cb6e6bd5b3e"},
		// nfold(64, "Rough Consensus, and Running Code") — RFC 3961 reference vector
		{[]byte("Rough Consensus, and Running Code"), 64, "bb6ed30870b7f0e0"},
		// nfold(192, "MASSACHVSETTS INSTITVTE OF TECHNOLOGY") — RFC 3961 reference vector
		{[]byte("MASSACHVSETTS INSTITVTE OF TECHNOLOGY"), 192, "db3b0d8f0b061e603282b308a50841229ad798fab9540c1b"},
		// nfold(168, "Q") — RFC 3961 reference vector
		{[]byte("Q"), 168, "518a54a215a8452a518a54a215a8452a518a54a215"},
		// nfold(168, "ba") — RFC 3961 reference vector
		{[]byte("ba"), 168, "fb25d531ae8974499f52fd92ea9857c4ba24cf297e"},
		// nfold(64, "012345") — RFC 3961 reference vector
		{[]byte("012345"), 64, "be072631276b1955"},
	}

	for _, tt := range tests {
		result := NFold(tt.input, tt.n)
		if len(result) != tt.n/8 {
			t.Errorf("NFold(%q, %d) length = %d, want %d", tt.input, tt.n, len(result), tt.n/8)
			continue
		}
		got := hex.EncodeToString(result)
		if got != tt.expected {
			t.Errorf("NFold(%q, %d) = %s, want %s", tt.input, tt.n, got, tt.expected)
		}
	}
}

// TestNFoldDeterministic verifies that NFold produces deterministic output.
func TestNFoldDeterministic(t *testing.T) {
	input := []byte("test-input-data")
	r1 := NFold(input, 64)
	r2 := NFold(input, 64)

	if hex.EncodeToString(r1) != hex.EncodeToString(r2) {
		t.Error("NFold is not deterministic")
	}
}

// Folding an empty input or to a non-positive width must not divide by zero.
func TestNFoldDegenerateInputs(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		n    int
		want []byte
	}{
		{"empty input to 64 bits", []byte{}, 64, make([]byte, 8)},
		{"nil input to 128 bits", nil, 128, make([]byte, 16)},
		{"zero width", []byte("kerberos"), 0, []byte{}},
		{"negative width", []byte("kerberos"), -8, []byte{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NFold(tt.in, tt.n); !bytes.Equal(got, tt.want) {
				t.Errorf("NFold(%x, %d) = %x, want %x", tt.in, tt.n, got, tt.want)
			}
		})
	}
}
