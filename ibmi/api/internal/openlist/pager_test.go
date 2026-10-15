package openlist_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/api/internal/openlist"
	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgygtle"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgyolaus"
	listinfo "github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/openlist"
)

var codec, _ = ccsid.Lookup(ccsid.CCSID37)

const recLen = 12 // AUTU0100

// listServer is an in-memory QGY list server behind rpc.ExclusiveCaller.
type listServer struct {
	t       *testing.T
	records int
	handle  listinfo.Handle

	openStatus   byte  // default 0xF2
	openBadInfo  bool  // malformed open metadata, handle recoverable
	openErr      error // returned instead of the open result
	perPage      int   // server-side cap on records per page (0 = none)
	pageHook     func(start int) (listinfo.Info, error, bool)
	closeErr     error
	pageErrStart int
	pageErr      error

	// Asynchronous build simulation: the open reports pending, and
	// list-information polls report pending buildPolls times, then
	// buildFinal (default complete).
	async      bool
	buildPolls int
	buildFinal byte
	infoCalls  int

	mu        sync.Mutex
	held      bool
	log       []string
	reserveID int
}

func (s *listServer) ExecuteExclusive(ctx context.Context, fn func(rpc.Caller) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.held = true
	s.reserveID++
	defer func() { s.held = false }()
	return fn(listCaller{s, s.reserveID})
}

type listCaller struct {
	s  *listServer
	id int
}

func info(total, returned int, h listinfo.Handle, length, first int, status byte) []byte {
	b := make([]byte, listinfo.Length)
	binary.BigEndian.PutUint32(b[0:], uint32(total))
	binary.BigEndian.PutUint32(b[4:], uint32(returned))
	copy(b[8:], h[:])
	binary.BigEndian.PutUint32(b[12:], uint32(length))
	b[16] = 0xC3
	b[30] = status
	binary.BigEndian.PutUint32(b[32:], uint32(returned*length))
	binary.BigEndian.PutUint32(b[36:], uint32(first))
	return b
}

func record(i int) []byte {
	name, _ := ccsid.EncodeFixed(codec, fmt.Sprintf("U%04d", i), 10)
	return append(name, 0xF0, 0xF0)
}

