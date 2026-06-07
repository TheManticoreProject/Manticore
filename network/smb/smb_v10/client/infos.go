package client

import (
	"fmt"
	"time"
)

// GetRemoteServerTime returns the server's system clock as reported in the
// SMB_COM_NEGOTIATE response.
//
// SMBv1 has no dedicated query-time command; the server's time is delivered in the
// NEGOTIATE response and captured by Negotiate as Connection.Server.SystemTime (a
// FILETIME, i.e. 100ns ticks since 1601-01-01 UTC). The returned time therefore
// reflects the server's clock at the moment of negotiation, in UTC.
//
// Returns:
//   - The server time in UTC.
//   - An error if no connection exists or negotiation has not populated the time.
func (c *Client) GetRemoteServerTime() (time.Time, error) {
	if c.Connection == nil || c.Connection.Server == nil {
		return time.Time{}, fmt.Errorf("no connection established")
	}

	systemTime := c.Connection.Server.SystemTime
	if systemTime.DwLowDateTime == 0 && systemTime.DwHighDateTime == 0 {
		return time.Time{}, fmt.Errorf("server time not available; negotiate first")
	}

	return systemTime.GetTime().UTC(), nil
}
