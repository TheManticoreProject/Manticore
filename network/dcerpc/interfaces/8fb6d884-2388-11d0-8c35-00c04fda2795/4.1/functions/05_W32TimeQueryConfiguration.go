package functions

import (
	"fmt"

	W32Time "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8fb6d884-2388-11d0-8c35-00c04fda2795/4.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msw32t "github.com/TheManticoreProject/Manticore/windows/protocols/ms-w32t"
)

// w32TimeQueryConfigurationRequest carries the [in] parameters of W32TimeQueryConfiguration.
type w32TimeQueryConfigurationRequest struct {
}

func (*w32TimeQueryConfigurationRequest) Opnum() uint16 {
	return W32Time.OpnumW32TimeQueryConfiguration
}

// w32TimeQueryConfigurationResponse carries the [out] parameters and return value of W32TimeQueryConfiguration.
type w32TimeQueryConfigurationResponse struct {
	PConfigurationInfo *msw32t.W32TIME_CONFIGURATION_INFO `ndr:"unique"`
	Status             ndr.DWORD                          `ndr:"retval"`
}

// W32TimeQueryConfiguration calls W32TimeQueryConfiguration (opnum 5) ([MS-W32T] section 3.2.4).
func W32TimeQueryConfiguration(rpc ndr.Invoker) (PConfigurationInfo *msw32t.W32TIME_CONFIGURATION_INFO, err error) {
	req := &w32TimeQueryConfigurationRequest{}
	var resp w32TimeQueryConfigurationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("W32TimeQueryConfiguration: %w", err)
		return
	}
	PConfigurationInfo = resp.PConfigurationInfo
	if uint32(resp.Status) != W32Time.StatusSuccess {
		err = fmt.Errorf("W32TimeQueryConfiguration failed: %s", W32Time.StatusString(uint32(resp.Status)))
	}
	return
}
