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

func TestManageChannel_Tier1_CreateWithName(t *testing.T) {
	var receivedName string
	var receivedReason string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedReason = r.Header.Get("X-Audit-Log-Reason")
		var payload rest.CreateChannelPayload
		_ = json.NewDecoder(r.Body).Decode(&payload)
		receivedName = payload.Name
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.Channel{ID: 999, Name: &payload.Name})
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	guildID := types.Snowflake(100)
	c := &Context{
		ctx:       context.Background(),
		GuildID:   &guildID,
		ChannelID: 200,
		REST:      client,
	}

	err := c.ManageChannel("create", "general")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedName != "general" {
		t.Fatalf("expected channel name 'general', got '%s'", receivedName)
	}

	err = c.ManageChannel("create", "general-two", "audit reason")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedName != "general-two" {
		t.Fatalf("expected channel name 'general-two', got '%s'", receivedName)
	}
	if receivedReason != "audit reason" {
		t.Fatalf("expected reason 'audit reason', got '%s'", receivedReason)
	}

	err = c.ManageChannel(ChannelCreate, "announcements")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedName != "announcements" {
		t.Fatalf("expected channel name 'announcements', got '%s'", receivedName)
	}
}

func TestManageChannel_Tier1_LockUnlock(t *testing.T) {
	var lastMethod string
	var lastReason string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastReason = r.Header.Get("X-Audit-Log-Reason")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	guildID := types.Snowflake(100)
	c := &Context{
		ctx:       context.Background(),
		GuildID:   &guildID,
		ChannelID: 200,
		REST:      client,
	}

	err := c.ManageChannel("lock", "lockdown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastMethod != http.MethodPut {
		t.Fatalf("expected PUT method for lock, got %s", lastMethod)
	}
	if lastReason != "lockdown" {
		t.Fatalf("expected reason lockdown, got %s", lastReason)
	}

	err = c.ManageChannel("unlock")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastMethod != http.MethodDelete {
		t.Fatalf("expected DELETE method for unlock, got %s", lastMethod)
	}
}
