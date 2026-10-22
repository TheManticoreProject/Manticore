package server

import (
	"strings"
	"sync"

	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// Generic rights as they map onto a file ([MS-SMB2] 3.3.5.9, [MS-DTYP] 2.4.3).
const (
	fileGenericRead    uint32 = 0x00120089
	fileGenericWrite   uint32 = 0x00120116
	fileGenericExecute uint32 = 0x001200A0

	// writeRights are the rights a read-only share never grants.
	writeRights = fileflags.FILE_WRITE_DATA | fileflags.FILE_APPEND_DATA | fileflags.FILE_WRITE_EA |
		fileflags.FILE_WRITE_ATTRIBUTES | fileflags.FILE_DELETE_CHILD | fileflags.DELETE |
		fileflags.WRITE_DAC | fileflags.WRITE_OWNER

	// dataRights are the rights the share modes arbitrate between opens.
	readDataRights  = fileflags.FILE_READ_DATA | fileflags.FILE_EXECUTE
	writeDataRights = fileflags.FILE_WRITE_DATA | fileflags.FILE_APPEND_DATA
	dataRights      = readDataRights | writeDataRights | fileflags.DELETE
)

// grantedAccess expands a CREATE's DesiredAccess into the rights the open
// holds. MAXIMUM_ALLOWED grants whatever the share allows; asking a read-only
// share for anything that modifies is refused.
func grantedAccess(desired uint32, readOnly bool) (uint32, nt_status.NT_STATUS) {
	granted := desired &^ (fileflags.GENERIC_READ | fileflags.GENERIC_WRITE | fileflags.GENERIC_EXECUTE | fileflags.GENERIC_ALL | maximumAllowed)
	for generic, rights := range map[uint32]uint32{
		fileflags.GENERIC_READ:    fileGenericRead,
		fileflags.GENERIC_WRITE:   fileGenericWrite,
		fileflags.GENERIC_EXECUTE: fileGenericExecute,
		fileflags.GENERIC_ALL:     fileAllAccess,
	} {
		if desired&generic != 0 {
			granted |= rights
		}
	}
	if readOnly && granted&writeRights != 0 {
		return 0, nt_status.NT_STATUS_ACCESS_DENIED
	}
	if desired&maximumAllowed != 0 {
		granted |= fileAllAccess
		if readOnly {
			granted = fileReadAccess
		}
	}
	return granted & fileAllAccess, nt_status.NT_STATUS_SUCCESS
}

// require refuses an operation the open was not granted any of rights for.
func (o *Open) require(rights uint32) nt_status.NT_STATUS {
	if o.access&rights == 0 {
		return nt_status.NT_STATUS_ACCESS_DENIED
	}
	return nt_status.NT_STATUS_SUCCESS
}

// shareModes records the opens of every file that hold data access, across
// connections, so an open whose access or share mode conflicts with theirs is
// refused with STATUS_SHARING_VIOLATION ([MS-FSA] 2.1.5.1.2).
type shareModes struct {
	mutex sync.Mutex
	opens map[string]map[*Open]struct{}
}

func newShareModes() *shareModes {
	return &shareModes{opens: make(map[string]map[*Open]struct{})}
}

// fileKey identifies a file across connections.
func fileKey(share *Share, path string) string {
	return strings.ToUpper(share.Name) + "/" + strings.ToUpper(path)
}

// conflicts reports whether a's access is something b does not share.
func conflicts(a, b *Open) bool {
	return a.access&readDataRights != 0 && b.shareAccess&fileflags.FILE_SHARE_READ == 0 ||
		a.access&writeDataRights != 0 && b.shareAccess&fileflags.FILE_SHARE_WRITE == 0 ||
		a.access&fileflags.DELETE != 0 && b.shareAccess&fileflags.FILE_SHARE_DELETE == 0
}

// acquire registers an open unless it conflicts with one already there. An
// open without data access neither conflicts nor counts.
func (t *shareModes) acquire(open *Open) bool {
	if open.access&dataRights == 0 {
		return true
	}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	key := fileKey(open.tree.Share, open.path)
	for other := range t.opens[key] {
		if conflicts(open, other) || conflicts(other, open) {
			return false
		}
	}
	if t.opens[key] == nil {
		t.opens[key] = make(map[*Open]struct{})
	}
	t.opens[key][open] = struct{}{}
	return true
}

// release forgets an open.
func (t *shareModes) release(open *Open) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	key := fileKey(open.tree.Share, open.path)
	delete(t.opens[key], open)
	if len(t.opens[key]) == 0 {
		delete(t.opens, key)
	}
}

// rename moves an open to the file's new name.
func (t *shareModes) rename(open *Open, target string) {
	if open.access&dataRights == 0 {
		return
	}
	t.release(open)
	t.mutex.Lock()
	defer t.mutex.Unlock()
	key := fileKey(open.tree.Share, target)
	if t.opens[key] == nil {
		t.opens[key] = make(map[*Open]struct{})
	}
	t.opens[key][open] = struct{}{}
}
