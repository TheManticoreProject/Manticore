package msdnsp

// DNS_RPC_RECORD_NODE_NAME is the record-data payload for records whose data is a single
// FQDN. Per [MS-DNSP] it is used for the following record types:
//
//	DNS_TYPE_PTR, DNS_TYPE_NS, DNS_TYPE_CNAME, DNS_TYPE_DNAME,
//	DNS_TYPE_MB, DNS_TYPE_MR, DNS_TYPE_MG, DNS_TYPE_MD, DNS_TYPE_MF
//
// Although the specification documents nameNode in DNS_RPC_NAME (section 2.2.2.2.1) form, in a
// dnsRecord attribute value carried over LDAP the name is stored in DNS_COUNT_NAME
// (section 2.2.2.2.2) form.
//
// Source: [MS-DNSP] DNS_RPC_RECORD_NODE_NAME (section 2.2.2.2.4.2)
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/8f986756-f151-4f5b-bfcf-0d85be8b0d7e
type DNS_RPC_RECORD_NODE_NAME struct {
	// NameNode (variable): The FQDN of this node.
	NameNode DNS_COUNT_NAME
}

// NewDNS_RPC_RECORD_NODE_NAME creates a new, empty DNS_RPC_RECORD_NODE_NAME.
//
// Returns:
// - A pointer to the new DNS_RPC_RECORD_NODE_NAME structure
func NewDNS_RPC_RECORD_NODE_NAME() *DNS_RPC_RECORD_NODE_NAME {
	return &DNS_RPC_RECORD_NODE_NAME{}
}

// Marshal marshals the DNS_RPC_RECORD_NODE_NAME structure into a byte array.
//
// Returns:
// - A byte array representing the DNS_RPC_RECORD_NODE_NAME structure
// - An error if the marshaling fails
func (r *DNS_RPC_RECORD_NODE_NAME) Marshal() ([]byte, error) {
	return r.NameNode.Marshal()
}

// Unmarshal unmarshals a byte array into the DNS_RPC_RECORD_NODE_NAME structure.
//
// Parameters:
// - rawData: The byte array to unmarshal
//
// Returns:
// - The number of bytes read
// - An error if the unmarshaling fails
func (r *DNS_RPC_RECORD_NODE_NAME) Unmarshal(rawData []byte) (int, error) {
	return r.NameNode.Unmarshal(rawData)
}
