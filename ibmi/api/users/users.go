// Package users lists IBM i user and group profiles.
//
// Client depends only on rpc.ExclusiveCaller, so it works with
// network/ibmi/client.Client or any other implementation. List opens a
// QGY/QGYOLAUS list in format AUTU0150, reads it page by page with
// QGYGTLE and closes it with QGYCLST, all inside one exclusive reservation
// of the server job; the list handle never leaves this package.
//
// IBM i returns only the profiles the signed-on user has *READ authority
// to, so ListResult.TotalReported is not proof of a complete inventory;
// every result carries AuthorityCaveat as a warning.
package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/api/internal/openlist"
	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgyolaus"
	listinfo "github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/openlist"
)

// AuthorityCaveat is attached to every ListResult.
const AuthorityCaveat = "IBM i lists only the profiles the signed-on user has *READ authority to; the result is not proof of a complete inventory"

// Options configures a Client.
type Options struct {
	// SystemCCSID (from Sign-on) encodes the selection parameters and
	// decodes profile names and descriptions. Required.
	SystemCCSID uint32
	// JobCCSID (from Remote Command) decodes API error IDs. Required.
	JobCCSID uint32
	// Library of the QGY programs; zero means QGY (V5R4).
	Library rpc.ObjectName
	// CleanupTimeout bounds closing the list; zero means 5 seconds.
	CleanupTimeout time.Duration
	// MaxReceiver caps one page receiver; zero means 16 MiB.
	MaxReceiver uint32
}

// Client lists profiles over an exclusive caller.
type Client struct {
	caller      rpc.ExclusiveCaller
	opts        Options
	system, job ccsid.Codec
}

// New returns a Client. It performs no I/O.
func New(caller rpc.ExclusiveCaller, opts Options) (*Client, error) {
	if caller == nil {
		return nil, errors.New("users: nil caller")
	}
	system, err := ccsid.Lookup(ccsid.CCSID(opts.SystemCCSID))
	if err != nil {
		return nil, fmt.Errorf("users: system CCSID: %w", err)
	}
	job, err := ccsid.Lookup(ccsid.CCSID(opts.JobCCSID))
	if err != nil {
		return nil, fmt.Errorf("users: job CCSID: %w", err)
	}
	return &Client{caller: caller, opts: opts, system: system, job: job}, nil
}

// Selection chooses which profiles to list.
type Selection uint8

const (
	// AllProfiles lists users and groups.
	AllProfiles Selection = iota
	// UserProfiles lists profiles that are not groups.
	UserProfiles
	// GroupProfiles lists group profiles.
	GroupProfiles
	// GroupMembers lists the members of ListOptions.Group.
	GroupMembers
)

func (s Selection) String() string {
	switch s {
	case AllProfiles:
		return "all"
	case UserProfiles:
		return "users"
	case GroupProfiles:
		return "groups"
	case GroupMembers:
		return "group members"
	}
	return fmt.Sprintf("Selection(%d)", uint8(s))
}

func (s Selection) qgy() (qgyolaus.Selection, error) {
	switch s {
	case AllProfiles:
		return qgyolaus.SelectAll, nil
	case UserProfiles:
		return qgyolaus.SelectUsers, nil
	case GroupProfiles:
		return qgyolaus.SelectGroups, nil
	case GroupMembers:
		return qgyolaus.SelectMembers, nil
	}
	return 0, fmt.Errorf("users: undefined selection %d", s)
}

// ListOptions bounds and filters one List. Zero values select defaults.
type ListOptions struct {
	// PageSize is the profiles fetched per page (default 100).
	PageSize int
	// MaxUsers caps the profiles returned (default 10000); a longer list
	// returns a partial result.
	MaxUsers  int
	Selection Selection
	// Group is required for GroupMembers.
	Group string
	// ProfilePattern is a profile name or IBM generic name ("AB*"); empty
	// means all.
	ProfilePattern string
	// Asynchronous asks IBM i to build the list in a server job and polls
	// until it is complete instead of building it during the open call.
	Asynchronous bool
	// MaxBuildWait bounds an asynchronous build (default 1 minute).
	MaxBuildWait time.Duration
}

// UserSummary is one AUTU0150 entry.
type UserSummary struct {
	Name            string
	GroupProfile    bool
	GroupHasMembers bool
	Description     string
}

// ListResult is the outcome of List.
type ListResult struct {
	Users []UserSummary
	// TotalReported is the list size IBM i reported, limited to what the
	// caller may see.
	TotalReported uint32
	Selection     Selection
	// SnapshotStatus is the list status once built: complete, or too much
	// data when IBM i stopped building it early.
	SnapshotStatus listinfo.Status
	// Partial means Users stops before the end of the list. It is only set
	// together with a non-nil error from List.
	Partial  bool
	Warnings []string
}

// List enumerates profiles. On success it returns every matching profile.
// If a limit or cancellation stops it after profiles were read, it returns
// those profiles with Partial set and a non-nil error; callers must opt in
// to using partial data. Malformed server data returns no result.
func (c *Client) List(ctx context.Context, opts ListOptions) (*ListResult, error) {
	sel, err := opts.Selection.qgy()
	if err != nil {
		return nil, err
	}
	format := qgyolaus.AUTU0150()
	req := qgyolaus.Request{Library: c.opts.Library, Selection: sel, Group: opts.Group, Profile: opts.ProfilePattern}
	if opts.Asynchronous {
		req.Build = qgyolaus.BuildAsynchronous
	}
	if _, err := qgyolaus.NewCall(format, req, c.system); err != nil {
		return nil, err // reject bad options before taking the reservation
	}
	cfg := openlist.Config{
		Library:        c.opts.Library,
		RecordLength:   format.RecordLength(),
		PageSize:       opts.PageSize,
		MaxRecords:     opts.MaxUsers,
		MaxReceiver:    c.opts.MaxReceiver,
		CleanupTimeout: c.opts.CleanupTimeout,
		JobCodec:       c.job,
		MaxBuildWait:   opts.MaxBuildWait,
	}
	read, readErr := openlist.Read(ctx, c.caller, cfg, func(ctx context.Context, caller rpc.Caller) (openlist.Opened, error) {
		res, recovered, err := qgyolaus.Open(ctx, caller, format, req, c.system, c.job)
		var opened openlist.Opened
		if res != nil {
			opened = openlist.Opened{Info: res.Info, Handle: res.Handle}
		}
		opened.HandleRecovered = recovered
		return opened, err
	})
	if read == nil {
		return nil, readErr
	}
	records, err := format.DecodeRecords(read.Records, read.Count, c.system)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("users: %w", err), readErr)
	}
	result := &ListResult{
		Users:          make([]UserSummary, len(records)),
		TotalReported:  uint32(read.Total),
		Selection:      opts.Selection,
		SnapshotStatus: read.Status,
		Partial:        read.Partial,
		Warnings:       append([]string{AuthorityCaveat}, read.Warnings...),
	}
	for i, r := range records {
		result.Users[i] = UserSummary{Name: r.Name, GroupProfile: r.GroupProfile, GroupHasMembers: r.GroupHasMembers, Description: r.Description}
	}
	return result, readErr
}
