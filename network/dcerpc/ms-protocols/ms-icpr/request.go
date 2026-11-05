package ms_icpr

import (
	"encoding/binary"
	"strings"
	"unicode/utf16"

	ICertPassage "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/91ae6020-9e3c-11cf-8d7c-00aa00c091be/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/91ae6020-9e3c-11cf-8d7c-00aa00c091be/0.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswcce "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wcce"
)

// Certificate request disposition values a CA reports in Result.Disposition, re-exported
// from the ICertPassage interface so callers need not import it ([MS-WCCE] 3.2.1.4.2.1).
const (
	DispIncomplete      = ICertPassage.CR_DISP_INCOMPLETE
	DispError           = ICertPassage.CR_DISP_ERROR
	DispDenied          = ICertPassage.CR_DISP_DENIED
	DispIssued          = ICertPassage.CR_DISP_ISSUED
	DispIssuedOutOfBand = ICertPassage.CR_DISP_ISSUED_OUT_OF_BAND
	DispUnderSubmission = ICertPassage.CR_DISP_UNDER_SUBMISSION
	DispRevoked         = ICertPassage.CR_DISP_REVOKED
)

// Result is the outcome of a CertServerRequest.
type Result struct {
	// RequestID is the CA's request id: the id assigned to a newly submitted request, or
	// echoed for a pending one.
	RequestID uint32
	// Disposition is the CA's verdict, one of the Disp* constants. DispIssued means the
	// certificate was issued and is in Certificate.
	Disposition uint32
	// Certificate is the issued certificate, DER-encoded (the CA's pctbEncodedCert). It is
	// empty unless the certificate was issued.
	Certificate []byte
	// CertChain is the CA's pctbCert blob: the certificate or chain in the CA's chosen
	// encoding (a PKCS#7 full response for some requests). Often empty for a plain issuance.
	CertChain []byte
	// DispositionMessage is the CA's human-readable status message (e.g. "Issued", or the
	// reason for a denial).
	DispositionMessage string
}

// Issued reports whether the CA issued the certificate.
func (r *Result) Issued() bool { return r.Disposition == DispIssued }

// Request submits a certificate enrollment request to the CA named caName (the CA's common
// name, e.g. "CORP-DC-CA") over the \cert pipe and returns the CA's disposition and, on
// success, the issued certificate.
//
// csr is the raw request as the CA expects it: a DER PKCS#10 for a direct request, or a
// CMC / PKCS#7 for a renewal or on-behalf-of (ESC-style) request. attributes is the
// request attribute string, newline-separated, empty if none — for example
// "CertificateTemplate:User" or "CertificateTemplate:Machine\nSAN:upn=victim@corp.local".
// This is the enrollment primitive an NTLM relay uses to obtain a certificate as the
// relayed identity (AD CS ESC8 and related).
//
// A non-nil error is returned when the RPC fails or the CA returns a failure status; a CA
// that merely denies the request returns a nil error with Disposition set to DispDenied,
// so callers inspect Result (Issued, Disposition, DispositionMessage) for the enrollment
// outcome. The returned Result is always non-nil, so its DispositionMessage is available
// to explain a failure even when err is non-nil.
func (c *Client) Request(caName string, csr []byte, attributes string) (*Result, error) {
	rpc, done, err := c.bind()
	if err != nil {
		return &Result{}, err
	}
	defer done()
	return request(rpc, caName, csr, attributes)
}

// request issues CertServerRequest over an already-bound invoker and projects the CA's
// CERTTRANSBLOB replies into a Result. It is split out so it can be unit-tested with an
// in-memory transport.
func request(rpc ndr.Invoker, caName string, csr []byte, attributes string) (*Result, error) {
	var attribBlob mswcce.CERTTRANSBLOB
	if attributes != "" {
		// pwszAttributes is a null-terminated wide string ([MS-WCCE] 3.2.1.4.2.1.4):
		// the CA parses the blob as a C-string, so it must end in a UTF-16 NUL, and
		// Cb counts that terminator. Without it the CA reads past the Cb-sized buffer.
		attr := append(encodeUTF16(attributes), 0x00, 0x00)
		attribBlob = mswcce.CERTTRANSBLOB{Cb: ndr.DWORD(len(attr)), Pb: attr}
	}
	requestBlob := mswcce.CERTTRANSBLOB{Cb: ndr.DWORD(len(csr)), Pb: csr}

	reqID, disposition, certChain, encodedCert, dispMsg, err := functions.CertServerRequest(
		rpc,
		0, // dwFlags: none
		wstrPtr(caName),
		0, // pdwRequestId: 0 submits a new request
		attribBlob,
		requestBlob,
	)
	// The CA may populate the disposition and message even on a failure status, so the
	// Result is built (and returned) regardless of err.
	res := &Result{
		RequestID:          uint32(reqID),
		Disposition:        uint32(disposition),
		Certificate:        encodedCert.Pb,
		CertChain:          certChain.Pb,
		DispositionMessage: decodeUTF16(dispMsg.Pb),
	}
	return res, err
}

// wstrPtr returns a *ndr.WSTR for s, or nil for the empty string so an optional [unique]
// pointer parameter is sent as a null pointer.
func wstrPtr(s string) *ndr.WSTR {
	if s == "" {
		return nil
	}
	w := ndr.WSTR(s)
	return &w
}

// encodeUTF16 encodes s as a UTF-16LE byte string, the wire form of the attribute blob a
// CA expects.
func encodeUTF16(s string) []byte {
	u := utf16.Encode([]rune(s))
	b := make([]byte, len(u)*2)
	for i, c := range u {
		binary.LittleEndian.PutUint16(b[i*2:], c)
	}
	return b
}

// decodeUTF16 decodes a UTF-16LE byte string (the CA's disposition message) into a Go
// string, trimming any trailing NUL.
func decodeUTF16(b []byte) string {
	if len(b) < 2 {
		return ""
	}
	u := make([]uint16, len(b)/2)
	for i := range u {
		u[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	return strings.TrimRight(string(utf16.Decode(u)), "\x00")
}
