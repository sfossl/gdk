package timeutil

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/sfossl/gda/types"
)

var (
	ErrUnsupportedType = errors.New("unsupported time input type")
	ErrInvalidTime     = errors.New("invalid time format")
)

type Item struct {
	instant  time.Time
	duration time.Duration
	isDur    bool
	err      error
}

func New(input any) Item {
	if input == nil {
		return Item{err: ErrInvalidTime}
	}

	switch v := input.(type) {
	case Item:
		return v
	case time.Time:
		utc := v.UTC()
		return Item{
			instant:  utc,
			duration: utc.Sub(time.Now().UTC()),
		}
	case time.Duration:
		return Item{
			instant:  time.Now().UTC().Add(v),
			duration: v,
			isDur:    true,
		}
	case types.Snowflake:
		if v == 0 {
			return Item{err: ErrInvalidTime}
		}
		t := v.Time().UTC()
		return Item{
			instant:  t,
			duration: t.Sub(time.Now().UTC()),
		}
	case int:
		return fromInt(int64(v))
	case int64:
		return fromInt(v)
	case uint64:
		return fromUint(v)
	case string:
		return fromString(v)
	default:
		return Item{err: ErrUnsupportedType}
	}
}

func fromInt(n int64) Item {
	if n <= 0 {
		return Item{err: ErrInvalidTime}
	}
	return fromUint(uint64(n))
}

func fromUint(n uint64) Item {
	if n == 0 {
		return Item{err: ErrInvalidTime}
	}
	if n <= 10000000000 {
		t := time.Unix(int64(n), 0).UTC()
		return Item{
			instant:  t,
			duration: t.Sub(time.Now().UTC()),
		}
	}
	if n <= 10000000000000 {
		t := time.UnixMilli(int64(n)).UTC()
		return Item{
			instant:  t,
			duration: t.Sub(time.Now().UTC()),
		}
	}
	sf := types.Snowflake(n)
	t := sf.Time().UTC()
	return Item{
		instant:  t,
		duration: t.Sub(time.Now().UTC()),
	}
}

func fromString(s string) Item {
	s = strings.TrimSpace(s)
	if s == "" {
		return Item{err: ErrInvalidTime}
	}

	if dur, err := ParseDuration(s); err == nil {
		return Item{
			instant:  time.Now().UTC().Add(dur),
			duration: dur,
			isDur:    true,
		}
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			utc := t.UTC()
			return Item{
				instant:  utc,
				duration: utc.Sub(time.Now().UTC()),
			}
		}
	}

	if u, err := strconv.ParseUint(s, 10, 64); err == nil {
		return fromUint(u)
	}

	return Item{err: ErrInvalidTime}
}

func (i Item) Time() time.Time {
	return i.instant
}

func (i Item) Duration() time.Duration {
	return i.duration
}

func (i Item) Unix() int64 {
	return i.instant.Unix()
}

func (i Item) UnixMilli() int64 {
	return i.instant.UnixMilli()
}

func (i Item) ISO() string {
	return i.instant.Format(time.RFC3339)
}

func (i Item) Discord(style ...string) string {
	code := "R"
	if len(style) > 0 {
		switch strings.ToLower(style[0]) {
		case "t", "time":
			code = "t"
		case "t_long", "time_long", "long_time":
			code = "T"
		case "d", "date":
			code = "d"
		case "d_long", "date_long", "long_date":
			code = "D"
		case "f", "full", "short_date_time":
			code = "f"
		case "f_long", "full_long", "long_date_time":
			code = "F"
		case "r", "relative":
			code = "R"
		}
	}
	return fmt.Sprintf("<t:%d:%s>", i.instant.Unix(), code)
}

func (i Item) Human() string {
	if i.isDur {
		return FormatDurationHuman(i.duration)
	}
	return FormatDurationHuman(i.duration)
}

func (i Item) Jitter(maxJitter any) Item {
	if i.err != nil {
		return i
	}
	var maxDur time.Duration
	switch v := maxJitter.(type) {
	case time.Duration:
		maxDur = v
	case string:
		d, err := ParseDuration(v)
		if err != nil {
			return i
		}
		maxDur = d
	default:
		return i
	}

	if maxDur <= 0 {
		return i
	}

	ns := maxDur.Nanoseconds()
	doubleRange := big.NewInt(2 * ns)
	r, err := rand.Int(rand.Reader, doubleRange)
	if err != nil {
		return i
	}
	offset := r.Int64() - ns
	jitterDuration := time.Duration(offset)

	return Item{
		instant:  i.instant.Add(jitterDuration),
		duration: i.duration + jitterDuration,
		isDur:    i.isDur,
	}
}

func (i Item) Error() error {
	return i.err
}

func Format(input any, format ...string) (string, error) {
	item := New(input)
	if item.Error() != nil {
		return "", item.Error()
	}

	f := "unix"
	if len(format) > 0 && strings.TrimSpace(format[0]) != "" {
		f = strings.ToLower(strings.TrimSpace(format[0]))
	}

	switch f {
	case "unix":
		return strconv.FormatInt(item.Unix(), 10), nil
	case "unix_ms", "unix-ms", "milli", "millis":
		return strconv.FormatInt(item.UnixMilli(), 10), nil
	case "iso", "iso8601", "rfc3339":
		return item.ISO(), nil
	case "discord", "discord_relative", "discord-relative", "relative", "r":
		return item.Discord("R"), nil
	case "discord_date", "discord-date", "date", "d":
		return item.Discord("d"), nil
	case "discord_date_long", "discord-date-long", "long_date":
		return item.Discord("D"), nil
	case "discord_time", "discord-time", "time", "t":
		return item.Discord("t"), nil
	case "discord_time_long", "discord-time-long", "long_time":
		return item.Discord("T"), nil
	case "discord_full", "discord-full", "full", "f":
		return item.Discord("f"), nil
	case "discord_full_long", "discord-full-long", "long_full":
		return item.Discord("F"), nil
	case "human":
		return item.Human(), nil
	case "duration":
		return FormatDuration(item.Duration()), nil
	default:
		return item.instant.Format(f), nil
	}
}

