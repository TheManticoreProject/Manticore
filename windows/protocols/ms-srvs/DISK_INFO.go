package mssrvs

// DISK_INFO contains the drive letter of a disk on the server ([MS-SRVS]
// 2.2.4.86). Disk is a fixed array of three WCHARs (drive letter, ':' and a
// null terminator).
type DISK_INFO struct {
	Disk [3]uint16
}
