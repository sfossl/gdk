package util

import "testing"

func TestPtr(t *testing.T) {
	s := "test"
	if p := Ptr(s); p == nil || *p != s {
		t.Fatalf("expected %s, got %v", s, p)
	}

	i := 42
	if p := Ptr(i); p == nil || *p != i {
		t.Fatalf("expected %d, got %v", i, p)
	}

	b := true
	if p := Ptr(b); p == nil || *p != b {
		t.Fatalf("expected %v, got %v", b, p)
	}

	f := 3.14
	if p := Ptr(f); p == nil || *p != f {
		t.Fatalf("expected %f, got %v", f, p)
	}
}

