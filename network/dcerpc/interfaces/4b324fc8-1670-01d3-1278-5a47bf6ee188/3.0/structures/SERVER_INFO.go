package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO is the SERVER_INFO union used by NetrServerGetInfo and
// NetrServerSetInfo ([MS-SRVS] 2.2.3.7). It is switched on an unsigned long
// (the method's Level argument); the discriminant is transmitted inline as a
// 4-byte DWORD ahead of the selected arm. Every arm is an LP pointer to its
// SERVER_INFO_<level> structure, modeled here as a [unique] pointer (nil when
// not selected). Set Tag to the level and populate the matching arm before
// marshalling.
type SERVER_INFO struct {
	Tag            ndr.DWORD         `ndr:"switch"`
	ServerInfo100  *SERVER_INFO_100  `ndr:"case=100,unique"`
	ServerInfo101  *SERVER_INFO_101  `ndr:"case=101,unique"`
	ServerInfo102  *SERVER_INFO_102  `ndr:"case=102,unique"`
	ServerInfo103  *SERVER_INFO_103  `ndr:"case=103,unique"`
	ServerInfo502  *SERVER_INFO_502  `ndr:"case=502,unique"`
	ServerInfo503  *SERVER_INFO_503  `ndr:"case=503,unique"`
	ServerInfo599  *SERVER_INFO_599  `ndr:"case=599,unique"`
	ServerInfo1005 *SERVER_INFO_1005 `ndr:"case=1005,unique"`
	ServerInfo1107 *SERVER_INFO_1107 `ndr:"case=1107,unique"`
	ServerInfo1010 *SERVER_INFO_1010 `ndr:"case=1010,unique"`
	ServerInfo1016 *SERVER_INFO_1016 `ndr:"case=1016,unique"`
	ServerInfo1017 *SERVER_INFO_1017 `ndr:"case=1017,unique"`
	ServerInfo1018 *SERVER_INFO_1018 `ndr:"case=1018,unique"`
	ServerInfo1501 *SERVER_INFO_1501 `ndr:"case=1501,unique"`
	ServerInfo1502 *SERVER_INFO_1502 `ndr:"case=1502,unique"`
	ServerInfo1503 *SERVER_INFO_1503 `ndr:"case=1503,unique"`
	ServerInfo1506 *SERVER_INFO_1506 `ndr:"case=1506,unique"`
	ServerInfo1510 *SERVER_INFO_1510 `ndr:"case=1510,unique"`
	ServerInfo1511 *SERVER_INFO_1511 `ndr:"case=1511,unique"`
	ServerInfo1512 *SERVER_INFO_1512 `ndr:"case=1512,unique"`
	ServerInfo1513 *SERVER_INFO_1513 `ndr:"case=1513,unique"`
	ServerInfo1514 *SERVER_INFO_1514 `ndr:"case=1514,unique"`
	ServerInfo1515 *SERVER_INFO_1515 `ndr:"case=1515,unique"`
	ServerInfo1516 *SERVER_INFO_1516 `ndr:"case=1516,unique"`
	ServerInfo1518 *SERVER_INFO_1518 `ndr:"case=1518,unique"`
	ServerInfo1523 *SERVER_INFO_1523 `ndr:"case=1523,unique"`
	ServerInfo1528 *SERVER_INFO_1528 `ndr:"case=1528,unique"`
	ServerInfo1529 *SERVER_INFO_1529 `ndr:"case=1529,unique"`
	ServerInfo1530 *SERVER_INFO_1530 `ndr:"case=1530,unique"`
	ServerInfo1533 *SERVER_INFO_1533 `ndr:"case=1533,unique"`
	ServerInfo1534 *SERVER_INFO_1534 `ndr:"case=1534,unique"`
	ServerInfo1535 *SERVER_INFO_1535 `ndr:"case=1535,unique"`
	ServerInfo1536 *SERVER_INFO_1536 `ndr:"case=1536,unique"`
	ServerInfo1538 *SERVER_INFO_1538 `ndr:"case=1538,unique"`
	ServerInfo1539 *SERVER_INFO_1539 `ndr:"case=1539,unique"`
	ServerInfo1540 *SERVER_INFO_1540 `ndr:"case=1540,unique"`
	ServerInfo1541 *SERVER_INFO_1541 `ndr:"case=1541,unique"`
	ServerInfo1542 *SERVER_INFO_1542 `ndr:"case=1542,unique"`
	ServerInfo1543 *SERVER_INFO_1543 `ndr:"case=1543,unique"`
	ServerInfo1544 *SERVER_INFO_1544 `ndr:"case=1544,unique"`
	ServerInfo1545 *SERVER_INFO_1545 `ndr:"case=1545,unique"`
	ServerInfo1546 *SERVER_INFO_1546 `ndr:"case=1546,unique"`
	ServerInfo1547 *SERVER_INFO_1547 `ndr:"case=1547,unique"`
	ServerInfo1548 *SERVER_INFO_1548 `ndr:"case=1548,unique"`
	ServerInfo1549 *SERVER_INFO_1549 `ndr:"case=1549,unique"`
	ServerInfo1550 *SERVER_INFO_1550 `ndr:"case=1550,unique"`
	ServerInfo1552 *SERVER_INFO_1552 `ndr:"case=1552,unique"`
	ServerInfo1553 *SERVER_INFO_1553 `ndr:"case=1553,unique"`
	ServerInfo1554 *SERVER_INFO_1554 `ndr:"case=1554,unique"`
	ServerInfo1555 *SERVER_INFO_1555 `ndr:"case=1555,unique"`
	ServerInfo1556 *SERVER_INFO_1556 `ndr:"case=1556,unique"`
}
