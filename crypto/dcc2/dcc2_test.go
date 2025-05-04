package dcc2

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
)

func TestDCC2Hash(t *testing.T) {
	testCases := []struct {
		username string
		password string
		rounds   int
		expected string
	}{
		{
			username: "podalirius",
			password: "TheManticoreProject",
			rounds:   10240,
			expected: "$DCC2$10240#podalirius#3c66bcdbefb145a04296693e42e79a91",
		},
		{
			username: "pOdAlIrIuS",
			password: "TheManticoreProject",
			rounds:   10240,
			expected: "$DCC2$10240#pOdAlIrIuS#3c66bcdbefb145a04296693e42e79a91",
		},
		{
			username: "",
			password: "TheManticoreProject",
			rounds:   10240,
			expected: "$DCC2$10240##d80d4f235b96f66e98b87cb003c649e3",
		},
		{
			username: "podalirius",
			password: "",
			rounds:   10240,
			expected: "$DCC2$10240#podalirius#476e061d8cb936fa9e86c5e78feb32b3",
		},
		{
			username: "podalirius",
			password: "TheManticoreProject",
			rounds:   1,
			expected: "$DCC2$1#podalirius#1fd0baf181cfb95bb9cbbf88514efb74",
		},
		{
			username: "podalirius",
			password: "TheManticoreProject",
			rounds:   2,
			expected: "$DCC2$2#podalirius#5ba8ea5dc77bb8c761e861f6f4d9805e",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.username, func(t *testing.T) {
			result := DCC2Hash(tc.username, tc.password, tc.rounds)
			if result != tc.expected {
				t.Errorf("DCC2Hash(%s, %s, %d) = %s; expected %s", tc.username, tc.password, tc.rounds, result, tc.expected)
			}
		})
	}
}

func TestDCC2HashWithPassword(t *testing.T) {
	testCases := []struct {
		username string
		password string
		rounds   int
		expected string
	}{
		{
			username: "podalirius",
			password: "TheManticoreProject",
			rounds:   10240,
			expected: "$DCC2$10240#podalirius#3c66bcdbefb145a04296693e42e79a91",
		},
		{
			username: "pOdAlIrIuS",
			password: "TheManticoreProject",
			rounds:   10240,
			expected: "$DCC2$10240#pOdAlIrIuS#3c66bcdbefb145a04296693e42e79a91",
		},
		{
			username: "",
			password: "TheManticoreProject",
			rounds:   10240,
			expected: "$DCC2$10240##d80d4f235b96f66e98b87cb003c649e3",
		},
		{
			username: "podalirius",
			password: "",
			rounds:   10240,
			expected: "$DCC2$10240#podalirius#476e061d8cb936fa9e86c5e78feb32b3",
		},
		{
			username: "podalirius",
			password: "TheManticoreProject",
			rounds:   1,
			expected: "$DCC2$1#podalirius#1fd0baf181cfb95bb9cbbf88514efb74",
		},
		{
			username: "podalirius",
			password: "TheManticoreProject",
			rounds:   2,
			expected: "$DCC2$2#podalirius#5ba8ea5dc77bb8c761e861f6f4d9805e",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.username, func(t *testing.T) {
			result := DCC2HashWithPassword(tc.username, tc.password, tc.rounds)
			if result != tc.expected {
				t.Errorf("DCC2HashWithPassword(%s, %s, %d) = %s; expected %s", tc.username, tc.password, tc.rounds, result, tc.expected)
			}
		})
	}
}

func TestDCC2HashWithNTHash(t *testing.T) {
	testCases := []struct {
		username string
		ntHash   [16]byte
		rounds   int
		expected string
	}{
		{
			username: "podalirius",
			ntHash:   nt.NTHash("TheManticoreProject"),
			rounds:   10240,
			expected: "$DCC2$10240#podalirius#3c66bcdbefb145a04296693e42e79a91",
		},
		{
			username: "pOdAlIrIuS",
			ntHash:   nt.NTHash("TheManticoreProject"),
			rounds:   10240,
			expected: "$DCC2$10240#pOdAlIrIuS#3c66bcdbefb145a04296693e42e79a91",
		},
		{
			username: "",
			ntHash:   nt.NTHash("TheManticoreProject"),
			rounds:   10240,
			expected: "$DCC2$10240##d80d4f235b96f66e98b87cb003c649e3",
		},
		{
			username: "podalirius",
			ntHash:   nt.NTHash(""),
			rounds:   10240,
			expected: "$DCC2$10240#podalirius#476e061d8cb936fa9e86c5e78feb32b3",
		},
		{
			username: "podalirius",
			ntHash:   nt.NTHash("TheManticoreProject"),
			rounds:   1,
			expected: "$DCC2$1#podalirius#1fd0baf181cfb95bb9cbbf88514efb74",
		},
		{
			username: "podalirius",
			ntHash:   nt.NTHash("TheManticoreProject"),
			rounds:   2,
			expected: "$DCC2$2#podalirius#5ba8ea5dc77bb8c761e861f6f4d9805e",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.username, func(t *testing.T) {
			result := DCC2HashWithNTHash(tc.username, tc.ntHash, tc.rounds)
			if result != tc.expected {
				t.Errorf("DCC2HashWithNTHash(%s, %s, %d) = %s; expected %s", tc.username, tc.ntHash, tc.rounds, result, tc.expected)
			}
		})
	}
}
