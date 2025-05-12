package ntlm

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
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/utils/encoding/utf16"
)

// NTLM message types
const (
	NTLM_NEGOTIATE    = 1
	NTLM_CHALLENGE    = 2
	NTLM_AUTHENTICATE = 3
)

// Target info field types
const (
	MsvAvEOL             uint16 = 0x0000
	MsvAvNbComputerName  uint16 = 0x0001
	MsvAvNbDomainName    uint16 = 0x0002
	MsvAvDnsComputerName uint16 = 0x0003
	MsvAvDnsDomainName   uint16 = 0x0004
	MsvAvDnsTreeName     uint16 = 0x0005
	MsvAvFlags           uint16 = 0x0006
	MsvAvTimestamp       uint16 = 0x0007
	MsvAvSingleHost      uint16 = 0x0008
	MsvAvTargetName      uint16 = 0x0009
	MsvAvChannelBindings uint16 = 0x000A
)

// NTLM signature
var NTLM_SIGNATURE = []byte("NTLMSSP\x00")

// NegotiateMessage is the first message in NTLM authentication
type NegotiateMessage struct {
	Signature      [8]byte
	MessageType    uint32
	NegotiateFlags uint32
	DomainName     []byte
	Workstation    []byte
	Version        version.Version
}

// ChallengeMessage is the second message in NTLM authentication
type ChallengeMessage struct {
	Signature       [8]byte
	MessageType     uint32
	TargetName      []byte
	NegotiateFlags  uint32
	ServerChallenge [8]byte
	Reserved        [8]byte
	TargetInfo      []byte
	Version         version.Version
}

// AuthenticateMessage is the third message in NTLM authentication
type AuthenticateMessage struct {
	Signature                 [8]byte
	MessageType               uint32
	LmChallengeResponse       []byte
	NtChallengeResponse       []byte
	DomainName                []byte
	UserName                  []byte
	Workstation               []byte
	EncryptedRandomSessionKey []byte
	NegotiateFlags            uint32
	Version                   version.Version
	MIC                       [16]byte
}

// AvPair represents a Target Info AV_PAIR structure
type AvPair struct {
	AvID   uint16
	AvLen  uint16
	AvData []byte
}

// CreateNegotiateMessage creates an NTLM NEGOTIATE message
func CreateNegotiateMessage(domain, workstation string, useUnicode bool) ([]byte, error) {
	flags := NTLMSSP_NEGOTIATE_NTLM |
		NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
		NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
		NTLMSSP_NEGOTIATE_128 |
		NTLMSSP_NEGOTIATE_56 |
		NTLMSSP_REQUEST_TARGET |
		NTLMSSP_NEGOTIATE_TARGET_INFO |
		NTLMSSP_NEGOTIATE_VERSION

	if useUnicode {
		flags |= NTLMSSP_NEGOTIATE_UNICODE
	} else {
		flags |= NTLMSSP_NEGOTIATE_OEM
	}

	var domainBytes, workstationBytes []byte

	if domain != "" {
		flags |= NTLMSSP_NEGOTIATE_OEM_DOMAIN_SUPPLIED
		if useUnicode {
			domainBytes = utf16.EncodeUTF16LE(domain)
		} else {
			domainBytes = []byte(strings.ToUpper(domain))
		}
	}

	if workstation != "" {
		flags |= NTLMSSP_NEGOTIATE_OEM_WORKSTATION_SUPPLIED
		if useUnicode {
			workstationBytes = utf16.EncodeUTF16LE(workstation)
		} else {
			workstationBytes = []byte(strings.ToUpper(workstation))
		}
	}

	// Calculate offsets and sizes
	headerSize := 40 // 8 (signature) + 4 (type) + 4 (flags) + 8 (domain fields) + 8 (workstation fields) + 8 (version)
	domainOffset := headerSize
	workstationOffset := domainOffset + len(domainBytes)

	// Create data
	data := []byte{}

	// Write signature
	data = append(data, NTLM_SIGNATURE...)

	// Write message type
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(NTLM_NEGOTIATE))
	data = append(data, buf...)

	// Write negotiate flags
	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, flags)
	data = append(data, buf...)

	// Write domain name fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(domainBytes)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(domainBytes)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(domainOffset))
	data = append(data, buf...)

	// Write workstation fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(workstationBytes)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(workstationBytes)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(workstationOffset))
	data = append(data, buf...)

	// Write version
	v := version.DefaultVersion()
	byteStream, err := v.Marshal()
	if err != nil {
		return nil, err
	}
	data = append(data, byteStream...)

	// Write domain and workstation
	data = append(data, domainBytes...)
	data = append(data, workstationBytes...)

	return data, nil
}

