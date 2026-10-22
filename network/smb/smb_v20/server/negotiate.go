package server

import (
	"bytes"
	"crypto/rand"
	"encoding/asn1"
	"encoding/binary"
	"slices"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/security"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// dialectRank orders the dialects this server speaks.
var dialectRank = map[dialects.Dialect]int{
	dialects.SMB2_DIALECT_2_0_2: 1,
	dialects.SMB2_DIALECT_2_1_0: 2,
	dialects.SMB2_DIALECT_3_0_0: 3,
	dialects.SMB2_DIALECT_3_0_2: 4,
	dialects.SMB2_DIALECT_3_1_1: 5,
}

// Server preference among the SMB 3.1.1 ciphers and signing algorithms.
var (
	cipherPreference = []uint16{
		commands.SMB2_ENCRYPTION_AES128_GCM, commands.SMB2_ENCRYPTION_AES128_CCM,
		commands.SMB2_ENCRYPTION_AES256_GCM, commands.SMB2_ENCRYPTION_AES256_CCM,
	}
	signingPreference = []uint16{
		commands.SMB2_SIGNING_ALG_AES_GMAC, commands.SMB2_SIGNING_ALG_AES_CMAC, commands.SMB2_SIGNING_ALG_HMAC_SHA256,
	}
)

// offers reports whether the server is configured to speak dialect.
func (s *Server) offers(dialect dialects.Dialect) bool {
	return slices.Contains(s.config.Dialects, dialect)
}

// highestDialect returns the highest dialect the server offers.
func (s *Server) highestDialect() dialects.Dialect {
	return slices.MaxFunc(s.config.Dialects, func(a, b dialects.Dialect) int { return dialectRank[a] - dialectRank[b] })
}

// negotiate answers SMB2 NEGOTIATE (MS-SMB2 3.3.5.4).
func (c *Connection) negotiate(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.NegotiateRequest)

	var selected dialects.Dialect
	for _, d := range req.Dialects {
		if c.server.offers(d) && dialectRank[d] > dialectRank[selected] {
			selected = d
		}
	}
	if selected == 0 {
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}

	c.clientGUID = req.ClientGuid
	c.clientCapabilities = req.Capabilities
	c.clientSecurityMode = req.SecurityMode
	c.clientDialects = req.Dialects

	var contexts []*commands.NegotiateContext
	switch selected {
	case dialects.SMB2_DIALECT_3_0_0, dialects.SMB2_DIALECT_3_0_2:
		if req.Capabilities.Has(capabilities.SMB2_GLOBAL_CAP_ENCRYPTION) {
			c.cipher = commands.SMB2_ENCRYPTION_AES128_CCM
		}
	case dialects.SMB2_DIALECT_3_1_1:
		var ok bool
		if contexts, ok = c.negotiateContexts(req.Contexts); !ok {
			return nt_status.NT_STATUS_INVALID_PARAMETER
		}
	}

	c.dialect = selected
	resp.command = c.negotiateResponse(selected, contexts)
	if selected == dialects.SMB2_DIALECT_3_1_1 {
		c.preauthHash = security.PreauthUpdate(make([]byte, security.PreauthHashSize), r.raw)
		resp.onSent = func(wire []byte) { c.preauthHash = security.PreauthUpdate(c.preauthHash, wire) }
	}
	logger.Debugf("SMB2 server: %s negotiated %s", c.remote, selected)
	return nt_status.NT_STATUS_SUCCESS
}

// negotiateContexts selects the SMB 3.1.1 hash, cipher and signing algorithm
// and returns the contexts answering them. A client that offers no SHA-512
// pre-authentication integrity is refused.
func (c *Connection) negotiateContexts(offered []*commands.NegotiateContext) ([]*commands.NegotiateContext, bool) {
	salt := make([]byte, 32)
	rand.Read(salt)
	contexts := []*commands.NegotiateContext{commands.NewPreauthIntegrityContext(salt)}

	sha512 := false
	for _, ctx := range offered {
		switch ctx.ContextType {
		case commands.SMB2_PREAUTH_INTEGRITY_CAPABILITIES:
			sha512 = slices.Contains(listedIDs(ctx.Data, 0, 4), commands.SMB2_PREAUTH_HASH_SHA_512)
		case commands.SMB2_ENCRYPTION_CAPABILITIES:
			// A cipher of zero tells the client none was agreed.
			c.cipher, _ = pick(cipherPreference, listedIDs(ctx.Data, 0, 2))
			contexts = append(contexts, commands.NewEncryptionContext([]uint16{c.cipher}))
		case commands.SMB2_SIGNING_CAPABILITIES:
			if alg, ok := pick(signingPreference, listedIDs(ctx.Data, 0, 2)); ok {
				c.signingAlgorithm = int(alg)
				contexts = append(contexts, commands.NewSigningCapabilitiesContext([]uint16{alg}))
			}
		}
	}
	return contexts, sha512
}

