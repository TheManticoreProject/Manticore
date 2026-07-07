package dnsrecord

import (
	"encoding/binary"
	"fmt"
	"time"
)

// ticksPerSecond is the number of 100-nanosecond intervals in one second.
const ticksPerSecond = 10_000_000

// epochDeltaSeconds is the number of seconds between the Windows FILETIME epoch
// (1601-01-01 00:00:00 UTC) and the Unix epoch (1970-01-01 00:00:00 UTC).
const epochDeltaSeconds = 11_644_473_600

// DNS_RPC_RECORD_TS is the tombstone record-data payload. When the last record set for a node
// in a directory-server-integrated zone is deleted, the DNS server sets the node's
// dnsTombstoned attribute to TRUE and writes a dnsRecord value of type DNS_TYPE_ZERO whose
// data is a DNS_RPC_RECORD_TS.
//
// EntombedTime is stored little-endian, unlike the big-endian numeric fields of the SOA, SRV,
// and NAME_PREFERENCE payloads.
//
// Source: [MS-DNSP] DNS_RPC_RECORD_TS (section 2.2.2.2.4.23); tombstone handling is described
// in Creating and Deleting a DNS Record,
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/c87b1caf-904d-4d8a-a7d1-3a6908112cd3
type DNS_RPC_RECORD_TS struct {
	// EntombedTime (8 bytes): The time at which the record was tombstoned, expressed as the
	// number of 100-nanosecond intervals since 1601-01-01 00:00:00 UTC. Little-endian.
	EntombedTime uint64
}

// NewDNS_RPC_RECORD_TS creates a new, empty DNS_RPC_RECORD_TS.
//
// Returns:
// - A pointer to the new DNS_RPC_RECORD_TS structure
func NewDNS_RPC_RECORD_TS() *DNS_RPC_RECORD_TS {
	return &DNS_RPC_RECORD_TS{}
}

// SetTime sets EntombedTime from a time.Time.
//
// Parameters:
// - t: The tombstone time
func (r *DNS_RPC_RECORD_TS) SetTime(t time.Time) {
	utc := t.UTC()
	ticks := (utc.Unix()+epochDeltaSeconds)*ticksPerSecond + int64(utc.Nanosecond())/100
	r.EntombedTime = uint64(ticks)
}

// GetTime returns EntombedTime as a UTC time.Time.
//
// Returns:
// - The tombstone time in UTC
func (r *DNS_RPC_RECORD_TS) GetTime() time.Time {
	ticks := int64(r.EntombedTime)
	secs := ticks/ticksPerSecond - epochDeltaSeconds
	nsec := (ticks % ticksPerSecond) * 100
	return time.Unix(secs, nsec).UTC()
}

// Marshal marshals the DNS_RPC_RECORD_TS structure into a byte array.
//
// Returns:
// - A byte array representing the DNS_RPC_RECORD_TS structure
// - An error if the marshaling fails
func (r *DNS_RPC_RECORD_TS) Marshal() ([]byte, error) {
	marshalled := make([]byte, 8)
	binary.LittleEndian.PutUint64(marshalled, r.EntombedTime)
	return marshalled, nil
}

// Unmarshal unmarshals a byte array into the DNS_RPC_RECORD_TS structure.
//
// Parameters:
// - rawData: The byte array to unmarshal
//
// Returns:
// - The number of bytes read
// - An error if the unmarshaling fails
func (r *DNS_RPC_RECORD_TS) Unmarshal(rawData []byte) (int, error) {
	if len(rawData) < 8 {
		return 0, fmt.Errorf("rawData too short for DNS_RPC_RECORD_TS: need 8 bytes, have %d", len(rawData))
	}
	r.EntombedTime = binary.LittleEndian.Uint64(rawData[:8])
	return 8, nil
}
