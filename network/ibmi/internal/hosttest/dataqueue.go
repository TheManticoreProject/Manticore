package hosttest

import (
	"bytes"
	"encoding/binary"
	"net"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// dataQueue is one in-memory queue of the scripted Data Queue server.
type dataQueue struct {
	kind       byte // 0 FIFO, 1 LIFO, 2 keyed
	maxEntry   uint32
	keyLength  uint16
	saveSender bool
	forceAux   bool
	desc       []byte
	entries    []dataEntry
}

type dataEntry struct{ data, key []byte }

// DataQueueEntries returns the entries of the queue named by its 10-byte
// queue and library fields, oldest first, or nil if it does not exist.
func (s *Server) DataQueueEntries(queue, library []byte) [][]byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	q := s.dataQueues[string(queue)+string(library)]
	if q == nil {
		return nil
	}
	out := [][]byte{}
	for _, e := range q.entries {
		out = append(out, append([]byte(nil), e.data...))
	}
	return out
}

func (s *Server) serveDataQueue(c net.Conn) {
	serverSeed := [8]byte{0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58}
	var clientSeed [8]byte
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
		if s.startup(c, f, &req, hostserver.ServerIDDataQueue, serverSeed, &clientSeed) {
			continue
		}
		s.record(req)
		if s.DataQueueReply != nil {
			if reply, ok := s.DataQueueReply(f); ok {
				write(c, reply)
				continue
			}
		}
		write(c, s.dataQueueReply(f))
	}
}

// dataQueueReply implements docs/16 against the in-memory queues. Command
// checks carry a CCSID 37 message that starts with the message ID, as the
// V5R4 server's do.
func (s *Server) dataQueueReply(f hostserver.Frame) hostserver.Frame {
	h := f.Header
	b := append(append([]byte(nil), f.Template...), f.Variable...)
	reply := func(id uint16, body []byte) hostserver.Frame {
		return hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDDataQueue, RequestID: id, CorrelationID: h.CorrelationID}, Variable: body}
	}
	common := func(rc uint16, msg string) hostserver.Frame {
		body := binary.BigEndian.AppendUint16(nil, rc)
		if msg != "" {
			codec, _ := ccsid.Lookup(ccsid.CCSID37)
			text, _ := codec.Encode(msg)
			body, _ = hostserver.AppendField(body, 0x1102, text)
		}
		return reply(0x8002, body)
	}
	if h.RequestID == 0x0000 {
		return reply(0x8000, make([]byte, 6))
	}
	if len(b) < 20 {
		return common(0xF003, "")
	}
	name := string(b[:20])
	var fields hostserver.Fields
	if at, ok := map[uint16]int{0x0002: 28, 0x0005: 22, 0x0006: 21}[h.RequestID]; ok {
		var err error
		if fields, err = hostserver.ParseFields(b[min(len(b), at):], hostserver.FieldPolicy{}); err != nil {
			return common(0xF002, "")
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dataQueues == nil {
		s.dataQueues = map[string]*dataQueue{}
	}
	q := s.dataQueues[name]
	if q == nil && h.RequestID != 0x0003 {
		return common(0xF001, "CPF2105 Object not found.")
	}
	switch h.RequestID {
	case 0x0001:
		body := make([]byte, 11, 61)
		binary.BigEndian.PutUint32(body[2:], q.maxEntry)
		body[6] = flagByte(q.saveSender)
		body[7] = 0xF0 | q.kind
		binary.BigEndian.PutUint16(body[8:], q.keyLength)
		body[10] = flagByte(q.forceAux)
		return reply(0x8001, append(body, q.desc...))
	case 0x0002:
		keyed := b[20] == 0xF1
		if keyed != (q.kind == 2) {
			return common(0xF001, "CPF9506 Wrong queue type.")
		}
		peek := b[27] == 0xF1
		search := string(b[21:23])
		key, _ := fields.Get(0x5002)
		i := -1
		for j := range q.entries {
			k := j
			if q.kind == 1 {
				k = len(q.entries) - 1 - j
			}
			if !keyed || compareKey(q.entries[k].key, key, search) {
				i = k
				break
			}
		}
		if i < 0 {
			return common(0xF006, "")
		}
		e := q.entries[i]
		if !peek {
			q.entries = append(q.entries[:i], q.entries[i+1:]...)
		}
		body := make([]byte, 2, 38)
		if q.saveSender {
			codec, _ := ccsid.Lookup(ccsid.CCSID37)
			for _, v := range []struct {
				s string
				n int
			}{{s.JobName, 10}, {"QUSER", 10}, {"000042", 6}, {s.Profile, 10}} {
				enc, _ := ccsid.EncodeFixed(codec, v.s, v.n)
				body = append(body, enc...)
			}
		} else {
			body = append(body, bytes.Repeat([]byte{0x40}, 36)...)
		}
		body, _ = hostserver.AppendField(body, 0x5001, e.data)
		if keyed {
			body, _ = hostserver.AppendField(body, 0x5002, e.key)
		}
		return reply(0x8003, body)
	case 0x0003:
		if q != nil {
			return common(0xF001, "CPF9870 Object already exists.")
		}
		s.dataQueues[name] = &dataQueue{
			maxEntry: binary.BigEndian.Uint32(b[20:]), saveSender: b[25] == 0xF1, kind: b[26] & 0x0F,
			keyLength: binary.BigEndian.Uint16(b[27:]), forceAux: b[29] == 0xF1, desc: append([]byte(nil), b[30:80]...),
		}
	case 0x0004:
		delete(s.dataQueues, name)
	case 0x0005:
		data, _ := fields.Get(0x5001)
		key, keyed := fields.Get(0x5002)
		switch {
		case keyed != (q.kind == 2):
			return common(0xF001, "CPF9506 Wrong queue type.")
		case len(data) == 0 || uint32(len(data)) > q.maxEntry || keyed && len(key) != int(q.keyLength):
			return common(0xF001, "CPF2498 Entry length not valid.")
		}
		q.entries = append(q.entries, dataEntry{append([]byte(nil), data...), append([]byte(nil), key...)})
	case 0x0006:
		key, keyed := fields.Get(0x5002)
		kept := q.entries[:0]
		for _, e := range q.entries {
			if keyed && !bytes.Equal(e.key, key) {
				kept = append(kept, e)
			}
		}
		q.entries = kept
	default:
		return common(0xF002, "")
	}
	return common(0xF000, "")
}

func flagByte(b bool) byte {
	if b {
		return 0xF1
	}
	return 0xF0
}

// compareKey applies a CCSID 37 search operator bytewise.
func compareKey(entry, key []byte, search string) bool {
	c := bytes.Compare(entry, key)
	switch search {
	case "\xC5\xD8": // EQ
		return c == 0
	case "\xD5\xC5": // NE
		return c != 0
	case "\xD3\xE3": // LT
		return c < 0
	case "\xD3\xC5": // LE
		return c <= 0
	case "\xC7\xE3": // GT
		return c > 0
	case "\xC7\xC5": // GE
		return c >= 0
	}
	return false
}