func (c listCaller) Call(ctx context.Context, call rpc.Call) (*rpc.Result, error) {
	s := c.s
	if !s.held {
		s.t.Errorf("%s called outside a reservation", call.Program())
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.log = append(s.log, fmt.Sprintf("%d:%s", c.id, call.Program()))
	res := &rpc.Result{Parameters: make([]rpc.ParameterResult, call.ParameterCount()), RequestedMessageOption: call.MessageOption(), EffectiveMessageOption: call.MessageOption()}
	params := call.Parameters()
	switch call.Program().String() {
	case "QGYOLAUS":
		if s.openErr != nil {
			return nil, s.openErr
		}
		status := s.openStatus
		if status == 0 {
			status = 0xF2
		}
		if s.async {
			status = 0xF0 // pending
		}
		b := info(s.records, 0, s.handle, recLen, 0, status)
		if s.openBadInfo {
			b[16] = 0x00
		}
		res.Parameters[2] = rpc.ParameterResult{Present: true, Data: b}
	case "QGYGTLE":
		count := int(int32(binary.BigEndian.Uint32(params[4].InputBytes())))
		start := int(int32(binary.BigEndian.Uint32(params[5].InputBytes())))
		if count == 0 && start == 0 { // list-information request
			s.infoCalls++
			status := byte(0xF0)
			if s.infoCalls > s.buildPolls {
				status = s.buildFinal
				if status == 0 {
					status = 0xF2
				}
			}
			res.Parameters[3] = rpc.ParameterResult{Present: true, Data: info(s.records, 0, s.handle, recLen, 0, status)}
			return res, nil
		}
		if s.pageErr != nil && start >= s.pageErrStart {
			return nil, s.pageErr
		}
		n := min(count, s.records-start+1)
		if s.perPage > 0 {
			n = min(n, s.perPage)
		}
		if n < 0 {
			n = 0
		}
		receiver := make([]byte, params[0].Maximum())
		for i := 0; i < n; i++ {
			copy(receiver[i*recLen:], record(start+i))
		}
		li := info(s.records, n, s.handle, recLen, start, 0xF2)
		if s.pageHook != nil {
			if custom, err, ok := s.pageHook(start); ok {
				if err != nil {
					return nil, err
				}
				li = marshalInfo(custom)
			}
		}
		res.Parameters[0] = rpc.ParameterResult{Present: true, Data: receiver}
		res.Parameters[3] = rpc.ParameterResult{Present: true, Data: li}
	case "QGYCLST":
		if !bytes.Equal(params[0].InputBytes(), s.handle[:]) {
			s.t.Errorf("QGYCLST for handle % x, want % x", params[0].InputBytes(), s.handle)
		}
		if s.closeErr != nil {
			return nil, s.closeErr
		}
	}
	return res, nil
}

func newServer(t *testing.T, records int) *listServer {
	return &listServer{t: t, records: records, handle: listinfo.Handle{1, 2, 3, 4}}
}

func open(ctx context.Context, caller rpc.Caller) (openlist.Opened, error) {
	return openWith(qgyolaus.Request{})(ctx, caller)
}

func openAsync(ctx context.Context, caller rpc.Caller) (openlist.Opened, error) {
	return openWith(qgyolaus.Request{Build: qgyolaus.BuildAsynchronous})(ctx, caller)
}

func openWith(req qgyolaus.Request) openlist.OpenFunc {
	return func(ctx context.Context, caller rpc.Caller) (openlist.Opened, error) {
		return openReq(ctx, caller, req)
	}
}

func openReq(ctx context.Context, caller rpc.Caller, req qgyolaus.Request) (openlist.Opened, error) {
	res, recovered, err := qgyolaus.Open(ctx, caller, qgyolaus.AUTU0100(), req, codec, codec)
	o := openlist.Opened{HandleRecovered: recovered}
	if res != nil {
		o.Info, o.Handle = res.Info, res.Handle
	}
	return o, err
}

func cfg(pageSize int) openlist.Config {
	return openlist.Config{RecordLength: recLen, PageSize: pageSize, JobCodec: codec}
}

// programs strips the "reservation:" prefix from each log entry.
func programs(log []string) []string {
	out := make([]string, len(log))
	for i, l := range log {
		_, out[i], _ = strings.Cut(l, ":")
	}
	return out
}

func reservation(entry string) string {
	id, _, _ := strings.Cut(entry, ":")
	return id
}

func wantLog(t *testing.T, s *listServer, want ...string) {
	t.Helper()
	got := programs(s.log)
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("calls = %v, want %v", got, want)
	}
}

func checkRecords(t *testing.T, res *openlist.Result, first, n int) {
	t.Helper()
	if res.Count != n || len(res.Records) != n*recLen {
		t.Fatalf("count %d, %d bytes; want %d records", res.Count, len(res.Records), n)
	}
	for i := 0; i < n; i++ {
		if !bytes.Equal(res.Records[i*recLen:(i+1)*recLen], record(first+i)) {
			t.Fatalf("record %d out of order", i)
		}
	}
}

func TestReadPageShapes(t *testing.T) {
	for _, tc := range []struct {
		name          string
		records, size int
		perPage       int
		wantGets      int
	}{
		{"empty", 0, 10, 0, 0},
		{"one page", 5, 10, 0, 1},
		{"exact page", 10, 10, 0, 1},
		{"multi page", 25, 10, 0, 3},
		{"short pages from server", 10, 10, 3, 4},
	} {
		s := newServer(t, tc.records)
		s.perPage = tc.perPage
		res, err := openlist.Read(context.Background(), s, cfg(tc.size), open)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		checkRecords(t, res, 1, tc.records)
		if res.Partial || res.Total != int32(tc.records) || res.Status != listinfo.StatusComplete {
			t.Errorf("%s: result %+v", tc.name, res)
		}
		want := []string{"QGYOLAUS"}
		for i := 0; i < tc.wantGets; i++ {
			want = append(want, "QGYGTLE")
		}
		wantLog(t, s, append(want, "QGYCLST")...)
	}
}

