// Package openlist runs the same-session lifecycle shared by IBM i open-list
// APIs: open, read pages with QGYGTLE, close with QGYCLST.
//
// Read holds one rpc.ExclusiveCaller reservation for the whole sequence, so
// no other call can run on the server job between open and close (the list
// handle belongs to that job, and QGY APIs are not thread-safe). The open
// itself is a program-specific callback; this package never assumes a record
// schema, only a record length.
//
// Cleanup: once the open has produced a handle, QGYCLST runs on every return
// path, on a fresh context bounded by CleanupTimeout, unless the error says
// the session is lost (rpc.ErrSessionLost). Then no further frame is
// written: the server job ends with the connection and IBM i reclaims the
// list. A close failure is joined to the primary error.
package openlist

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgyclst"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgygtle"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/errorcode"
	listinfo "github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/openlist"
)

// Defaults and bounds.
const (
	DefaultPageSize       = 100
	MaxPageSize           = 10000
	DefaultMaxRecords     = 10000
	MaxMaxRecords         = 1 << 20
	DefaultMaxReceiver    = 16 << 20
	DefaultCleanupTimeout = 5 * time.Second
)

// Config bounds one Read. Zero values select the defaults.
type Config struct {
	// Library of QGYGTLE/QGYCLST; zero means QGY.
	Library rpc.ObjectName
	// RecordLength is the exact record length of the opened format.
	RecordLength uint32
	// PageSize is the records requested per QGYGTLE call.
	PageSize int
	// MaxRecords caps the records read; a longer list yields a partial
	// result with ErrLimit.
	MaxRecords int
	// MaxReceiver caps one page receiver (PageSize * RecordLength).
	MaxReceiver uint32
	// CleanupTimeout bounds QGYCLST on its fresh context.
	CleanupTimeout time.Duration
	// JobCodec decodes error-code exception IDs.
	JobCodec ccsid.Codec
	// ErrorCode is used for QGYGTLE and QGYCLST; zero is exception mode.
	ErrorCode errorcode.Parameter
}

func (c Config) resolve() (Config, error) {
	if c.PageSize == 0 {
		c.PageSize = DefaultPageSize
	}
	if c.MaxRecords == 0 {
		c.MaxRecords = DefaultMaxRecords
	}
	if c.MaxReceiver == 0 {
		c.MaxReceiver = DefaultMaxReceiver
	}
	if c.CleanupTimeout == 0 {
		c.CleanupTimeout = DefaultCleanupTimeout
	}
	switch {
	case c.RecordLength == 0:
		return Config{}, errors.New("openlist: zero record length")
	case c.JobCodec == nil:
		return Config{}, errors.New("openlist: nil job codec")
	case c.PageSize < 1 || c.PageSize > MaxPageSize:
		return Config{}, fmt.Errorf("openlist: page size %d outside 1..%d", c.PageSize, MaxPageSize)
	case c.MaxRecords < 1 || c.MaxRecords > MaxMaxRecords:
		return Config{}, fmt.Errorf("openlist: maximum records %d outside 1..%d", c.MaxRecords, MaxMaxRecords)
	case c.CleanupTimeout < 0:
		return Config{}, fmt.Errorf("openlist: negative cleanup timeout %s", c.CleanupTimeout)
	}
	receiver := uint64(c.PageSize) * uint64(c.RecordLength)
	if receiver > uint64(c.MaxReceiver) || receiver > math.MaxInt32 {
		return Config{}, fmt.Errorf("%w: a %d-record page needs %d bytes, limit %d", ErrLimit, c.PageSize, receiver, c.MaxReceiver)
	}
	return c, nil
}

// Opened is what the program-specific open produced.
type Opened struct {
	Info   listinfo.Info
	Handle listinfo.Handle
	// HandleRecovered is true when the list exists on the server and must
	// be closed, even if the open also returned an error.
	HandleRecovered bool
}

// OpenFunc opens a list with the reservation's caller.
type OpenFunc func(ctx context.Context, caller rpc.Caller) (Opened, error)

// Result is what Read collected.
type Result struct {
	// Records holds Count records of RecordLength bytes, in list order.
	Records []byte
	Count   int
	// Total is the list size the server reported at open.
	Total int32
	// Status is the list status at open.
	Status listinfo.Status
	// Partial means Records stops before the end of the list; an error is
	// always returned with a partial result.
	Partial  bool
	Warnings []string
}