// listedIDs reads a context's list of 16-bit identifiers: a count at countAt,
// the identifiers from listAt.
func listedIDs(data []byte, countAt, listAt int) []uint16 {
	if len(data) < countAt+2 {
		return nil
	}
	count := int(binary.LittleEndian.Uint16(data[countAt:]))
	var ids []uint16
	for i := 0; i < count && listAt+2*i+2 <= len(data); i++ {
		ids = append(ids, binary.LittleEndian.Uint16(data[listAt+2*i:]))
	}
	return ids
}

// pick returns the first of preference the client offered.
func pick(preference, offered []uint16) (uint16, bool) {
	for _, id := range preference {
		if slices.Contains(offered, id) {
			return id, true
		}
	}
	return 0, false
}

// negotiateResponse builds the NEGOTIATE response for a dialect revision.
func (c *Connection) negotiateResponse(revision dialects.Dialect, contexts []*commands.NegotiateContext) *commands.NegotiateResponse {
	config := c.server.config

	c.securityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	if config.RequireSigning {
		c.securityMode |= securitymode.SMB2_NEGOTIATE_SIGNING_REQUIRED
	}
	c.capabilities = 0
	if dialectRank[revision] >= dialectRank[dialects.SMB2_DIALECT_2_1_0] || revision == dialects.SMB2_DIALECT_WILDCARD {
		c.capabilities |= capabilities.SMB2_GLOBAL_CAP_LEASING | capabilities.SMB2_GLOBAL_CAP_LARGE_MTU
	}
	if (revision == dialects.SMB2_DIALECT_3_0_0 || revision == dialects.SMB2_DIALECT_3_0_2) && c.cipher != 0 {
		c.capabilities |= capabilities.SMB2_GLOBAL_CAP_ENCRYPTION
	}

	maxRead, maxWrite := config.MaxReadSize, config.MaxWriteSize
	if !c.capabilities.Has(capabilities.SMB2_GLOBAL_CAP_LARGE_MTU) {
		maxRead, maxWrite = min(maxRead, 1<<16), min(maxWrite, 1<<16)
	}

	resp := commands.NewNegotiateResponse()
	resp.SecurityMode = c.securityMode
	resp.DialectRevision = revision
	resp.ServerGuid = config.ServerGUID
	resp.Capabilities = c.capabilities
	resp.MaxTransactSize = max(maxRead, maxWrite)
	resp.MaxReadSize = maxRead
	resp.MaxWriteSize = maxWrite
	resp.SystemTime = filetime(time.Now())
	resp.SecurityBuffer = c.server.negotiateHint()
	resp.Contexts = contexts
	return resp
}

// negotiateHint is the SPNEGO token listing the mechanisms the server accepts,
// so a client knows which to start with.
func (s *Server) negotiateHint() []byte {
	mechs := []asn1.ObjectIdentifier{spnego.NtlmOID}
	if s.config.Kerberos != nil {
		mechs = []asn1.ObjectIdentifier{spnego.MSKerberosOID, spnego.KerberosOID, spnego.NtlmOID}
	}
	hint, err := spnego.NewNegTokenInit(mechs, asn1.BitString{}, nil, nil).Marshal()
	if err != nil {
		return nil
	}
	return hint
}

// validateNegotiateInfo answers FSCTL_VALIDATE_NEGOTIATE_INFO, with which a 3.0
// or 3.0.2 client checks over a signed channel that the negotiation was not
// tampered with (MS-SMB2 3.3.5.15.12). A mismatch drops the connection.
func (c *Connection) validateNegotiateInfo(input []byte) ([]byte, bool) {
	if c.dialect == dialects.SMB2_DIALECT_3_1_1 || len(input) < 24 {
		return nil, false
	}
	count := int(binary.LittleEndian.Uint16(input[22:24]))
	offered := listedIDs(input, 22, 24)
	if len(offered) != count ||
		capabilities.Capabilities(binary.LittleEndian.Uint32(input[0:4])) != c.clientCapabilities ||
		!bytes.Equal(input[4:20], c.clientGUID[:]) ||
		securitymode.SecurityMode(binary.LittleEndian.Uint16(input[20:22])) != c.clientSecurityMode {
		return nil, false
	}
	for i, d := range offered {
		if i >= len(c.clientDialects) || dialects.Dialect(d) != c.clientDialects[i] {
			return nil, false
		}
	}

	output := binary.LittleEndian.AppendUint32(nil, uint32(c.capabilities))
	output = append(output, c.server.config.ServerGUID[:]...)
	output = binary.LittleEndian.AppendUint16(output, uint16(c.securityMode))
	return binary.LittleEndian.AppendUint16(output, uint16(c.dialect)), true
}
