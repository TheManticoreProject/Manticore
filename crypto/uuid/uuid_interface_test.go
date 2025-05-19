package uuid_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/uuid"
	"github.com/TheManticoreProject/Manticore/crypto/uuid/uuid_v1"
	"github.com/TheManticoreProject/Manticore/crypto/uuid/uuid_v2"
)

func TestUUIDInterfaceCompatibility(t *testing.T) {
	// Test that UUIDv1 implements UUIDInterface
	var v1 uuid_v1.UUIDv1
	var _ uuid.UUIDInterface = &v1

	// Test that UUIDv1 implements UUIDInterface
	var v2 uuid_v2.UUIDv2
	var _ uuid.UUIDInterface = &v2

	// // Test that UUIDv1 implements UUIDInterface
	// var v3 uuid_v3.UUIDv3
	// var _ uuid.UUIDInterface = &v3

	// // Test that UUIDv1 implements UUIDInterface
	// var v4 uuid_v4.UUIDv4
	// var _ uuid.UUIDInterface = &v4

	// // Test that UUIDv1 implements UUIDInterface
	// var v5 uuid_v5.UUIDv5
	// var _ uuid.UUIDInterface = &v5

	// // Test that UUIDv1 implements UUIDInterface
	// var v6 uuid_v6.UUIDv6
	// var _ uuid.UUIDInterface = &v6

	// // Test that UUIDv1 implements UUIDInterface
	// var v7 uuid_v7.UUIDv7
	// var _ uuid.UUIDInterface = &v7

	// // Test that UUIDv1 implements UUIDInterface
	// var v8 uuid_v8.UUIDv8
	// var _ uuid.UUIDInterface = &v8

}