var (
	// ErrLimit is matched when a configured limit stopped reading.
	ErrLimit = errors.New("openlist: limit reached")
	// ErrNoProgress is matched when a page returned no record before the
	// end of the list.
	ErrNoProgress = errors.New("openlist: page made no progress")
	// ErrUnstable is matched when the list changed between pages.
	ErrUnstable = errors.New("openlist: list changed while paging")
)

// Read opens a list with open, reads up to cfg.MaxRecords records, and
// closes it, all inside one reservation of ex.
//
// A non-nil Result together with an error is returned only when reading
// stopped at a limit or a cancellation; the records read so far are then
// marked Partial. Any malformed or inconsistent page returns no records.
func Read(ctx context.Context, ex rpc.ExclusiveCaller, cfg Config, open OpenFunc) (*Result, error) {
	cfg, err := cfg.resolve()
	if err != nil {
		return nil, err
	}
	var result *Result
	err = ex.ExecuteExclusive(ctx, func(caller rpc.Caller) (err error) {
		opened, openErr := open(ctx, caller)
		if opened.HandleRecovered {
			defer func() { err = closeList(ctx, caller, cfg, opened.Handle, err) }()
		}
		if openErr != nil {
			return openErr
		}
		result, err = readPages(ctx, caller, cfg, opened)
		return err
	})
	return result, err
}

func readPages(ctx context.Context, caller rpc.Caller, cfg Config, opened Opened) (*Result, error) {
	total := opened.Info.TotalRecords
	res := &Result{Total: total, Status: opened.Info.Status}
	limit := int(total)
	var limitErr error
	if limit > cfg.MaxRecords {
		limit = cfg.MaxRecords
		limitErr = fmt.Errorf("%w: the list has %d records, maximum %d", ErrLimit, total, cfg.MaxRecords)
	}
	for start := 1; start <= limit; {
		if err := ctx.Err(); err != nil {
			res.Partial = true
			return res, err
		}
		count := min(cfg.PageSize, limit-start+1)
		req := qgygtle.Request{
			Library: cfg.Library, Handle: opened.Handle, RecordLength: cfg.RecordLength,
			Count: int32(count), Start: int32(start), ErrorCode: cfg.ErrorCode,
		}
		page, err := qgygtle.Get(ctx, caller, req, cfg.JobCodec)
		if err != nil {
			return nil, err
		}
		switch {
		case page.Info.TotalRecords != total:
			return nil, fmt.Errorf("%w: total changed from %d to %d", ErrUnstable, total, page.Info.TotalRecords)
		case page.Info.Status == listinfo.StatusTooMuchData:
			res.Warnings = appendOnce(res.Warnings, "the server stopped building the list (too much data); it is incomplete")
		case page.Info.Status != listinfo.StatusComplete:
			return nil, fmt.Errorf("%w: list status %s while paging a synchronous list", ErrUnstable, page.Info.Status)
		}
		if page.Info.RecordsReturned == 0 {
			return nil, fmt.Errorf("%w at record %d of %d", ErrNoProgress, start, total)
		}
		res.Records = append(res.Records, page.Records...)
		res.Count += int(page.Info.RecordsReturned)
		start += int(page.Info.RecordsReturned)
	}
	if limitErr != nil {
		res.Partial = true
		return res, limitErr
	}
	return res, nil
}

func appendOnce(list []string, s string) []string {
	for _, v := range list {
		if v == s {
			return list
		}
	}
	return append(list, s)
}

// closeList releases the handle unless the session is lost, joining any
// close failure to the primary error.
func closeList(ctx context.Context, caller rpc.Caller, cfg Config, handle listinfo.Handle, primary error) error {
	if errors.Is(primary, rpc.ErrSessionLost) {
		return primary
	}
	cctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.CleanupTimeout)
	defer cancel()
	err := qgyclst.Close(cctx, caller, qgyclst.Request{Library: cfg.Library, Handle: handle, ErrorCode: cfg.ErrorCode}, cfg.JobCodec)
	if err != nil {
		return errors.Join(primary, fmt.Errorf("openlist: close list: %w", err))
	}
	return primary
}
