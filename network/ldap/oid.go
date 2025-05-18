package ldap

import (
	goldapv3 "github.com/go-ldap/ldap/v3"
)

// OID constants for LDAP controls
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/3c5e87db-4728-4f29-b164-01dd7d7391ea
const (
	// LDAP_PAGED_RESULT_OID_STRING is the OID for the LDAP paged results control
	// which allows clients to control the rate at which an LDAP server returns the
	// results of a search operation.
	LDAP_PAGED_RESULT_OID_STRING = "1.2.840.113556.1.4.319"

	// LDAP_SERVER_CROSSDOM_MOVE_TARGET_OID is the OID for the LDAP cross-domain move target control
	// which is used during cross-domain move operations.
	LDAP_SERVER_CROSSDOM_MOVE_TARGET_OID = "1.2.840.113556.1.4.521"

	// LDAP_SERVER_DIRSYNC_OID is the OID for the LDAP directory synchronization control
	// which is used to maintain directory synchronization.
	LDAP_SERVER_DIRSYNC_OID = "1.2.840.113556.1.4.841"

	// LDAP_SERVER_DOMAIN_SCOPE_OID is the OID for the LDAP domain scope control
	// which limits the scope of an operation to a single domain.
	LDAP_SERVER_DOMAIN_SCOPE_OID = "1.2.840.113556.1.4.1339"

	// LDAP_SERVER_EXTENDED_DN_OID is the OID for the LDAP extended DN control
	// which requests that the LDAP server return extended DNs with additional information.
	LDAP_SERVER_EXTENDED_DN_OID = "1.2.840.113556.1.4.529"

	// LDAP_SERVER_GET_STATS_OID is the OID for the LDAP get stats control
	// which retrieves statistics from the LDAP server.
	LDAP_SERVER_GET_STATS_OID = "1.2.840.113556.1.4.970"

	// LDAP_SERVER_LAZY_COMMIT_OID is the OID for the LDAP lazy commit control
	// which allows for delayed commitment of changes to the directory.
	LDAP_SERVER_LAZY_COMMIT_OID = "1.2.840.113556.1.4.619"

	// LDAP_SERVER_PERMISSIVE_MODIFY_OID is the OID for the LDAP permissive modify control
	// which allows modify operations to succeed when they would otherwise fail due to
	// attempts to add existing attributes or delete non-existent attributes.
	LDAP_SERVER_PERMISSIVE_MODIFY_OID = "1.2.840.113556.1.4.1413"

	// LDAP_SERVER_NOTIFICATION_OID is the OID for the LDAP notification control
	// which enables event notification for changes to the directory.
	LDAP_SERVER_NOTIFICATION_OID = "1.2.840.113556.1.4.528"

	// LDAP_SERVER_RESP_SORT_OID is the OID for the LDAP response sort control
	// which is used in the server's response to a sort request control.
	LDAP_SERVER_RESP_SORT_OID = "1.2.840.113556.1.4.474"

	// LDAP_SERVER_SD_FLAGS_OID is the OID for the LDAP security descriptor flags control
	// which specifies which portions of a security descriptor to retrieve.
	LDAP_SERVER_SD_FLAGS_OID = "1.2.840.113556.1.4.801"

	// LDAP_SERVER_SEARCH_OPTIONS_OID is the OID for the LDAP search options control
	// which provides additional options for search operations.
	LDAP_SERVER_SEARCH_OPTIONS_OID = "1.2.840.113556.1.4.1340"

	// LDAP_SERVER_SORT_OID is the OID for the LDAP sort control
	// which requests that the LDAP server sort search results before returning them.
	LDAP_SERVER_SORT_OID = "1.2.840.113556.1.4.473"

	// LDAP_SERVER_SHOW_DELETED_OID is the OID for the LDAP show deleted control
	// which requests that the LDAP server return deleted objects in search results.
	LDAP_SERVER_SHOW_DELETED_OID = "1.2.840.113556.1.4.417"

	// LDAP_SERVER_TREE_DELETE_OID is the OID for the LDAP tree delete control
	// which requests that the LDAP server delete an entire subtree in a single operation.
	LDAP_SERVER_TREE_DELETE_OID = "1.2.840.113556.1.4.805"

	// LDAP_SERVER_VERIFY_NAME_OID is the OID for the LDAP verify name control
	// which verifies the name of the target object for an operation.
	LDAP_SERVER_VERIFY_NAME_OID = "1.2.840.113556.1.4.1338"

	// LDAP_CONTROL_VLVREQUEST is the OID for the LDAP virtual list view request control
	// which allows clients to retrieve a subset of search results based on positional information.
	LDAP_CONTROL_VLVREQUEST = "2.16.840.1.113730.3.4.9"

	// LDAP_CONTROL_VLVRESPONSE is the OID for the LDAP virtual list view response control
	// which is used in the server's response to a virtual list view request control.
	LDAP_CONTROL_VLVRESPONSE = "2.16.840.1.113730.3.4.10"

	// LDAP_SERVER_ASQ_OID is the OID for the LDAP attribute scoped query control
	// which allows searching based on values of attributes in a specific entry.
	LDAP_SERVER_ASQ_OID = "1.2.840.113556.1.4.1504"

	// LDAP_SERVER_QUOTA_CONTROL_OID is the OID for the LDAP quota control
	// which is used to manage directory quotas.
	LDAP_SERVER_QUOTA_CONTROL_OID = "1.2.840.113556.1.4.1852"

	// LDAP_SERVER_RANGE_OPTION_OID is the OID for the LDAP range option control
	// which is used to retrieve large multi-valued attributes in chunks.
	LDAP_SERVER_RANGE_OPTION_OID = "1.2.840.113556.1.4.802"

	// LDAP_SERVER_SHUTDOWN_NOTIFY_OID is the OID for the LDAP shutdown notification control
	// which notifies clients of server shutdown.
	LDAP_SERVER_SHUTDOWN_NOTIFY_OID = "1.2.840.113556.1.4.1907"

	// LDAP_SERVER_FORCE_UPDATE_OID is the OID for the LDAP force update control
	// which forces updates to the directory even when they would normally be rejected.
	LDAP_SERVER_FORCE_UPDATE_OID = "1.2.840.113556.1.4.1974"

	// LDAP_SERVER_RANGE_RETRIEVAL_NOERR_OID is the OID for the LDAP range retrieval no error control
	// which suppresses errors during range retrieval operations.
	LDAP_SERVER_RANGE_RETRIEVAL_NOERR_OID = "1.2.840.113556.1.4.1948"

	// LDAP_SERVER_RODC_DCPROMO_OID is the OID for the LDAP RODC DCPromo control
	// which is used during the promotion of a read-only domain controller.
	LDAP_SERVER_RODC_DCPROMO_OID = "1.2.840.113556.1.4.1341"

	// LDAP_SERVER_DN_INPUT_OID is the OID for the LDAP DN input control
	// which specifies the format of distinguished names in the request.
	LDAP_SERVER_DN_INPUT_OID = "1.2.840.113556.1.4.2026"

	// LDAP_SERVER_SHOW_DEACTIVATED_LINK_OID is the OID for the LDAP show deactivated link control
	// which requests that the LDAP server return deactivated links in search results.
	LDAP_SERVER_SHOW_DEACTIVATED_LINK_OID = "1.2.840.113556.1.4.2065"

	// LDAP_SERVER_SHOW_RECYCLED_OID is the OID for the LDAP show recycled control
	// which requests that the LDAP server return recycled objects in search results.
	LDAP_SERVER_SHOW_RECYCLED_OID = "1.2.840.113556.1.4.2064"

	// LDAP_SERVER_POLICY_HINTS_DEPRECATED_OID is the OID for the deprecated LDAP policy hints control
	// which provided hints about policy enforcement to the LDAP server.
	LDAP_SERVER_POLICY_HINTS_DEPRECATED_OID = "1.2.840.113556.1.4.2066"

	// LDAP_SERVER_DIRSYNC_EX_OID is the OID for the LDAP extended directory synchronization control
	// which provides enhanced directory synchronization capabilities.
	LDAP_SERVER_DIRSYNC_EX_OID = "1.2.840.113556.1.4.2090"

	// LDAP_SERVER_UPDATE_STATS_OID is the OID for the LDAP update stats control
	// which updates statistics on the LDAP server.
	LDAP_SERVER_UPDATE_STATS_OID = "1.2.840.113556.1.4.2205"

	// LDAP_SERVER_TREE_DELETE_EX_OID is the OID for the LDAP extended tree delete control
	// which provides enhanced capabilities for deleting entire subtrees.
	LDAP_SERVER_TREE_DELETE_EX_OID = "1.2.840.113556.1.4.2204"

	// LDAP_SERVER_SEARCH_HINTS_OID is the OID for the LDAP search hints control
	// which provides hints to optimize search operations.
	LDAP_SERVER_SEARCH_HINTS_OID = "1.2.840.113556.1.4.2206"

	// LDAP_SERVER_EXPECTED_ENTRY_COUNT_OID is the OID for the LDAP expected entry count control
	// which specifies the expected number of entries in a search result.
	LDAP_SERVER_EXPECTED_ENTRY_COUNT_OID = "1.2.840.113556.1.4.2211"

	// LDAP_SERVER_POLICY_HINTS_OID is the OID for the LDAP policy hints control
	// which provides hints about policy enforcement to the LDAP server.
	LDAP_SERVER_POLICY_HINTS_OID = "1.2.840.113556.1.4.2239"

	// LDAP_SERVER_SET_OWNER_OID is the OID for the LDAP set owner control
	// which sets the owner of an object during creation or modification.
	LDAP_SERVER_SET_OWNER_OID = "1.2.840.113556.1.4.2255"

	// LDAP_SERVER_BYPASS_QUOTA_OID is the OID for the LDAP bypass quota control
	// which allows operations to bypass quota restrictions.
	LDAP_SERVER_BYPASS_QUOTA_OID = "1.2.840.113556.1.4.2256"

	// LDAP_SERVER_LINK_TTL_OID is the OID for the LDAP link TTL control
	// which specifies the time-to-live for linked attributes.
	LDAP_SERVER_LINK_TTL_OID = "1.2.840.113556.1.4.2309"

	// LDAP_SERVER_SET_CORRELATION_ID_OID is the OID for the LDAP set correlation ID control
	// which sets a correlation ID for tracking operations across multiple requests.
	LDAP_SERVER_SET_CORRELATION_ID_OID = "1.2.840.113556.1.4.2330"

	// LDAP_SERVER_THREAD_TRACE_OVERRIDE_OID is the OID for the LDAP thread trace override control
	// which overrides thread tracing settings for diagnostic purposes.
	LDAP_SERVER_THREAD_TRACE_OVERRIDE_OID = "1.2.840.113556.1.4.2354"
)

// NewControlsWithOIDs creates multiple controls with the specified OIDs.
// This function returns a slice of structs that implement the goldapv3.Control interface.
//
// Parameters:
//   - oids: A slice of strings representing the Object Identifiers for the controls.
//   - criticality: A boolean indicating whether the controls are critical.
//
// Returns:
//   - A slice of goldapv3.Control that can be used with LDAP operations.
//
// Example usage:
//
//	controls := NewControlsWithOIDs([]string{LDAP_SERVER_PERMISSIVE_MODIFY_OID}, false)
//	modifyRequest := NewModifyRequest("cn=John Doe,dc=example,dc=com")
//	for _, control := range controls {
//	    modifyRequest.AddControl(control)
//	}
func NewControlsWithOIDs(oids []string, criticality bool) []goldapv3.Control {
	controls := make([]goldapv3.Control, len(oids))
	for i, oid := range oids {
		controls[i] = &goldapv3.ControlString{
			ControlType:  oid,
			Criticality:  criticality,
			ControlValue: "",
		}
	}
	return controls
}
