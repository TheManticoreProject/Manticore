// Package hosttest provides scripted, in-memory IBM i Sign-on, Remote
// Command, Data Queue and File host servers for tests of the network/ibmi
// packages. It is test
// infrastructure only: production code must never import it.
//
// A Server is reached through its Dial method (an ibminet.DialFunc), which
// routes by port and returns one side of a net.Pipe. The servers verify
// every password substitute independently from their own copy of the
// credential and the seeds they saw, record every request, and track open
// connections so tests can assert that sockets are closed.
package hosttest

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/auth"
)

// Default ports the Dial method routes on.
const (
	SignonPort        = 8476
	RemoteCommandPort = 8475
	DataQueuePort     = 8472
	FilePort          = 8473
)

// Request records one frame a server received.
type Request struct {
	ServerID      uint16
	RequestID     uint16
	CorrelationID uint32
	// ClientSeed is set for 0x7003 and 0x7001.
	ClientSeed [8]byte
	// Substitute is set for 0x7004 and 0x7002.
	Substitute []byte
	// Attributes is header byte 4 (client attributes).
	Attributes byte
	// CredentialType is the first template byte of 0x7004 and 0x7002.
	CredentialType byte
}

// Server scripts both host servers. Configure the exported fields before
// the first Dial; zero values select a healthy V5R4-like server.
type Server struct {
	Profile  string
	Password string
	// PasswordLevel is reported in CP 0x1119 (default 0).
	PasswordLevel byte
	// SignonRC and StartRC force a 0xF004 or 0xF002 return code. A wrong
	// substitute yields 0x0003000B regardless.
	SignonRC uint32
	StartRC  uint32
	// AttributesRC is the 0x8001 return code.
	AttributesRC uint16
	// SystemCCSID (0xF004) and JobCCSID (0x8001) default to 37.
	SystemCCSID uint32
	JobCCSID    uint32
	// RemoteCommandLevel is the 0x8001 data-stream level (default 7).
	RemoteCommandLevel uint16
	// JobName is returned in 0xF002 CP 0x111F (default QZRCSRVS).
	JobName string
	// Handle answers Remote Command requests other than the startup ones.
	// Returning ok=false sends no reply.
	Handle func(req hostserver.Frame) (reply hostserver.Frame, ok bool)
	// SilentOn suppresses the reply to this request ID.
	SilentOn uint16
	// AttributesCorrelationDelta is added to the 0x8001 correlation ID, to
	// script a mismatched reply.
	AttributesCorrelationDelta uint32

	// TokenRC forces a 0xF007 return code.
	TokenRC uint32

	// DataQueueReply, when set, answers Data Queue requests after startup
	// in place of the in-memory queues; ok=false falls through to them.
	DataQueueReply func(req hostserver.Frame) (reply hostserver.Frame, ok bool)
	dataQueues     map[string]*dataQueue

	// Files maps IFS paths to their content for the File server.
	Files map[string][]byte
	// FileLevel is the 0x8009 data-stream level (default 12, what V5R4
	// returns; V7R5 returns 24); FilePathCCSID the pathname CCSID (default
	// 1200). Like a real server, the fake lays packets out by the level the
	// client requested and rejects a level-8 open after a request of 16+.
	FileLevel     uint16
	FilePathCCSID uint16
	// FileChunk, when nonzero, splits read replies into chained packets of
	// at most this many bytes.
	FileChunk int
	// FileEOFReturnCode signals end of file with 0x8001 return code 22
	// instead of an empty 0x8003.
	FileEOFReturnCode bool
	// FileReply, when set, answers File requests after startup in place
	// of the in-memory files; ok=false falls through to them.
	FileReply func(req hostserver.Frame) (replies []hostserver.Frame, ok bool)

	mu       sync.Mutex
	requests []Request
	open     int
	wg       sync.WaitGroup
	// tokens maps issued profile tokens to their state.
	tokens map[string]*issuedToken
}

type issuedToken struct {
	singleUse bool
	used      bool
}

// tokenRC checks a presented profile token, consuming a single-use one.
func (s *Server) tokenRC(token []byte) uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tokens[string(token)]
	if !ok || (t.singleUse && t.used) {
		return 0x00060001 // profile token not valid
	}
	t.used = true
	return 0
}

// Tokens returns how many profile tokens the server issued.
func (s *Server) Tokens() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.tokens)
}

