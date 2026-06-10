package ntlmv1

import (
	"crypto/des"
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/lm"
	"github.com/TheManticoreProject/Manticore/crypto/nt"
)

// NTLMv1Ctx represents the components needed for NTLMv1 authentication.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/464551a8-9fc4-428e-b3d3-bc5bfb2e73a5
type NTLMv1Ctx struct {
	Username string // Username for authentication
	Password string // Password for authentication
	Domain   string // Domain name

	NTHash [16]byte // NT hash of the password
	LMHash [16]byte // LM hash of the password

	ServerChallenge [8]byte // 8-byte challenge from server
}

// NewNTLMv1CtxWithPassword creates a new NTLMv1 instance with the provided credentials and challenge.
// It calculates both the NT and LM hashes from the provided password.
//
// Parameters:
//   - domain: The domain name
//   - username: The username
//   - password: The plaintext password
//   - serverChallenge: 8-byte challenge from the server
//
// Returns:
//   - *NTLMv1Ctx: The initialized NTLMv1 context
//   - error: If server challenge is not 8 bytes
func NewNTLMv1CtxWithPassword(domain, username, password string, serverChallenge [8]byte) (*NTLMv1Ctx, error) {
	if len(serverChallenge) != 8 {
		return nil, fmt.Errorf("server challenge must be 8 bytes")
	}

	ntHash := nt.NTHash(password)
	lmHash := lm.LMHash(password)

	ntlm := &NTLMv1Ctx{
		Domain:   domain,
		Username: username,
		Password: password,

		ServerChallenge: serverChallenge,

		NTHash: ntHash,
		LMHash: lmHash,
	}

	return ntlm, nil
}

// NewNTLMv1CtxWithNTHash creates a new NTLMv1 instance using a pre-computed NT hash.
// The LM hash will be computed from an empty password.
//
// Parameters:
//   - domain: The domain name
//   - username: The username
//   - nthash: The pre-computed NT hash
//   - serverChallenge: 8-byte challenge from the server
//
// Returns:
//   - *NTLMv1Ctx: The initialized NTLMv1 context
//   - error: If server challenge is not 8 bytes
func NewNTLMv1CtxWithNTHash(domain, username string, nthash [16]byte, serverChallenge [8]byte) (*NTLMv1Ctx, error) {
	if len(serverChallenge) != 8 {
		return nil, fmt.Errorf("server challenge must be 8 bytes")
	}

	ntlmv1Ctx := &NTLMv1Ctx{
		Domain:   domain,
		Username: username,
		Password: "",

		ServerChallenge: serverChallenge,

		NTHash: nthash,
		LMHash: lm.LMHash(""),
	}

	return ntlmv1Ctx, nil
}

// NewNTLMv1CtxWithLMHash creates a new NTLMv1 instance using a pre-computed LM hash.
// The NT hash will be computed from an empty password.
//
// Parameters:
//   - domain: The domain name
//   - username: The username
//   - lmhash: The pre-computed LM hash
//   - serverChallenge: 8-byte challenge from the server
//
// Returns:
//   - *NTLMv1Ctx: The initialized NTLMv1 context
//   - error: If server challenge is not 8 bytes
func NewNTLMv1CtxWithLMHash(domain, username string, lmhash [16]byte, serverChallenge [8]byte) (*NTLMv1Ctx, error) {
	if len(serverChallenge) != 8 {
		return nil, fmt.Errorf("server challenge must be 8 bytes")
	}

	ntHash := nt.NTHash("")

	ntlmv1Ctx := &NTLMv1Ctx{
		Domain:          domain,
		Username:        username,
		Password:        "",
		ServerChallenge: serverChallenge,
		NTHash:          ntHash,
		LMHash:          lmhash,
	}

	return ntlmv1Ctx, nil
}

