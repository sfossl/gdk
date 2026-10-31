package middleware

import (
	"errors"
	"runtime/debug"

	"github.com/sfossl/gdk/context"
	"github.com/sfossl/gdk/defaults"
)

var ErrPanicRecovered = errors.New("panic recovered in interaction handler")

type Handler func(ctx *context.Context) error
type Middleware func(next Handler) Handler
type PanicHandler func(ctx *context.Context, recovered any, stack []byte)

type GuardConfig struct {
	OnPanic   PanicHandler
	Message   string
	Ephemeral bool
}

type GuardOption func(*GuardConfig)

func WithPanicHandler(fn PanicHandler) GuardOption {
	return func(c *GuardConfig) {
		c.OnPanic = fn
	}
}

func WithMessage(msg string) GuardOption {
	return func(c *GuardConfig) {
		c.Message = msg
	}
}

func WithEphemeral(ephemeral bool) GuardOption {
	return func(c *GuardConfig) {
		c.Ephemeral = ephemeral
	}
}

func Guard(opts ...GuardOption) Middleware {
	cfg := GuardConfig{
		Message:   defaults.PanicMessage,
		Ephemeral: true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	return func(next Handler) Handler {
		return func(ctx *context.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					stack := debug.Stack()
					if cfg.OnPanic != nil {
						func() {
							defer func() {
								_ = recover()
							}()
							cfg.OnPanic(ctx, r, stack)
						}()
					}

					if ctx != nil && (!ctx.IsAcknowledged() || ctx.IsDeferred()) {
						_, _ = ctx.Respond(context.Response{
							Content:   cfg.Message,
							Ephemeral: cfg.Ephemeral,
						})
					}

					err = ErrPanicRecovered
				}
			}()

			if next != nil {
				return next(ctx)
			}
			return nil
		}
	}
}
