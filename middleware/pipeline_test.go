package middleware

import (
	stdctx "context"
	"errors"
	"reflect"
	"testing"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/context"
)

func TestPipeline_ExecutionOrder(t *testing.T) {
	var trace []string

	m1 := func(next Handler) Handler {
		return func(ctx *context.Context) error {
			trace = append(trace, "m1_before")
			err := next(ctx)
			trace = append(trace, "m1_after")
			return err
		}
	}

	m2 := func(next Handler) Handler {
		return func(ctx *context.Context) error {
			trace = append(trace, "m2_before")
			err := next(ctx)
			trace = append(trace, "m2_after")
			return err
		}
	}

	p := NewPipeline(m1, m2)
	ctx := context.New(stdctx.Background(), &types.Interaction{}, nil)

	err := p.Execute(ctx, func(c *context.Context) error {
		trace = append(trace, "handler")
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"m1_before", "m2_before", "handler", "m2_after", "m1_after"}
	if !reflect.DeepEqual(trace, expected) {
		t.Fatalf("expected trace %v, got %v", expected, trace)
	}
}

func TestPipeline_ShortCircuit(t *testing.T) {
	errBlocked := errors.New("blocked by m1")
	m2Called := false
	handlerCalled := false

	m1 := func(next Handler) Handler {
		return func(ctx *context.Context) error {
			return errBlocked
		}
	}

	m2 := func(next Handler) Handler {
		return func(ctx *context.Context) error {
			m2Called = true
			return next(ctx)
		}
	}

	p := NewPipeline(m1, m2)
	ctx := context.New(stdctx.Background(), &types.Interaction{}, nil)

	err := p.Execute(ctx, func(c *context.Context) error {
		handlerCalled = true
		return nil
	})

	if !errors.Is(err, errBlocked) {
		t.Fatalf("expected %v, got %v", errBlocked, err)
	}
	if m2Called {
		t.Fatal("m2 should not have been called")
	}
	if handlerCalled {
		t.Fatal("handler should not have been called")
	}
}

func TestPipeline_CloneAndBranch(t *testing.T) {
	base := NewPipeline()
	var baseTrace []string

	mBase := func(next Handler) Handler {
		return func(ctx *context.Context) error {
			baseTrace = append(baseTrace, "base")
			return next(ctx)
		}
	}
	base.Use(mBase)

	branch := base.Clone()
	mBranch := func(next Handler) Handler {
		return func(ctx *context.Context) error {
			baseTrace = append(baseTrace, "branch")
			return next(ctx)
		}
	}
	branch.Use(mBranch)

	ctx := context.New(stdctx.Background(), &types.Interaction{}, nil)

	_ = base.Execute(ctx, func(c *context.Context) error { return nil })
	if len(baseTrace) != 1 || baseTrace[0] != "base" {
		t.Fatalf("base pipeline should only execute base middleware, got %v", baseTrace)
	}

	baseTrace = nil
	_ = branch.Execute(ctx, func(c *context.Context) error { return nil })
	expected := []string{"base", "branch"}
	if !reflect.DeepEqual(baseTrace, expected) {
		t.Fatalf("branch pipeline should execute both, got %v", baseTrace)
	}
}

func TestChain_Helper(t *testing.T) {
	var trace []int

	m1 := func(next Handler) Handler {
		return func(ctx *context.Context) error {
			trace = append(trace, 1)
			return next(ctx)
		}
	}
	m2 := func(next Handler) Handler {
		return func(ctx *context.Context) error {
			trace = append(trace, 2)
			return next(ctx)
		}
	}

	chained := Chain(m1, nil, m2)
	handler := chained(func(ctx *context.Context) error {
		trace = append(trace, 3)
		return nil
	})

	ctx := context.New(stdctx.Background(), &types.Interaction{}, nil)
	if err := handler(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(trace, expected) {
		t.Fatalf("expected trace %v, got %v", expected, trace)
	}
}

func TestPipeline_NilHandlerAndMiddlewares(t *testing.T) {
	p := NewPipeline(nil)
	wrapped := p.Wrap(nil)
	if err := wrapped(nil); err != nil {
		t.Fatalf("expected nil error for nil handler, got %v", err)
	}
}

