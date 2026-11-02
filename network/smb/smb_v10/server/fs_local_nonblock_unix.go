//go:build unix

package server

import "syscall"

// openNonBlock makes an open of a FIFO return at once instead of waiting for
// the other end; the entry's type is checked on the handle afterwards. It has no
// effect on a regular file or a directory.
const openNonBlock = syscall.O_NONBLOCK
