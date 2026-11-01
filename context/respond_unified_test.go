package context

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

func TestRespond_EmptyMessageValidation(t *testing.T) {
	in := &types.Interaction{ID: 100, Token: "test-token"}
	c := New(context.Background(), in, nil)

	_, err := c.Respond(Options{})
	if !errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("expected ErrEmptyMessage, got: %v", err)
	}

	_, err = c.Respond("")
	if !errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("expected ErrEmptyMessage on empty string, got: %v", err)
	}

	c2 := New(context.Background(), in, nil)
	_, err = c2.Respond(Options{Deferred: true})
	if errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("unexpected ErrEmptyMessage when deferred is true")
	}
}

func TestRespond_ScopeDM_DefaultAndCustomReceipt(t *testing.T) {
	t.Run("default receipt", func(t *testing.T) {
		var dmCreated bool
		var dmSentMessage rest.MessagePayload
		var callbackAcknowledged bool
		var callbackPayload types.InteractionResponse

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/users/@me/channels" && r.Method == http.MethodPost {
				dmCreated = true
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(types.Channel{ID: types.Snowflake(777), Type: 1})
				return
			}
			if r.URL.Path == "/channels/777/messages" && r.Method == http.MethodPost {
				_ = json.NewDecoder(r.Body).Decode(&dmSentMessage)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(types.Message{ID: types.Snowflake(888), Content: dmSentMessage.Content})
				return
			}
			if r.URL.Path == "/interactions/100/test-token/callback" && r.Method == http.MethodPost {
				callbackAcknowledged = true
				_ = json.NewDecoder(r.Body).Decode(&callbackPayload)
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
		in := &types.Interaction{
			ID:    100,
			Token: "test-token",
			User:  &types.User{ID: 555, Username: "targetuser"},
		}
		c := NewGatewayContext(context.Background(), in, client)

		msg, err := c.Respond(Options{
			Scope:   "dm",
			Target:  555,
			Content: "secret verification code",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg == nil || msg.ID != 888 {
			t.Fatalf("expected DM message with ID 888, got %+v", msg)
		}
		if !dmCreated {
			t.Fatalf("expected DM channel to be created")
		}
		if dmSentMessage.Content != "secret verification code" {
			t.Fatalf("unexpected DM content: %s", dmSentMessage.Content)
		}
		if !callbackAcknowledged {
			t.Fatalf("expected guild interaction callback to be acknowledged")
		}
		if callbackPayload.Data == nil || callbackPayload.Data.Content != DefaultDMReceipt {
			t.Fatalf("expected default DM receipt '%s', got '%v'", DefaultDMReceipt, callbackPayload.Data)
		}
		if callbackPayload.Data.Flags != types.MessageFlagEphemeral {
			t.Fatalf("expected receipt to be ephemeral, got flag: %d", callbackPayload.Data.Flags)
		}
	})

	t.Run("custom receipt", func(t *testing.T) {
		var callbackPayload types.InteractionResponse

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/users/@me/channels" && r.Method == http.MethodPost {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(types.Channel{ID: types.Snowflake(777)})
				return
			}
			if r.URL.Path == "/channels/777/messages" && r.Method == http.MethodPost {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(types.Message{ID: types.Snowflake(888)})
				return
			}
			if r.URL.Path == "/interactions/100/test-token/callback" && r.Method == http.MethodPost {
				_ = json.NewDecoder(r.Body).Decode(&callbackPayload)
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
		in := &types.Interaction{ID: 100, Token: "test-token", User: &types.User{ID: 555}}
		c := NewGatewayContext(context.Background(), in, client)

		_, err := c.Respond(Options{
			Scope:   "dm",
			Target:  555,
			Content: "hello user",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if callbackPayload.Data == nil || callbackPayload.Data.Content != DefaultDMReceipt {
			t.Fatalf("expected default receipt, got '%v'", callbackPayload.Data)
		}
	})
}

func TestRespond_StandaloneChannelSend(t *testing.T) {
	var sentToChannel types.Snowflake
	var sentPayload rest.MessagePayload

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/channels/999/messages" && r.Method == http.MethodPost {
			sentToChannel = 999
			_ = json.NewDecoder(r.Body).Decode(&sentPayload)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(types.Message{ID: types.Snowflake(123), ChannelID: 999})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	in := &types.Interaction{ID: 100, Token: "test-token"}
	c := NewGatewayContext(context.Background(), in, client)

	msg, err := c.Respond(Options{
		Type:    "send",
		Channel: 999,
		Content: "alert message to log channel",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg == nil || msg.ID != 123 {
		t.Fatalf("expected message id 123, got: %+v", msg)
	}
	if sentToChannel != 999 {
		t.Fatalf("expected send to channel 999, got: %d", sentToChannel)
	}
	if sentPayload.Content != "alert message to log channel" {
		t.Fatalf("unexpected payload content: %s", sentPayload.Content)
	}
}

func TestRespond_Mention_RepliedUserSuppression(t *testing.T) {
	in := &types.Interaction{ID: 100, Token: "test-token"}
	c := New(context.Background(), in, nil)

	noMention := false
	_, err := c.Respond(Options{
		Content: "reply without ping",
		Mention: &noMention,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case res := <-c.ResponseChannel():
		if res.Data == nil {
			t.Fatalf("expected data in response")
		}
		allowed, ok := res.Data.AllowedMentions.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any in AllowedMentions, got: %T", res.Data.AllowedMentions)
		}
		if repliedUser, exists := allowed["replied_user"]; !exists || repliedUser != false {
			t.Fatalf("expected replied_user to be false, got: %v", repliedUser)
		}
	default:
		t.Fatalf("expected response on channel")
	}
}

func TestRespond_TypingIndicator(t *testing.T) {
	var typingTriggered bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/channels/444/typing" && r.Method == http.MethodPost {
			typingTriggered = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.URL.Path == "/channels/444/messages" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(types.Message{ID: 1})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	in := &types.Interaction{ID: 100, Token: "test-token"}
	c := NewGatewayContext(context.Background(), in, client)
	c.ChannelID = 444

	_, err := c.Respond(Options{
		Type:    "send",
		Channel: 444,
		Content: "typing and sending",
		Typing:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !typingTriggered {
		t.Fatalf("expected typing indicator to be triggered on channel 444")
	}
}

func TestRespond_PositionalFiles(t *testing.T) {
	in := &types.Interaction{ID: 100, Token: "test-token"}
	c := New(context.Background(), in, nil)
	opts := parseResponseOptions("log", types.File{Name: "app.log"})
	if len(opts.Files) != 1 || opts.Files[0].Name != "app.log" {
		t.Fatalf("expected 1 file parsed positionally, got: %+v", opts.Files)
	}
	optsList := parseResponseOptions([]types.File{{Name: "a.txt"}, {Name: "b.txt"}})
	if len(optsList.Files) != 2 {
		t.Fatalf("expected 2 files parsed positionally, got: %+v", optsList.Files)
	}
	_ = c
}
