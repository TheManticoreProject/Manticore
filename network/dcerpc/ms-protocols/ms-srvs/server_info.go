package mssrvs

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ServerInfo is the configuration of a server: platform id, name, OS version, SV_TYPE_*
// flags, and comment ([MS-SRVS] 2.2.4.41 SERVER_INFO_101).
type ServerInfo struct {
	PlatformID   uint32
	Name         string
	VersionMajor uint32
	VersionMinor uint32
	Type         uint32
	Comment      string
}

// GetServerInfo queries the server's configuration via NetrServerGetInfo at info level
// 101 ([MS-SRVS] 3.1.4.17). It binds a fresh \srvsvc pipe for the call.
func (c *Client) GetServerInfo() (*ServerInfo, error) {
	rpc, done, err := c.bind()
	if err != nil {
		return nil, err
	}
	defer done()
	return getServerInfo(rpc)
}

// getServerInfo performs the NetrServerGetInfo call over an already-bound invoker and
// projects the level-101 union arm into a ServerInfo.
func getServerInfo(rpc ndr.Invoker) (*ServerInfo, error) {
	out, err := functions.NetrServerGetInfo(rpc, "", 101)
	if err != nil {
		return nil, err
	}

	si := out.ServerInfo101
	if si == nil {
		return nil, fmt.Errorf("mssrvs: NetrServerGetInfo(101) returned no SERVER_INFO_101")
	}
	return &ServerInfo{
		PlatformID:   uint32(si.Sv101PlatformId),
		Name:         string(si.Sv101Name),
		VersionMajor: uint32(si.Sv101VersionMajor),
		VersionMinor: uint32(si.Sv101VersionMinor),
		Type:         uint32(si.Sv101Type),
		Comment:      string(si.Sv101Comment),
	}, nil
}
