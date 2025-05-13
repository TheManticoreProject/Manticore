package challenge

import (
	"bytes"
	"crypto/des"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/ntlmv1"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/negotiate"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/utils/encoding/utf16"
)

// ChallengeMessage is the second message in NTLM authentication
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/801a4681-8809-4be9-ab0d-61dcfe762786
type ChallengeMessage struct {
	// Signature (8 bytes): An 8-byte character array that MUST contain the ASCII string ('N', 'T', 'L', 'M', 'S', 'S', 'P', '\0').
	Signature [8]byte

	// MessageType (4 bytes): A 32-bit unsigned integer that indicates the message type. This field MUST be set to 0x00000002.
	MessageType message.MessageType

	// TargetNameFields (8 bytes): A field containing TargetName information. The field diagram for TargetNameFields is as follows.
	TargetName []byte

	// NegotiateFlags (4 bytes): A NEGOTIATE structure that contains a set of flags, as defined by section 2.2.2.5. The server sets flags to indicate options it supports or, if there has been a NEGOTIATE_MESSAGE (section 2.2.1.1), the choices it has made from the options offered by the client. If the client has set the NTLMSSP_NEGOTIATE_SIGN in the NEGOTIATE_MESSAGE the Server MUST return it.
	NegotiateFlags negotiate.NegotiateFlags

	// ServerChallenge (8 bytes): A 64-bit value that contains the NTLM challenge. The challenge is a 64-bit nonce. The processing of the ServerChallenge is specified in sections 3.1.5 and 3.2.5.
	ServerChallenge [8]byte

	// Reserved (8 bytes): An 8-byte array whose elements MUST be zero when sent and MUST be ignored on receipt.
	Reserved [8]byte

	// TargetInfo (variable): A field containing TargetInfo information. The field diagram for TargetInfo is as follows.
	TargetInfo []byte

	// Version (8 bytes): A VERSION structure that contains version information. The field diagram for Version is as follows.
	Version version.Version
}

// ParseChallengeMessage parses an NTLM CHALLENGE message
func ParseChallengeMessage(data []byte) (*ChallengeMessage, error) {
	if len(data) < 56 {
		return nil, errors.New("challenge message too short")
	}

	// Verify signature
	if !bytes.Equal(data[0:8], ntlm.NTLM_SIGNATURE[:]) {
		return nil, errors.New("invalid NTLM signature")
	}

	// Verify message type
	messageType := message.MessageType(binary.LittleEndian.Uint32(data[8:12]))
	if messageType != message.NTLM_CHALLENGE {
		return nil, fmt.Errorf("expected CHALLENGE message type (2), got %d", messageType)
	}

	challenge := &ChallengeMessage{}
	copy(challenge.Signature[:], data[0:8])
	challenge.MessageType = messageType

	// Parse target name
	targetNameLen := binary.LittleEndian.Uint16(data[12:14])
	// targetNameMaxLen := binary.LittleEndian.Uint16(data[14:16])
	targetNameOffset := binary.LittleEndian.Uint32(data[16:20])
	if targetNameLen > 0 && targetNameOffset+uint32(targetNameLen) <= uint32(len(data)) {
		challenge.TargetName = data[targetNameOffset : targetNameOffset+uint32(targetNameLen)]
	}

	// Parse flags
	challenge.NegotiateFlags = negotiate.NegotiateFlags(binary.LittleEndian.Uint32(data[20:24]))

	// Parse server challenge
	copy(challenge.ServerChallenge[:], data[24:32])

	// Parse reserved
	copy(challenge.Reserved[:], data[32:40])

	// Parse target info
	targetInfoLen := binary.LittleEndian.Uint16(data[40:42])
	// targetInfoMaxLen := binary.LittleEndian.Uint16(data[42:44])
	targetInfoOffset := binary.LittleEndian.Uint32(data[44:48])
	if targetInfoLen > 0 && targetInfoOffset+uint32(targetInfoLen) <= uint32(len(data)) {
		challenge.TargetInfo = data[targetInfoOffset : targetInfoOffset+uint32(targetInfoLen)]
	}

	// Parse version if present
	if challenge.NegotiateFlags&negotiate.NTLMSSP_NEGOTIATE_VERSION != 0 && len(data) >= 56 {
		bytesRead, err := challenge.Version.Unmarshal(data[48:56])
		if err != nil {
			return nil, err
		}
		if bytesRead != 8 {
			return nil, errors.New("expected 8 bytes, got " + strconv.Itoa(bytesRead))
		}
	}

	return challenge, nil
}

// calculateNTLMv1Response calculates the LM and NT responses for NTLMv1
func calculateNTLMv1Response(challenge []byte, password string) ([]byte, []byte, error) {
	// Create NTLMv1 instance
	ntlmv1instance, err := ntlmv1.NewNTLMv1WithPassword("", "", password, challenge)
	if err != nil {
		return nil, nil, err
	}

	// Calculate LM response
	lmResponse, err := ntlmv1instance.LMResponse()
	if err != nil {
		return nil, nil, err
	}

	// Calculate NT response
	ntResponse, err := ntlmv1instance.NTResponse()
	if err != nil {
		return nil, nil, err
	}

	return lmResponse, ntResponse, nil
}

