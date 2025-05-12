package ntlmv1

import (
	"bytes"
	"crypto/des"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/crypto/lm"
	"github.com/TheManticoreProject/Manticore/crypto/nt"
)

// NTLMv1 represents the components needed for NTLMv1 authentication
type NTLMv1 struct {
	Username        string
	Password        string
	Domain          string
	ServerChallenge []byte
	NTHash          []byte
}

// NewNTLMv1WithPassword creates a new NTLMv1 instance with the provided credentials and challenge
func NewNTLMv1WithPassword(domain, username, password string, serverChallenge []byte) (*NTLMv1, error) {
	if len(serverChallenge) != 8 {
		return nil, fmt.Errorf("server challenge must be 8 bytes")
	}

	ntHash := nt.NTHash(password)

	ntlm := &NTLMv1{
		Domain:          domain,
		Username:        username,
		Password:        password,
		ServerChallenge: serverChallenge,
		NTHash:          ntHash[:],
	}

	return ntlm, nil
}

// NewNTLMv1WithNTHash creates a new NTLMv1 instance with the provided credentials and challenge
func NewNTLMv1WithNTHash(domain, username string, nthash []byte, serverChallenge []byte) (*NTLMv1, error) {
	if len(serverChallenge) != 8 {
		return nil, fmt.Errorf("server challenge must be 8 bytes")
	}

	ntlm := &NTLMv1{
		Domain:          domain,
		Username:        username,
		Password:        "",
		ServerChallenge: serverChallenge,
		NTHash:          nthash,
	}

	return ntlm, nil
}

// NTLMv1Hash calculates the NTLMv1 hash
//
// Returns:
//   - The NTLMv1 hash as a byte slice
//   - An error if the NTHash or Password is not set
func (h *NTLMv1) Hash() ([]byte, error) {
	// Start with the NT hash of the password
	if len(h.NTHash) == 0 && len(h.Password) == 0 {
		return nil, fmt.Errorf("NTHash and Password are not set")
	}
	if len(h.NTHash) == 0 {
		ntHash := nt.NTHash(h.Password)
		h.NTHash = ntHash[:]
	}

	rawKeys := h.NTHash
	// Pad the hash with zeros to get 21 bytes (3 * 7)
	rawKeys = append(rawKeys, bytes.Repeat([]byte{0}, 21-len(rawKeys))...)

	// Compute block 1
	key1 := rawKeys[0:7]
	key1Adjusted, err := ParityAdjust(key1)
	if err != nil {
		return nil, fmt.Errorf("failed to adjust parity for K1: %v", err)
	}

	block1, err := des.NewCipher(key1Adjusted)
	if err != nil {
		return nil, fmt.Errorf("failed to create DES cipher for block 1: %v", err)
	}

	ct1 := make([]byte, 8)
	block1.Encrypt(ct1, h.ServerChallenge)

	// Compute block 2
	key2 := rawKeys[7:14]
	key2Adjusted, err := ParityAdjust(key2)
	if err != nil {
		return nil, fmt.Errorf("failed to adjust parity for K2: %v", err)
	}

	block2, err := des.NewCipher(key2Adjusted)
	if err != nil {
		return nil, fmt.Errorf("failed to create DES cipher for block 2: %v", err)
	}

	ct2 := make([]byte, 8)
	block2.Encrypt(ct2, h.ServerChallenge)

	// Compute block 3
	key3 := rawKeys[14:21]
	key3Adjusted, err := ParityAdjust(key3)
	if err != nil {
		return nil, fmt.Errorf("failed to adjust parity for K3: %v", err)
	}

	block3, err := des.NewCipher(key3Adjusted)
	if err != nil {
		return nil, fmt.Errorf("failed to create DES cipher for block 3: %v", err)
	}

	ct3 := make([]byte, 8)
	block3.Encrypt(ct3, h.ServerChallenge)

	// Combine the results
	response := append(append(ct1, ct2...), ct3...)

	return response, nil

}

// String returns the NTLMv1 hash as a hexadecimal string
//
// Returns:
//   - The NTLMv1 hash as an uppercase hexadecimal string
func (h *NTLMv1) String() string {
	hash, err := h.Hash()
	if err != nil {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(hash))
}

// ParityBit calculates the parity bit for a given integer.
// It returns 1 if the number of set bits (1s) in the binary representation
// of the input is even, and 0 if the number is odd.
// This is used in DES key generation to ensure each byte has odd parity.
func ParityBit(n int) int {
	parity := 1
	for n != 0 {
		if (n & 1) == 1 {
			parity ^= 1
		}
		n >>= 1
	}
	return parity
}

