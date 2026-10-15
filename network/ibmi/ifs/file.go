package ifs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/ifs/message"
)

// File is an open stream file. Its handle belongs to the Session's server
// job. Close it when done; closing the Session also releases it.
type File struct {
	s      *Session
	handle uint32
	closed bool // guarded by s.mu
	info   FileInfo
}

// FileInfo is the metadata returned by open.
type FileInfo struct {
	// Size is the file size at open, from the 32-bit level-8 field.
	Size int64
	// DataCCSID is the file's data CCSID; the bytes read are not converted.
	DataCCSID uint16
	FileID    uint64
	Created   time.Time
	Modified  time.Time
	Accessed  time.Time
}

// Info returns the metadata returned by open.
func (f *File) Info() FileInfo { return f.info }

// Open opens an existing stream file for reading, sharing it with every
// other reader and writer, without data conversion. The path is sent
// exactly as given. A missing file fails with ErrNotFound.
func (s *Session) Open(ctx context.Context, path string) (*File, error) {
	const op = "open"
	p, err := encodePath(path)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	corr, err := s.corr.Next()
	if err != nil {
		return nil, err
	}
	req, err := message.OpenRequest{
		Correlation: corr, PathCCSID: s.info.PathCCSID, DataCCSID: s.info.PathCCSID,
		Access: message.AccessRead, Deny: message.DenyNone, Conversion: message.ConvertNone,
		Duplicate: message.OpenExistingOnly, Path: p,
	}.Frame()
	if err != nil {
		return nil, fmt.Errorf("ifs: %s: %w", op, err)
	}
	reply, err := s.exchange(ctx, op, req, message.ReplyOpen, message.ReplyReturnCode)
	if err != nil {
		return nil, err
	}
	if reply.Header.RequestID == message.ReplyReturnCode {
		if err := s.returnCode(op, reply); err != nil {
			return nil, err
		}
		return nil, s.replyOutcome(op, &hostserver.MalformedError{What: "reply", Offset: 18, Reason: "return code 0 in place of an open reply"})
	}
	r, err := message.ParseOpenReply(reply)
	if err != nil {
		return nil, s.replyOutcome(op, err)
	}
	return &File{s: s, handle: r.Handle, info: FileInfo{
		Size: int64(r.Size), DataCCSID: r.DataCCSID, FileID: r.FileID,
		Created: r.Created, Modified: r.Modified, Accessed: r.Accessed,
	}}, nil
}

