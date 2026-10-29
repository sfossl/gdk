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

func TestUpdateMessage_InitialWebhook(t *testing.T) {
	in := &types.Interaction{
		ID:    types.Snowflake(100),
		Token: "test-token",
	}
	c := New(context.Background(), in, nil)

	msg, err := c.UpdateMessage(Response{
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

func TestUpdateMessage_InitialGateway(t *testing.T) {
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

	msg, err := c.UpdateMessage(Response{
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

func TestUpdateMessage_DeferUpdateAndEditOriginal(t *testing.T) {
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

	msg, err := c.UpdateMessage(Response{
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

	secondEdit, err := c.UpdateMessage(Response{
		Content: "third update in-place",
	})
	if err != nil {
		t.Fatalf("unexpected second update error: %v", err)
	}
	if secondEdit == nil || secondEdit.Content != "third update in-place" {
		t.Fatalf("unexpected second edit message: %+v", secondEdit)
	}
}

func TestUpdateValidation(t *testing.T) {
	cNoGuild := &Context{}
	if err := cNoGuild.Update(UpdateOptions{Target: 123}); !errors.Is(err, ErrNotInGuild) {
		t.Fatalf("expected ErrNotInGuild, got %v", err)
	}

	gID := types.Snowflake(10)
	cNoREST := &Context{GuildID: &gID}
	if err := cNoREST.Update(UpdateOptions{Target: 123}); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()
	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	c := &Context{GuildID: &gID, REST: client}

	if err := c.Update(UpdateOptions{Target: 0}); !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("expected ErrInvalidTarget, got %v", err)
	}

	nick := "test"
	if err := c.Update(UpdateOptions{Target: 123, Nickname: &nick}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.Update(UpdateOptions{Target: 123}); err != nil {
		t.Fatalf("expected no-op when no options set, got: %v", err)
	}
}

func TestUpdateOperations(t *testing.T) {
	gID := types.Snowflake(555)
	var lastPayload rest.ModifyMemberPayload
	var lastAuditReason string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/555/members/777":
			lastAuditReason = r.Header.Get("X-Audit-Log-Reason")
			_ = json.NewDecoder(r.Body).Decode(&lastPayload)
			_ = json.NewEncoder(w).Encode(types.Member{})
			return
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/555/members/777":
			_ = json.NewEncoder(w).Encode(types.Member{
				Roles: []types.Snowflake{1, 2},
			})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	targetID := types.Snowflake(777)
	c := &Context{
		ctx:     context.Background(),
		GuildID: &gID,
		REST:    client,
		Interaction: &types.Interaction{
			Data: &types.InteractionData{
				TargetID: &targetID,
			},
		},
	}

	nick := "NewNick"
	err := c.Update(UpdateOptions{
		Nickname: &nick,
		Reason:   "Nickname update",
	})
	if err != nil {
		t.Fatalf("unexpected error updating nickname: %v", err)
	}
	if lastPayload.Nick == nil || *lastPayload.Nick != "NewNick" {
		t.Fatalf("expected nickname NewNick, got %+v", lastPayload.Nick)
	}
	if lastAuditReason != "Nickname update" {
		t.Fatalf("expected audit reason 'Nickname update', got %s", lastAuditReason)
	}

	err = c.Update(UpdateOptions{
		Target:      777,
		AddRoles:    []types.Snowflake{3},
		RemoveRoles: []types.Snowflake{2},
	})
	if err != nil {
		t.Fatalf("unexpected error updating roles: %v", err)
	}
	if lastPayload.Roles == nil {
		t.Fatalf("expected roles payload to be populated")
	}
	roles := *lastPayload.Roles
	hasRole1, hasRole3, hasRole2 := false, false, false
	for _, r := range roles {
		if r == 1 {
			hasRole1 = true
		}
		if r == 3 {
			hasRole3 = true
		}
		if r == 2 {
			hasRole2 = true
		}
	}
	if !hasRole1 || !hasRole3 || hasRole2 {
		t.Fatalf("expected roles [1, 3] and not 2, got %+v", roles)
	}

	mute := true
	deaf := false
	vCh := types.Snowflake(999)
	err = c.Update(UpdateOptions{
		Target:       777,
		VoiceMute:    &mute,
		VoiceDeaf:    &deaf,
		VoiceChannel: &vCh,
	})
	if err != nil {
		t.Fatalf("unexpected error updating voice state: %v", err)
	}
	if lastPayload.Mute == nil || *lastPayload.Mute != true {
		t.Fatalf("expected voice mute true")
	}
	if lastPayload.Deaf == nil || *lastPayload.Deaf != false {
		t.Fatalf("expected voice deaf false")
	}
	if lastPayload.ChannelID == nil || *lastPayload.ChannelID != 999 {
		t.Fatalf("expected voice channel 999")
	}
}

func TestUpdate_OptionsAndShortcuts(t *testing.T) {
	gID := types.Snowflake(555)
	var lastPayload rest.ModifyMemberPayload

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch && r.URL.Path == "/guilds/555/members/777" {
			_ = json.NewDecoder(r.Body).Decode(&lastPayload)
			_ = json.NewEncoder(w).Encode(types.Member{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	targetID := types.Snowflake(777)
	c := &Context{
		ctx:     context.Background(),
		GuildID: &gID,
		REST:    client,
		Interaction: &types.Interaction{
			Data: &types.InteractionData{
				TargetID: &targetID,
			},
		},
	}

	err := c.Update(targetID, "ShortcutNick")
	if err != nil {
		t.Fatalf("unexpected update shortcut error: %v", err)
	}
	if lastPayload.Nick == nil || *lastPayload.Nick != "ShortcutNick" {
		t.Fatalf("expected ShortcutNick, got %+v", lastPayload.Nick)
	}

	nick := "OptionsNick"
	err = c.Update(Options{
		Nickname: &nick,
	})
	if err != nil {
		t.Fatalf("unexpected update options error: %v", err)
	}
	if lastPayload.Nick == nil || *lastPayload.Nick != "OptionsNick" {
		t.Fatalf("expected OptionsNick, got %+v", lastPayload.Nick)
	}
}
