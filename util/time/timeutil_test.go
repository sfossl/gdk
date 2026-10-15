package timeutil

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

	d3, err := ParseDuration("0.5h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected3 := 30 * time.Minute
	if d3 != expected3 {
		t.Fatalf("expected %v, got %v", expected3, d3)
	}
}

func TestParseDuration_Errors(t *testing.T) {
	if _, err := ParseDuration(""); !errors.Is(err, ErrEmptyDuration) {
		t.Fatalf("expected ErrEmptyDuration, got: %v", err)
	}
	if _, err := ParseDuration("abc"); !errors.Is(err, ErrInvalidDuration) {
		t.Fatalf("expected ErrInvalidDuration, got: %v", err)
	}
	if _, err := ParseDuration("10unknown"); !errors.Is(err, ErrUnknownUnit) {
		t.Fatalf("expected ErrUnknownUnit, got: %v", err)
	}
}

func TestFormatDuration_And_Human(t *testing.T) {
	d := 2*Day + 3*time.Hour + 10*time.Minute
	compact := FormatDuration(d)
	if compact != "2d 3h 10m" {
		t.Fatalf("unexpected compact format: %s", compact)
	}

	human := FormatDurationHuman(d)
	if human != "2 days, 3 hours, 10 minutes" {
		t.Fatalf("unexpected human format: %s", human)
	}

	dSingle := 1*Day + 1*time.Hour
	humanSingle := FormatDurationHuman(dSingle)
	if humanSingle != "1 day, 1 hour" {
		t.Fatalf("unexpected single human format: %s", humanSingle)
	}
}

func TestTime_Snowflake(t *testing.T) {
	sf := types.Snowflake(175928847299117063)
	item := New(sf)
	if item.Error() != nil {
		t.Fatalf("unexpected error: %v", item.Error())
	}

	expectedYear := 2016
	if item.Time().Year() != expectedYear {
		t.Fatalf("expected year %d, got %d", expectedYear, item.Time().Year())
	}

	itemStr := New("175928847299117063")
	if itemStr.Time().Year() != expectedYear {
		t.Fatalf("expected year %d from string, got %d", expectedYear, itemStr.Time().Year())
	}
}

func TestTime_DiscordTags(t *testing.T) {
	tFixed := time.Date(2026, 10, 15, 20, 0, 0, 0, time.UTC)
	item := New(tFixed)

	cases := []struct {
		style    string
		expected string
	}{
		{"", "<t:1792094400:R>"},
		{"relative", "<t:1792094400:R>"},
		{"d", "<t:1792094400:d>"},
		{"date_long", "<t:1792094400:D>"},
		{"time", "<t:1792094400:t>"},
		{"time_long", "<t:1792094400:T>"},
		{"full", "<t:1792094400:f>"},
		{"full_long", "<t:1792094400:F>"},
	}

	for _, c := range cases {
		var tag string
		if c.style == "" {
			tag = item.Discord()
		} else {
			tag = item.Discord(c.style)
		}
		if tag != c.expected {
			t.Fatalf("for style %s expected %s, got %s", c.style, c.expected, tag)
		}
	}
}

func TestTime_FormatHelper(t *testing.T) {
	tFixed := time.Date(2026, 10, 15, 20, 0, 0, 0, time.UTC)

	unixStr, err := Format(tFixed, "unix")
	if err != nil || unixStr != "1792094400" {
		t.Fatalf("unexpected unix output: %s, err: %v", unixStr, err)
	}

	isoStr, err := Format(tFixed, "iso")
	if err != nil || isoStr != "2026-10-15T20:00:00Z" {
		t.Fatalf("unexpected iso output: %s, err: %v", isoStr, err)
	}

	discordStr, err := Format(tFixed, "discord")
	if err != nil || discordStr != "<t:1792094400:R>" {
		t.Fatalf("unexpected discord output: %s, err: %v", discordStr, err)
	}

	durStr, err := Format("2d 4h", "duration")
	if err != nil || durStr != "2d 4h" {
		t.Fatalf("unexpected duration output: %s, err: %v", durStr, err)
	}

	customStr, err := Format(tFixed, "2006/01/02")
	if err != nil || customStr != "2026/10/15" {
		t.Fatalf("unexpected custom format output: %s, err: %v", customStr, err)
	}
}

func TestTime_Jitter(t *testing.T) {
	base := New("1h")
	jittered := base.Jitter("10s")

	diff := jittered.Time().Sub(base.Time())
	if diff < -10*time.Second || diff > 10*time.Second {
		t.Fatalf("jitter exceeded boundary: %v", diff)
	}

	invJitter := base.Jitter("invalid")
	if invJitter.Time() != base.Time() {
		t.Fatalf("invalid jitter should be no-op")
	}
}

func TestTime_Errors(t *testing.T) {
	itemNil := New(nil)
	if itemNil.Error() == nil {
		t.Fatalf("expected error for nil input")
	}

	_, err := Format("invalid!!date", "unix")
	if err == nil {
		t.Fatalf("expected error for invalid string")
	}

	itemStruct := New(struct{}{})
	if !errors.Is(itemStruct.Error(), ErrUnsupportedType) {
		t.Fatalf("expected ErrUnsupportedType, got: %v", itemStruct.Error())
	}

	itemZeroSnow := New(types.Snowflake(0))
	if itemZeroSnow.Error() == nil {
		t.Fatalf("expected error for zero snowflake")
	}

	itemZeroInt := New(0)
	if itemZeroInt.Error() == nil {
		t.Fatalf("expected error for zero int")
	}
}

func TestParseDuration_Signs(t *testing.T) {
	dPos, err := ParseDuration("+15m")
	if err != nil || dPos != 15*time.Minute {
		t.Fatalf("expected +15m, got: %v, err: %v", dPos, err)
	}

	dNeg, err := ParseDuration("-2h")
	if err != nil || dNeg != -2*time.Hour {
		t.Fatalf("expected -2h, got: %v, err: %v", dNeg, err)
	}

	dNegCompound, err := ParseDuration("-1d 2h")
	if err != nil || dNegCompound != -(24*time.Hour + 2*time.Hour) {
		t.Fatalf("expected -26h, got: %v, err: %v", dNegCompound, err)
	}
}
