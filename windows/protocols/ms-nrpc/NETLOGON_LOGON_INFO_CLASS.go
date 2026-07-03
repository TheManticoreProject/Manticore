package msnrpc

// NETLOGON_LOGON_INFO_CLASS is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-NRPC]).
type NETLOGON_LOGON_INFO_CLASS uint16

const (
	NetlogonInteractiveInformation           NETLOGON_LOGON_INFO_CLASS = 1
	NetlogonNetworkInformation               NETLOGON_LOGON_INFO_CLASS = 2
	NetlogonServiceInformation               NETLOGON_LOGON_INFO_CLASS = 3
	NetlogonGenericInformation               NETLOGON_LOGON_INFO_CLASS = 4
	NetlogonInteractiveTransitiveInformation NETLOGON_LOGON_INFO_CLASS = 5
	NetlogonNetworkTransitiveInformation     NETLOGON_LOGON_INFO_CLASS = 6
	NetlogonServiceTransitiveInformation     NETLOGON_LOGON_INFO_CLASS = 7
	NetlogonTicketLogonInformation           NETLOGON_LOGON_INFO_CLASS = 8
)
