// Package crypto implements the Netlogon secure-channel cryptographic primitives ([MS-NRPC]
// 3.1.4.3/3.1.4.4/3.1.4.5): session-key derivation, credential computation, and client
// authenticator computation, for both the AES and the legacy (strong-key/DES) cipher suites.
// It depends only on the NDR wire structures in the parent package, never on the RPC opnums,
// so it stays reusable and free of import cycles.
package crypto

import (
	"crypto/aes"
	"crypto/des"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/crypto/aes/cfb8"
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

// AddToCredential adds delta to a credential ([MS-NRPC] 3.1.4.5): the least-significant 4
// octets are treated as a little-endian 32-bit integer and delta is added with overflow
// ignored; the most-significant 4 octets are left unchanged. This is the arithmetic used to
// advance the stored credential by a timestamp or by the constant 1.
func AddToCredential(cred msnrpc.NETLOGON_CREDENTIAL, delta uint32) msnrpc.NETLOGON_CREDENTIAL {
	var out msnrpc.NETLOGON_CREDENTIAL
	low := binary.LittleEndian.Uint32(cred[0:4]) + delta // uint32 wraps mod 2^32
	binary.LittleEndian.PutUint32(out[0:4], low)
	copy(out[4:8], cred[4:8])
	return out
}

// ComputeNetlogonAuthenticatorAES computes a client Netlogon authenticator ([MS-NRPC]
// 3.1.4.5) for the AES cipher suite: it adds timestamp to the stored credential (low 32
// bits, overflow ignored) and encrypts the sum with the session key
// (ComputeNetlogonCredentialAES). This is a pure function — it does not advance the caller's
// stored credential; a caller that maintains a rolling secure channel should use
// SecureChannel, which applies the stored-credential updates the protocol requires.
//
// Parameters:
//   - storedCredential: The current stored client credential.
//   - timestamp: The authenticator timestamp (seconds since 1970-01-01 UTC).
//   - sessionKey: The 16-byte AES session key.
//
// Returns:
//   - The NETLOGON_AUTHENTICATOR to send with the request.
func ComputeNetlogonAuthenticatorAES(storedCredential msnrpc.NETLOGON_CREDENTIAL, timestamp uint32, sessionKey [16]byte) msnrpc.NETLOGON_AUTHENTICATOR {
	return msnrpc.NETLOGON_AUTHENTICATOR{
		Credential: ComputeNetlogonCredentialAES(AddToCredential(storedCredential, timestamp), sessionKey),
		Timestamp:  timestamp,
	}
}

// ComputeSessionKeyStrongKey derives the legacy "strong-key" Netlogon session key ([MS-NRPC]
// 3.1.4.3.1, the non-AES branch), used when AES is not negotiated but strong keys are. It is
// HMAC-MD5(NTOWFv1, MD5(0x00000000 || ClientChallenge || ServerChallenge)); the 16-byte
// HMAC-MD5 output is the session key. Exactly one of password or ntHash supplies the key
// material (see ComputeSessionKeyAES).
func ComputeSessionKeyStrongKey(password string, ntHash []byte, clientChallenge, serverChallenge msnrpc.NETLOGON_CREDENTIAL) [16]byte {
	key := ntHash
	if key == nil {
		h := nt.NTHash(password)
		key = h[:]
	}

	inner := md5.New()
	inner.Write([]byte{0, 0, 0, 0})
	inner.Write(clientChallenge[:])
	inner.Write(serverChallenge[:])

	mac := hmac.New(md5.New, key)
	mac.Write(inner.Sum(nil))

	var sessionKey [16]byte
	copy(sessionKey[:], mac.Sum(nil)) // HMAC-MD5 is exactly 16 octets
	return sessionKey
}

// transformKey expands a 7-byte value into an 8-byte DES key by spreading the bits across
// the high 7 bits of each output byte, leaving the low (parity) bit zero ([MS-SAMR]
// 2.2.11.1.2; Go's DES ignores parity). It is the key schedule for ComputeNetlogonCredential.
func transformKey(in [7]byte) [8]byte {
	var out [8]byte
	out[0] = in[0] >> 1
	out[1] = ((in[0] & 0x01) << 6) | (in[1] >> 2)
	out[2] = ((in[1] & 0x03) << 5) | (in[2] >> 3)
	out[3] = ((in[2] & 0x07) << 4) | (in[3] >> 4)
	out[4] = ((in[3] & 0x0F) << 3) | (in[4] >> 5)
	out[5] = ((in[4] & 0x1F) << 2) | (in[5] >> 6)
	out[6] = ((in[5] & 0x3F) << 1) | (in[6] >> 7)
	out[7] = in[6] & 0x7F
	for i := range out {
		out[i] = (out[i] << 1) & 0xFE
	}
	return out
}

// ComputeNetlogonCredential computes a Netlogon credential for the legacy (non-AES) cipher
// suite ([MS-NRPC] 3.1.4.4.2): the 8-byte input is encrypted with two DES-ECB passes whose
// keys are the first and second 7 octets of the session key, each expanded by transformKey.
// It is used by both the strong-key (RC4) and DES paths, in place of
// ComputeNetlogonCredentialAES.
func ComputeNetlogonCredential(input msnrpc.NETLOGON_CREDENTIAL, sessionKey [16]byte) msnrpc.NETLOGON_CREDENTIAL {
	var k1, k2 [7]byte
	copy(k1[:], sessionKey[0:7])
	copy(k2[:], sessionKey[7:14])
	t1 := transformKey(k1)
	t2 := transformKey(k2)
	b1, err := des.NewCipher(t1[:])
	if err != nil {
		panic(err) // 8-byte key, cannot fail
	}
	b2, err := des.NewCipher(t2[:])
	if err != nil {
		panic(err)
	}
	var tmp, out msnrpc.NETLOGON_CREDENTIAL
	b1.Encrypt(tmp[:], input[:])
	b2.Encrypt(out[:], tmp[:])
	return out
}

// ComputeNetlogonAuthenticator computes a client Netlogon authenticator ([MS-NRPC] 3.1.4.5)
// for the legacy (non-AES) cipher suite: identical arithmetic to ComputeNetlogonAuthenticatorAES
// but using the DES-based ComputeNetlogonCredential.
func ComputeNetlogonAuthenticator(storedCredential msnrpc.NETLOGON_CREDENTIAL, timestamp uint32, sessionKey [16]byte) msnrpc.NETLOGON_AUTHENTICATOR {
	return msnrpc.NETLOGON_AUTHENTICATOR{
		Credential: ComputeNetlogonCredential(AddToCredential(storedCredential, timestamp), sessionKey),
		Timestamp:  timestamp,
	}
}
