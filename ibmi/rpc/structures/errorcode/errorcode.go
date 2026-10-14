// Package errorcode implements the IBM i API error-code parameter shared by
// many unrelated programs.
//
// Two modes exist:
//
//   - Exception mode (bytes provided = 0): a failure is raised as an escape
//     message, which the transport reports as an IBM message. On the wire
//     this uses the JTOpen 4.8 legacy form, an input-only parameter of eight
//     zero bytes, matching the V5R4 golden traffic.
//   - Buffer mode (ERRC0100 or ERRC0200): an input/output allocation whose
//     bytes-provided field equals its size. On failure the program fills it
//     and returns normally; Check turns a nonzero bytes-available into an
//     *APIError.
//
// An *APIError is an IBM API result, distinct from a Remote Command return
// code (rmtcmd.NativeError) and from transport failures.
//
// Layouts: IBM i "Error code parameter" (ERRC0100 and ERRC0200 tables),
// https://www.ibm.com/docs/en/i/7.6.0?topic=parameter-error-code-format
// (retrieved 2026-10-14).
package errorcode

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
)

// Format selects the error-code layout.
type Format uint8

const (
	formatException Format = iota
	FormatERRC0100
	FormatERRC0200
)

func (f Format) String() string {
	switch f {
	case formatException:
		return "exception"
	case FormatERRC0100:
		return "ERRC0100"
	case FormatERRC0200:
		return "ERRC0200"
	}
	return fmt.Sprintf("Format(%d)", uint8(f))
}

// Minimum buffer sizes accepted here: large enough for the fixed fields
// through the exception ID (IBM accepts smaller ERRC0100 = 8 and ERRC0200 =
// 12 buffers, which return only bytes available).
const (
	MinERRC0100 uint32 = 16
	MinERRC0200 uint32 = 32
	// MaxSize is the largest useful buffer: 32 KiB of exception data plus
	// the ERRC0200 fixed fields.
	MaxSize uint32 = 32767 + 32
)

// legacyExceptionLength is the size of the JTOpen 4.8 exception-mode
// parameter: eight zero bytes, sent input-only.
const legacyExceptionLength = 8

// Parameter is one error-code parameter. The zero value is exception mode.
type Parameter struct {
	format Format
	size   uint32
}

// Exception returns exception mode (bytes provided = 0).
func Exception() Parameter { return Parameter{} }

// ERRC0100 returns a buffer-mode parameter of size bytes.
func ERRC0100(size uint32) (Parameter, error) {
	if size < MinERRC0100 || size > MaxSize {
		return Parameter{}, fmt.Errorf("errorcode: ERRC0100 size %d outside %d..%d", size, MinERRC0100, MaxSize)
	}
	return Parameter{format: FormatERRC0100, size: size}, nil
}

// ERRC0200 returns a buffer-mode parameter of size bytes with key -1 (CCHAR
// support requested).
func ERRC0200(size uint32) (Parameter, error) {
	if size < MinERRC0200 || size > MaxSize {
		return Parameter{}, fmt.Errorf("errorcode: ERRC0200 size %d outside %d..%d", size, MinERRC0200, MaxSize)
	}
	return Parameter{format: FormatERRC0200, size: size}, nil
}

// Format reports the mode.
func (p Parameter) Format() Format { return p.format }

// RPC returns the parameter for an rpc.Call.
func (p Parameter) RPC() rpc.Parameter {
	switch p.format {
	case FormatERRC0100:
		init := make([]byte, 4)
		binary.BigEndian.PutUint32(init, p.size) // bytes provided
		param, _ := rpc.InputOutput(init, p.size)
		return param
	case FormatERRC0200:
		init := make([]byte, 8)
		binary.BigEndian.PutUint32(init[0:], 0xFFFFFFFF) // key -1
		binary.BigEndian.PutUint32(init[4:], p.size)     // bytes provided
		param, _ := rpc.InputOutput(init, p.size)
		return param
	}
	param, _ := rpc.Input(make([]byte, legacyExceptionLength))
	return param
}

// ErrAPI is matched (via errors.Is) by every *APIError.
var ErrAPI = errors.New("errorcode: IBM i API reported an error")

// APIError is an error returned through a buffer-mode error-code parameter.
// Its text contains the program and exception ID only; ExceptionData can
// hold object or profile names and is never formatted.
type APIError struct {
	Library, Program rpc.ObjectName
	Format           Format
	ExceptionID      string
	BytesProvided    uint32
	BytesAvailable   uint32
	// CCSID is the ERRC0200 CCSID of the CCHAR data (0 = job CCSID).
	CCSID uint32
	// ExceptionData is a copy of the returned insert data, possibly
	// truncated to the buffer.
	ExceptionData []byte
}

func (e *APIError) Error() string {
	return fmt.Sprintf("errorcode: %s/%s failed with %s", e.Library, e.Program, e.ExceptionID)
}

// Is makes errors.Is(err, ErrAPI) true.
func (e *APIError) Is(target error) bool { return target == ErrAPI }

// ErrMalformed is matched when a returned error-code buffer is inconsistent.
var ErrMalformed = errors.New("errorcode: malformed error-code buffer")

func malformed(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrMalformed}, args...)...)
}

// Check interprets the returned slot of p in a call to library/program. It
// returns nil when no error was reported, an *APIError when one was, and an
// ErrMalformed error when the buffer is inconsistent. In exception mode the
// slot is input-only and Check always returns nil. codec decodes the
// exception ID (job CCSID).
func (p Parameter) Check(library, program rpc.ObjectName, slot rpc.ParameterResult, codec ccsid.Codec) error {
	if p.format == formatException {
		return nil
	}
	b := slot.Data
	if !slot.Present || uint64(len(b)) != uint64(p.size) {
		return malformed("%s slot of %d bytes (present %t), want %d", p.format, len(b), slot.Present, p.size)
	}
	e := &APIError{Library: library, Program: program, Format: p.format}
	var idOff, dataStart, dataEnd int
	switch p.format {
	case FormatERRC0100:
		e.BytesProvided = binary.BigEndian.Uint32(b[0:])
		e.BytesAvailable = binary.BigEndian.Uint32(b[4:])
		idOff, dataStart = 8, 16
		dataEnd = int(min(uint64(e.BytesAvailable), uint64(len(b))))
	case FormatERRC0200:
		e.BytesProvided = binary.BigEndian.Uint32(b[4:])
		e.BytesAvailable = binary.BigEndian.Uint32(b[8:])
		idOff = 12
		if e.BytesAvailable != 0 {
			e.CCSID = binary.BigEndian.Uint32(b[20:])
			off := uint64(binary.BigEndian.Uint32(b[24:]))
			n := uint64(binary.BigEndian.Uint32(b[28:]))
			if n > 0 && off < uint64(MinERRC0200) {
				return malformed("ERRC0200 exception data offset %d inside the fixed fields", off)
			}
			if off > uint64(len(b)) {
				off = uint64(len(b))
			}
			dataStart, dataEnd = int(off), int(min(off+n, uint64(len(b))))
		}
	}
	if e.BytesProvided != p.size {
		return malformed("%s bytes provided changed from %d to %d", p.format, p.size, e.BytesProvided)
	}
	if e.BytesAvailable == 0 {
		return nil
	}
	id, err := ccsid.DecodeFixed(codec, b[idOff:idOff+7])
	if err != nil {
		return err
	}
	e.ExceptionID = strings.TrimRight(id, " ")
	if dataEnd > dataStart {
		e.ExceptionData = append([]byte(nil), b[dataStart:dataEnd]...)
	}
	return e
}
