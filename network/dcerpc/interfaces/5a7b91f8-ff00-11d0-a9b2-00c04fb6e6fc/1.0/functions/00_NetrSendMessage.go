package functions

import (
	"fmt"

	msgsvcsend "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5a7b91f8-ff00-11d0-a9b2-00c04fb6e6fc/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrSendMessageRequest carries the [in] parameters of NetrSendMessage.
//
// From/To/Text are [in, string] LPSTR (char*) — OEM-charset ASCII strings
// ([MS-MSRP] 3.2.4.1), so each is an ndr.STR, not an ndr.WSTR. They have no explicit
// pointer attribute, so as top-level [ref] string parameters they are transmitted
// inline (no referent id). hRpcBinding is an explicit primitive binding handle
// (handle_t) and is not marshalled, so it is absent from this struct.
type netrSendMessageRequest struct {
	From ndr.STR
	To   ndr.STR
	Text ndr.STR
}

func (*netrSendMessageRequest) Opnum() uint16 { return msgsvcsend.OpnumNetrSendMessage }

// netrSendMessageResponse carries the return value of NetrSendMessage. NetrSendMessage
// has no [out] parameters; error_status_t is the only result.
type netrSendMessageResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrSendMessage calls NetrSendMessage (opnum 0), sending a text message to a message
// server ([MS-MSRP] 3.2.4.1). From, To and Text are OEM-charset strings.
func NetrSendMessage(rpc ndr.Invoker, from ndr.STR, to ndr.STR, text ndr.STR) (err error) {
	req := &netrSendMessageRequest{
		From: from,
		To:   to,
		Text: text,
	}
	var resp netrSendMessageResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrSendMessage: %w", err)
		return
	}
	if uint32(resp.Status) != msgsvcsend.StatusSuccess {
		err = fmt.Errorf("NetrSendMessage failed: %s", msgsvcsend.StatusString(uint32(resp.Status)))
	}
	return
}
