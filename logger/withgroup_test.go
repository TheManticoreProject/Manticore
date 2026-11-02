package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

type tokenValuer struct{}

func (tokenValuer) LogValue() slog.Value { return slog.StringValue("resolved") }

// TestSlogGroupsQualifyKeys checks WithGroup qualifies the keys of attributes
// added afterwards (on the record and through WithAttrs), that nested groups and
// slog.Group attributes are flattened into dotted keys, and that empty groups,
// empty attributes and LogValuers follow the slog.Handler rules. WithGroup used
// to return the handler unchanged, so grouped keys came out bare.
func TestSlogGroupsQualifyKeys(t *testing.T) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf), WithTimestamp(false), WithNoColors(true)).Slog()

	for _, tc := range []struct {
		log  func()
		want string
	}{
		{func() { l.WithGroup("req").Info("hit", "id", 1) }, "hit req.id=1"},
		{func() { l.With("app", "x").WithGroup("req").With("ip", "1.2.3.4").Info("hit", "id", 2) }, "hit app=x req.ip=1.2.3.4 req.id=2"},
		{func() { l.WithGroup("a").WithGroup("b").Info("m", "k", "v") }, "m a.b.k=v"},
		{func() { l.Info("m", slog.Group("g", "k", 1, slog.Group("h", "z", 2))) }, "m g.k=1 g.h.z=2"},
		{func() { l.WithGroup("").Info("m", "k", 1) }, "m k=1"},
		{func() { l.Info("m", slog.Group("empty"), slog.Attr{}, "k", 1) }, "m k=1"},
		{func() { l.Info("m", slog.Group("", "inline", 1)) }, "m inline=1"},
		{func() { l.Info("m", "tok", tokenValuer{}) }, "m tok=resolved"},
	} {
		buf.Reset()
		tc.log()
		got := strings.TrimSpace(buf.String())
		if !strings.HasSuffix(got, tc.want) {
			t.Errorf("got %q, want it to end with %q", got, tc.want)
		}
	}
}
