package functions

// IDL source: [MS-DNSP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/83136c8e-f5ea-4ec5-bf33-2134053d33bd
// A fetched copy is kept at ms-dnsp.idl in the interface directory.

import (
	"fmt"

	DnsServer "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/50abc2a4-574d-40b3-9d66-ee4fd5fba076/5.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DnssrvEnumRecords2Request carries the [in] parameters of R_DnssrvEnumRecords2.
type r_DnssrvEnumRecords2Request struct {
	DwClientVersion ndr.DWORD
	DwSettingFlags  ndr.DWORD
	PwszServerName  *ndr.WSTR `ndr:"unique"`
	PszZone         *ndr.STR  `ndr:"unique"`
	PszNodeName     *ndr.STR  `ndr:"unique"`
	PszStartChild   *ndr.STR  `ndr:"unique"`
	WRecordType     uint16
	FSelectFlag     ndr.DWORD
	PszFilterStart  *ndr.STR `ndr:"unique"`
	PszFilterStop   *ndr.STR `ndr:"unique"`
}

func (*r_DnssrvEnumRecords2Request) Opnum() uint16 { return DnsServer.OpnumR_DnssrvEnumRecords2 }

// r_DnssrvEnumRecords2Response carries the [out] parameters and return value of R_DnssrvEnumRecords2.
type r_DnssrvEnumRecords2Response struct {
	PdwBufferLength ndr.DWORD
	PpBuffer        []uint8   `ndr:"unique,size_is=PdwBufferLength"`
	Status          ndr.DWORD `ndr:"retval"`
}

// R_DnssrvEnumRecords2 calls R_DnssrvEnumRecords2 (opnum 8) ([MS-DNSP] — verify the parameter
// modeling and status handling).
func R_DnssrvEnumRecords2(rpc ndr.Invoker, dwClientVersion ndr.DWORD, dwSettingFlags ndr.DWORD, pwszServerName *ndr.WSTR, pszZone *ndr.STR, pszNodeName *ndr.STR, pszStartChild *ndr.STR, wRecordType uint16, fSelectFlag ndr.DWORD, pszFilterStart *ndr.STR, pszFilterStop *ndr.STR) (PdwBufferLength ndr.DWORD, PpBuffer []uint8, err error) {
	req := &r_DnssrvEnumRecords2Request{
		DwClientVersion: dwClientVersion,
		DwSettingFlags:  dwSettingFlags,
		PwszServerName:  pwszServerName,
		PszZone:         pszZone,
		PszNodeName:     pszNodeName,
		PszStartChild:   pszStartChild,
		WRecordType:     wRecordType,
		FSelectFlag:     fSelectFlag,
		PszFilterStart:  pszFilterStart,
		PszFilterStop:   pszFilterStop,
	}
	var resp r_DnssrvEnumRecords2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DnssrvEnumRecords2: %w", err)
		return
	}
	PdwBufferLength = resp.PdwBufferLength
	PpBuffer = resp.PpBuffer
	if uint32(resp.Status) != DnsServer.StatusSuccess {
		err = fmt.Errorf("R_DnssrvEnumRecords2 failed: %s", DnsServer.StatusString(uint32(resp.Status)))
	}
	return
}