// ReadAt reads up to n bytes (1 to StartOptions.MaxChunk) at offset (0 to
// 2^31-1). The server may return fewer bytes than requested. At end of
// file it returns the bytes received, possibly none, together with io.EOF.
func (f *File) ReadAt(ctx context.Context, offset int64, n int) ([]byte, error) {
	const op = "read"
	s := f.s
	if n < 1 || n > s.maxChunk {
		return nil, fmt.Errorf("ifs: %s: length %d outside 1-%d", op, n, s.maxChunk)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if f.closed {
		return nil, ErrFileClosed
	}
	corr, err := s.corr.Next()
	if err != nil {
		return nil, err
	}
	req, err := message.ReadFrame(corr, f.handle, offset, uint32(n))
	if err != nil {
		return nil, fmt.Errorf("ifs: %s: %w", op, err)
	}
	reply, err := s.exchange(ctx, op, req, message.ReplyRead, message.ReplyReturnCode)
	if err != nil {
		return nil, err
	}

	// Consume the whole chain before returning, whatever it carries.
	var (
		data    []byte
		eof     bool
		failure error
	)
	for packets := 1; ; packets++ {
		chained := message.Chained(reply)
		switch reply.Header.RequestID {
		case message.ReplyRead:
			r, err := message.ParseReadReply(reply)
			if err != nil {
				return nil, s.replyOutcome(op, err)
			}
			switch {
			case len(r.Data) == 0:
				eof = true
			case eof:
				return nil, s.replyOutcome(op, &hostserver.MalformedError{What: "0x8003 reply", Offset: hostserver.HeaderLength, Reason: "data after end of file"})
			case len(data)+len(r.Data) > n:
				return nil, s.replyOutcome(op, &hostserver.MalformedError{What: "0x8003 reply", Offset: hostserver.HeaderLength, Reason: fmt.Sprintf("more than the %d bytes requested", n)})
			default:
				data = append(data, r.Data...)
			}
		default:
			rc, err := message.ParseReturnCode(reply)
			if err != nil {
				return nil, s.replyOutcome(op, err)
			}
			switch rc {
			case RCNoMoreData:
				eof = true
			case RCSuccess:
			default:
				if failure == nil {
					failure = &Error{Operation: op, ReturnCode: rc}
				}
			}
		}
		if !chained {
			break
		}
		if packets == s.maxChain {
			return nil, s.replyOutcome(op, fmt.Errorf("ifs: %s: reply chain longer than %d packets", op, s.maxChain))
		}
		if reply, err = s.receiveChained(ctx, corr); err != nil {
			return nil, s.replyOutcome(op, err)
		}
	}
	if failure != nil {
		return nil, failure
	}
	if eof {
		return data, io.EOF
	}
	if data == nil {
		data = []byte{}
	}
	return data, nil
}

// receiveChained reads the next packet of a reply chain and checks that it
// continues the chain for corr.
func (s *Session) receiveChained(ctx context.Context, corr uint32) (hostserver.Frame, error) {
	f, err := s.conn.Receive(ctx)
	if err != nil {
		return hostserver.Frame{}, err
	}
	h := f.Header
	if h.ServerID != hostserver.ServerIDFile || h.CorrelationID != corr || (h.RequestID != message.ReplyRead && h.RequestID != message.ReplyReturnCode) {
		return hostserver.Frame{}, fmt.Errorf("%w: chained reply 0x%04X, server 0x%04X, correlation %d (want %d)", hostserver.ErrUnexpectedReply, h.RequestID, h.ServerID, h.CorrelationID, corr)
	}
	return f, nil
}

// Close releases the file handle. It is sent once; later calls return nil.
// On a session that is already unusable it only marks the File closed.
func (f *File) Close(ctx context.Context) error {
	const op = "close"
	s := f.s
	s.mu.Lock()
	defer s.mu.Unlock()
	if f.closed {
		return nil
	}
	f.closed = true
	if s.conn.Err() != nil {
		return nil // the server job, and with it the handle, is gone
	}
	corr, err := s.corr.Next()
	if err != nil {
		return err
	}
	reply, err := s.exchange(ctx, op, message.CloseFrame(corr, f.handle), message.ReplyClose, message.ReplyReturnCode)
	if err != nil {
		return err
	}
	return s.returnCode(op, reply)
}

// ReadFile opens path, reads it to end of file and closes it. limit (at
// least 1) caps the content: a longer file fails with ErrTooLarge, which is
// detected by reading at most limit+1 bytes whatever size open reported.
func (s *Session) ReadFile(ctx context.Context, path string, limit int) ([]byte, error) {
	if limit < 1 || limit >= message.MaxOffset {
		return nil, fmt.Errorf("ifs: limit %d outside 1-%d", limit, message.MaxOffset-1)
	}
	f, err := s.Open(ctx, path)
	if err != nil {
		return nil, err
	}
	data, err := readAll(ctx, f, limit)
	if cerr := f.Close(ctx); err == nil {
		err = cerr
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func readAll(ctx context.Context, f *File, limit int) ([]byte, error) {
	if f.info.Size > int64(limit) {
		return nil, ErrTooLarge
	}
	var data []byte
	for {
		n := min(f.s.maxChunk, limit+1-len(data))
		chunk, err := f.ReadAt(ctx, int64(len(data)), n)
		data = append(data, chunk...)
		if len(data) > limit {
			return nil, ErrTooLarge
		}
		if errors.Is(err, io.EOF) {
			return data, nil
		}
		if err != nil {
			return nil, err
		}
		if len(chunk) == 0 {
			return nil, f.s.protocolStall()
		}
	}
}

// protocolStall reports a read that returned no data without end of file,
// which would otherwise loop forever.
func (s *Session) protocolStall() error {
	return s.replyOutcome("read", &hostserver.MalformedError{What: "0x8003 reply", Offset: hostserver.HeaderLength, Reason: "no data and no end of file"})
}
