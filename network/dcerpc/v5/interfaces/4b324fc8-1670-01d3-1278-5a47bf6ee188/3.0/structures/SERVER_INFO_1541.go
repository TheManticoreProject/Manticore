package structures

// SERVER_INFO_1541 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1541Minfreeconnections.
type SERVER_INFO_1541 struct {
	Sv1541Minfreeconnections int32
}
