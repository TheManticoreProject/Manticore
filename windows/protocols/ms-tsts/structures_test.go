package mststs

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// wstr returns a pointer to an ndr.WSTR carrying s, for building [unique] string fields.
func wstr(s string) *ndr.WSTR {
	w := ndr.WSTR(s)
	return &w
}

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts deep
// equality — the wire-shape acceptance check for a structure's NDR tags.
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}
	var out T
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("%s: round trip mismatch:\n in:  %+v\n out: %+v", name, in, out)
	}
}

// TestLSMSESSIONINFORMATION_RoundTrip covers a struct of [unique] wide-string pointers
// interleaved with fixed scalars (TermSrvSession RpcGetSessionInformation).
func TestLSMSESSIONINFORMATION_RoundTrip(t *testing.T) {
	roundTrip(t, "LSMSESSIONINFORMATION", LSMSESSIONINFORMATION{
		PszUserName:     wstr("alice"),
		PszDomain:       wstr("CONTOSO"),
		PszTerminalName: wstr("RDP-Tcp#7"),
		SessionState:    4,
		DesktopLocked:   1,
		ConnectTime:     0x01d9a1b2c3d4e5f6,
		DisconnectTime:  0,
		LogonTime:       0x01d9a1b200000000,
	})
	roundTrip(t, "LSMSESSIONINFORMATION/nil-strings", LSMSESSIONINFORMATION{SessionState: 1})
}

// TestSESSIONENUM_RoundTrip covers the switch_is(Level) union wrapper selecting level 2,
// which carries a GUID.
func TestSESSIONENUM_RoundTrip(t *testing.T) {
	roundTrip(t, "SESSIONENUM/level2", SESSIONENUM{
		Level: 2,
		Data: SessionInfo{
			Tag: 2,
			SessionEnum_Level2: SESSIONENUM_LEVEL2{
				SessionId:    3,
				State:        1,
				Name:         [33]uint16{'R', 'D', 'P'},
				Source:       0,
				BFullDesktop: 1,
			},
		},
	})
	roundTrip(t, "SESSIONENUM/level1", SESSIONENUM{
		Level: 1,
		Data: SessionInfo{
			Tag:                1,
			SessionEnum_Level1: SESSIONENUM_LEVEL1{SessionId: 9, State: 0, Name: [33]uint16{'c', 'o', 'n'}},
		},
	})
}

// TestSESSIONENUM_LEVEL3_RoundTrip covers a [unique] size_is byte blob field.
func TestSESSIONENUM_LEVEL3_RoundTrip(t *testing.T) {
	roundTrip(t, "SESSIONENUM_LEVEL3", SESSIONENUM_LEVEL3{
		SessionId:     11,
		State:         1,
		Name:          [33]uint16{'s', 'e', 's'},
		ProtoDataSize: 4,
		PProtocolData: []uint8{0xde, 0xad, 0xbe, 0xef},
	})
	roundTrip(t, "SESSIONENUM_LEVEL3/nil-blob", SESSIONENUM_LEVEL3{SessionId: 1})
}

// TestLISTENERENUM_RoundTrip covers the RCMPublic listener union wrapper.
func TestLISTENERENUM_RoundTrip(t *testing.T) {
	roundTrip(t, "LISTENERENUM", LISTENERENUM{
		Level: 1,
		Data: ListenerInfo{
			Tag:                 1,
			ListenerEnum_Level1: LISTENERENUM_LEVEL1{Id: 5, BListening: 1, Name: [33]uint16{'R', 'D', 'P'}},
		},
	})
}

// TestEXECENVDATA_RoundTrip covers the TermSrvEnumeration exec-env union wrapper.
func TestEXECENVDATA_RoundTrip(t *testing.T) {
	roundTrip(t, "EXECENVDATA/level2", EXECENVDATA{
		Level: 2,
		Data: ExecEnvData{
			Tag: 2,
			ExecEnvEnum_Level2: EXECENVDATA_LEVEL2{
				ExecEnvId:    2,
				State:        1,
				SessionName:  [33]uint16{'s'},
				AbsSessionId: 2,
				HostName:     [33]uint16{'h'},
				UserName:     [33]uint16{'u'},
				DomainName:   [33]uint16{'d'},
				FarmName:     [33]uint16{'f'},
			},
		},
	})
}

// TestTS_COUNTER_RoundTrip covers a nested fixed struct with a LARGE_INTEGER.
func TestTS_COUNTER_RoundTrip(t *testing.T) {
	roundTrip(t, "TS_COUNTER", TS_COUNTER{
		CounterHead: TS_COUNTER_HEADER{DwCounterID: 42, BResult: true},
		DwValue:     1000,
		StartTime:   dtyp.LARGE_INTEGER(0x01d9a1b2c3d4e5f6),
	})
}