func (s *Server) defaults() {
	if s.SystemCCSID == 0 {
		s.SystemCCSID = 37
	}
	if s.JobCCSID == 0 {
		s.JobCCSID = 37
	}
	if s.RemoteCommandLevel == 0 {
		s.RemoteCommandLevel = 7
	}
	if s.JobName == "" {
		s.JobName = "QZRCSRVS"
	}
	if s.FileLevel == 0 {
		s.FileLevel = 12
	}
	if s.FilePathCCSID == 0 {
		s.FilePathCCSID = 1200
	}
}

// Dial is an ibminet.DialFunc routing SignonPort and RemoteCommandPort to
// the scripted servers. Any other port fails.
func (s *Server) Dial(ctx context.Context, network, address string) (net.Conn, error) {
	_, portText, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	port, _ := strconv.Atoi(portText)
	var serve func(net.Conn)
	switch port {
	case SignonPort:
		serve = s.serveSignon
	case RemoteCommandPort:
		serve = s.serveRemoteCommand
	case DataQueuePort:
		serve = s.serveDataQueue
	case FilePort:
		serve = s.serveFile
	default:
		return nil, fmt.Errorf("hosttest: nothing listens on port %d", port)
	}
	s.mu.Lock()
	s.defaults()
	s.open++
	s.mu.Unlock()
	client, server := net.Pipe()
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer func() {
			server.Close()
			s.mu.Lock()
			s.open--
			s.mu.Unlock()
		}()
		serve(server)
	}()
	return client, nil
}

// Requests returns a copy of every recorded request.
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Request(nil), s.requests...)
}

// Count returns how many requests with id were received.
func (s *Server) Count(id uint16) int {
	n := 0
	for _, r := range s.Requests() {
		if r.RequestID == id {
			n++
		}
	}
	return n
}

// WaitClosed waits until every connection the client opened has been closed
// by the client, or fails after timeout.
func (s *Server) WaitClosed(timeout time.Duration) error {
	done := make(chan struct{})
	go func() { s.wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		s.mu.Lock()
		defer s.mu.Unlock()
		return fmt.Errorf("hosttest: %d connection(s) still open", s.open)
	}
}

func (s *Server) record(r Request) {
	s.mu.Lock()
	s.requests = append(s.requests, r)
	s.mu.Unlock()
}

func (s *Server) credential() *auth.PasswordCredential {
	c, err := auth.NewPasswordCredential(s.Profile, []byte(s.Password))
	if err != nil {
		panic("hosttest: invalid scripted credential: " + err.Error())
	}
	return c
}

func (s *Server) family() auth.Family {
	f, err := auth.FamilyForLevel(s.PasswordLevel)
	if err != nil {
		return auth.FamilyDES
	}
	return f
}

// readFrame reads one frame; io.EOF means the client closed the socket.
func readFrame(c net.Conn) (hostserver.Frame, error) {
	var hb [hostserver.HeaderLength]byte
	if _, err := io.ReadFull(c, hb[:]); err != nil {
		return hostserver.Frame{}, err
	}
	b := make([]byte, binary.BigEndian.Uint32(hb[:]))
	copy(b, hb[:])
	if _, err := io.ReadFull(c, b[hostserver.HeaderLength:]); err != nil {
		return hostserver.Frame{}, err
	}
	return hostserver.UnmarshalFrame(b, 0)
}

func write(c net.Conn, f hostserver.Frame) {
	b, err := f.Marshal(0)
	if err != nil {
		panic("hosttest: " + err.Error())
	}
	_, _ = c.Write(b)
}

// Reply builds a reply frame with a 4-byte template and LL/CP fields.
func Reply(serverID, replyID uint16, correlation uint32, template []byte, fields ...hostserver.Field) hostserver.Frame {
	var v []byte
	for _, f := range fields {
		v, _ = hostserver.AppendField(v, f.CodePoint, f.Value)
	}
	return hostserver.Frame{
		Header:   hostserver.Header{ServerID: serverID, RequestID: replyID, CorrelationID: correlation},
		Template: template, Variable: v,
	}
}

func u32(v uint32) []byte { return binary.BigEndian.AppendUint32(nil, v) }

