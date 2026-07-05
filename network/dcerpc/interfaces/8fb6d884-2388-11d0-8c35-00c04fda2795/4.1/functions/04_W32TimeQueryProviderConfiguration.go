package functions

import (
	"fmt"

	W32Time "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8fb6d884-2388-11d0-8c35-00c04fda2795/4.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msw32t "github.com/TheManticoreProject/Manticore/windows/protocols/ms-w32t"
)

// w32TimeQueryProviderConfigurationRequest carries the [in] parameters of W32TimeQueryProviderConfiguration.
type w32TimeQueryProviderConfigurationRequest struct {
	UlFlags      ndr.DWORD
	PwszProvider ndr.WSTR
}

func (*w32TimeQueryProviderConfigurationRequest) Opnum() uint16 {
	return W32Time.OpnumW32TimeQueryProviderConfiguration
}

// w32TimeQueryProviderConfigurationResponse carries the [out] parameters and return value of W32TimeQueryProviderConfiguration.
type w32TimeQueryProviderConfigurationResponse struct {
	PConfigurationProviderInfo *msw32t.W32TIME_CONFIGURATION_PROVIDER `ndr:"unique"`
	Status                     ndr.DWORD                              `ndr:"retval"`
}

// W32TimeQueryProviderConfiguration calls W32TimeQueryProviderConfiguration (opnum 4) ([MS-W32T] section 3.2.4).
func W32TimeQueryProviderConfiguration(rpc ndr.Invoker, ulFlags ndr.DWORD, pwszProvider ndr.WSTR) (PConfigurationProviderInfo *msw32t.W32TIME_CONFIGURATION_PROVIDER, err error) {
	req := &w32TimeQueryProviderConfigurationRequest{
		UlFlags:      ulFlags,
		PwszProvider: pwszProvider,
	}
	var resp w32TimeQueryProviderConfigurationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("W32TimeQueryProviderConfiguration: %w", err)
		return
	}
	PConfigurationProviderInfo = resp.PConfigurationProviderInfo
	if uint32(resp.Status) != W32Time.StatusSuccess {
		err = fmt.Errorf("W32TimeQueryProviderConfiguration failed: %s", W32Time.StatusString(uint32(resp.Status)))
	}
	return
}
