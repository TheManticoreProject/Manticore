package types

import "github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_types"

// SMB_FEA_LIST
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/1ca1684e-6552-432c-bdd0-f559814bbaef
type SMB_FEA_LIST struct {
	SizeOfListInBytes data_types.ULONG
	FEAList           []data_types.UCHAR
}
