package middleware

import (
	stdctx "context"
	"errors"
	"testing"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/context"
)

func TestGuard_PassThrough(t *testing.T) {
	guard := Guard()
	called := false

	handler := guard(func(ctx *context.Context) error {
		called = true
		return nil
	})

	ctx := context.New(stdctx.Background(), &types.Interaction{}, nil)
	if err := handler(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected downstream handler to be called")
	}
}

func TestGuard_PassThroughError(t *testing.T) {
	guard := Guard()
	expectedErr := errors.New("business logic failure")

	handler := guard(func(ctx *context.Context) error {
		return expectedErr
	})

	ctx := context.New(stdctx.Background(), &types.Interaction{}, nil)
	err := handler(ctx)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestGuard_RecoversPanicAndInvokesCallback(t *testing.T) {
	var capturedPanic any
	var capturedStack []byte
	var capturedCtx *context.Context

	guard := Guard(
		WithPanicHandler(func(ctx *context.Context, r any, stack []byte) {
			capturedCtx = ctx
			capturedPanic = r
			capturedStack = stack
		}),
		WithMessage("Custom error notice"),
		WithEphemeral(false),
	)

	handler := guard(func(ctx *context.Context) error {
		panic("fatal handler crash")
	})

	in := &types.Interaction{
		ID:    types.Snowflake(111),
		AppID: types.Snowflake(222),
		Token: "test_token",
	}
	ctx := context.New(stdctx.Background(), in, nil)

	err := handler(ctx)
	if !errors.Is(err, ErrPanicRecovered) {
		t.Fatalf("expected ErrPanicRecovered, got %v", err)
	}

	if capturedCtx != ctx {
		t.Fatal("expected captured context to match")
	}
	if str, ok := capturedPanic.(string); !ok || str != "fatal handler crash" {
		t.Fatalf("unexpected captured panic value %v", capturedPanic)
	}
	if len(capturedStack) == 0 {
		t.Fatal("expected non-empty stack trace")
	}

	select {
	case resp := <-ctx.ResponseChannel():
		if resp.Type != types.CallbackTypeChannelMessageWithSource {
			t.Fatalf("expected message callback type, got %v", resp.Type)
		}
		if resp.Data == nil || resp.Data.Content != "Custom error notice" {
			t.Fatalf("unexpected response content: %v", resp.Data)
		}
		if resp.Data.Flags != 0 {
			t.Fatalf("expected non-ephemeral response, got flags %v", resp.Data.Flags)
		}
	default:
		t.Fatal("expected response payload on channel")
	}
}

func TestGuard_DeferredInteractionUnblocked(t *testing.T) {
	guard := Guard(WithMessage("Recovered after defer"))

	handler := guard(func(ctx *context.Context) error {
		if err := ctx.Defer(); err != nil {
			return err
		}
		panic("crash after defer")
	})

	in := &types.Interaction{
		ID:    types.Snowflake(333),
		AppID: types.Snowflake(444),
		Token: "defer_token",
	}
	ctx := context.New(stdctx.Background(), in, nil)

	err := handler(ctx)
	if !errors.Is(err, ErrPanicRecovered) {
		t.Fatalf("expected ErrPanicRecovered, got %v", err)
	}

	select {
	case resp := <-ctx.ResponseChannel():
		if resp.Type != types.CallbackTypeDeferredChannelMessageWithSource {
			t.Fatalf("expected deferred callback type, got %v", resp.Type)
		}
	default:
		t.Fatal("expected deferred response on channel")
	}
}

func TestGuard_NilContextHandling(t *testing.T) {
	guard := Guard(nil)

	handler := guard(func(ctx *context.Context) error {
		panic("nil context panic")
	})

	err := handler(nil)
	if !errors.Is(err, ErrPanicRecovered) {
		t.Fatalf("expected ErrPanicRecovered for nil context, got %v", err)
	}
}