func (s *Server) serveSignon(c net.Conn) {
	cred := s.credential()
	defer cred.Destroy()
	serverSeed := [8]byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38}
	var clientSeed [8]byte
	for {
		f, err := readFrame(c)
		if err != nil {
			return
		}
		req := Request{ServerID: f.Header.ServerID, RequestID: f.Header.RequestID, CorrelationID: f.Header.CorrelationID}
		if f.Header.RequestID == 0x7004 && len(f.Template) > 0 {
			req.CredentialType = f.Template[0]
		}
		switch f.Header.RequestID {
		case 0x7003:
			fs, _ := hostserver.ParseFields(f.Variable, hostserver.FieldPolicy{})
			seed, _ := fs.Get(0x1103)
			copy(clientSeed[:], seed)
			req.ClientSeed = clientSeed
			s.record(req)
			if s.SilentOn == 0x7003 {
				continue
			}
			write(c, Reply(hostserver.ServerIDSignon, 0xF003, 0, u32(0),
				hostserver.Field{CodePoint: 0x1101, Value: u32(0x00050400)},
				hostserver.Field{CodePoint: 0x1102, Value: []byte{0, 2}},
				hostserver.Field{CodePoint: 0x1103, Value: serverSeed[:]},
				hostserver.Field{CodePoint: 0x1119, Value: []byte{s.PasswordLevel}}))
		case 0x7004:
			fs, _ := hostserver.ParseFields(f.Variable, hostserver.FieldPolicy{})
			rc := s.SignonRC
			if req.CredentialType == 0x02 {
				token, _ := fs.Get(0x1115)
				req.Substitute = append([]byte(nil), token...)
				s.record(req)
				if _, hasProfile := fs.Get(0x1104); rc == 0 && hasProfile {
					rc = 0x00010010 // a token request carries no profile
				}
				if rc == 0 {
					rc = s.tokenRC(token)
				}
				write(c, Reply(hostserver.ServerIDSignon, 0xF004, 0, u32(rc),
					hostserver.Field{CodePoint: 0x1114, Value: u32(s.SystemCCSID)}))
				continue
			}
			got, _ := fs.Get(0x1105)
			req.Substitute = append([]byte(nil), got...)
			s.record(req)
			if want, err := auth.Derive(cred, s.family(), clientSeed, serverSeed); rc == 0 && (err != nil || !bytes.Equal(got, want.Value)) {
				rc = 0x0003000B
			}
			write(c, Reply(hostserver.ServerIDSignon, 0xF004, 0, u32(rc),
				hostserver.Field{CodePoint: 0x1114, Value: u32(s.SystemCCSID)}))
		case 0x7007:
			fs, _ := hostserver.ParseFields(f.Variable, hostserver.FieldPolicy{})
			got, _ := fs.Get(0x1105)
			typ, _ := fs.Get(0x1116)
			req.Substitute = append([]byte(nil), got...)
			req.CredentialType = f.Template[0]
			s.record(req)
			rc := s.TokenRC
			if want, err := auth.Derive(cred, s.family(), clientSeed, serverSeed); rc == 0 && (err != nil || !bytes.Equal(got, want.Value)) {
				rc = 0x0003000B
			}
			if rc != 0 {
				write(c, Reply(hostserver.ServerIDSignon, 0xF007, 0, u32(rc)))
				continue
			}
			token := make([]byte, 32)
			_, _ = rand.Read(token)
			s.mu.Lock()
			if s.tokens == nil {
				s.tokens = map[string]*issuedToken{}
			}
			s.tokens[string(token)] = &issuedToken{singleUse: len(typ) == 1 && typ[0] == 0xF1}
			s.mu.Unlock()
			// Return code, a 6-byte prefix (CP 0x1115 envelope), the token.
			body := append(u32(0), 0, 0, 0, 0x26, 0x11, 0x15)
			write(c, hostserver.Frame{
				Header:   hostserver.Header{ServerID: hostserver.ServerIDSignon, RequestID: 0xF007},
				Template: body[:4], Variable: append(body[4:], token...),
			})
		default:
			s.record(req) // 0x7006 and anything else: no reply
		}
	}
}

// startup answers 0x7001 and 0x7002 for serverID, verifying the
// substitute or token independently; it reports whether f was one of them.
func (s *Server) startup(c net.Conn, f hostserver.Frame, req *Request, serverID uint16, serverSeed [8]byte, clientSeed *[8]byte) bool {
	switch f.Header.RequestID {
	case 0x7001:
		copy(clientSeed[:], f.Template)
		req.ClientSeed = *clientSeed
		s.record(*req)
		write(c, Reply(serverID, 0xF001, 0, append(u32(0), serverSeed[:]...)))
	case 0x7002:
		cred := s.credential()
		defer cred.Destroy()
		fs, _ := hostserver.ParseFields(f.Variable, hostserver.FieldPolicy{})
		rc := s.StartRC
		if req.CredentialType == 0x02 {
			token, _ := fs.Get(0x1115)
			req.Substitute = append([]byte(nil), token...)
			if _, hasProfile := fs.Get(0x1104); rc == 0 && hasProfile {
				rc = 0x00010010
			}
			if rc == 0 {
				rc = s.tokenRC(token)
			}
		} else {
			got, _ := fs.Get(0x1105)
			req.Substitute = append([]byte(nil), got...)
			if want, err := auth.Derive(cred, s.family(), *clientSeed, serverSeed); rc == 0 && (err != nil || !bytes.Equal(got, want.Value)) {
				rc = 0x0003000B
			}
		}
		s.record(*req)
		codec, _ := ccsid.Lookup(ccsid.CCSID37)
		job, _ := ccsid.EncodeFixed(codec, s.JobName, 28)
		write(c, Reply(serverID, 0xF002, 0, u32(rc),
			hostserver.Field{CodePoint: 0x111F, Value: append([]byte{0, 0, 0, 0x25}, job...)}))
	default:
		return false
	}
	return true
}