// NewNTLMv1CtxWithHashes creates a new NTLMv1 instance using pre-computed NT and LM hashes.
//
// Parameters:
//   - domain: The domain name
//   - username: The username
//   - lmhash: The pre-computed LM hash
//   - nthash: The pre-computed NT hash
//   - serverChallenge: 8-byte challenge from the server
//
// Returns:
//   - *NTLMv1Ctx: The initialized NTLMv1 context
//   - error: If server challenge is not 8 bytes
func NewNTLMv1CtxWithHashes(domain, username string, lmhash [16]byte, nthash [16]byte, serverChallenge [8]byte) (*NTLMv1Ctx, error) {
	if len(serverChallenge) != 8 {
		return nil, fmt.Errorf("server challenge must be 8 bytes")
	}

	ntlmv1Ctx := &NTLMv1Ctx{
		Domain:   domain,
		Username: username,
		Password: "",

		ServerChallenge: serverChallenge,

		NTHash: nthash,
		LMHash: lmhash,
	}

	return ntlmv1Ctx, nil
}

// ComputeResponse calculates the NTLMv1 response for the given context.
//
// The response is calculated by encrypting the server challenge with the LM and NT hashes
// using DES encryption. The LM and NT hashes are split into three 7-byte keys, each adjusted
// for DES parity, and each key is used to encrypt the challenge.
//
// Returns:
//   - *NTLMv1Response: The NTLMv1 response
//   - error: If the response cannot be computed
func (h *NTLMv1Ctx) ComputeResponse() (*NTLMv1Response, error) {
	// Start with the NT hash of the password
	if len(h.NTHash) == 0 && len(h.Password) == 0 {
		return nil, fmt.Errorf("NTHash and Password are not set")
	}
	if len(h.NTHash) == 0 {
		ntHash := nt.NTHash(h.Password)
		h.NTHash = ntHash
	}

	ntResponse, err := h.ComputeNtChallengeResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to compute NT response: %v", err)
	}

	lmResponse, err := h.ComputeLmChallengeResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to compute LM response: %v", err)
	}

	serverChallenge := [8]byte(h.ServerChallenge[:])
	lmResponseBytes := [24]byte(lmResponse[:])
	ntResponseBytes := [24]byte(ntResponse[:])
	response := NewNTLMv1Response(h.Username, h.Domain, serverChallenge, lmResponseBytes, ntResponseBytes)

	return response, nil
}

// NtChallengeResponse calculates the NT response for NTLMv1 authentication.
//
// The NT response is calculated by encrypting the server challenge with the NT hash
// using DES encryption. The NT hash is split into three 7-byte keys, each adjusted
// for DES parity, and each key is used to encrypt the challenge.
//
// The process:
// 1. Split the NT hash into three 7-byte keys (padding the last key with zeros)
// 2. Adjust each key for DES parity
// 3. Create DES ciphers with each key
// 4. Encrypt the server challenge with each cipher
// 5. Concatenate the results
//
// Returns:
//   - []byte: The 24-byte NT response
//   - error: If key adjustment or encryption fails
func (n *NTLMv1Ctx) ComputeNtChallengeResponse() ([24]byte, error) {
	// Split the NT hash into three 7-byte keys
	key1 := n.NTHash[:7]
	key2 := n.NTHash[7:14]
	key3 := n.NTHash[14:16]
	// Pad the third key to 7 bytes with zeros
	key3 = append(key3, make([]byte, 5)...)

	// Adjust keys for DES parity
	key1, err := ParityAdjust(key1)
	if err != nil {
		return [24]byte{}, fmt.Errorf("failed to adjust key1 parity: %v", err)
	}
	key2, err = ParityAdjust(key2)
	if err != nil {
		return [24]byte{}, fmt.Errorf("failed to adjust key2 parity: %v", err)
	}
	key3, err = ParityAdjust(key3)
	if err != nil {
		return [24]byte{}, fmt.Errorf("failed to adjust key3 parity: %v", err)
	}

	// Create DES ciphers with each key
	cipher1, err := des.NewCipher(key1)
	if err != nil {
		return [24]byte{}, fmt.Errorf("failed to create DES cipher with key1: %v", err)
	}
	cipher2, err := des.NewCipher(key2)
	if err != nil {
		return [24]byte{}, fmt.Errorf("failed to create DES cipher with key2: %v", err)
	}
	cipher3, err := des.NewCipher(key3)
	if err != nil {
		return [24]byte{}, fmt.Errorf("failed to create DES cipher with key3: %v", err)
	}

	// Encrypt the challenge with each cipher
	result1 := make([]byte, 8)
	result2 := make([]byte, 8)
	result3 := make([]byte, 8)
	cipher1.Encrypt(result1, n.ServerChallenge[:])
	cipher2.Encrypt(result2, n.ServerChallenge[:])
	cipher3.Encrypt(result3, n.ServerChallenge[:])

	// Concatenate the results
	ntResponse := append(result1, result2...)
	ntResponse = append(ntResponse, result3...)

	return [24]byte(ntResponse), nil
}

