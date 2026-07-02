package msdssp

// DSROLER_PRIMARY_DOMAIN_INFORMATION is the [switch_type(DSROLE_PRIMARY_DOMAIN_INFO_LEVEL)]
// discriminated union returned by DsRolerGetPrimaryDomainInformation ([MS-DSSP] 2.2.7).
// The 16-bit NDR enum discriminant (Tag) precedes the selected arm ([C706] 14.3.8); the
// case labels are the DSROLE_PRIMARY_DOMAIN_INFO_LEVEL values (Basic=1, UpgradeStatus=2,
// OperationState=3).
type DSROLER_PRIMARY_DOMAIN_INFORMATION struct {
	Tag                DSROLE_PRIMARY_DOMAIN_INFO_LEVEL  `ndr:"switch"`
	DomainInfoBasic    DSROLER_PRIMARY_DOMAIN_INFO_BASIC `ndr:"case=1"`
	UpgradStatusInfo   DSROLE_UPGRADE_STATUS_INFO        `ndr:"case=2"`
	OperationStateInfo DSROLE_OPERATION_STATE_INFO       `ndr:"case=3"`
}
