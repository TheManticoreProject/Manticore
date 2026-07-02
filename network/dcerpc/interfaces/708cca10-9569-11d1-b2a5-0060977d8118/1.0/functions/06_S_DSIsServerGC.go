package functions

import (
	"fmt"

	dscomm2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/708cca10-9569-11d1-b2a5-0060977d8118/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// s_DSIsServerGCRequest carries the [in] parameters of S_DSIsServerGC.
type s_DSIsServerGCRequest struct {
}

func (*s_DSIsServerGCRequest) Opnum() uint16 { return dscomm2.OpnumS_DSIsServerGC }

// s_DSIsServerGCResponse carries the return value of S_DSIsServerGC. Unlike most dscomm2
// methods this one does not return an HRESULT: the return is a Boolean-valued long that is
// nonzero when the server is a global catalog server ([MS-MQDS] 3.1.4.16).
type s_DSIsServerGCResponse struct {
	Result int32 `ndr:"retval"`
}

// S_DSIsServerGC calls S_DSIsServerGC (opnum 6) and reports whether the server is a global
// catalog server ([MS-MQDS] 3.1.4.16).
func S_DSIsServerGC(rpc ndr.Invoker) (IsGC bool, err error) {
	req := &s_DSIsServerGCRequest{}
	var resp s_DSIsServerGCResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSIsServerGC: %w", err)
		return
	}
	IsGC = resp.Result != 0
	return
}
