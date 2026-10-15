package timeutil

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	ErrInvalidDuration = errors.New("invalid duration format")
	ErrEmptyDuration   = errors.New("empty duration string")
	ErrUnknownUnit     = errors.New("unknown duration unit")
)

const (
	Day   = 24 * time.Hour
	Month = 30 * Day
	Year  = 365 * Day
)

func ParseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrEmptyDuration
	}

	var total time.Duration
	runes := []rune(s)
	n := len(runes)
	i := 0

	for i < n && unicode.IsSpace(runes[i]) {
		i++
	}
	if i >= n {
		return 0, ErrEmptyDuration
	}

	isNegative := false
	switch runes[i] {
	case '-':
		isNegative = true
		i++
	case '+':
		i++
	}

	parsedParts := 0
	for i < n {
		for i < n && unicode.IsSpace(runes[i]) {
			i++
		}
		if i >= n {
			break
		}

		startNum := i
		hasDot := false
		for i < n && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
			if runes[i] == '.' {
				if hasDot {
					return 0, ErrInvalidDuration
				}
				hasDot = true
			}
			i++
		}

		if startNum == i {
			return 0, ErrInvalidDuration
		}

		numStr := string(runes[startNum:i])

		for i < n && unicode.IsSpace(runes[i]) {
			i++
		}

		startUnit := i
		for i < n && unicode.IsLetter(runes[i]) {
			i++
		}

		if startUnit == i {
			return 0, ErrInvalidDuration
		}

		unitStr := strings.ToLower(string(runes[startUnit:i]))
		unitDur, err := parseUnit(unitStr)
		if err != nil {
			return 0, err
		}

		var partDur time.Duration
		if hasDot {
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, ErrInvalidDuration
			}
			nanos := val * float64(unitDur)
			if nanos >= 9223372036854775808.0 || nanos < 0 {
				return 0, ErrInvalidDuration
			}
			partDur = time.Duration(nanos)
		} else {
			intVal, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil || intVal < 0 {
				return 0, ErrInvalidDuration
			}
			unitNanos := int64(unitDur)
			if intVal > math.MaxInt64/unitNanos {
				return 0, ErrInvalidDuration
			}
			partDur = time.Duration(intVal * unitNanos)
		}

		if partDur < 0 || math.MaxInt64-total < partDur {
			return 0, ErrInvalidDuration
		}
		total += partDur
		parsedParts++
	}

	if parsedParts == 0 {
		return 0, ErrInvalidDuration
	}

	if isNegative {
		total = -total
	}

	return total, nil
}

func parseUnit(u string) (time.Duration, error) {
	switch u {
	case "ns", "nanosecond", "nanoseconds":
		return time.Nanosecond, nil
	case "us", "µs", "microsecond", "microseconds":
		return time.Microsecond, nil
	case "ms", "millisecond", "milliseconds":
		return time.Millisecond, nil
	case "s", "sec", "secs", "second", "seconds":
		return time.Second, nil
	case "m", "min", "mins", "minute", "minutes":
		return time.Minute, nil
	case "h", "hr", "hrs", "hour", "hours":
		return time.Hour, nil
	case "d", "day", "days":
		return Day, nil
	case "w", "wk", "wks", "week", "weeks":
		return 7 * Day, nil
	case "mo", "month", "months":
		return Month, nil
	case "y", "yr", "yrs", "year", "years":
		return Year, nil
	default:
		return 0, ErrUnknownUnit
	}
}

func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}

	var parts []string

	years := d / Year
	if years > 0 {
		parts = append(parts, strconv.FormatInt(int64(years), 10)+"y")
		d %= Year
	}

	months := d / Month
	if months > 0 {
		parts = append(parts, strconv.FormatInt(int64(months), 10)+"mo")
		d %= Month
	}

	days := d / Day
	if days > 0 {
		parts = append(parts, strconv.FormatInt(int64(days), 10)+"d")
		d %= Day
	}

	hours := d / time.Hour
	if hours > 0 {
		parts = append(parts, strconv.FormatInt(int64(hours), 10)+"h")
		d %= time.Hour
	}

	minutes := d / time.Minute
	if minutes > 0 {
		parts = append(parts, strconv.FormatInt(int64(minutes), 10)+"m")
		d %= time.Minute
	}

	seconds := d / time.Second
	if seconds > 0 {
		parts = append(parts, strconv.FormatInt(int64(seconds), 10)+"s")
		d %= time.Second
	}

	if len(parts) == 0 {
		if d >= time.Millisecond {
			return strconv.FormatInt(int64(d/time.Millisecond), 10) + "ms"
		}
		return strconv.FormatInt(int64(d), 10) + "ns"
	}

	return strings.Join(parts, " ")
}

func FormatDurationHuman(d time.Duration) string {
	if d <= 0 {
		return "0 seconds"
	}

	var parts []string

	years := d / Year
	if years > 0 {
		parts = append(parts, pluralize(int64(years), "year"))
		d %= Year
	}

	months := d / Month
	if months > 0 {
		parts = append(parts, pluralize(int64(months), "month"))
		d %= Month
	}

	days := d / Day
	if days > 0 {
		parts = append(parts, pluralize(int64(days), "day"))
		d %= Day
	}

	hours := d / time.Hour
	if hours > 0 {
		parts = append(parts, pluralize(int64(hours), "hour"))
		d %= time.Hour
	}

	minutes := d / time.Minute
	if minutes > 0 {
		parts = append(parts, pluralize(int64(minutes), "minute"))
		d %= time.Minute
	}

	seconds := d / time.Second
	if seconds > 0 {
		parts = append(parts, pluralize(int64(seconds), "second"))
	}

	if len(parts) == 0 {
		return "less than a second"
	}

	return strings.Join(parts, ", ")
}

func pluralize(n int64, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return strconv.FormatInt(n, 10) + " " + unit + "s"
}
