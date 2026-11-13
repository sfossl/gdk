package context

import (
	"context"
	"testing"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/component"
)

func TestRespond_AutoFlagComponentsV2(t *testing.T) {
	in := &types.Interaction{
		ID:    types.Snowflake(200),
		Token: "test-token",
	}
	c := New(context.Background(), in, nil)

	card := component.Container(component.Text("Hello from layout!")).Build()
	_, err := c.Respond(Options{
		Components: []types.Component{card},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case res := <-c.ResponseChannel():
		if res.Data == nil {
			t.Fatalf("expected response data")
		}
		if res.Data.Flags&types.MessageFlagIsComponentsV2 == 0 {
			t.Fatalf("expected MessageFlagIsComponentsV2 in flags, got %d", res.Data.Flags)
		}
	default:
		t.Fatalf("expected response on channel")
	}
}

func TestRespond_ValidationFailure(t *testing.T) {
	in := &types.Interaction{
		ID:    types.Snowflake(201),
		Token: "test-token",
	}
	c := New(context.Background(), in, nil)

	badCard := component.Container(
		component.Container(component.Text("nested")).Build(),
	).Build()

	_, err := c.Respond(Options{
		Components: []types.Component{badCard},
	})
	if err != component.ErrNestedContainers {
		t.Fatalf("expected ErrNestedContainers, got %v", err)
	}
}

func TestUpdateMessage_AutoFlagComponentsV2(t *testing.T) {
	in := &types.Interaction{
		ID:    types.Snowflake(202),
		Token: "test-token",
	}
	c := New(context.Background(), in, nil)

	card := component.Container(component.Text("Updated content")).Build()
	_, err := c.UpdateMessage(Options{
		Components: []types.Component{card},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case res := <-c.ResponseChannel():
		if res.Data == nil {
			t.Fatalf("expected response data")
		}
		if res.Data.Flags&types.MessageFlagIsComponentsV2 == 0 {
			t.Fatalf("expected MessageFlagIsComponentsV2 in flags, got %d", res.Data.Flags)
		}
	default:
		t.Fatalf("expected response on channel")
	}
}
