//go:build !unix

package server

// openNonBlock is zero where the host has no FIFOs to block an open.
const openNonBlock = 0