// TestNT6_TS_UNICODE_STRING_RoundTrip exercises the size_is(MaximumLength/2) /
// length_is(Length/2) divisor bounds on a conformant-varying wide buffer.
func TestNT6_TS_UNICODE_STRING_RoundTrip(t *testing.T) {
	roundTrip(t, "NT6_TS_UNICODE_STRING", NT6_TS_UNICODE_STRING{
		Length:        8, // bytes; 4 wide chars valid
		MaximumLength: 8,
		Buffer:        []uint16{'c', 'm', 'd', 0},
	})
	roundTrip(t, "NT6_TS_UNICODE_STRING/nil", NT6_TS_UNICODE_STRING{})
}

// TestSESSION_CHANGE_RoundTrip covers the notification change record.
func TestSESSION_CHANGE_RoundTrip(t *testing.T) {
	roundTrip(t, "SESSION_CHANGE", SESSION_CHANGE{SessionId: 7, NotificationId: WTS_NOTIFY_LOGON})
}

// TestTSVIP_SOCKADDR_RoundTrip covers the encapsulated union, selecting each arm.
func TestTSVIP_SOCKADDR_RoundTrip(t *testing.T) {
	roundTrip(t, "TSVIP_SOCKADDR/ipv4", TSVIP_SOCKADDR{
		SinFamily: 2,
		Ipv4:      TSVIP_SOCKADDR_IPV4{SinPort: 0x0d3d, InAddr: 0x7f000001, SinZero: [8]uint8{}},
	})
	roundTrip(t, "TSVIP_SOCKADDR/ipv6", TSVIP_SOCKADDR{
		SinFamily: 23,
		Ipv6:      TSVIP_SOCKADDR_IPV6{Sin6Port: 0x0d3d, Sin6Flowinfo: 1, Sin6Addr: [8]uint16{0xfe80, 0, 0, 0, 0, 0, 0, 1}, Sin6ScopeId: 2},
	})
}

// TestRCM_REMOTEADDRESS_RoundTrip covers the RpcGetRemoteAddress union.
func TestRCM_REMOTEADDRESS_RoundTrip(t *testing.T) {
	roundTrip(t, "RCM_REMOTEADDRESS/ipv4", RCM_REMOTEADDRESS{
		SinFamily: 2,
		Ipv4:      RCM_REMOTEADDRESS_IPV4{SinPort: 0x0d3d, InAddr: 0x0a000001},
	})
	roundTrip(t, "RCM_REMOTEADDRESS/ipv6", RCM_REMOTEADDRESS{
		SinFamily: 23,
		Ipv6:      RCM_REMOTEADDRESS_IPV6{Sin6Port: 1, Sin6Addr: [8]uint16{0x2001, 0xdb8}},
	})
}

// TestTSVIPSession_RoundTrip covers the nested TSVIPAddress (union + varying byte array).
func TestTSVIPSession_RoundTrip(t *testing.T) {
	roundTrip(t, "TSVIPSession", TSVIPSession{
		DwVersion: 1,
		SessionId: 5,
		SessionIP: TSVIPAddress{
			DwVersion:             1,
			IPAddress:             TSVIP_SOCKADDR{SinFamily: 2, Ipv4: TSVIP_SOCKADDR_IPV4{SinPort: 0, InAddr: 0xc0a80001}},
			PrefixOrSubnetMask:    24,
			PhysicalAddressLength: 6,
			PhysicalAddress:       [TSVIP_MAX_ADAPTER_ADDRESS_LENGTH]uint8{0x00, 0x0c, 0x29, 0xab, 0xcd, 0xef},
			LeaseExpires:          3600,
			T1:                    1800,
			T2:                    3150,
		},
	})
}

// TestTS_ALL_PROCESSES_INFO_RoundTrip covers a [unique] pointer to a struct plus a
// [unique] size_is SID blob.
func TestTS_ALL_PROCESSES_INFO_RoundTrip(t *testing.T) {
	roundTrip(t, "TS_ALL_PROCESSES_INFO", TS_ALL_PROCESSES_INFO{
		PTsProcessInfo: &TS_SYS_PROCESS_INFORMATION{
			NumberOfThreads: 4,
			UniqueProcessId: 1234,
			SessionId:       2,
			ImageName: TS_UNICODE_STRING{
				Length:        8,
				MaximumLength: 8,
				Buffer:        []uint16{'c', 'm', 'd', '\x00', 'e', 'x', 'e', '\x00'},
			},
			VirtualSize: 0x1000,
		},
		SizeOfSid: 4,
		PSid:      []byte{1, 5, 0, 0},
	})
	roundTrip(t, "TS_ALL_PROCESSES_INFO/nil", TS_ALL_PROCESSES_INFO{})
}
