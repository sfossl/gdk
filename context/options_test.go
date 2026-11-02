package context

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

func TestOptionsPointer_Respond(t *testing.T) {
	in := &types.Interaction{
		ID:    types.Snowflake(100),
		Token: "test-token",
	}
	c := New(context.Background(), in, nil)

	opts := &Options{
		Content:   "pointer reply",
		Ephemeral: true,
	}

	_, err := c.Respond(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case res := <-c.ResponseChannel():
		if res.Data == nil || res.Data.Content != "pointer reply" {
			t.Fatalf("expected content 'pointer reply', got: %+v", res.Data)
		}
		if res.Data.Flags != types.MessageFlagEphemeral {
			t.Fatalf("expected ephemeral flag, got: %d", res.Data.Flags)
		}
	default:
		t.Fatalf("expected response on channel")
	}
}

func TestOptionsPointer_UnifiedMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"123","owner_id":"999","roles":[]}`))
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"456","name":"new"}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPatch:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"123","name":"edited"}`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer server.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(server.URL))
	gid := types.Snowflake(123)
	in := &types.Interaction{
		ID:      types.Snowflake(100),
		GuildID: &gid,
		Data: &types.InteractionData{
			TargetID: &gid,
		},
	}
	c := New(context.Background(), in, client)

	modOpts := &Options{
		Type:   "kick",
		Target: types.Snowflake(456),
		Reason: "pointer test",
	}
	_ = c.Mod(modOpts)

	chOpts := &Options{
		Type:    "lock",
		Channel: types.Snowflake(789),
		Reason:  "pointer lock",
	}
	if err := c.ManageChannel(chOpts); err != nil {
		t.Fatalf("unexpected channel error: %v", err)
	}

	roleOpts := &Options{
		Type: "create",
		Name: "test-role",
	}
	if _, err := c.ManageRole(roleOpts); err != nil {
		t.Fatalf("unexpected role error: %v", err)
	}

	msgOpts := &Options{
		Type:    "delete",
		Channel: types.Snowflake(789),
		Message: types.Snowflake(999),
	}
	if err := c.ManageMessage(msgOpts); err != nil {
		t.Fatalf("unexpected message error: %v", err)
	}

	updateOpts := &Options{
		Target:   types.Snowflake(456),
		Nickname: func(s string) *string { return &s }("test-nick"),
	}
	if err := c.Update(updateOpts); err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}
}

func TestOptionsResolveType(t *testing.T) {
	optSend := Options{Type: "SEND"}
	if optSend.ResolveType() != "send" {
		t.Fatalf("expected send, got %s", optSend.ResolveType())
	}
	optNil := Options{Type: nil}
	if optNil.ResolveType() != "" {
		t.Fatalf("expected empty, got %s", optNil.ResolveType())
	}
	optInt := Options{Type: 123}
	if optInt.ResolveType() != "" {
		t.Fatalf("expected empty, got %s", optInt.ResolveType())
	}
}