// LmChallengeResponse calculates the LM response for NTLMv1 authentication.
//
// The LM response is calculated by encrypting the server challenge with the LM hash
// using DES encryption. The LM hash is split into three 7-byte keys, each adjusted
// for DES parity, and each key is used to encrypt the challenge.
//
// The process:
// 1. Split the LM hash into three 7-byte keys (padding the last key with zeros)
// 2. Adjust each key for DES parity
// 3. Create DES ciphers with each key
// 4. Encrypt the server challenge with each cipher
// 5. Concatenate the results
//
// Returns:
//   - []byte: The 24-byte LM response
//   - error: If key adjustment or encryption fails
func (n *NTLMv1Ctx) ComputeLmChallengeResponse() ([24]byte, error) {
	// Use the stored LM hash, which the constructors populate either from the
	// password or directly from a caller-supplied hash (pass-the-hash). This
	// mirrors ComputeNtChallengeResponse, which keys off n.NTHash; recomputing
	// from n.Password here would discard a supplied hash, since the hash-based
	// constructors leave Password empty.
	lmHash := n.LMHash

	// Split the LM hash into three 7-byte keys
	key1 := lmHash[:7]
	key2 := lmHash[7:14]
	key3 := lmHash[14:16]
	// Pad the third key to 7 bytes with zeros
	key3 = append(key3, make([]byte, 5)...)

	// Adjust keys for DES parity
	key1, err := ParityAdjust(key1)
	if err != nil {
		return [24]byte{}, fmt.Errorf("failed to adjust key1 parity: %v", err)
	}
	key2, err = ParityAdjust(key2)
	if err != nil {
		return [24]byte{}, fmt.Errorf("failed to adjust key2 parity: %v", err)
	}
	key3, err = ParityAdjust(key3)
	if err != nil {
		return [24]byte{}, fmt.Errorf("failed to adjust key3 parity: %v", err)
	}

	// Create DES ciphers with each key
	cipher1, err := des.NewCipher(key1)
	if err != nil {
		return [24]byte{}, fmt.Errorf("failed to create DES cipher with key1: %v", err)
	}
	cipher2, err := des.NewCipher(key2)
	if err != nil {
		return [24]byte{}, fmt.Errorf("failed to create DES cipher with key2: %v", err)
	}
	cipher3, err := des.NewCipher(key3)
	if err != nil {
		return [24]byte{}, fmt.Errorf("failed to create DES cipher with key3: %v", err)
	}

	// Encrypt the challenge with each cipher
	result1 := make([]byte, 8)
	result2 := make([]byte, 8)
	result3 := make([]byte, 8)
	cipher1.Encrypt(result1, n.ServerChallenge[:])
	cipher2.Encrypt(result2, n.ServerChallenge[:])
	cipher3.Encrypt(result3, n.ServerChallenge[:])

	// Concatenate the results
	lmResponse := append(result1, result2...)
	lmResponse = append(lmResponse, result3...)

	return [24]byte(lmResponse), nil
}
