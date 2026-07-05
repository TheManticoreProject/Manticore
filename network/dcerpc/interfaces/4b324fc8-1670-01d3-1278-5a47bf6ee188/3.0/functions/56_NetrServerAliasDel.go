package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// netrServerAliasDelRequest is the [in] parameter set of NetrServerAliasDel: the
// [unique] server name, the level, and the [in, switch_is(Level)] SERVER_ALIAS_INFO
// union (its Tag is set to Level before marshalling).
type netrServerAliasDelRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Level      ndr.DWORD
	InfoStruct mssrvs.SERVER_ALIAS_INFO
}

func (*netrServerAliasDelRequest) Opnum() uint16 { return srvsvc.OpnumNetrServerAliasDel }

// NetrServerAliasDel calls NetrServerAliasDel (opnum 56), deleting an alias name from
// the server ([MS-SRVS] 3.1.4.30).
func NetrServerAliasDel(rpc ndr.Invoker, serverName string, level uint32, info mssrvs.SERVER_ALIAS_INFO) error {
	info.Tag = ndr.DWORD(level)
	req := &netrServerAliasDelRequest{
		ServerName: optWStr(serverName),
		Level:      ndr.DWORD(level),
		InfoStruct: info,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrServerAliasDel: %w", err)
	}
	if status := uint32(resp.Status); status != srvsvc.NERR_Success {
		return fmt.Errorf("NetrServerAliasDel failed: %s", srvsvc.StatusString(status))
	}
	return nil
}