func TestReadRejectsInconsistentPagesAndCloses(t *testing.T) {
	h := listinfo.Handle{1, 2, 3, 4}
	for _, tc := range []struct {
		name string
		hook func(start int) (listinfo.Info, error, bool)
		want error
	}{
		{"zero progress", func(start int) (listinfo.Info, error, bool) {
			return listinfo.Info{TotalRecords: 25, Handle: h, RecordLength: recLen, Complete: listinfo.CompleteAccurate, Status: listinfo.StatusComplete}, nil, true
		}, openlist.ErrNoProgress},
		{"total changed", func(start int) (listinfo.Info, error, bool) {
			return listinfo.Info{TotalRecords: 26, RecordsReturned: 10, Handle: h, RecordLength: recLen, Complete: listinfo.CompleteAccurate, Status: listinfo.StatusComplete, InfoLength: 120, FirstRecord: int32(start)}, nil, true
		}, openlist.ErrUnstable},
		{"status building", func(start int) (listinfo.Info, error, bool) {
			return listinfo.Info{TotalRecords: 25, RecordsReturned: 10, Handle: h, RecordLength: recLen, Complete: listinfo.CompleteAccurate, Status: listinfo.StatusBuilding, InfoLength: 120, FirstRecord: int32(start)}, nil, true
		}, openlist.ErrUnstable},
		{"handle changed", func(start int) (listinfo.Info, error, bool) {
			return listinfo.Info{TotalRecords: 25, RecordsReturned: 10, Handle: listinfo.Handle{9}, RecordLength: recLen, Complete: listinfo.CompleteAccurate, Status: listinfo.StatusComplete, InfoLength: 120, FirstRecord: int32(start)}, nil, true
		}, qgygtle.ErrInvalidPage},
		{"wrong first record", func(start int) (listinfo.Info, error, bool) {
			return listinfo.Info{TotalRecords: 25, RecordsReturned: 10, Handle: h, RecordLength: recLen, Complete: listinfo.CompleteAccurate, Status: listinfo.StatusComplete, InfoLength: 120, FirstRecord: int32(start + 1)}, nil, true
		}, qgygtle.ErrInvalidPage},
	} {
		s := newServer(t, 25)
		s.pageHook = tc.hook
		res, err := openlist.Read(context.Background(), s, cfg(10), open)
		if res != nil || !errors.Is(err, tc.want) {
			t.Errorf("%s: result %v, error %v", tc.name, res, err)
		}
		wantLog(t, s, "QGYOLAUS", "QGYGTLE", "QGYCLST")
	}
}

func TestReadTooMuchDataWarns(t *testing.T) {
	s := newServer(t, 5)
	s.pageHook = func(start int) (listinfo.Info, error, bool) {
		return listinfo.Info{TotalRecords: 5, RecordsReturned: 5, Handle: s.handle, RecordLength: recLen, Complete: listinfo.PartialAccurate, Status: listinfo.StatusTooMuchData, InfoLength: 60, FirstRecord: 1}, nil, true
	}
	res, err := openlist.Read(context.Background(), s, cfg(10), open)
	if err != nil || len(res.Warnings) != 1 {
		t.Fatalf("result %+v, %v", res, err)
	}
}

func TestReadLimits(t *testing.T) {
	s := newServer(t, 25)
	c := cfg(10)
	c.MaxRecords = 15
	res, err := openlist.Read(context.Background(), s, c, open)
	if !errors.Is(err, openlist.ErrLimit) || res == nil || !res.Partial || res.Total != 25 {
		t.Fatalf("result %+v, %v", res, err)
	}
	checkRecords(t, res, 1, 15)
	wantLog(t, s, "QGYOLAUS", "QGYGTLE", "QGYGTLE", "QGYCLST")

	for name, bad := range map[string]openlist.Config{
		"receiver over limit": {RecordLength: recLen, PageSize: 100, MaxReceiver: 1000, JobCodec: codec},
		"page size":           {RecordLength: recLen, PageSize: openlist.MaxPageSize + 1, JobCodec: codec},
		"max records":         {RecordLength: recLen, MaxRecords: openlist.MaxMaxRecords + 1, JobCodec: codec},
		"receiver past int32": {RecordLength: 1 << 20, PageSize: 4096, MaxReceiver: 1 << 31, JobCodec: codec},
		"zero record length":  {JobCodec: codec},
		"no codec":            {RecordLength: recLen},
	} {
		s := newServer(t, 5)
		if _, err := openlist.Read(context.Background(), s, bad, open); err == nil || len(s.log) != 0 {
			t.Errorf("%s: error %v after %d calls", name, err, len(s.log))
		}
	}
}

