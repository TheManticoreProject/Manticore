package msscmr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVICE_FAILURE_ACTIONSA ([MS-SCMR] 2.2.18) is the ANSI failure-actions record. The
// message and command are [unique] pointers to conformant-varying ASCII arrays and the
// action array is a [unique] pointer to a conformant array sized by CActions.
type SERVICE_FAILURE_ACTIONSA struct {
	DwResetPeriod ndr.DWORD
	LpRebootMsg   *ndr.STR `ndr:"unique"`
	LpCommand     *ndr.STR `ndr:"unique"`
	CActions      ndr.DWORD
	LpsaActions   []SC_ACTION `ndr:"unique,size_is=CActions"`
}
