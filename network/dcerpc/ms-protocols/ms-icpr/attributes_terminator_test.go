package ms_icpr

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	mswcce "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wcce"
)

// certServerRequestArgs mirrors the generated certServerRequestRequest [in]
// parameter layout, so the attribute blob actually put on the wire can be read
// back from the sent stub.
type certServerRequestArgs struct {
	DwFlags       ndr.DWORD
	PwszAuthority *ndr.WSTR `ndr:"unique"`
	PdwRequestId  ndr.DWORD
	PctbAttribs   mswcce.CERTTRANSBLOB
	PctbRequest   mswcce.CERTTRANSBLOB
}

// TestRequestAttributesAreNulTerminated asserts the pwszAttributes CERTTRANSBLOB
// put on the wire is a UTF-16 string ending in a wide NUL, with Cb counting the
// terminator. pwszAttributes is a null-terminated wide string ([MS-WCCE]
// 3.2.1.4.2.1.4); a CA parses it as a C-string, so an unterminated blob makes it
// read past the Cb-sized buffer.
func TestRequestAttributesAreNulTerminated(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)

	ft.queue(responsePDU(t, 2, stub(t, &certReqResp{
		PdwDisposition: DispIssued,
		Status:         0,
	})))

	const attributes = "CertificateTemplate:User"
	if _, err := request(c, "CORP-CA", []byte{0x30, 0x01}, attributes); err != nil {
		t.Fatalf("request() error = %v", err)
	}

	var sent pdu.Request
	if _, err := sent.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request PDU does not parse: %v", err)
	}
	var args certServerRequestArgs
	if err := ndr.Unmarshal(sent.Stub, &args); err != nil {
		t.Fatalf("request stub does not parse: %v", err)
	}

	pb := args.PctbAttribs.Pb
	if int(args.PctbAttribs.Cb) != len(pb) {
		t.Errorf("Cb = %d, want %d (len(Pb))", args.PctbAttribs.Cb, len(pb))
	}
	if len(pb) < 2 || pb[len(pb)-2] != 0x00 || pb[len(pb)-1] != 0x00 {
		t.Fatalf("attribute blob does not end in a UTF-16 NUL terminator: % x", pb)
	}
	// The terminator is extra: the payload before it must still be the attributes.
	if got := decodeUTF16(pb); got != attributes {
		t.Errorf("decoded attributes = %q, want %q", got, attributes)
	}
	if want := len(encodeUTF16(attributes)) + 2; len(pb) != want {
		t.Errorf("attribute blob is %d bytes, want %d (string + wide NUL)", len(pb), want)
	}
}
