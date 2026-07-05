package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	srvstypes "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// SessionInfo is an active session on the server: the connecting client's name, the
// authenticated user, and the active/idle times in seconds ([MS-SRVS] 2.2.4.15
// SESSION_INFO_10).
type SessionInfo struct {
	ClientName string
	UserName   string
	ActiveSecs uint32
	IdleSecs   uint32
}

// ListSessions enumerates the active sessions on the server via NetrSessionEnum at info
// level 10 ([MS-SRVS] 3.1.4.6). It binds a fresh \srvsvc pipe for the call. The server
// typically requires administrative rights to enumerate sessions.
func (c *Client) ListSessions() ([]SessionInfo, error) {
	rpc, done, err := c.bind()
	if err != nil {
		return nil, err
	}
	defer done()
	return listSessions(rpc)
}

// listSessions performs the NetrSessionEnum call over an already-bound invoker and
// projects the level-10 container into SessionInfo values.
func listSessions(rpc ndr.Invoker) ([]SessionInfo, error) {
	info := srvstypes.SESSION_ENUM_STRUCT{
		Level: 10,
		SessionInfo: srvstypes.SESSION_ENUM_UNION{
			Tag:     10,
			Level10: &srvstypes.SESSION_INFO_10_CONTAINER{},
		},
	}

	out, _, _, err := functions.NetrSessionEnum(rpc, "", "", "", info, MaxPreferredLength, 0)
	if err != nil {
		return nil, err
	}

	container := out.SessionInfo.Level10
	if container == nil {
		return nil, nil
	}
	sessions := make([]SessionInfo, 0, len(container.Buffer))
	for _, s := range container.Buffer {
		sessions = append(sessions, SessionInfo{
			ClientName: string(s.Sesi10Cname),
			UserName:   string(s.Sesi10Username),
			ActiveSecs: uint32(s.Sesi10Time),
			IdleSecs:   uint32(s.Sesi10IdleTime),
		})
	}
	return sessions, nil
}
