package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ShareInfo is a shared resource exported by the server: its name, STYPE_* type flags,
// and comment ([MS-SRVS] 2.2.4.23 SHARE_INFO_1). Type carries the raw STYPE_* value,
// including the STYPE_SPECIAL (0x80000000) and STYPE_TEMPORARY (0x40000000) high bits.
type ShareInfo struct {
	Name    string
	Type    uint32
	Comment string
}

// ListShares enumerates the shares exported by the server via NetrShareEnum at info
// level 1 ([MS-SRVS] 3.1.4.8). It binds a fresh \srvsvc pipe for the call.
func (c *Client) ListShares() ([]ShareInfo, error) {
	rpc, done, err := c.bind()
	if err != nil {
		return nil, err
	}
	defer done()
	return listShares(rpc)
}

// listShares performs the NetrShareEnum call over an already-bound invoker and projects
// the level-1 container into ShareInfo values. It is split out so it can be unit-tested
// with an in-memory transport.
func listShares(rpc ndr.Invoker) ([]ShareInfo, error) {
	resume := ndr.DWORD(0)
	info := structures.SHARE_ENUM_STRUCT{
		Level: 1,
		ShareInfo: structures.SHARE_ENUM_UNION{
			Tag:    1,
			Level1: &structures.SHARE_INFO_1_CONTAINER{},
		},
	}

	out, _, _, err := functions.NetrShareEnum(rpc, "", info, ndr.DWORD(MaxPreferredLength), &resume)
	if err != nil {
		return nil, err
	}

	container := out.ShareInfo.Level1
	if container == nil {
		return nil, nil
	}
	shares := make([]ShareInfo, 0, len(container.Buffer))
	for _, s := range container.Buffer {
		shares = append(shares, ShareInfo{
			Name:    string(s.Shi1Netname),
			Type:    uint32(s.Shi1Type),
			Comment: string(s.Shi1Remark),
		})
	}
	return shares, nil
}