func TestReadCancellationReturnsPartialAndCloses(t *testing.T) {
	s := newServer(t, 25)
	ctx, cancel := context.WithCancel(context.Background())
	s.pageHook = func(start int) (listinfo.Info, error, bool) {
		if start == 11 {
			cancel() // cancelled after the second page is served
		}
		return listinfo.Info{}, nil, false
	}
	res, err := openlist.Read(ctx, s, cfg(10), open)
	if !errors.Is(err, context.Canceled) || res == nil || !res.Partial {
		t.Fatalf("result %v, %v", res, err)
	}
	checkRecords(t, res, 1, 20)
	// The close runs on a fresh context despite the cancellation.
	wantLog(t, s, "QGYOLAUS", "QGYGTLE", "QGYGTLE", "QGYCLST")
}

func TestReadJoinsCloseFailure(t *testing.T) {
	s := newServer(t, 25)
	s.pageErr, s.pageErrStart = errors.New("primary page failure"), 11
	s.closeErr = errors.New("close failure")
	res, err := openlist.Read(context.Background(), s, cfg(10), open)
	if res != nil || !errors.Is(err, s.pageErr) || !errors.Is(err, s.closeErr) {
		t.Fatalf("result %v, error %v", res, err)
	}
	// A successful read with a failing close still reports the close.
	s2 := newServer(t, 3)
	s2.closeErr = errors.New("close failure")
	if _, err := openlist.Read(context.Background(), s2, cfg(10), open); !errors.Is(err, s2.closeErr) {
		t.Errorf("close-only failure = %v", err)
	}
}

func TestReadClosesAfterMalformedOpenMetadata(t *testing.T) {
	s := newServer(t, 5)
	s.openBadInfo = true
	if _, err := openlist.Read(context.Background(), s, cfg(10), open); !errors.Is(err, qgyolaus.ErrMetadata) {
		t.Fatalf("error = %v", err)
	}
	wantLog(t, s, "QGYOLAUS", "QGYCLST")

	s = newServer(t, 5)
	s.openStatus = 0xF1 // building: not acceptable for a synchronous open
	if _, err := openlist.Read(context.Background(), s, cfg(10), open); !errors.Is(err, qgyolaus.ErrMetadata) {
		t.Fatalf("error = %v", err)
	}
	wantLog(t, s, "QGYOLAUS", "QGYCLST")
}

func TestReadDoesNotCloseWithoutHandle(t *testing.T) {
	s := newServer(t, 5)
	s.openErr = errors.New("CPF2204 from the server")
	if _, err := openlist.Read(context.Background(), s, cfg(10), open); !errors.Is(err, s.openErr) {
		t.Fatalf("error = %v", err)
	}
	wantLog(t, s, "QGYOLAUS")
}

func TestReadSkipsCloseOnLostSession(t *testing.T) {
	s := newServer(t, 25)
	s.pageErr, s.pageErrStart = fmt.Errorf("malformed reply: %w", rpc.ErrSessionLost), 11
	res, err := openlist.Read(context.Background(), s, cfg(10), open)
	if res != nil || !errors.Is(err, rpc.ErrSessionLost) {
		t.Fatalf("result %v, error %v", res, err)
	}
	wantLog(t, s, "QGYOLAUS", "QGYGTLE", "QGYGTLE") // no QGYCLST on a lost session
}

func TestReadHoldsOneReservation(t *testing.T) {
	s := newServer(t, 25)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := openlist.Read(context.Background(), s, cfg(10), open); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	// Every read's five calls are contiguous and share one reservation.
	if len(s.log) != 20 {
		t.Fatalf("%d calls", len(s.log))
	}
	for i := 0; i < len(s.log); i += 5 {
		id := reservation(s.log[i])
		for _, l := range s.log[i : i+5] {
			if reservation(l) != id {
				t.Fatalf("calls interleaved: %v", s.log)
			}
		}
	}
}

func marshalInfo(i listinfo.Info) []byte {
	b := info(int(i.TotalRecords), int(i.RecordsReturned), i.Handle, int(i.RecordLength), int(i.FirstRecord), 0xF0+byte(i.Status))
	b[16] = map[listinfo.Complete]byte{listinfo.CompleteAccurate: 0xC3, listinfo.Incomplete: 0xC9, listinfo.PartialAccurate: 0xD7}[i.Complete]
	binary.BigEndian.PutUint32(b[32:], uint32(i.InfoLength))
	return b
}

