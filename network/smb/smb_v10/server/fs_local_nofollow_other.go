//go:build !unix

package server

// openNoFollow is zero where the host has no flag refusing to follow a symbolic
// link in the final path element; hostPath's own check still applies there.
const openNoFollow = 0
