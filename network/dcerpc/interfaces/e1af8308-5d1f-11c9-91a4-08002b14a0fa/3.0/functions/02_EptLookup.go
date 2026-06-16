package functions

import (
	"fmt"

	epm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// eptLookupRequest carries the [in] parameters of ept_lookup (opnum 2). The binding
// handle (handle_t) is not part of the NDR stub and is omitted. Object and Ifid are the
// optional [in, ptr] object-UUID and interface filters (nil for an all-elements
// enumeration); EntryHandle is the 20-octet context handle (null on the first call);
// MaxEnts caps the batch ([MS-RPCE] range(0,500)).
type eptLookupRequest struct {
	InquiryType ndr.DWORD
	Object      *structures.EptUUID `ndr:"ptr"`
	Ifid        *structures.RpcIfID `ndr:"ptr"`
	VersOption  ndr.DWORD
	EntryHandle structures.ContextHandle
	MaxEnts     ndr.DWORD
}

func (*eptLookupRequest) Opnum() uint16 { return epm.OpnumEptLookup }

// eptLookupResponse carries the [out] parameters of ept_lookup: the advanced context
// handle, the number of entries returned, the entries themselves (a full pointer to a
// conformant-varying array of ept_entry_t — size_is(max_ents), length_is(num_ents)), and
// the status.
type eptLookupResponse struct {
	EntryHandle structures.ContextHandle
	NumEnts     ndr.DWORD
	Entries     []structures.EptEntry `ndr:"ptr,varying"`
	Status      ndr.DWORD
}

// EptLookup calls ept_lookup (opnum 2) ([C706] Appendix O, [MS-RPCE] 2.2.1.2.4) once,
// returning a single batch of endpoint-map entries. inquiryType selects what to match
// (use epm.EptInquiryAllElts to enumerate everything); object and ifid are the optional
// filters; versOption is the interface-version constraint; entryHandle is null on the
// first call and the value returned by the previous call thereafter. It returns the batch
// of entries, the advanced entry handle, and the raw [out] status (the caller decides how
// to treat ept_s_not_registered); err is non-nil only for transport/decoding failures.
// Lookup wraps this with the paging loop for the common "enumerate everything" case.
func EptLookup(rpc ndr.Invoker, inquiryType uint32, object *guid.GUID, ifid *structures.RpcIfID, versOption uint32, entryHandle structures.ContextHandle, maxEnts uint32) (entries []structures.EptEntry, next structures.ContextHandle, status uint32, err error) {
	req := &eptLookupRequest{
		InquiryType: ndr.DWORD(inquiryType),
		Ifid:        ifid,
		VersOption:  ndr.DWORD(versOption),
		EntryHandle: entryHandle,
		MaxEnts:     ndr.DWORD(maxEnts),
	}
	if object != nil {
		u := structures.NewEptUUID(*object)
		req.Object = &u
	}

	var resp eptLookupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		return nil, structures.ContextHandle{}, 0, fmt.Errorf("ept_lookup: %w", err)
	}

	// num_ents is the authoritative count; clamp to the decoded array as a guard.
	n := int(resp.NumEnts)
	if n > len(resp.Entries) {
		n = len(resp.Entries)
	}
	return resp.Entries[:n], resp.EntryHandle, uint32(resp.Status), nil
}

// eptLookupHandleFreeRequest carries the [in, out] context handle of
// ept_lookup_handle_free (opnum 1).
type eptLookupHandleFreeRequest struct {
	EntryHandle structures.ContextHandle
}

func (*eptLookupHandleFreeRequest) Opnum() uint16 { return epm.OpnumEptLookupHandleFree }

// eptLookupHandleFreeResponse carries the [out] (nulled) handle and status.
type eptLookupHandleFreeResponse struct {
	EntryHandle structures.ContextHandle
	Status      ndr.DWORD
}

// EptLookupHandleFree calls ept_lookup_handle_free (opnum 1), releasing a lookup context
// handle obtained from EptLookup. A full enumeration via Lookup nulls the handle on
// completion and needs no explicit free; this is for abandoning a partial walk early.
func EptLookupHandleFree(rpc ndr.Invoker, handle structures.ContextHandle) (structures.ContextHandle, error) {
	req := &eptLookupHandleFreeRequest{EntryHandle: handle}
	var resp eptLookupHandleFreeResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.ContextHandle{}, fmt.Errorf("ept_lookup_handle_free: %w", err)
	}
	if uint32(resp.Status) != epm.EptStatusSuccess {
		return resp.EntryHandle, fmt.Errorf("ept_lookup_handle_free failed: %s", epm.StatusString(uint32(resp.Status)))
	}
	return resp.EntryHandle, nil
}

// Lookup enumerates the entire endpoint map by paging ept_lookup to completion. It issues
// EptInquiryAllElts lookups (no object/interface filter) from a null entry handle,
// accumulating every returned ept_entry_t until the server returns a null entry handle or
// reports ept_s_not_registered (no further matches) — both a normal end of enumeration.
// Each entry's binding is available via EptEntry.DecodeTower / Tower.Binding. For finer
// control (a filter, a custom batch size, or one page at a time) call EptLookup directly.
func Lookup(rpc ndr.Invoker) ([]structures.EptEntry, error) {
	var (
		entries []structures.EptEntry
		handle  structures.ContextHandle
	)
	for {
		batch, next, status, err := EptLookup(rpc, epm.EptInquiryAllElts, nil, nil, epm.EptVersAll, handle, DefaultMaxEnts)
		if err != nil {
			return nil, err
		}
		if status != epm.EptStatusSuccess && status != epm.EptStatusNotRegistered {
			return nil, fmt.Errorf("ept_lookup failed: %s", epm.StatusString(status))
		}
		entries = append(entries, batch...)
		handle = next

		if status == epm.EptStatusNotRegistered || handle.IsNull() {
			return entries, nil
		}
		if len(batch) == 0 {
			// The server advanced no entries but left a live handle; release it rather than
			// loop forever (best-effort).
			_, _ = EptLookupHandleFree(rpc, handle)
			return entries, nil
		}
	}
}