// calculateNTLMv2Response calculates the LM and NT responses for NTLMv2
func calculateNTLMv2Response(challenge *ChallengeMessage, username, password, domain string) ([]byte, []byte, error) {
	// Calculate NTLMv2 hash
	ntlmv2Hash := ntowfv2(username, password, domain)

	// Create client challenge
	clientChallenge := make([]byte, 8)
	_, err := rand.Read(clientChallenge)
	if err != nil {
		return nil, nil, err
	}

	// Extract target info
	targetInfo := challenge.TargetInfo

	// Create blob
	blob := createNTLMv2Blob(clientChallenge, targetInfo)

	// Calculate proof
	serverChallenge := challenge.ServerChallenge[:]
	proof := calculateNTLMv2Proof(ntlmv2Hash, serverChallenge, blob)

	// NTLMv2 response is the proof followed by the blob
	ntResponse := append(proof, blob...)

	// For LMv2, we use a different client challenge
	lmClientChallenge := make([]byte, 8)
	_, err = rand.Read(lmClientChallenge)
	if err != nil {
		return nil, nil, err
	}

	// Calculate LMv2 response
	lmProof := calculateNTLMv2Proof(ntlmv2Hash, serverChallenge, lmClientChallenge)
	lmResponse := append(lmProof, lmClientChallenge...)

	return lmResponse, ntResponse, nil
}

// createNTLMv2Blob creates the NTLMv2 blob
func createNTLMv2Blob(clientChallenge, targetInfo []byte) []byte {
	data := []byte{}

	// Blob signature
	data = append(data, 0x01)
	data = append(data, 0x01)

	// Reserved
	data = append(data, []byte{0, 0, 0, 0, 0, 0}...)

	// Timestamp (Windows file time format)
	now := time.Now()
	// Convert to Windows file time (100ns intervals since Jan 1, 1601)
	// First get seconds since Unix epoch (Jan 1, 1970)
	unixSecs := now.Unix()
	// Add seconds between 1601 and 1970
	windowsSecs := unixSecs + 11644473600
	// Convert to 100ns intervals
	windowsTime := windowsSecs * 10000000

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(windowsTime))
	data = append(data, buf...)

	// Client challenge
	data = append(data, clientChallenge...)

	// Unknown (4 bytes of zeros)
	data = append(data, []byte{0, 0, 0, 0}...)

	// Target info
	data = append(data, targetInfo...)

	// Unknown (4 bytes of zeros)
	data = append(data, []byte{0, 0, 0, 0}...)

	return data
}

// calculateNTLMv2Proof calculates the NTLMv2 proof
func calculateNTLMv2Proof(ntlmv2Hash, serverChallenge, blob []byte) []byte {
	// Concatenate server challenge and blob
	data := append(serverChallenge, blob...)

	// Calculate HMAC-MD5
	h := hmac.New(md5.New, ntlmv2Hash)
	h.Write(data)
	return h.Sum(nil)
}

// ntowfv2 calculates the NTLMv2 hash
func ntowfv2(username, password, domain string) []byte {
	// Calculate NT hash
	ntlm, _ := ntlmv1.NewNTLMv1WithPassword("", "", password, make([]byte, 8))
	ntHash := ntlm.NTHash

	// Convert username and domain to uppercase and UTF-16LE
	upperUsername := strings.ToUpper(username)
	upperDomain := strings.ToUpper(domain)
	usernameAndDomain := utf16.EncodeUTF16LE(upperUsername + upperDomain)

	// Calculate HMAC-MD5
	h := hmac.New(md5.New, ntHash)
	h.Write(usernameAndDomain)
	return h.Sum(nil)
}

// createDesKey creates a DES key from a 7-byte input
func createDesKey(bytes []byte) ([]byte, error) {
	if len(bytes) != 7 {
		return nil, errors.New("input must be 7 bytes")
	}

	// Convert 7 bytes to 8 bytes (56 bits to 64 bits)
	key := make([]byte, 8)

	// Byte 0
	key[0] = bytes[0] >> 1
	// Byte 1
	key[1] = ((bytes[0] & 0x01) << 6) | (bytes[1] >> 2)
	// Byte 2
	key[2] = ((bytes[1] & 0x03) << 5) | (bytes[2] >> 3)
	// Byte 3
	key[3] = ((bytes[2] & 0x07) << 4) | (bytes[3] >> 4)
	// Byte 4
	key[4] = ((bytes[3] & 0x0F) << 3) | (bytes[4] >> 5)
	// Byte 5
	key[5] = ((bytes[4] & 0x1F) << 2) | (bytes[5] >> 6)
	// Byte 6
	key[6] = ((bytes[5] & 0x3F) << 1) | (bytes[6] >> 7)
	// Byte 7
	key[7] = bytes[6] & 0x7F

	// Set parity bits
	for i := 0; i < 8; i++ {
		key[i] = (key[i] << 1)

		// Count bits set
		var bitCount int
		for j := 0; j < 7; j++ {
			if (key[i] & (1 << j)) != 0 {
				bitCount++
			}
		}

		// Set parity bit
		if bitCount%2 == 0 {
			key[i] |= 1
		}
	}

	return key, nil
}

// desEncrypt encrypts a challenge with a hash using DES
func desEncrypt(hash, challenge []byte) []byte {
	if len(hash) != 16 || len(challenge) != 8 {
		return nil
	}

	// Split hash into three parts
	k1 := hash[:7]
	k2 := hash[7:14]
	k3 := append(hash[14:16], make([]byte, 5)...)

	// Create DES keys
	key1, _ := createDesKey(k1)
	key2, _ := createDesKey(k2)
	key3, _ := createDesKey(k3)

	// Encrypt challenge with each key
	cipher1, _ := des.NewCipher(key1)
	result1 := make([]byte, 8)
	cipher1.Encrypt(result1, challenge)

	cipher2, _ := des.NewCipher(key2)
	result2 := make([]byte, 8)
	cipher2.Encrypt(result2, challenge)

	cipher3, _ := des.NewCipher(key3)
	result3 := make([]byte, 8)
	cipher3.Encrypt(result3, challenge)

	// Concatenate results
	return append(append(result1, result2...), result3...)
}
