package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// POLICY_DOMAIN_QUALITY_OF_SERVICE_INFO is an obsolete policy domain information class
// ([MS-LSAD] 2.2.4.18).
type POLICY_DOMAIN_QUALITY_OF_SERVICE_INFO struct {
	QualityOfService ndr.DWORD
}
