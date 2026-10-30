package pac

import (
	"encoding/binary"
	"testing"
)

// pacWithSignatureBuffers builds a PAC holding only a server checksum (type 6)
// and a KDC checksum (type 7) buffer with the given contents.
func pacWithSignatureBuffers(srv, kdc []byte) []byte {
	hdr := 8 + 2*16
	b := make([]byte, hdr)
	binary.LittleEndian.PutUint32(b[0:], 2)
	off := hdr
	for i, d := range [][]byte{srv, kdc} {
		e := 8 + i*16
		binary.LittleEndian.PutUint32(b[e:], uint32(BufferServerChecksum+i))
		binary.LittleEndian.PutUint32(b[e+4:], uint32(len(d)))
		binary.LittleEndian.PutUint64(b[e+8:], uint64(off))
		off += len(d)
	}
	b = append(b, srv...)
	return append(b, kdc...)
}

// Signature buffers shorter than their type field, or than the checksum their
// type announces, must be reported as errors rather than indexed past the end.
func TestVerifySignaturesRejectShortBuffers(t *testing.T) {
	hmacMD5Only := make([]byte, 4) // signature type only, no signature bytes
	binary.LittleEndian.PutUint32(hmacMD5Only, 0xFFFFFF76)
	hmacMD5Full := make([]byte, 20)
	binary.LittleEndian.PutUint32(hmacMD5Full, 0xFFFFFF76)

	for _, tc := range []struct {
		name   string
		srv    []byte
		kdc    []byte
		verify func(*PAC) error
	}{
		{"KDC signature with empty server buffer", nil, nil, func(p *PAC) error { return p.VerifyKDCSignature(make([]byte, 16)) }},
		{"KDC signature with empty KDC buffer", hmacMD5Full, nil, func(p *PAC) error { return p.VerifyKDCSignature(make([]byte, 16)) }},
		{"server signature with type only", hmacMD5Only, hmacMD5Full, func(p *PAC) error { return p.VerifyServerSignature(make([]byte, 16)) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(pacWithSignatureBuffers(tc.srv, tc.kdc))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			var verr error
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("verification panicked: %v", r)
					}
				}()
				verr = tc.verify(p)
			}()
			if verr == nil {
				t.Fatal("verification accepted a truncated signature buffer")
			}
		})
	}
}
