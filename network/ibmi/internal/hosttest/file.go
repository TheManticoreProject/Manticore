package hosttest

import (
	"encoding/binary"
	"net"
	"unicode/utf16"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

func (s *Server) serveFile(c net.Conn) {
	serverSeed := [8]byte{0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68}
	var clientSeed [8]byte
	open := map[uint32][]byte{} // handle -> content
	next := uint32(0x100)
	var requested uint16 // the level asked for in 0x0016
	for {
		f, err := readFrame(c)
		if err != nil {
			return
		}
		h := f.Header
		req := Request{ServerID: h.ServerID, RequestID: h.RequestID, CorrelationID: h.CorrelationID, Attributes: byte(h.Attributes >> 8)}
		if len(f.Template) > 0 && h.RequestID == 0x7002 {
			req.CredentialType = f.Template[0]
		}
		if h.RequestID == s.SilentOn && s.SilentOn != 0 {
			s.record(req)
			continue
		}
		if s.startup(c, f, &req, hostserver.ServerIDFile, serverSeed, &clientSeed) {
			continue
		}
		s.record(req)
		if s.FileReply != nil {
			if replies, ok := s.FileReply(f); ok {
				for _, r := range replies {
					write(c, r)
				}
				continue
			}
		}
		if h.RequestID == 0x0016 && len(f.Template) >= 4 {
			requested = binary.BigEndian.Uint16(f.Template[2:])
		}
		for _, r := range s.fileReplies(f, open, &next, requested) {
			write(c, r)
		}
	}
}

// fileReplies implements the docs/08 level-below-16 read path against
// s.Files, with the reply forms observed on V5R4.
func (s *Server) fileReplies(f hostserver.Frame, open map[uint32][]byte, next *uint32, requested uint16) []hostserver.Frame {
	h := f.Header
	b := append(append([]byte(nil), f.Template...), f.Variable...)
	reply := func(id uint16, body []byte) hostserver.Frame {
		return hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDFile, RequestID: id, CorrelationID: h.CorrelationID}, Variable: body}
	}
	rc := func(code uint16) []hostserver.Frame {
		return []hostserver.Frame{reply(0x8001, []byte{0, 0, byte(code >> 8), byte(code)})}
	}
	switch h.RequestID {
	case 0x0016:
		body := binary.BigEndian.AppendUint16([]byte{0, 0}, s.FileLevel)
		body = append(body, 0, 0, 0x40, 0, 0, 0)
		body = append(body, 0, 0, 0, 8, 0, 0x0A)
		return []hostserver.Frame{reply(0x8009, binary.BigEndian.AppendUint16(body, s.FilePathCCSID))}
	case 0x0002:
		if len(b) < 42 || requested >= 16 && len(f.Template) == 36 {
			return rc(17)
		}
		fields, err := hostserver.ParseFields(b[36:], hostserver.FieldPolicy{})
		name, ok := fields.Get(0x0002)
		if err != nil || !ok || len(name)%2 != 0 {
			return rc(17)
		}
		units := make([]uint16, len(name)/2)
		for i := range units {
			units[i] = binary.BigEndian.Uint16(name[2*i:])
		}
		s.mu.Lock()
		content, exists := s.Files[string(utf16.Decode(units))]
		s.mu.Unlock()
		if !exists {
			return rc(2)
		}
		handle := *next
		*next++
		open[handle] = content
		body := make([]byte, 69)
		binary.BigEndian.PutUint32(body[2:], handle)
		binary.BigEndian.PutUint64(body[6:], uint64(handle))
		binary.BigEndian.PutUint16(body[14:], 819)
		binary.BigEndian.PutUint32(body[26:], 1700000000)
		binary.BigEndian.PutUint32(body[42:], uint32(len(content)))
		return []hostserver.Frame{reply(0x8002, body)}
	case 0x0003:
		content, ok := open[binary.BigEndian.Uint32(b[2:])]
		if !ok {
			return rc(6)
		}
		off, n := int(binary.BigEndian.Uint32(b[10:])), int(binary.BigEndian.Uint32(b[14:]))
		if off >= len(content) {
			if s.FileEOFReturnCode {
				return rc(22)
			}
			return []hostserver.Frame{readPacket(reply, false, nil)}
		}
		data := content[off:min(len(content), off+n)]
		chunk := len(data)
		if s.FileChunk > 0 {
			chunk = s.FileChunk
		}
		var out []hostserver.Frame
		for len(data) > chunk {
			out = append(out, readPacket(reply, true, data[:chunk]))
			data = data[chunk:]
		}
		return append(out, readPacket(reply, false, data))
	case 0x0009:
		handle := binary.BigEndian.Uint32(b[2:])
		code := uint16(0)
		if _, ok := open[handle]; !ok {
			code = 6
		}
		delete(open, handle)
		return []hostserver.Frame{reply(0x8004, []byte{0, 0, byte(code >> 8), byte(code), 0, 0, 0, 0})}
	}
	return rc(23)
}

// readPacket builds one 0x8003: chain, CCSID 0xFFFF and the data field
// under code point 0x0005, as V5R4 sends them.
func readPacket(reply func(uint16, []byte) hostserver.Frame, chained bool, data []byte) hostserver.Frame {
	body := []byte{0, 0, 0xFF, 0xFF}
	if chained {
		body[1] = 1
	}
	body, _ = hostserver.AppendField(body, 0x0005, data)
	return reply(0x8003, body)
}
