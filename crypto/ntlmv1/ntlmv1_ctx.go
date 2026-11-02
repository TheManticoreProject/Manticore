package ntlmv1

import (
	"crypto/des"
	"crypto/hmac"
	"crypto/md5"
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/lm"
	"github.com/TheManticoreProject/Manticore/crypto/md4"
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
	// Start with the NT hash of the password. NTHash and LMHash are arrays,
	// so an unset hash is the all-zero value, not an empty one.
	if h.NTHash == [16]byte{} && len(h.Password) == 0 {
		return nil, fmt.Errorf("NTHash and Password are not set")
	}
	if h.NTHash == [16]byte{} {
		h.NTHash = nt.NTHash(h.Password)
	}
	if h.LMHash == [16]byte{} && len(h.Password) != 0 {
		h.LMHash = lm.LMHash(h.Password)
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
	return desl(n.NTHash, n.ServerChallenge[:])
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
	return desl(n.LMHash, n.ServerChallenge[:])
}

// desl is the DESL function of [MS-NLMP] 6: the 16-byte key is split into three
// 7-byte DES keys (the last zero-padded), each parity-adjusted, and the 8-byte
// data is encrypted under each in turn; the three ciphertexts are concatenated.
func desl(key [16]byte, data []byte) ([24]byte, error) {
	var out [24]byte
	parts := [3][]byte{key[:7], key[7:14], append(append([]byte(nil), key[14:16]...), make([]byte, 5)...)}
	for i, part := range parts {
		desKey, err := ParityAdjust(part)
		if err != nil {
			return [24]byte{}, fmt.Errorf("failed to adjust key%d parity: %v", i+1, err)
		}
		c, err := des.NewCipher(desKey)
		if err != nil {
			return [24]byte{}, fmt.Errorf("failed to create DES cipher with key%d: %v", i+1, err)
		}
		c.Encrypt(out[8*i:8*i+8], data)
	}
	return out, nil
}

// ComputeResponseESS computes the NTLMv1 response with extended session security
// ([MS-NLMP] 3.3.1, NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY with NTLMv1), also
// called the NTLM2 session response:
//
//	NtChallengeResponse = DESL(NTOWFv1, MD5(ServerChallenge || ClientChallenge)[0:8])
//	LmChallengeResponse = ClientChallenge || Z(16)
//
// The LM hash is not used. hashcat cracks the result as mode 5500, recognizing
// the extended session security form from the zero-padded LM response.
//
// Parameters:
//   - clientChallenge: the client's 8-byte random challenge
//
// Returns:
//   - *NTLMv1Response: the responses, with the server challenge
//   - error: if no NT hash or password is set
func (h *NTLMv1Ctx) ComputeResponseESS(clientChallenge [8]byte) (*NTLMv1Response, error) {
	if h.NTHash == [16]byte{} && len(h.Password) == 0 {
		return nil, fmt.Errorf("NTHash and Password are not set")
	}
	if h.NTHash == [16]byte{} {
		h.NTHash = nt.NTHash(h.Password)
	}

	sum := md5.Sum(append(append([]byte(nil), h.ServerChallenge[:]...), clientChallenge[:]...))
	ntResponse, err := desl(h.NTHash, sum[:8])
	if err != nil {
		return nil, fmt.Errorf("failed to compute NT response: %v", err)
	}
	var lmResponse [24]byte
	copy(lmResponse[:8], clientChallenge[:])

	return NewNTLMv1Response(h.Username, h.Domain, h.ServerChallenge, lmResponse, ntResponse), nil
}

// SessionBaseKey returns the NTLMv1 SessionBaseKey, MD4(NTOWFv1) ([MS-NLMP]
// 3.3.1). When the NT hash is unset it is derived from the password, as
// ComputeResponse does.
func (h *NTLMv1Ctx) SessionBaseKey() [16]byte {
	if h.NTHash == [16]byte{} && len(h.Password) != 0 {
		h.NTHash = nt.NTHash(h.Password)
	}
	m := md4.New()
	m.Write(h.NTHash[:])
	return m.Sum()
}

// KeyExchangeKeyESS returns the KeyExchangeKey of an NTLMv1 exchange with
// extended session security ([MS-NLMP] 3.4.5.1):
// HMAC_MD5(SessionBaseKey, ServerChallenge || LmChallengeResponse[0:8]), where
// the first 8 bytes of the LM response are the client challenge.
func (h *NTLMv1Ctx) KeyExchangeKeyESS(clientChallenge [8]byte) [16]byte {
	sbk := h.SessionBaseKey()
	mac := hmac.New(md5.New, sbk[:])
	mac.Write(h.ServerChallenge[:])
	mac.Write(clientChallenge[:])
	var key [16]byte
	copy(key[:], mac.Sum(nil))
	return key
}
