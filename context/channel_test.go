package context

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func setupChannelTestContext(handler http.HandlerFunc) (*Context, func()) {
	ts := httptest.NewServer(handler)
	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))

	guildID := types.Snowflake(100)
	channelID := types.Snowflake(200)

	c := &Context{
		ctx:       context.Background(),
		GuildID:   &guildID,
		ChannelID: channelID,
		REST:      client,
	}
	return c, ts.Close
}

func TestChannelDefaultsAndTargeting(t *testing.T) {
	c := &Context{
		ChannelID: types.Snowflake(200),
	}

	act := c.Channel()
	if act.ID() != types.Snowflake(200) {
		t.Fatalf("expected channel id 200, got %v", act.ID())
	}

	actTarget := c.Channel(types.Snowflake(300))
	if actTarget.ID() != types.Snowflake(300) {
		t.Fatalf("expected channel id 300, got %v", actTarget.ID())
	}

	actManage := c.ManageChannel(types.Snowflake(400))
	if actManage.ID() != types.Snowflake(400) {
		t.Fatalf("expected channel id 400, got %v", actManage.ID())
	}
}

func TestChannelValidationErrors(t *testing.T) {
	cNoTarget := &Context{}
	actNoTarget := cNoTarget.Channel()

	if _, err := actNoTarget.Get(); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}
	if err := actNoTarget.Delete(); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}
	if _, err := actNoTarget.SetName("test"); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}
	if _, err := actNoTarget.SetTopic("test"); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}
	if _, err := actNoTarget.SetSlowmode(10); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}
	if _, err := actNoTarget.Edit(rest.ModifyChannelPayload{}); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}

	cNoREST := &Context{ChannelID: types.Snowflake(200)}
	actNoREST := cNoREST.Channel()

	if _, err := actNoREST.Get(); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}
	if err := actNoREST.Delete(); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}
	if _, err := actNoREST.SetName("test"); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}
	if _, err := actNoREST.SetTopic("test"); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}
	if _, err := actNoREST.SetSlowmode(10); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}
	if _, err := actNoREST.Edit(rest.ModifyChannelPayload{}); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}
	if _, err := actNoREST.Create(rest.CreateChannelPayload{Name: "test"}); !errors.Is(err, ErrNotInGuild) {
		t.Fatalf("expected ErrNotInGuild, got %v", err)
	}
}

func TestChannelCreate(t *testing.T) {
	c, cleanup := setupChannelTestContext(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/guilds/100/channels" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("X-Audit-Log-Reason") != "setup channels" {
			t.Fatalf("unexpected audit reason %s", r.Header.Get("X-Audit-Log-Reason"))
		}
		var payload rest.CreateChannelPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}
		if payload.Name != "general" {
			t.Fatalf("expected name general, got %s", payload.Name)
		}

		_ = json.NewEncoder(w).Encode(types.Channel{
			ID:   types.Snowflake(555),
			Name: strPtr("general"),
		})
	})
	defer cleanup()

	if _, err := c.Channel().Create(rest.CreateChannelPayload{Name: ""}); !errors.Is(err, ErrChannelNameRequired) {
		t.Fatalf("expected ErrChannelNameRequired, got %v", err)
	}

	ch, err := c.Channel().Create(rest.CreateChannelPayload{Name: "general"}, "setup", "channels")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if ch.ID != types.Snowflake(555) {
		t.Fatalf("expected channel ID 555, got %v", ch.ID)
	}
}

func TestChannelGet(t *testing.T) {
	c, cleanup := setupChannelTestContext(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/channels/200" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(types.Channel{
			ID:   types.Snowflake(200),
			Name: strPtr("general"),
		})
	})
	defer cleanup()

	ch, err := c.Channel().Get()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if ch.Name == nil || *ch.Name != "general" {
		t.Fatalf("expected general, got %v", ch.Name)
	}
}

