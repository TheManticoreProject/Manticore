package functions

import (
	"fmt"

	tapsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/2f5f6520-ca46-1067-b319-00dd010662da/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-trp"
)

// clientAttachRequest carries the [in] parameters of ClientAttach. pszDomainUser and
// pszMachine are top-level [ref] [string] wchar_t* parameters ([MS-TRP] 3.2.4.1), so they
// marshal inline as conformant-varying UTF-16LE strings with no referent id.
type clientAttachRequest struct {
	LProcessID    int32
	PszDomainUser ndr.WSTR
	PszMachine    ndr.WSTR
}

func (*clientAttachRequest) Opnum() uint16 { return tapsrv.OpnumClientAttach }

// clientAttachResponse carries the [out] parameters and return value of ClientAttach.
type clientAttachResponse struct {
	PphContext         mstrp.PCONTEXT_HANDLE_TYPE
	PhAsyncEventsEvent int32
	Return             ndr.DWORD `ndr:"retval"`
}

// ClientAttach calls ClientAttach (opnum 0) ([MS-TRP] 3.2.4.1). The client uses it to
// establish a session with the telephony server: lProcessID identifies the client
// process (0xFFFFFFFF / 0xFFFFFFFD select the reduced-privilege forms described by the
// spec), pszDomainUser is the caller's domain\user, and pszMachine the client machine
// name. On success the server returns the tapsrv context handle plus phAsyncEventsEvent,
// the event value used to signal asynchronous completions. The method returns 0 on
// success or a nonzero error code.
func ClientAttach(rpc ndr.Invoker, lProcessID int32, pszDomainUser ndr.WSTR, pszMachine ndr.WSTR) (PphContext mstrp.PCONTEXT_HANDLE_TYPE, PhAsyncEventsEvent int32, err error) {
	req := &clientAttachRequest{
		LProcessID:    lProcessID,
		PszDomainUser: pszDomainUser,
		PszMachine:    pszMachine,
	}
	var resp clientAttachResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ClientAttach: %w", err)
		return
	}
	PphContext = resp.PphContext
	PhAsyncEventsEvent = resp.PhAsyncEventsEvent
	if uint32(resp.Return) != tapsrv.StatusSuccess {
		err = fmt.Errorf("ClientAttach failed: %s", tapsrv.StatusString(uint32(resp.Return)))
	}
	return
}
