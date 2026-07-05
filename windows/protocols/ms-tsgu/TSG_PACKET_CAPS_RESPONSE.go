package mstsgu

// TSG_PACKET_CAPS_RESPONSE is the capabilities response returned by the RDG server
// ([MS-TSGU] 2.2.9.2.1.6): the quarantine-encapsulated response followed by the
// consent/service message response.
//
// Wire-order caveat: PktQuarEncResponse carries [unique] pointers (CertChainData,
// VersionCaps) whose referent bodies are deferred to the end of this struct, while
// PktConsentMessage embeds a union (TSG_PACKET_TYPE_MESSAGE_UNION) whose pointer arm is
// flushed locally by the codec (network/dcerpc/ndr/union.go). When both the QUARENC
// pointers and the message-union arm are populated at once, the arm body is emitted
// ahead of the QUARENC referent bodies rather than in strict field order. This is a
// shared-codec deferral limitation for a union nested beside sibling pointers, not a
// modeling error here; round-trips are self-consistent, but this specific mix is not
// wire-validated against a live RDG server.
type TSG_PACKET_CAPS_RESPONSE struct {
	PktQuarEncResponse TSG_PACKET_QUARENC_RESPONSE
	PktConsentMessage  TSG_PACKET_MSG_RESPONSE
}
