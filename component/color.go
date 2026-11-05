package component

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidColor    = errors.New("invalid color specification")
	ErrColorOutOfRange = errors.New("color value out of range 0 to 16777215")
)

const (
	ColorBlurple = 0x5865F2
	ColorGreen   = 0x57F287
	ColorYellow  = 0xFEE75C
	ColorFuchsia = 0xEB459E
	ColorRed     = 0xED4245
	ColorWhite   = 0xFFFFFF
	ColorBlack   = 0x000000
	ColorDark    = 0x2B2D31
	ColorNavy    = 0x001F3F
	ColorAqua    = 0x7FDBFF
	ColorTeal    = 0x39CCCC
	ColorOlive   = 0x3D9970
	ColorLime    = 0x01FF70
	ColorOrange  = 0xFF851B
	ColorMaroon  = 0x85144B
	ColorPurple  = 0xB10DC9
)

var namedColors = map[string]int{
	"blurple": ColorBlurple,
	"green":   ColorGreen,
	"yellow":  ColorYellow,
	"fuchsia": ColorFuchsia,
	"red":     ColorRed,
	"white":   ColorWhite,
	"black":   ColorBlack,
	"dark":    ColorDark,
	"navy":    ColorNavy,
	"aqua":    ColorAqua,
	"teal":    ColorTeal,
	"olive":   ColorOlive,
	"lime":    ColorLime,
	"orange":  ColorOrange,
	"maroon":  ColorMaroon,
	"purple":  ColorPurple,
}

func RGB(r, g, b int) int {
	return (clampColorByte(r) << 16) | (clampColorByte(g) << 8) | clampColorByte(b)
}

func Hex(s string) (int, error) {
	clean := strings.TrimSpace(s)
	clean = strings.TrimPrefix(clean, "#")
	clean = strings.TrimPrefix(clean, "0x")
	clean = strings.TrimPrefix(clean, "0X")

	if len(clean) == 3 {
		expanded := fmt.Sprintf("%c%c%c%c%c%c",
			clean[0], clean[0], clean[1], clean[1], clean[2], clean[2])
		clean = expanded
	}

	if len(clean) != 6 {
		return 0, ErrInvalidColor
	}

	val, err := strconv.ParseUint(clean, 16, 32)
	if err != nil {
		return 0, ErrInvalidColor
	}

	if val > 0xFFFFFF {
		return 0, ErrColorOutOfRange
	}

	return int(val), nil
}

func ResolveColor(c any) (int, error) {
	if c == nil {
		return 0, nil
	}

	switch val := c.(type) {
	case int:
		if val < 0 || val > 0xFFFFFF {
			return 0, ErrColorOutOfRange
		}
		return val, nil
	case int64:
		if val < 0 || val > 0xFFFFFF {
			return 0, ErrColorOutOfRange
		}
		return int(val), nil
	case string:
		clean := strings.ToLower(strings.TrimSpace(val))
		if clean == "" {
			return 0, nil
		}
		if named, ok := namedColors[clean]; ok {
			return named, nil
		}
		return Hex(clean)
	default:
		return 0, ErrInvalidColor
	}
}

func clampColorByte(val int) int {
	if val < 0 {
		return 0
	}
	if val > 255 {
		return 255
	}
	return val
}
