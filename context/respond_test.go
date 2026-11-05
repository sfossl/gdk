package context

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

func TestRespond_InitialWebhook(t *testing.T) {
	in := &types.Interaction{
		ID:    types.Snowflake(100),
		Token: "test-token",
	}
	c := New(context.Background(), in, nil)

	msg, err := c.Respond(Response{
		Content:   "hello webhook",
		Ephemeral: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != nil {
		t.Fatalf("expected nil message for initial response, got: %+v", msg)
	}
	if !c.IsAcknowledged() {
		t.Fatalf("expected context to be acknowledged")
	}

	select {
	case res := <-c.ResponseChannel():
		if res.Type != types.CallbackTypeChannelMessageWithSource {
			t.Fatalf("expected callback type %d, got: %d", types.CallbackTypeChannelMessageWithSource, res.Type)
		}
		if res.Data == nil || res.Data.Content != "hello webhook" {
			t.Fatalf("unexpected response data: %+v", res.Data)
		}
		if res.Data.Flags != types.MessageFlagEphemeral {
			t.Fatalf("expected ephemeral flag, got: %d", res.Data.Flags)
		}
	default:
		t.Fatalf("expected response on channel")
	}
}

func TestRespond_InitialGateway(t *testing.T) {
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

	msg, err := c.Respond(Response{
		Content: "hello gateway",
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
	if receivedResponse.Data == nil || receivedResponse.Data.Content != "hello gateway" {
		t.Fatalf("unexpected received response: %+v", receivedResponse.Data)
	}
}

func TestRespond_DeferAndEditOriginal(t *testing.T) {
	var editedContent string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/interactions/100/test-token/callback":
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

	err := c.Defer(false)
	if err != nil {
		t.Fatalf("unexpected defer error: %v", err)
	}
	if !c.IsDeferred() {
		t.Fatalf("expected context to be deferred")
	}

	msg, err := c.Respond(Response{
		Content: "updated after defer",
	})
	if err != nil {
		t.Fatalf("unexpected edit error: %v", err)
	}
	if msg == nil || msg.Content != "updated after defer" {
		t.Fatalf("unexpected message: %+v", msg)
	}
	if c.IsDeferred() {
		t.Fatalf("expected context to no longer be deferred")
	}
	if editedContent != "updated after defer" {
		t.Fatalf("expected edited content 'updated after defer', got: %s", editedContent)
	}
}

func TestRespond_AutomaticFollowup(t *testing.T) {
	var followupContent string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/interactions/100/test-token/callback":
			w.WriteHeader(http.StatusNoContent)
		case "/webhooks/50/test-token":
			var payload rest.MessagePayload
			_ = json.NewDecoder(r.Body).Decode(&payload)
			followupContent = payload.Content
			_ = json.NewEncoder(w).Encode(types.Message{
				ID:      types.Snowflake(999),
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

	_, err := c.Respond(Response{Content: "first message"})
	if err != nil {
		t.Fatalf("unexpected error on first response: %v", err)
	}

	followup, err := c.Respond(Response{Content: "second message followup"})
	if err != nil {
		t.Fatalf("unexpected error on followup: %v", err)
	}
	if followup == nil || followup.Content != "second message followup" {
		t.Fatalf("unexpected followup message: %+v", followup)
	}
	if followupContent != "second message followup" {
		t.Fatalf("expected followup content 'second message followup', got: %s", followupContent)
	}
}

func TestRespond_AlreadyRespondedDefer(t *testing.T) {
	in := &types.Interaction{
		ID:    types.Snowflake(100),
		Token: "test-token",
	}
	c := New(context.Background(), in, nil)
	_, _ = c.Respond(Response{Content: "initial"})

	err := c.Defer()
	if !errors.Is(err, ErrAlreadyResponded) {
		t.Fatalf("expected ErrAlreadyResponded, got: %v", err)
	}
}

func TestRespond_MissingREST(t *testing.T) {
	in := &types.Interaction{
		ID:    types.Snowflake(100),
		Token: "test-token",
	}
	c := New(context.Background(), in, nil)
	_, _ = c.Respond(Response{Content: "initial"})

	_, err := c.Respond(Response{Content: "followup"})
	if !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got: %v", err)
	}
}

func TestRespond_FailureResetAndRetry(t *testing.T) {
	failRequest := true
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failRequest {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	in := &types.Interaction{
		ID:    types.Snowflake(100),
		Token: "test-token",
	}
	c := NewGatewayContext(context.Background(), in, client)

	_, err := c.Respond(Response{Content: "try 1"})
	if err == nil {
		t.Fatalf("expected failure on try 1")
	}
	if c.IsAcknowledged() {
		t.Fatalf("expected context to not be marked acknowledged after failure")
	}

	failRequest = false
	_, err = c.Respond(Response{Content: "try 2"})
	if err != nil {
		t.Fatalf("expected retry to succeed, got: %v", err)
	}
	if !c.IsAcknowledged() {
		t.Fatalf("expected context to be marked acknowledged after successful retry")
	}
}

func TestRespond_ConcurrentDeferred(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/interactions/100/test-token/callback":
			w.WriteHeader(http.StatusNoContent)
		case "/webhooks/50/test-token/messages/@original":
			_ = json.NewEncoder(w).Encode(types.Message{ID: types.Snowflake(111), Content: "original"})
		case "/webhooks/50/test-token":
			_ = json.NewEncoder(w).Encode(types.Message{ID: types.Snowflake(222), Content: "followup"})
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
	_ = c.Defer()

	errChan := make(chan error, 2)
	go func() {
		_, err := c.Respond(Response{Content: "goroutine 1"})
		errChan <- err
	}()
	go func() {
		_, err := c.Respond(Response{Content: "goroutine 2"})
		errChan <- err
	}()

	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			t.Fatalf("concurrent respond failed: %v", err)
		}
	}
}

func TestContext_Modal(t *testing.T) {
	t.Run("success webhook", func(t *testing.T) {
		in := &types.Interaction{ID: 100, Token: "token"}
		c := New(context.Background(), in, nil)
		err := c.Modal(Modal{
			CustomID: "my_modal",
			Title:    "Feedback",
			Components: []types.Component{
				{Type: types.ComponentTypeActionRow},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !c.IsAcknowledged() {
			t.Fatalf("expected context to be acknowledged")
		}
		select {
		case res := <-c.ResponseChannel():
			if res.Type != types.CallbackTypeModal {
				t.Fatalf("expected modal callback, got %d", res.Type)
			}
			if res.Data.CustomID != "my_modal" || res.Data.Title != "Feedback" {
				t.Fatalf("unexpected modal data: %+v", res.Data)
			}
		default:
			t.Fatalf("expected response on channel")
		}
	})

	t.Run("modal after defer", func(t *testing.T) {
		in := &types.Interaction{ID: 100, Token: "token"}
		c := New(context.Background(), in, nil)
		if err := c.Defer(); err != nil {
			t.Fatalf("unexpected defer error: %v", err)
		}
		if err := c.Modal(Modal{CustomID: "modal"}); !errors.Is(err, ErrModalAfterDefer) {
			t.Fatalf("expected ErrModalAfterDefer, got %v", err)
		}
	})

	t.Run("modal after respond", func(t *testing.T) {
		in := &types.Interaction{ID: 100, Token: "token"}
		c := New(context.Background(), in, nil)
		if _, err := c.Respond(Response{Content: "hi"}); err != nil {
			t.Fatalf("unexpected respond error: %v", err)
		}
		if err := c.Modal(Modal{CustomID: "modal"}); !errors.Is(err, ErrAlreadyResponded) {
			t.Fatalf("expected ErrAlreadyResponded, got %v", err)
		}
	})
}

func TestRespond_Tier1Shortcuts(t *testing.T) {
	t.Run("simple string", func(t *testing.T) {
		in := &types.Interaction{ID: 100, Token: "token"}
		c := New(context.Background(), in, nil)
		_, err := c.Respond("Pong!")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		select {
		case res := <-c.ResponseChannel():
			if res.Data == nil || res.Data.Content != "Pong!" {
				t.Fatalf("expected Pong!, got %+v", res.Data)
			}
			if res.Data.Flags != 0 {
				t.Fatalf("expected 0 flags, got %d", res.Data.Flags)
			}
		default:
			t.Fatalf("expected response on channel")
		}
	})

	t.Run("string with ephemeral bool", func(t *testing.T) {
		in := &types.Interaction{ID: 100, Token: "token"}
		c := New(context.Background(), in, nil)
		_, err := c.Respond("Secret", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		select {
		case res := <-c.ResponseChannel():
			if res.Data == nil || res.Data.Content != "Secret" {
				t.Fatalf("expected Secret, got %+v", res.Data)
			}
			if res.Data.Flags != types.MessageFlagEphemeral {
				t.Fatalf("expected ephemeral flag, got %d", res.Data.Flags)
			}
		default:
			t.Fatalf("expected response on channel")
		}
	})

	t.Run("direct error", func(t *testing.T) {
		in := &types.Interaction{ID: 100, Token: "token"}
		c := New(context.Background(), in, nil)
		testErr := errors.New("something broke")
		_, err := c.Respond(testErr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		select {
		case res := <-c.ResponseChannel():
			if res.Data == nil || res.Data.Content != "something broke" {
				t.Fatalf("expected 'something broke', got %+v", res.Data)
			}
			if res.Data.Flags != types.MessageFlagEphemeral {
				t.Fatalf("expected ephemeral flag, got %d", res.Data.Flags)
			}
		default:
			t.Fatalf("expected response on channel")
		}
	})

	t.Run("options struct", func(t *testing.T) {
		in := &types.Interaction{ID: 100, Token: "token"}
		c := New(context.Background(), in, nil)
		_, err := c.Respond(Options{
			Content:   "Tier 2 Options",
			Ephemeral: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		select {
		case res := <-c.ResponseChannel():
			if res.Data == nil || res.Data.Content != "Tier 2 Options" {
				t.Fatalf("expected 'Tier 2 Options', got %+v", res.Data)
			}
			if res.Data.Flags != types.MessageFlagEphemeral {
				t.Fatalf("expected ephemeral flag, got %d", res.Data.Flags)
			}
		default:
			t.Fatalf("expected response on channel")
		}
	})

	t.Run("merged string and options", func(t *testing.T) {
		in := &types.Interaction{ID: 101, Token: "token"}
		c := New(context.Background(), in, nil)
		_, err := c.Respond("hello world", Options{Ephemeral: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		select {
		case res := <-c.ResponseChannel():
			if res.Data == nil || res.Data.Content != "hello world" {
				t.Fatalf("expected 'hello world', got %+v", res.Data)
			}
			if res.Data.Flags != types.MessageFlagEphemeral {
				t.Fatalf("expected ephemeral flag, got %d", res.Data.Flags)
			}
		default:
			t.Fatalf("expected response on channel")
		}
	})

	t.Run("single embed option and pointer", func(t *testing.T) {
		in := &types.Interaction{ID: 102, Token: "token"}
		c := New(context.Background(), in, nil)
		emb := types.Embed{Title: "Embed Title"}
		_, err := c.Respond(Options{Embed: &emb})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		select {
		case res := <-c.ResponseChannel():
			if res.Data == nil || len(res.Data.Embeds) != 1 || res.Data.Embeds[0].Title != "Embed Title" {
				t.Fatalf("expected single embed, got %+v", res.Data)
			}
		default:
			t.Fatalf("expected response on channel")
		}
	})
}

func TestContext_Autocomplete_Cap25(t *testing.T) {
	in := &types.Interaction{ID: 100, Token: "token"}
	c := New(context.Background(), in, nil)
	var choices []types.Choice
	for i := 0; i < 30; i++ {
		choices = append(choices, types.Choice{Name: fmt.Sprintf("C%d", i), Value: fmt.Sprintf("V%d", i)})
	}
	err := c.Autocomplete(choices)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case res := <-c.ResponseChannel():
		if res.Type != types.CallbackTypeApplicationCommandAutocompleteResult {
			t.Fatalf("expected autocomplete result type %d, got %d", types.CallbackTypeApplicationCommandAutocompleteResult, res.Type)
		}
		if res.Data == nil || len(res.Data.Choices) != 25 {
			t.Fatalf("expected choices to be capped at 25, got %d", len(res.Data.Choices))
		}
	default:
		t.Fatalf("expected response on channel")
	}
}