func TestChannelDelete(t *testing.T) {
	c, cleanup := setupChannelTestContext(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/channels/200" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("X-Audit-Log-Reason") != "cleanup room" {
			t.Fatalf("unexpected audit header %s", r.Header.Get("X-Audit-Log-Reason"))
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := c.Channel().Delete("cleanup", "room")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestChannelSetNameAndTopic(t *testing.T) {
	c, cleanup := setupChannelTestContext(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/channels/200" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var payload rest.ModifyChannelPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		_ = json.NewEncoder(w).Encode(types.Channel{
			ID:    types.Snowflake(200),
			Name:  strPtr("updated"),
			Topic: strPtr("new topic"),
		})
	})
	defer cleanup()

	if _, err := c.Channel().SetName(""); !errors.Is(err, ErrChannelNameRequired) {
		t.Fatalf("expected ErrChannelNameRequired, got %v", err)
	}

	ch, err := c.Channel().SetName("updated", "routine", "rename")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if ch.Name == nil || *ch.Name != "updated" {
		t.Fatalf("expected updated, got %v", ch.Name)
	}

	ch2, err := c.Channel().SetTopic("new topic", "audit", "reason")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if ch2.Topic == nil || *ch2.Topic != "new topic" {
		t.Fatalf("expected new topic, got %v", ch2.Topic)
	}
}

func TestChannelSetSlowmode(t *testing.T) {
	var lastSlowmode int
	c, cleanup := setupChannelTestContext(func(w http.ResponseWriter, r *http.Request) {
		var payload rest.ModifyChannelPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if payload.RateLimitPerUser != nil {
			lastSlowmode = *payload.RateLimitPerUser
		}
		_ = json.NewEncoder(w).Encode(types.Channel{
			ID:               types.Snowflake(200),
			RateLimitPerUser: intPtr(lastSlowmode),
		})
	})
	defer cleanup()

	tests := []struct {
		input       any
		expectedSec int
		expectedErr error
	}{
		{input: 15, expectedSec: 15},
		{input: int64(30), expectedSec: 30},
		{input: 5 * time.Minute, expectedSec: 300},
		{input: "10m", expectedSec: 600},
		{input: "0", expectedSec: 0},
		{input: nil, expectedSec: 0},
		{input: -5, expectedErr: ErrSlowmodeNegative},
		{input: 25000, expectedErr: ErrSlowmodeTooLong},
		{input: "7h", expectedErr: ErrSlowmodeTooLong},
		{input: struct{}{}, expectedErr: ErrUnsupportedType},
	}

	for _, tt := range tests {
		ch, err := c.Channel().SetSlowmode(tt.input)
		if tt.expectedErr != nil {
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("input %v expected err %v, got %v", tt.input, tt.expectedErr, err)
			}
		} else {
			if err != nil {
				t.Fatalf("input %v unexpected err %v", tt.input, err)
			}
			if ch.RateLimitPerUser == nil || *ch.RateLimitPerUser != tt.expectedSec {
				t.Fatalf("input %v expected %d, got %v", tt.input, tt.expectedSec, ch.RateLimitPerUser)
			}
		}
	}
}

func TestChannelEdit(t *testing.T) {
	c, cleanup := setupChannelTestContext(func(w http.ResponseWriter, r *http.Request) {
		var payload rest.ModifyChannelPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if r.Header.Get("X-Audit-Log-Reason") != "full update" {
			t.Fatalf("unexpected audit header %s", r.Header.Get("X-Audit-Log-Reason"))
		}
		nsfw := true
		if payload.NSFW != nil {
			nsfw = *payload.NSFW
		}
		_ = json.NewEncoder(w).Encode(types.Channel{
			ID:   types.Snowflake(200),
			NSFW: nsfw,
		})
	})
	defer cleanup()

	isNSFW := true
	ch, err := c.Channel().Edit(rest.ModifyChannelPayload{
		NSFW: &isNSFW,
	}, "full", "update")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if !ch.NSFW {
		t.Fatalf("expected nsfw true")
	}
}

