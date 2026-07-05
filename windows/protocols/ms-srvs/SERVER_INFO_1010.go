package mssrvs

// SERVER_INFO_1010 contains the auto-disconnect time in minutes ([MS-SRVS]
// 2.2.4.49). Sv1010Disc is an IDL long (signed 32-bit).
type SERVER_INFO_1010 struct {
	Sv1010Disc int32
}
