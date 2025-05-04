package ntlmv1

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestParityBit(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{0b00000000, 1}, // 0 has 0 bits set (even), so parity bit should be 1
		{0b00000001, 0}, // 1 has 1 bit set (odd), so parity bit should be 0
		{0b00000010, 0}, // 2 (10 in binary) has 1 bit set (odd), so parity bit should be 0
		{0b00000011, 1}, // 3 (11 in binary) has 2 bits set (even), so parity bit should be 1
		{0b00000111, 0}, // 7 (111 in binary) has 3 bits set (odd), so parity bit should be 0
		{0b00001111, 1}, // 15 (1111 in binary) has 4 bits set (even), so parity bit should be 1
		{0b01111111, 0}, // 127 (1111111 in binary) has 7 bits set (odd), so parity bit should be 0
		{0b10000000, 0}, // 128 (10000000 in binary) has 1 bit set (odd), so parity bit should be 0
		{0b11111111, 1}, // 255 (11111111 in binary) has 8 bits set (even), so parity bit should be 1
	}

	for _, test := range tests {
		result := ParityBit(test.input)
		if result != test.expected {
			t.Errorf("ParityBit(%d) = %d; expected %d", test.input, result, test.expected)
		}
	}
}

func TestParityAdjust(t *testing.T) {
	tests := []struct {
		input    []byte
		expected []byte
	}{
		{[]byte("0"), []byte("1")},
		{[]byte("01"), []byte("1\x19")},
		{[]byte("012"), []byte("1\x19L")},
		{[]byte("0123"), []byte("1\x19LF")},
		{[]byte("01234"), []byte("1\x19LF2")},
		{[]byte("012345"), []byte("1\x19LF2\xa1")},
		{[]byte("0123456"), []byte("1\x19LF2\xa1\xd5m")},
		{[]byte("01234567"), []byte("1\x19LF2\xa1\xd5m7")},
		{[]byte("012345678"), []byte("1\x19LF2\xa1\xd5m7\x9d")},
		{[]byte("0123456789"), []byte("1\x19LF2\xa1\xd5m7\x9d\x0e")},
		{[]byte("0123456789a"), []byte("1\x19LF2\xa1\xd5m7\x9d\x0e,")},
		{[]byte("0123456789ab"), []byte("1\x19LF2\xa1\xd5m7\x9d\x0e,\x16")},
		{[]byte("0123456789abc"), []byte("1\x19LF2\xa1\xd5m7\x9d\x0e,\x16\x13")},
		{[]byte("0123456789abcd"), []byte("1\x19LF2\xa1\xd5m7\x9d\x0e,\x16\x13\x8c\xc8")},
		{[]byte("0123456789abcde"), []byte("1\x19LF2\xa1\xd5m7\x9d\x0e,\x16\x13\x8c\xc8d")},
	}

	for _, test := range tests {
		result, err := ParityAdjust(test.input)
		if err != nil {
			t.Errorf("ParityAdjust(%s) returned error: %v", test.input, err)
		}
		if !bytes.Equal(result, test.expected) {
			t.Errorf("ParityAdjust(%s) = %s; expected %s", test.input, hex.EncodeToString(result), hex.EncodeToString(test.expected))
		}
	}
}

func TestNTLMv1HashFromPassword(t *testing.T) {
	serverChallenge := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}
	tests := []struct {
		domain    string
		username  string
		password  string
		challenge []byte
		expected  string
	}{
		{
			domain:    "WORKGROUP",
			username:  "podalirius",
			password:  "Podalirius!",
			challenge: serverChallenge,
			expected:  "8110779A47517B1E6BD686317BD8BC395A07A640AD9E3E70",
		},
	}

	for _, test := range tests {
		ntlm, err := NewNTLMv1WithPassword(test.domain, test.username, test.password, test.challenge)
		if err != nil {
			t.Errorf("NewNTLMv1WithPassword(%s, %s, %s, %s) returned error: %v", test.domain, test.username, test.password, test.challenge, err)
		}
		result := ntlm.String()
		if result != test.expected {
			t.Errorf("NTLMv1Hash(%s, %s) = %s; expected %s", test.password, test.challenge, result, test.expected)
		}
	}
}
