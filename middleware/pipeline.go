package middleware

import (
	"github.com/sfossl/gdk/context"
)

type Pipeline struct {
	middlewares []Middleware
}

func NewPipeline(middlewares ...Middleware) *Pipeline {
	p := &Pipeline{
		middlewares: make([]Middleware, 0, len(middlewares)),
	}
	p.Use(middlewares...)
	return p
}

func (p *Pipeline) Use(middlewares ...Middleware) *Pipeline {
	if p == nil {
		return p
	}
	for _, m := range middlewares {
		if m != nil {
			p.middlewares = append(p.middlewares, m)
		}
	}
	return p
}

func (p *Pipeline) Clone() *Pipeline {
	if p == nil {
		return NewPipeline()
	}
	cloned := make([]Middleware, len(p.middlewares))
	copy(cloned, p.middlewares)
	return &Pipeline{
		middlewares: cloned,
	}
}

func (p *Pipeline) Wrap(handler Handler) Handler {
	if handler == nil {
		handler = func(ctx *context.Context) error {
			return nil
		}
	}
	if p == nil {
		return handler
	}
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		handler = p.middlewares[i](handler)
	}
	return handler
}

func (p *Pipeline) Execute(ctx *context.Context, handler Handler) error {
	return p.Wrap(handler)(ctx)
}

func Chain(middlewares ...Middleware) Middleware {
	return func(final Handler) Handler {
		if final == nil {
			final = func(ctx *context.Context) error {
				return nil
			}
		}
		for i := len(middlewares) - 1; i >= 0; i-- {
			if middlewares[i] != nil {
				final = middlewares[i](final)
			}
		}
		return final
	}
}
