package timeutil

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"
	"time"
)

func TestStress_DurationOverflowAndBounds(t *testing.T) {
	overflowInputs := []string{
		"9223372036854775808ns",
		"9223372036854776s",
		"1000000000000d",
		"999999999999y",
		"1000000000000000h",
	}

	for _, input := range overflowInputs {
		_, err := ParseDuration(input)
		if !errors.Is(err, ErrInvalidDuration) {
			t.Fatalf("expected ErrInvalidDuration for overflow input %s, got %v", input, err)
		}
	}

	maxValid := "9223372036854775807ns"
	d, err := ParseDuration(maxValid)
	if err != nil {
		t.Fatalf("unexpected error parsing max int64 nanos: %v", err)
	}
	if d != time.Duration(math.MaxInt64) {
		t.Fatalf("expected math.MaxInt64 duration, got %v", d)
	}
}

func TestStress_ConcurrentDurationParsing(t *testing.T) {
	inputs := []string{
		"1h 30m 15s",
		"500ms",
		"2w 3d 4h 5m 6s",
		"1.5h",
		"250µs",
		"100ns",
		"-15m",
		"+3h 45m",
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for _, in := range inputs {
				d, err := ParseDuration(in)
				if err != nil {
					t.Errorf("worker %d failed to parse %s: %v", workerID, in, err)
				}
				if d == 0 {
					t.Errorf("worker %d unexpected 0 duration for %s", workerID, in)
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestStress_MalformedAndFuzzDurations(t *testing.T) {
	malformed := []string{
		"",
		"   ",
		"abc",
		"10",
		"10.5.5s",
		"..s",
		"-",
		"+",
		"10xyz",
		"- -10s",
		"++10s",
		"10s -5s",
	}

	for _, m := range malformed {
		_, err := ParseDuration(m)
		if err == nil {
			t.Fatalf("expected error for malformed input %q, got nil", m)
		}
	}
}

func TestStress_ConcurrentItemFormatting(t *testing.T) {
	var wg sync.WaitGroup
	now := time.Now().UTC()

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			item := New(now.Add(time.Duration(idx) * time.Minute))
			if item.Error() != nil {
				t.Errorf("unexpected error creating item: %v", item.Error())
			}
			tag := item.Discord("f")
			expectedPrefix := "<t:"
			if len(tag) < len(expectedPrefix) || tag[:len(expectedPrefix)] != expectedPrefix {
				t.Errorf("unexpected format tag %s", tag)
			}
			rel := item.Discord("R")
			if len(rel) == 0 {
				t.Errorf("unexpected empty relative tag")
			}
		}(i)
	}
	wg.Wait()
}

func TestStress_FormatDurationCombinations(t *testing.T) {
	for sec := 1; sec <= 500; sec += 13 {
		d := time.Duration(sec) * time.Second
		formatted := FormatDuration(d)
		if len(formatted) == 0 {
			t.Fatalf("unexpected empty formatted duration for %v", d)
		}
		human := FormatDurationHuman(d)
		if len(human) == 0 {
			t.Fatalf("unexpected empty human formatted duration for %v", d)
		}
	}
}

func TestStress_SnowflakeConversion(t *testing.T) {
	itemZero := New(uint64(0))
	if itemZero.Error() == nil {
		t.Fatal("expected error for zero uint64 time")
	}

	itemValid := New(uint64(1700000000))
	if itemValid.Error() != nil {
		t.Fatalf("unexpected error for valid unix timestamp: %v", itemValid.Error())
	}
	if itemValid.Unix() != 1700000000 {
		t.Fatalf("expected unix timestamp 1700000000, got %d", itemValid.Unix())
	}

	formatted := fmt.Sprintf("%v", itemValid)
	if len(formatted) == 0 {
		t.Fatal("unexpected empty String() format")
	}
}
