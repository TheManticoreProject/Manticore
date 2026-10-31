package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/security"
)

// encryptMessage wraps a marshalled SMB2 message in an SMB2 TRANSFORM_HEADER
// sealed with the session's EncryptionKey (MS-SMB2 3.1.4.3). The nonce comes
// from a per-session counter so it never repeats.
func (c *Client) encryptMessage(plaintext []byte) ([]byte, error) {
	if c.Session == nil || len(c.Session.EncryptionKey) == 0 {
		return nil, fmt.Errorf("transform: no session encryption key")
	}
	c.Session.nonceCounter++
	return security.Encrypt(c.Connection.Cipher, c.Session.EncryptionKey, c.Session.SessionId, c.Session.nonceCounter, plaintext)
}

// decryptMessage unwraps an SMB2 TRANSFORM_HEADER frame with the session's
// DecryptionKey and returns the plaintext SMB2 message.
func (c *Client) decryptMessage(data []byte) ([]byte, error) {
	if c.Session == nil || len(c.Session.DecryptionKey) == 0 {
		return nil, fmt.Errorf("transform: no session decryption key")
	}
	if len(data) < security.TransformHeaderSize {
		return nil, fmt.Errorf("transform: frame shorter than TRANSFORM_HEADER")
	}
	if id := security.TransformSessionID(data); id != c.Session.SessionId {
		return nil, fmt.Errorf("transform: TRANSFORM_HEADER SessionId %d does not match session %d", id, c.Session.SessionId)
	}
	return security.Decrypt(c.Connection.Cipher, c.Session.DecryptionKey, data)
}

// deriveSMB3Keys fills a session's SMB 3.x keys from its SessionKey (and, for the
// SMB 3.1.1 AES-256 cipher keys, its FullSessionKey). The client encrypts with
// the client-to-server key and decrypts with the other.
func deriveSMB3Keys(session *Session, dialect dialects.Dialect, preauthHash []byte, cipher uint16, _ int) {
	keys := security.DeriveKeysWithFullSessionKey(session.SessionKey, session.FullSessionKey, dialect, preauthHash, cipher)
	session.SigningKey = keys.Signing
	session.ApplicationKey = keys.Application
	session.EncryptionKey = keys.ClientToServer
	session.DecryptionKey = keys.ServerToClient
}
