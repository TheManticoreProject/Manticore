//go:build unix

package server

import "syscall"

// openNoFollow makes an open fail rather than follow a symbolic link in the final
// path element.
const openNoFollow = syscall.O_NOFOLLOW