// ParityAdjust takes a byte slice as input and adjusts it for DES key parity.
// For each 7 bits of the input, it adds a parity bit as the 8th bit to ensure
// odd parity (each byte has an odd number of 1 bits).
// This is required for DES key generation in NTLM authentication.
//
// The function processes the input in 7-bit chunks, adding a parity bit to each chunk
// to form 8-bit bytes in the output. If the input length is not a multiple of 7 bits,
// it pads with zeros.
//
// Returns the parity-adjusted byte slice and any error encountered during processing.
func ParityAdjust(key []byte) ([]byte, error) {
	// Get a stream of bits from the key
	keyBits := []byte{}
	for _, b := range key {
		for i := 7; i >= 0; i-- {
			keyBits = append(keyBits, (b>>i)&1)
		}
	}
	keyBits = keyBits[:len(keyBits)-len(keyBits)%7]

	// Adjust the key for parity
	parityAdjustedKey := []byte{}
	for i := 0; i < len(keyBits); i += 7 {
		parityAdjustedByte := byte(0)
		for offset, bit := range keyBits[i : i+7] {
			if bit == 1 {
				parityAdjustedByte |= (1 << (7 - offset))
			}
		}
		parityAdjustedByte = parityAdjustedByte | byte(ParityBit(int(parityAdjustedByte)))
		parityAdjustedKey = append(parityAdjustedKey, parityAdjustedByte)
	}

	return parityAdjustedKey, nil
}

// NTResponse calculates the NT response for NTLMv1 authentication
//
// The NT response is calculated by encrypting the server challenge with the NT hash
// using DES encryption. The NT hash is split into three 7-byte keys, each adjusted
// for DES parity, and each key is used to encrypt the challenge.
//
// Returns the 24-byte NT response or an error if the encryption fails.
func (n *NTLMv1) NTResponse() ([]byte, error) {
	// Split the NT hash into three 7-byte keys
	key1 := n.NTHash[:7]
	key2 := n.NTHash[7:14]
	key3 := n.NTHash[14:16]
	// Pad the third key to 7 bytes with zeros
	key3 = append(key3, make([]byte, 5)...)

	// Adjust keys for DES parity
	key1, err := ParityAdjust(key1)
	if err != nil {
		return nil, fmt.Errorf("failed to adjust key1 parity: %v", err)
	}
	key2, err = ParityAdjust(key2)
	if err != nil {
		return nil, fmt.Errorf("failed to adjust key2 parity: %v", err)
	}
	key3, err = ParityAdjust(key3)
	if err != nil {
		return nil, fmt.Errorf("failed to adjust key3 parity: %v", err)
	}

	// Create DES ciphers with each key
	cipher1, err := des.NewCipher(key1)
	if err != nil {
		return nil, fmt.Errorf("failed to create DES cipher with key1: %v", err)
	}
	cipher2, err := des.NewCipher(key2)
	if err != nil {
		return nil, fmt.Errorf("failed to create DES cipher with key2: %v", err)
	}
	cipher3, err := des.NewCipher(key3)
	if err != nil {
		return nil, fmt.Errorf("failed to create DES cipher with key3: %v", err)
	}

	// Encrypt the challenge with each cipher
	result1 := make([]byte, 8)
	result2 := make([]byte, 8)
	result3 := make([]byte, 8)
	cipher1.Encrypt(result1, n.ServerChallenge)
	cipher2.Encrypt(result2, n.ServerChallenge)
	cipher3.Encrypt(result3, n.ServerChallenge)

	// Concatenate the results
	ntResponse := append(result1, result2...)
	ntResponse = append(ntResponse, result3...)

	return ntResponse, nil
}

// LMResponse calculates the LM response for NTLMv1 authentication
//
// The LM response is calculated by first creating the LM hash from the password,
// then encrypting the server challenge with the LM hash using DES encryption.
// The LM hash is split into three 7-byte keys, each adjusted for DES parity,
// and each key is used to encrypt the challenge.
//
// Returns the 24-byte LM response or an error if the encryption fails.
func (n *NTLMv1) LMResponse() ([]byte, error) {
	// Create the LM hash
	lmHash := lm.LMHash(n.Password)

	// Split the LM hash into three 7-byte keys
	key1 := lmHash[:7]
	key2 := lmHash[7:14]
	key3 := lmHash[14:16]
	// Pad the third key to 7 bytes with zeros
	key3 = append(key3, make([]byte, 5)...)

	// Adjust keys for DES parity
	key1, err := ParityAdjust(key1)
	if err != nil {
		return nil, fmt.Errorf("failed to adjust key1 parity: %v", err)
	}
	key2, err = ParityAdjust(key2)
	if err != nil {
		return nil, fmt.Errorf("failed to adjust key2 parity: %v", err)
	}
	key3, err = ParityAdjust(key3)
	if err != nil {
		return nil, fmt.Errorf("failed to adjust key3 parity: %v", err)
	}

	// Create DES ciphers with each key
	cipher1, err := des.NewCipher(key1)
	if err != nil {
		return nil, fmt.Errorf("failed to create DES cipher with key1: %v", err)
	}
	cipher2, err := des.NewCipher(key2)
	if err != nil {
		return nil, fmt.Errorf("failed to create DES cipher with key2: %v", err)
	}
	cipher3, err := des.NewCipher(key3)
	if err != nil {
		return nil, fmt.Errorf("failed to create DES cipher with key3: %v", err)
	}

	// Encrypt the challenge with each cipher
	result1 := make([]byte, 8)
	result2 := make([]byte, 8)
	result3 := make([]byte, 8)
	cipher1.Encrypt(result1, n.ServerChallenge)
	cipher2.Encrypt(result2, n.ServerChallenge)
	cipher3.Encrypt(result3, n.ServerChallenge)

	// Concatenate the results
	lmResponse := append(result1, result2...)
	lmResponse = append(lmResponse, result3...)

	return lmResponse, nil
}
