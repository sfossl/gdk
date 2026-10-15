package component

import (
	"errors"
	"testing"
)

func TestRGB(t *testing.T) {
	c := RGB(255, 0, 128)
	expected := (255 << 16) | (0 << 8) | 128
	if c != expected {
		t.Fatalf("expected %d, got %d", expected, c)
	}

	clamped := RGB(-10, 300, 50)
	expectedClamped := (0 << 16) | (255 << 8) | 50
	if clamped != expectedClamped {
		t.Fatalf("expected %d, got %d", expectedClamped, clamped)
	}
}

func TestHex(t *testing.T) {
	tests := []struct {
		input       string
		expected    int
		expectedErr error
	}{
		{input: "#5865F2", expected: 0x5865F2},
		{input: "5865f2", expected: 0x5865F2},
		{input: "0x5865F2", expected: 0x5865F2},
		{input: "#FFF", expected: 0xFFFFFF},
		{input: "#000", expected: 0x000000},
		{input: "#f00", expected: 0xFF0000},
		{input: "invalid", expectedErr: ErrInvalidColor},
		{input: "12345", expectedErr: ErrInvalidColor},
		{input: "1234567", expectedErr: ErrInvalidColor},
	}

	for _, tt := range tests {
		val, err := Hex(tt.input)
		if tt.expectedErr != nil {
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("input %s expected error %v, got %v", tt.input, tt.expectedErr, err)
			}
		} else {
			if err != nil {
				t.Fatalf("input %s unexpected error: %v", tt.input, err)
			}
			if val != tt.expected {
				t.Fatalf("input %s expected %d, got %d", tt.input, tt.expected, val)
			}
		}
	}
}

func TestResolveColor(t *testing.T) {
	tests := []struct {
		input       any
		expected    int
		expectedErr error
	}{
		{input: nil, expected: 0},
		{input: "", expected: 0},
		{input: 0x5865F2, expected: 0x5865F2},
		{input: int64(0x57F287), expected: 0x57F287},
		{input: -1, expectedErr: ErrColorOutOfRange},
		{input: 0x1000000, expectedErr: ErrColorOutOfRange},
		{input: int64(-5), expectedErr: ErrColorOutOfRange},
		{input: int64(0x1000000), expectedErr: ErrColorOutOfRange},
		{input: "blurple", expected: ColorBlurple},
		{input: "BLURPLE", expected: ColorBlurple},
		{input: "red", expected: ColorRed},
		{input: "green", expected: ColorGreen},
		{input: "#5865F2", expected: 0x5865F2},
		{input: 3.14, expectedErr: ErrInvalidColor},
		{input: struct{}{}, expectedErr: ErrInvalidColor},
	}

	for _, tt := range tests {
		val, err := ResolveColor(tt.input)
		if tt.expectedErr != nil {
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("input %v expected err %v, got %v", tt.input, tt.expectedErr, err)
			}
		} else {
			if err != nil {
				t.Fatalf("input %v unexpected err: %v", tt.input, err)
			}
			if val != tt.expected {
				t.Fatalf("input %v expected %d, got %d", tt.input, tt.expected, val)
			}
		}
	}
}

