package context

import (
	"context"
	"errors"
	"testing"

	"github.com/sfossl/gda/types"
)

func TestGatewayMode_NoRESTClientReturnsError(t *testing.T) {
	in := &types.Interaction{ID: 100, Token: "token"}

	t.Run("Defer without REST", func(t *testing.T) {
		c := NewGatewayContext(context.Background(), in, nil)
		err := c.Defer()
		if !errors.Is(err, ErrNoRESTClient) {
			t.Fatalf("expected ErrNoRESTClient, got %v", err)
		}
		if c.IsAcknowledged() || c.IsDeferred() {
			t.Fatalf("context state must reset after failed defer")
		}
	})

	t.Run("Respond without REST", func(t *testing.T) {
		c := NewGatewayContext(context.Background(), in, nil)
		_, err := c.Respond("test")
		if !errors.Is(err, ErrNoRESTClient) {
			t.Fatalf("expected ErrNoRESTClient, got %v", err)
		}
		if c.IsAcknowledged() {
			t.Fatalf("context state must reset after failed respond")
		}
	})

	t.Run("Pong without REST", func(t *testing.T) {
		c := NewGatewayContext(context.Background(), in, nil)
		err := c.Pong()
		if !errors.Is(err, ErrNoRESTClient) {
			t.Fatalf("expected ErrNoRESTClient, got %v", err)
		}
		if c.IsAcknowledged() {
			t.Fatalf("context state must reset after failed pong")
		}
	})

	t.Run("Autocomplete without REST", func(t *testing.T) {
		c := NewGatewayContext(context.Background(), in, nil)
		err := c.Autocomplete([]types.Choice{{Name: "opt", Value: "val"}})
		if !errors.Is(err, ErrNoRESTClient) {
			t.Fatalf("expected ErrNoRESTClient, got %v", err)
		}
		if c.IsAcknowledged() {
			t.Fatalf("context state must reset after failed autocomplete")
		}
	})

	t.Run("Modal without REST", func(t *testing.T) {
		c := NewGatewayContext(context.Background(), in, nil)
		err := c.Modal(Modal{CustomID: "form", Title: "Title"})
		if !errors.Is(err, ErrNoRESTClient) {
			t.Fatalf("expected ErrNoRESTClient, got %v", err)
		}
		if c.IsAcknowledged() {
			t.Fatalf("context state must reset after failed modal")
		}
	})

	t.Run("DeferUpdate without REST", func(t *testing.T) {
		c := NewGatewayContext(context.Background(), in, nil)
		err := c.DeferUpdate()
		if !errors.Is(err, ErrNoRESTClient) {
			t.Fatalf("expected ErrNoRESTClient, got %v", err)
		}
		if c.IsAcknowledged() || c.IsDeferred() {
			t.Fatalf("context state must reset after failed defer update")
		}
	})

	t.Run("UpdateMessage without REST", func(t *testing.T) {
		c := NewGatewayContext(context.Background(), in, nil)
		_, err := c.UpdateMessage(Response{Content: "edit"})
		if !errors.Is(err, ErrNoRESTClient) {
			t.Fatalf("expected ErrNoRESTClient, got %v", err)
		}
		if c.IsAcknowledged() {
			t.Fatalf("context state must reset after failed update message")
		}
	})
}
