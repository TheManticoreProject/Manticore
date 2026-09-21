package client

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
)

// RDMA protocol IDs (MS-SMB2 2.2.32.5).
const (
	RDMA_NONE  uint16 = 0x0000
	RDMA_iWARP uint16 = 0x0001
	RDMA_InfiniBand uint16 = 0x0002
)

// Interface capability flags (MS-SMB2 2.2.32.5).
const (
	INTERFACE_CAP_RSS  uint32 = 0x00000001
	INTERFACE_CAP_RDMA uint32 = 0x00000002
)

// NetworkInterface is a single entry from the
// FSCTL_QUERY_NETWORK_INTERFACE_INFO response (MS-SMB2 2.2.32.5).
type NetworkInterface struct {
	IfIndex      uint32
	Capability   uint32
	LinkSpeed    uint64
	SockAddrInet net.IP
	SockAddrPort uint16
	Family       uint16
}

// IsRSS reports whether the interface supports Receive Side Scaling.
func (ni *NetworkInterface) IsRSS() bool { return ni.Capability&INTERFACE_CAP_RSS != 0 }

// IsRDMA reports whether the interface supports RDMA.
func (ni *NetworkInterface) IsRDMA() bool { return ni.Capability&INTERFACE_CAP_RDMA != 0 }

// parseNetworkInterfaces decodes the chained NETWORK_INTERFACE_INFO response.
// Each entry is at least 64 bytes: Next(4) IfIndex(4) Capability(4) Reserved(4)
// LinkSpeed(8) SockAddrStorage(128).
func parseNetworkInterfaces(data []byte) ([]NetworkInterface, error) {
	const entryMinSize = 152 // 4+4+4+4+8+128
	var out []NetworkInterface
	offset := 0
	for offset < len(data) {
		if len(data)-offset < entryMinSize {
			return nil, fmt.Errorf("network interface entry at offset %d too short: %d bytes, need %d", offset, len(data)-offset, entryMinSize)
		}
		entry := data[offset:]
		next := binary.LittleEndian.Uint32(entry[0:4])

		ni := NetworkInterface{
			IfIndex:    binary.LittleEndian.Uint32(entry[4:8]),
			Capability: binary.LittleEndian.Uint32(entry[8:12]),
			LinkSpeed:  binary.LittleEndian.Uint64(entry[16:24]),
		}

		// SOCKADDR_STORAGE starts at offset 24, 128 bytes. The first two bytes
		// are the address family (AF_INET=2, AF_INET6=23).
		sockaddr := entry[24:152]
		family := binary.LittleEndian.Uint16(sockaddr[0:2])
		ni.Family = family

		switch family {
		case 2: // AF_INET
			ni.SockAddrPort = binary.BigEndian.Uint16(sockaddr[2:4])
			ni.SockAddrInet = net.IP(append([]byte{}, sockaddr[4:8]...))
		case 23: // AF_INET6
			ni.SockAddrPort = binary.BigEndian.Uint16(sockaddr[2:4])
			ni.SockAddrInet = net.IP(append([]byte{}, sockaddr[8:24]...))
		}

		out = append(out, ni)

		if next == 0 {
			break
		}
		offset += int(next)
	}
	return out, nil
}

// QueryNetworkInterfaceInfo queries the server for its available network
// interfaces via FSCTL_QUERY_NETWORK_INTERFACE_INFO. The caller must have an
// established session and tree connect. Wire: SMB2 IOCTL.
func (c *Client) QueryNetworkInterfaceInfo() ([]NetworkInterface, error) {
	fileId := types.SMB2_FILEID{
		Persistent: 0xFFFFFFFFFFFFFFFF,
		Volatile:   0xFFFFFFFFFFFFFFFF,
	}

	output, err := c.Ioctl(fileId, filesystem.FSCTL_QUERY_NETWORK_INTERFACE_INFO, nil, true, 65536)
	if err != nil {
		return nil, fmt.Errorf("FSCTL_QUERY_NETWORK_INTERFACE_INFO failed: %w", err)
	}
	return parseNetworkInterfaces(output)
}
