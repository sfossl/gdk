package util

import (
	"errors"
	"testing"
	"time"

	"github.com/sfossl/gda/types"
)

func TestParseDuration_Units(t *testing.T) {
	cases := []struct {
		input    string
		expected time.Duration
	}{
		{"10ns", 10 * time.Nanosecond},
		{"5us", 5 * time.Microsecond},
		{"500ms", 500 * time.Millisecond},
		{"30s", 30 * time.Second},
		{"15m", 15 * time.Minute},
		{"2h", 2 * time.Hour},
		{"3d", 3 * 24 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"2mo", 2 * 30 * 24 * time.Hour},
		{"1y", 365 * 24 * time.Hour},
	}

	for _, c := range cases {
		d, err := ParseDuration(c.input)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", c.input, err)
		}
		if d != c.expected {
			t.Fatalf("for %s expected %v, got %v", c.input, c.expected, d)
		}
	}
}

func TestParseDuration_CompoundAndDecimals(t *testing.T) {
	d, err := ParseDuration("1d 2h 30m 15s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := 24*time.Hour + 2*time.Hour + 30*time.Minute + 15*time.Second
	if d != expected {
		t.Fatalf("expected %v, got %v", expected, d)
	}

	d2, err := ParseDuration("1.5d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected2 := 36 * time.Hour
	if d2 != expected2 {
		t.Fatalf("expected %v, got %v", expected2, d2)
	}
}

func TestParseDuration_Errors(t *testing.T) {
	cases := []struct {
		input       string
		expectedErr error
	}{
		{"", ErrEmptyDuration},
		{"   ", ErrEmptyDuration},
		{"10", ErrInvalidDuration},
		{"abc", ErrInvalidDuration},
		{"10unknown", ErrUnknownUnit},
		{"1.2.3s", ErrInvalidDuration},
	}

	for _, c := range cases {
		_, err := ParseDuration(c.input)
		if !errors.Is(err, c.expectedErr) {
			t.Fatalf("for %q expected error %v, got %v", c.input, c.expectedErr, err)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		input    time.Duration
		expected string
	}{
		{0, "0s"},
		{-10 * time.Second, "0s"},
		{500 * time.Nanosecond, "500ns"},
		{500 * time.Microsecond, "500µs"},
		{50 * time.Millisecond, "50ms"},
		{10 * time.Second, "10s"},
		{2*time.Hour + 30*time.Minute, "2h 30m"},
		{365*24*time.Hour + 24*time.Hour, "1y 1d"},
	}

	for _, c := range cases {
		res := FormatDuration(c.input)
		if res != c.expected {
			t.Fatalf("for %v expected %q, got %q", c.input, c.expected, res)
		}
	}
}

func TestFormatDurationHuman(t *testing.T) {
	cases := []struct {
		input    time.Duration
		expected string
	}{
		{0, "0 seconds"},
		{-5 * time.Second, "0 seconds"},
		{500 * time.Millisecond, "less than a second"},
		{1 * time.Second, "1 second"},
		{10 * time.Second, "10 seconds"},
		{1 * time.Hour, "1 hour"},
		{2*time.Hour + 1*time.Minute, "2 hours, 1 minute"},
	}

	for _, c := range cases {
		res := FormatDurationHuman(c.input)
		if res != c.expected {
			t.Fatalf("for %v expected %q, got %q", c.input, c.expected, res)
		}
	}
}

func TestNew_Inputs(t *testing.T) {
	if item := New(nil); item.Error() != ErrInvalidTime {
		t.Fatalf("expected ErrInvalidTime for nil, got %v", item.Error())
	}

	now := time.Now().UTC()
	itemTime := New(now)
	if itemTime.Error() != nil || itemTime.Unix() != now.Unix() {
		t.Fatalf("unexpected error or unix mismatch for time.Time: %v", itemTime.Error())
	}

	itemDur := New(10 * time.Minute)
	if itemDur.Error() != nil || itemDur.Duration() != 10*time.Minute {
		t.Fatalf("unexpected error or duration mismatch for time.Duration: %v", itemDur.Error())
	}

	itemStr := New("15m")
	if itemStr.Error() != nil || itemStr.Duration() != 15*time.Minute {
		t.Fatalf("unexpected error for duration string: %v", itemStr.Error())
	}

	itemSnowflake := New(types.Snowflake(175928889953484800))
	if itemSnowflake.Error() != nil {
		t.Fatalf("unexpected error for snowflake: %v", itemSnowflake.Error())
	}

	itemUnix := New(int64(1700000000))
	if itemUnix.Error() != nil || itemUnix.Unix() != 1700000000 {
		t.Fatalf("unexpected error for unix timestamp: %v", itemUnix.Error())
	}

	itemUnsupported := New([]string{"bad"})
	if itemUnsupported.Error() != ErrUnsupportedType {
		t.Fatalf("expected ErrUnsupportedType, got %v", itemUnsupported.Error())
	}
}

func TestItem_Discord(t *testing.T) {
	fixed := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	item := New(fixed)

	if item.Discord() != "<t:1735732800:R>" {
		t.Fatalf("expected default relative discord format, got %s", item.Discord())
	}
	if item.Discord("t") != "<t:1735732800:t>" {
		t.Fatalf("expected short time discord format, got %s", item.Discord("t"))
	}
	if item.Discord("D") != "<t:1735732800:D>" {
		t.Fatalf("expected long date discord format, got %s", item.Discord("D"))
	}
}

func TestItem_Jitter(t *testing.T) {
	base := New(time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC))
	jittered := base.Jitter("10s")
	if jittered.Error() != nil {
		t.Fatalf("unexpected error jittering: %v", jittered.Error())
	}

	largeJitter := base.Jitter(200 * Year)
	if largeJitter.Error() != nil {
		t.Fatalf("unexpected error on large jitter: %v", largeJitter.Error())
	}

	invJitter := base.Jitter("invalid")
	if invJitter.Time() != base.Time() {
		t.Fatalf("expected invalid jitter to return unchanged item")
	}
}

func TestFormat(t *testing.T) {
	fixed := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	u, err := Format(fixed, "unix")
	if err != nil || u != "1735732800" {
		t.Fatalf("expected unix format 1735732800, got %s, err: %v", u, err)
	}

	d, err := Format(fixed, "discord")
	if err != nil || d != "<t:1735732800:R>" {
		t.Fatalf("expected discord format <t:1735732800:R>, got %s, err: %v", d, err)
	}

	dLong, err := Format(fixed, "D")
	if err != nil || dLong != "<t:1735732800:D>" {
		t.Fatalf("expected discord format <t:1735732800:D>, got %s, err: %v", dLong, err)
	}

	tLong, err := Format(fixed, "T")
	if err != nil || tLong != "<t:1735732800:T>" {
		t.Fatalf("expected discord format <t:1735732800:T>, got %s, err: %v", tLong, err)
	}

	fLong, err := Format(fixed, "F")
	if err != nil || fLong != "<t:1735732800:F>" {
		t.Fatalf("expected discord format <t:1735732800:F>, got %s, err: %v", fLong, err)
	}
}