// ParseChallengeMessage parses an NTLM CHALLENGE message
func ParseChallengeMessage(data []byte) (*ChallengeMessage, error) {
	if len(data) < 56 {
		return nil, errors.New("challenge message too short")
	}

	// Verify signature
	if !bytes.Equal(data[0:8], NTLM_SIGNATURE) {
		return nil, errors.New("invalid NTLM signature")
	}

	// Verify message type
	messageType := binary.LittleEndian.Uint32(data[8:12])
	if messageType != NTLM_CHALLENGE {
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
	challenge.NegotiateFlags = binary.LittleEndian.Uint32(data[20:24])

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
	if challenge.NegotiateFlags&NTLMSSP_NEGOTIATE_VERSION != 0 && len(data) >= 56 {
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

// CreateAuthenticateMessage creates an NTLM AUTHENTICATE message
func CreateAuthenticateMessage(challenge *ChallengeMessage, username, password, domain, workstation string) ([]byte, error) {
	// Determine if we should use Unicode
	useUnicode := (challenge.NegotiateFlags & NTLMSSP_NEGOTIATE_UNICODE) != 0

	// Prepare domain, username, and workstation
	var domainBytes, usernameBytes, workstationBytes []byte

	if useUnicode {
		domainBytes = utf16.EncodeUTF16LE(strings.ToUpper(domain))
		usernameBytes = utf16.EncodeUTF16LE(username)
		workstationBytes = utf16.EncodeUTF16LE(strings.ToUpper(workstation))
	} else {
		domainBytes = []byte(strings.ToUpper(domain))
		usernameBytes = []byte(username)
		workstationBytes = []byte(strings.ToUpper(workstation))
	}

	// Calculate NT response
	var lmResponse, ntResponse []byte
	var err error

	if (challenge.NegotiateFlags & NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY) != 0 {
		// Use NTLMv2
		lmResponse, ntResponse, err = calculateNTLMv2Response(challenge, username, password, domain)
	} else {
		// Use NTLMv1
		lmResponse, ntResponse, err = calculateNTLMv1Response(challenge.ServerChallenge[:], password)
	}

	if err != nil {
		return nil, err
	}

	// Prepare session key (empty for now)
	sessionKey := []byte{}

	// Calculate offsets
	headerSize := 88 // Fixed header size including MIC
	lmResponseOffset := headerSize
	ntResponseOffset := lmResponseOffset + len(lmResponse)
	domainOffset := ntResponseOffset + len(ntResponse)
	usernameOffset := domainOffset + len(domainBytes)
	workstationOffset := usernameOffset + len(usernameBytes)
	sessionKeyOffset := workstationOffset + len(workstationBytes)

	// Create data
	data := []byte{}

	// Write signature
	data = append(data, NTLM_SIGNATURE...)

	// Write message type
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(NTLM_AUTHENTICATE))
	data = append(data, buf...)

	// Write LM response fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(lmResponse)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(lmResponse)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(lmResponseOffset))
	data = append(data, buf...)

	// Write NT response fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(ntResponse)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(ntResponse)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(ntResponseOffset))
	data = append(data, buf...)

	// Write domain fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(domainBytes)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(domainBytes)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(domainOffset))
	data = append(data, buf...)

	// Write username fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(usernameBytes)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(usernameBytes)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(usernameOffset))
	data = append(data, buf...)

	// Write workstation fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(workstationBytes)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(workstationBytes)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(workstationOffset))
	data = append(data, buf...)

	// Write session key fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(sessionKey)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(sessionKey)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(sessionKeyOffset))
	data = append(data, buf...)

	// Write negotiate flags
	flags := challenge.NegotiateFlags
	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, flags)
	data = append(data, buf...)

	// Write version if needed
	if (flags & NTLMSSP_NEGOTIATE_VERSION) != 0 {
		v := version.DefaultVersion()
		byteStream, err := v.Marshal()
		if err != nil {
			return nil, err
		}
		data = append(data, byteStream...)
	} else {
		// Write 8 bytes of zeros
		data = append(data, make([]byte, 8)...)
	}

	// Write MIC (all zeros for now)
	data = append(data, make([]byte, 16)...)

	// Write payload data
	data = append(data, lmResponse...)
	data = append(data, ntResponse...)
	data = append(data, domainBytes...)
	data = append(data, usernameBytes...)
	data = append(data, workstationBytes...)
	data = append(data, sessionKey...)

	return data, nil
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

// ParseTargetInfo parses the target info from a challenge message
func ParseTargetInfo(targetInfo []byte) (map[uint16][]byte, error) {
	result := make(map[uint16][]byte)

	offset := 0
	for offset < len(targetInfo) {
		// Need at least 4 bytes for the AV_PAIR header
		if offset+4 > len(targetInfo) {
			return nil, errors.New("target info truncated")
		}

		avId := binary.LittleEndian.Uint16(targetInfo[offset : offset+2])
		avLen := binary.LittleEndian.Uint16(targetInfo[offset+2 : offset+4])

		offset += 4

		// Check if we have enough bytes for the value
		if offset+int(avLen) > len(targetInfo) {
			return nil, errors.New("target info value truncated")
		}

		// Extract the value
		if avId != MsvAvEOL {
			result[avId] = targetInfo[offset : offset+int(avLen)]
		}

		offset += int(avLen)

		// If we reached the end of list marker, we're done
		if avId == MsvAvEOL {
			break
		}
	}

	return result, nil
}
