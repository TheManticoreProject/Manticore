package functions

import (
	"fmt"

	DnsServer "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/50abc2a4-574d-40b3-9d66-ee4fd5fba076/5.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DnssrvEnumRecords4Request carries the [in] parameters of R_DnssrvEnumRecords4.
type r_DnssrvEnumRecords4Request struct {
	DwClientVersion              ndr.DWORD
	DwSettingFlags               ndr.DWORD
	PwszServerName               *ndr.WSTR `ndr:"unique"`
	PwszVirtualizationInstanceID *ndr.WSTR `ndr:"unique"`
	PszZone                      *ndr.STR  `ndr:"unique"`
	PwszZoneScope                *ndr.WSTR `ndr:"unique"`
	PszNodeName                  *ndr.STR  `ndr:"unique"`
	PszStartChild                *ndr.STR  `ndr:"unique"`
	WRecordType                  uint16
	FSelectFlag                  ndr.DWORD
	PszFilterStart               *ndr.STR `ndr:"unique"`
	PszFilterStop                *ndr.STR `ndr:"unique"`
}

func (*r_DnssrvEnumRecords4Request) Opnum() uint16 { return DnsServer.OpnumR_DnssrvEnumRecords4 }

// r_DnssrvEnumRecords4Response carries the [out] parameters and return value of R_DnssrvEnumRecords4.
type r_DnssrvEnumRecords4Response struct {
	PdwBufferLength ndr.DWORD
	PpBuffer        []uint8   `ndr:"unique,size_is=PdwBufferLength"`
	Status          ndr.DWORD `ndr:"retval"`
}

// R_DnssrvEnumRecords4 calls R_DnssrvEnumRecords4 (opnum 18) ([MS-DNSP] — verify the parameter
// modeling and status handling).
func R_DnssrvEnumRecords4(rpc ndr.Invoker, dwClientVersion ndr.DWORD, dwSettingFlags ndr.DWORD, pwszServerName *ndr.WSTR, pwszVirtualizationInstanceID *ndr.WSTR, pszZone *ndr.STR, pwszZoneScope *ndr.WSTR, pszNodeName *ndr.STR, pszStartChild *ndr.STR, wRecordType uint16, fSelectFlag ndr.DWORD, pszFilterStart *ndr.STR, pszFilterStop *ndr.STR) (PdwBufferLength ndr.DWORD, PpBuffer []uint8, err error) {
	req := &r_DnssrvEnumRecords4Request{
		DwClientVersion:              dwClientVersion,
		DwSettingFlags:               dwSettingFlags,
		PwszServerName:               pwszServerName,
		PwszVirtualizationInstanceID: pwszVirtualizationInstanceID,
		PszZone:                      pszZone,
		PwszZoneScope:                pwszZoneScope,
		PszNodeName:                  pszNodeName,
		PszStartChild:                pszStartChild,
		WRecordType:                  wRecordType,
		FSelectFlag:                  fSelectFlag,
		PszFilterStart:               pszFilterStart,
		PszFilterStop:                pszFilterStop,
	}
	var resp r_DnssrvEnumRecords4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DnssrvEnumRecords4: %w", err)
		return
	}
	PdwBufferLength = resp.PdwBufferLength
	PpBuffer = resp.PpBuffer
	if uint32(resp.Status) != DnsServer.StatusSuccess {
		err = fmt.Errorf("R_DnssrvEnumRecords4 failed: %s", DnsServer.StatusString(uint32(resp.Status)))
	}
	return
}
