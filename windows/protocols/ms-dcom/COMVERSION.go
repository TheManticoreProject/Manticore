package msdcom

// COMVERSION models COMVERSION ([MS-DCOM] 2.2.11): the major and minor version of the
// DCOM protocol used by a client or server.
type COMVERSION struct {
	MajorVersion uint16
	MinorVersion uint16
}
