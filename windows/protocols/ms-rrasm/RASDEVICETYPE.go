package msrrasm

// RASDEVICETYPE ([MS-RRASM] 2.2.1.1.9). Widened to a 32-bit base: the RDT_Tunnel..RDT_Broadband
// flag values (0x00010000+) exceed the 16-bit NDR enum range, so it is transmitted as 4 octets.
type RASDEVICETYPE uint32

const (
	RDT_Modem        RASDEVICETYPE = 0
	RDT_X25          RASDEVICETYPE = 1
	RDT_Isdn         RASDEVICETYPE = 2
	RDT_Serial       RASDEVICETYPE = 3
	RDT_FrameRelay   RASDEVICETYPE = 4
	RDT_Atm          RASDEVICETYPE = 5
	RDT_Sonet        RASDEVICETYPE = 6
	RDT_Sw56         RASDEVICETYPE = 7
	RDT_Tunnel_Pptp  RASDEVICETYPE = 8
	RDT_Tunnel_L2tp  RASDEVICETYPE = 9
	RDT_Irda         RASDEVICETYPE = 10
	RDT_Parallel     RASDEVICETYPE = 11
	RDT_Other        RASDEVICETYPE = 12
	RDT_PPPoE        RASDEVICETYPE = 13
	RDT_Tunnel_Sstp  RASDEVICETYPE = 14
	RDT_Tunnel_Ikev2 RASDEVICETYPE = 15
	RDT_Tunnel       RASDEVICETYPE = 0x00010000
	RDT_Direct       RASDEVICETYPE = 0x00020000
	RDT_Null_Modem   RASDEVICETYPE = 0x00040000
	RDT_Broadband    RASDEVICETYPE = 0x00080000
)
