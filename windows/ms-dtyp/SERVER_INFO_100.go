package msdtyp

// SERVER_INFO_100
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/007c654b-7d78-49d4-9f4d-0da7c1889727
type SERVER_INFO_100 struct {
	// Sv100PlatformId: The platform ID.
	Sv100PlatformId DWORD
	// Sv100Name: The server name.
	Sv100Name WCHAR
}

type PSERVER_INFO_100 *SERVER_INFO_100
type LPSERVER_INFO_100 *SERVER_INFO_100