func asyncCfg(pageSize int) openlist.Config {
	c := cfg(pageSize)
	c.PollInterval = openlist.MinPollInterval
	return c
}

func TestReadWaitsForAsynchronousBuild(t *testing.T) {
	s := newServer(t, 25)
	s.async, s.buildPolls = true, 3
	res, err := openlist.Read(context.Background(), s, asyncCfg(10), openAsync)
	if err != nil {
		t.Fatal(err)
	}
	checkRecords(t, res, 1, 25)
	if res.Status != listinfo.StatusComplete || s.infoCalls != 4 {
		t.Errorf("status %s after %d polls", res.Status, s.infoCalls)
	}
	want := []string{"QGYOLAUS", "QGYGTLE", "QGYGTLE", "QGYGTLE", "QGYGTLE", "QGYGTLE", "QGYGTLE", "QGYGTLE", "QGYCLST"}
	wantLog(t, s, want...)
}

func TestReadAsynchronousBuildFailures(t *testing.T) {
	s := newServer(t, 25)
	s.async, s.buildPolls, s.buildFinal = true, 1, 0xF3 // error
	if res, err := openlist.Read(context.Background(), s, asyncCfg(10), openAsync); res != nil || !errors.Is(err, openlist.ErrBuildFailed) {
		t.Errorf("build error: %v, %v", res, err)
	}
	wantLog(t, s, "QGYOLAUS", "QGYGTLE", "QGYGTLE", "QGYCLST")

	s = newServer(t, 25)
	s.async, s.buildPolls = true, 1<<30 // never completes
	c := asyncCfg(10)
	c.MaxBuildWait = 50 * time.Millisecond
	if res, err := openlist.Read(context.Background(), s, c, openAsync); res != nil || !errors.Is(err, openlist.ErrBuildTimeout) {
		t.Errorf("timeout: %v, %v", res, err)
	}
	if got := programs(s.log); got[len(got)-1] != "QGYCLST" {
		t.Errorf("list not closed after the timeout: %v", got)
	}

	s = newServer(t, 25)
	s.async, s.buildPolls = true, 1<<30
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, err := openlist.Read(ctx, s, asyncCfg(10), openAsync); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("cancellation: %v", err)
	}
	if got := programs(s.log); got[len(got)-1] != "QGYCLST" {
		t.Errorf("list not closed after cancellation: %v", got)
	}
}

func TestReadAsynchronousTooMuchData(t *testing.T) {
	s := newServer(t, 5)
	s.async, s.buildPolls, s.buildFinal = true, 0, 0xF5
	s.pageHook = func(start int) (listinfo.Info, error, bool) {
		return listinfo.Info{TotalRecords: 5, RecordsReturned: 5, Handle: s.handle, RecordLength: recLen, Complete: listinfo.PartialAccurate, Status: listinfo.StatusTooMuchData, InfoLength: 60, FirstRecord: 1}, nil, true
	}
	res, err := openlist.Read(context.Background(), s, asyncCfg(10), openAsync)
	if err != nil || len(res.Warnings) != 1 || res.Count != 5 {
		t.Fatalf("result %+v, %v", res, err)
	}
}

func TestSynchronousOpenRejectsPending(t *testing.T) {
	s := newServer(t, 5)
	s.async = true // the server answers pending even though -1 was sent
	if _, err := openlist.Read(context.Background(), s, cfg(10), open); !errors.Is(err, qgyolaus.ErrMetadata) {
		t.Errorf("pending synchronous open = %v", err)
	}
	wantLog(t, s, "QGYOLAUS", "QGYCLST")
}

func TestPollConfig(t *testing.T) {
	for name, c := range map[string]openlist.Config{
		"poll too fast": {RecordLength: recLen, JobCodec: codec, PollInterval: time.Millisecond},
		"build wait":    {RecordLength: recLen, JobCodec: codec, MaxBuildWait: 2 * time.Hour},
		"negative wait": {RecordLength: recLen, JobCodec: codec, MaxBuildWait: -time.Second},
	} {
		s := newServer(t, 5)
		if _, err := openlist.Read(context.Background(), s, c, open); err == nil || len(s.log) != 0 {
			t.Errorf("%s: %v after %d calls", name, err, len(s.log))
		}
	}
}