func (s *Server) serveRemoteCommand(c net.Conn) {
	serverSeed := [8]byte{0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48}
	var clientSeed [8]byte
	for {
		f, err := readFrame(c)
		if err != nil {
			return
		}
		h := f.Header
		req := Request{ServerID: h.ServerID, RequestID: h.RequestID, CorrelationID: h.CorrelationID, Attributes: byte(h.Attributes >> 8)}
		if len(f.Template) > 0 && (h.RequestID == 0x7002) {
			req.CredentialType = f.Template[0]
		}
		if h.RequestID == s.SilentOn && s.SilentOn != 0 {
			s.record(req)
			continue
		}
		if s.startup(c, f, &req, hostserver.ServerIDRemoteCommand, serverSeed, &clientSeed) {
			continue
		}
		switch h.RequestID {
		case 0x1001:
			s.record(req)
			body := make([]byte, 16)
			binary.BigEndian.PutUint16(body[0:], s.AttributesRC)
			binary.BigEndian.PutUint32(body[2:], s.JobCCSID)
			binary.BigEndian.PutUint16(body[14:], s.RemoteCommandLevel)
			write(c, hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDRemoteCommand, RequestID: 0x8001, CorrelationID: h.CorrelationID + s.AttributesCorrelationDelta}, Variable: body})
		default:
			s.record(req)
			if s.Handle != nil {
				if reply, ok := s.Handle(f); ok {
					write(c, reply)
				}
			}
		}
	}
}

// OperationReply builds a 0x8002/0x8003/0x9003 reply: return code, message
// count, then rest (message records or output envelopes).
func OperationReply(replyID uint16, correlation uint32, rc, count uint16, rest []byte) hostserver.Frame {
	t := binary.BigEndian.AppendUint16(nil, rc)
	t = binary.BigEndian.AppendUint16(t, count)
	return hostserver.Frame{
		Header:   hostserver.Header{ServerID: hostserver.ServerIDRemoteCommand, RequestID: replyID, CorrelationID: correlation},
		Template: t, Variable: rest,
	}
}

// MessageRecord builds a classic CCSID 37 message record. The message type
// is written as two zoned digits.
func MessageRecord(id string, messageType uint8, severity uint16, file, library string, substitution []byte, text string) []byte {
	codec, _ := ccsid.Lookup(ccsid.CCSID37)
	enc := func(s string, width int) []byte {
		b, err := ccsid.EncodeFixed(codec, s, width)
		if err != nil {
			panic("hosttest: " + err.Error())
		}
		return b
	}
	textBytes, err := codec.Encode(text)
	if err != nil {
		panic("hosttest: " + err.Error())
	}
	r := binary.BigEndian.AppendUint32(nil, uint32(41+len(substitution)+len(textBytes)))
	r = binary.BigEndian.AppendUint16(r, 0)
	r = append(r, enc(id, 7)...)
	r = append(r, 0xF0|messageType/10, 0xF0|messageType%10)
	r = binary.BigEndian.AppendUint16(r, severity)
	r = append(r, enc(file, 10)...)
	r = append(r, enc(library, 10)...)
	r = binary.BigEndian.AppendUint16(r, uint16(len(substitution)))
	r = binary.BigEndian.AppendUint16(r, uint16(len(textBytes)))
	return append(append(r, substitution...), textBytes...)
}

// OutputEnvelope builds one returned parameter envelope.
func OutputEnvelope(usage uint16, length uint32, payload []byte) []byte {
	b := binary.BigEndian.AppendUint32(nil, uint32(12+len(payload)))
	b = binary.BigEndian.AppendUint16(b, 0x1103)
	b = binary.BigEndian.AppendUint32(b, length)
	b = binary.BigEndian.AppendUint16(b, usage)
	return append(b, payload...)
}
