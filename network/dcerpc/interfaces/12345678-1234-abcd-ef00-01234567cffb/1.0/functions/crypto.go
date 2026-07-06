package functions

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

import (
	"crypto/aes"
	"crypto/hmac"
	"crypto/sha256"

	"github.com/TheManticoreProject/Manticore/crypto/cfb8"
	"github.com/TheManticoreProject/Manticore/crypto/nt"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// ComputeSessionKeyAES derives the AES Netlogon session key ([MS-NRPC] 3.1.4.4.1). The key
// material is the account's NT one-way function (NTOWFv1 = MD4(UTF16-LE(password))); the
// session key is the first 16 octets of HMAC-SHA256(NTOWFv1, ClientChallenge ||
// ServerChallenge).
//
// Exactly one of password or ntHash supplies the key material: when ntHash is non-nil it is
// used directly (the raw 16-byte NT hash, e.g. for pass-the-hash), otherwise it is derived
// from password.
//
// Parameters:
//   - password: The account cleartext password, or "" when ntHash is used.
//   - ntHash: The raw 16-byte NT hash, or nil when password is used.
//   - clientChallenge: The client challenge.
//   - serverChallenge: The server challenge.
//
// Returns:
//   - The 16-byte AES session key.
func ComputeSessionKeyAES(password string, ntHash []byte, clientChallenge, serverChallenge msnrpc.NETLOGON_CREDENTIAL) [16]byte {
	key := ntHash
	if key == nil {
		h := nt.NTHash(password)
		key = h[:]
	}

	mac := hmac.New(sha256.New, key)
	mac.Write(clientChallenge[:])
	mac.Write(serverChallenge[:])
	sum := mac.Sum(nil)

	var sessionKey [16]byte
	copy(sessionKey[:], sum[:16])
	return sessionKey
}

// ComputeNetlogonCredentialAES computes a Netlogon credential ([MS-NRPC] 3.1.4.4.1) by
// encrypting an 8-byte challenge with AES-128 in 8-bit cipher feedback (CFB8) mode under the
// session key and an all-zero IV.
//
// Parameters:
//   - challenge: The 8-byte input challenge.
//   - sessionKey: The 16-byte AES session key.
//
// Returns:
//   - The 8-byte Netlogon credential.
func ComputeNetlogonCredentialAES(challenge msnrpc.NETLOGON_CREDENTIAL, sessionKey [16]byte) msnrpc.NETLOGON_CREDENTIAL {
	block, err := aes.NewCipher(sessionKey[:])
	if err != nil {
		// sessionKey is always 16 bytes, so this cannot fail in practice.
		panic(err)
	}
	var iv [aes.BlockSize]byte // 16 zero octets
	var out msnrpc.NETLOGON_CREDENTIAL
	cfb8.NewEncrypter(block, iv[:]).XORKeyStream(out[:], challenge[:])
	return out
}
