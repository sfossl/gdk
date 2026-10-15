package context

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

func TestUpdate_InitialWebhook(t *testing.T) {
	in := &types.Interaction{
		ID:    types.Snowflake(100),
		Token: "test-token",
	}
	c := New(context.Background(), in, nil)

	msg, err := c.Update(Response{
		Content: "updated content",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != nil {
		t.Fatalf("expected nil message for initial update, got: %+v", msg)
	}
	if !c.IsAcknowledged() {
		t.Fatalf("expected context to be acknowledged")
	}

	select {
	case res := <-c.ResponseChannel():
		if res.Type != types.CallbackTypeUpdateMessage {
			t.Fatalf("expected callback type %d, got: %d", types.CallbackTypeUpdateMessage, res.Type)
		}
		if res.Data == nil || res.Data.Content != "updated content" {
			t.Fatalf("unexpected response data: %+v", res.Data)
		}
	default:
		t.Fatalf("expected response on channel")
	}
}

func TestUpdate_InitialGateway(t *testing.T) {
	var receivedResponse types.InteractionResponse
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/interactions/100/test-token/callback" {
			_ = json.NewDecoder(r.Body).Decode(&receivedResponse)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	in := &types.Interaction{
		ID:    types.Snowflake(100),
		Token: "test-token",
	}
	c := NewGatewayContext(context.Background(), in, client)

	msg, err := c.Update(Response{
		Content: "gateway update",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != nil {
		t.Fatalf("expected nil message, got: %+v", msg)
	}
	if !c.IsAcknowledged() {
		t.Fatalf("expected context to be acknowledged")
	}
	if receivedResponse.Type != types.CallbackTypeUpdateMessage {
		t.Fatalf("expected callback type %d, got %d", types.CallbackTypeUpdateMessage, receivedResponse.Type)
	}
	if receivedResponse.Data == nil || receivedResponse.Data.Content != "gateway update" {
		t.Fatalf("unexpected received response: %+v", receivedResponse.Data)
	}
}

func TestUpdate_DeferUpdateAndEditOriginal(t *testing.T) {
	var receivedDeferType types.InteractionCallbackType
	var editedContent string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/interactions/100/test-token/callback":
			var resp types.InteractionResponse
			_ = json.NewDecoder(r.Body).Decode(&resp)
			receivedDeferType = resp.Type
			w.WriteHeader(http.StatusNoContent)
		case "/webhooks/50/test-token/messages/@original":
			var payload rest.MessagePayload
			_ = json.NewDecoder(r.Body).Decode(&payload)
			editedContent = payload.Content
			_ = json.NewEncoder(w).Encode(types.Message{
				ID:      types.Snowflake(888),
				Content: payload.Content,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	in := &types.Interaction{
		ID:    types.Snowflake(100),
		AppID: types.Snowflake(50),
		Token: "test-token",
	}
	c := NewGatewayContext(context.Background(), in, client)

	err := c.DeferUpdate()
	if err != nil {
		t.Fatalf("unexpected defer update error: %v", err)
	}
	if receivedDeferType != types.CallbackTypeDeferredUpdateMessage {
		t.Fatalf("expected callback type %d, got: %d", types.CallbackTypeDeferredUpdateMessage, receivedDeferType)
	}

	msg, err := c.Update(Response{
		Content: "updated after defer update",
	})
	if err != nil {
		t.Fatalf("unexpected edit error: %v", err)
	}
	if msg == nil || msg.Content != "updated after defer update" {
		t.Fatalf("unexpected message: %+v", msg)
	}
	if editedContent != "updated after defer update" {
		t.Fatalf("expected edited content 'updated after defer update', got: %s", editedContent)
	}

	secondEdit, err := c.Update(Response{
		Content: "third update in-place",
	})
	if err != nil {
		t.Fatalf("unexpected second update error: %v", err)
	}
	if secondEdit == nil || secondEdit.Content != "third update in-place" {
		t.Fatalf("unexpected second edit message: %+v", secondEdit)
	}
}

