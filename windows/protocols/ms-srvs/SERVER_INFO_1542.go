package mssrvs

// SERVER_INFO_1542 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1542Maxfreeconnections.
type SERVER_INFO_1542 struct {
	Sv1542Maxfreeconnections int32
}
