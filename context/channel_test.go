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

	act := c.ChannelSuite()
	if act.ID() != types.Snowflake(200) {
		t.Fatalf("expected channel id 200, got %v", act.ID())
	}

	actTarget := c.ChannelSuite(types.Snowflake(300))
	if actTarget.ID() != types.Snowflake(300) {
		t.Fatalf("expected channel id 300, got %v", actTarget.ID())
	}
}

func TestChannelValidationErrors(t *testing.T) {
	cNoTarget := &Context{}
	actNoTarget := cNoTarget.ChannelSuite()

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
	actNoREST := cNoREST.ChannelSuite()

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

	if _, err := c.ChannelSuite().Create(rest.CreateChannelPayload{Name: ""}); !errors.Is(err, ErrChannelNameRequired) {
		t.Fatalf("expected ErrChannelNameRequired, got %v", err)
	}

	ch, err := c.ChannelSuite().Create(rest.CreateChannelPayload{Name: "general"}, "setup", "channels")
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

	ch, err := c.ChannelSuite().Get()
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

	err := c.ChannelSuite().Delete("cleanup", "room")
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

	if _, err := c.ChannelSuite().SetName(""); !errors.Is(err, ErrChannelNameRequired) {
		t.Fatalf("expected ErrChannelNameRequired, got %v", err)
	}

	ch, err := c.ChannelSuite().SetName("updated", "routine", "rename")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if ch.Name == nil || *ch.Name != "updated" {
		t.Fatalf("expected updated, got %v", ch.Name)
	}

	ch2, err := c.ChannelSuite().SetTopic("new topic", "audit", "reason")
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
		ch, err := c.ChannelSuite().SetSlowmode(tt.input)
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
	ch, err := c.ChannelSuite().Edit(rest.ModifyChannelPayload{
		NSFW: &isNSFW,
	}, "full", "update")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if !ch.NSFW {
		t.Fatalf("expected nsfw true")
	}
}

func TestManageChannel_Unified(t *testing.T) {
	var created, edited, deleted, locked, unlocked, purged bool
	var auditReason string

	c, cleanup := setupChannelTestContext(func(w http.ResponseWriter, r *http.Request) {
		auditReason = r.Header.Get("X-Audit-Log-Reason")
		switch {
		case r.Method == "POST" && r.URL.Path == "/guilds/100/channels":
			created = true
			_ = json.NewEncoder(w).Encode(types.Channel{ID: types.Snowflake(201), Name: strPtr("created-ch")})
		case r.Method == "PATCH" && r.URL.Path == "/channels/200":
			edited = true
			_ = json.NewEncoder(w).Encode(types.Channel{ID: types.Snowflake(200), Topic: strPtr("updated-topic")})
		case r.Method == "DELETE" && r.URL.Path == "/channels/200":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "PUT" && r.URL.Path == "/channels/200/permissions/100":
			locked = true
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "DELETE" && r.URL.Path == "/channels/200/permissions/100":
			unlocked = true
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "GET" && r.URL.Path == "/channels/200/messages":
			_ = json.NewEncoder(w).Encode([]types.Message{
				{ID: types.Snowflake(1001)},
				{ID: types.Snowflake(1002)},
			})
		case r.Method == "POST" && r.URL.Path == "/channels/200/messages/bulk-delete":
			purged = true
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer cleanup()

	err := c.ManageChannel(ChannelOptions{
		Type:   ChannelCreate,
		Name:   "created-ch",
		Reason: "create reason",
	})
	if err != nil || !created || auditReason != "create reason" {
		t.Fatalf("ChannelCreate failed: err=%v, created=%v", err, created)
	}

	err = c.ManageChannel(ChannelOptions{
		Type:   ChannelEdit,
		Topic:  "updated-topic",
		Reason: "edit reason",
	})
	if err != nil || !edited || auditReason != "edit reason" {
		t.Fatalf("ChannelEdit failed: err=%v, edited=%v", err, edited)
	}

	err = c.ManageChannel(ChannelOptions{
		Type:   ChannelLock,
		Locked: true,
	})
	if err != nil || !locked {
		t.Fatalf("ChannelLock (lock) failed: err=%v, locked=%v", err, locked)
	}

	err = c.ManageChannel(ChannelOptions{
		Type:   ChannelLock,
		Locked: false,
	})
	if err != nil || !unlocked {
		t.Fatalf("ChannelLock (unlock) failed: err=%v, unlocked=%v", err, unlocked)
	}

	err = c.ManageChannel(ChannelOptions{
		Type:       ChannelPurge,
		PurgeCount: 2,
		Reason:     "purge reason",
	})
	if err != nil || !purged || auditReason != "purge reason" {
		t.Fatalf("ChannelPurge failed: err=%v, purged=%v", err, purged)
	}

	err = c.ManageChannel(ChannelOptions{
		Type:       ChannelPurge,
		PurgeCount: 1,
	})
	if !errors.Is(err, ErrInvalidPurgeCount) {
		t.Fatalf("expected ErrInvalidPurgeCount, got %v", err)
	}

	err = c.ManageChannel(ChannelOptions{
		Type:   ChannelDelete,
		Reason: "delete reason",
	})
	if err != nil || !deleted || auditReason != "delete reason" {
		t.Fatalf("ChannelDelete failed: err=%v, deleted=%v", err, deleted)
	}

	err = c.ManageChannel(ChannelOptions{
		Type: ChannelAction(999),
	})
	if !errors.Is(err, ErrInvalidChannelAction) {
		t.Fatalf("expected ErrInvalidChannelAction, got %v", err)
	}
}

func TestManageChannel_OptionsAndShortcuts(t *testing.T) {
	var locked, deleted bool
	var auditReason string
	c, cleanup := setupChannelTestContext(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "PUT" && r.URL.Path == "/channels/200/permissions/100":
			locked = true
			auditReason = r.Header.Get("X-Audit-Log-Reason")
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "DELETE" && r.URL.Path == "/channels/200":
			deleted = true
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer cleanup()

	err := c.ManageChannel("lock", "emergency lockdown")
	if err != nil {
		t.Fatalf("unexpected lock shortcut error: %v", err)
	}
	if !locked || auditReason != "emergency lockdown" {
		t.Fatalf("expected locked with reason, got locked=%v reason=%s", locked, auditReason)
	}

	err = c.ManageChannel(Options{
		Type: "delete",
	})
	if err != nil {
		t.Fatalf("unexpected delete options error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected deleted true")
	}
}

func TestManageChannel_PurgeSingleMessage(t *testing.T) {
	var singleDeleted bool
	var auditReason string
	c, cleanup := setupChannelTestContext(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/channels/200/messages":
			_ = json.NewEncoder(w).Encode([]types.Message{
				{ID: types.Snowflake(5555)},
			})
		case r.Method == "DELETE" && r.URL.Path == "/channels/200/messages/5555":
			singleDeleted = true
			auditReason = r.Header.Get("X-Audit-Log-Reason")
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer cleanup()

	err := c.ManageChannel(Options{
		Type:   "purge",
		Count:  5,
		Reason: "purge single",
	})
	if err != nil {
		t.Fatalf("unexpected purge single error: %v", err)
	}
	if !singleDeleted {
		t.Fatal("expected single message 5555 to be deleted via DeleteMessage")
	}
	if auditReason != "purge single" {
		t.Fatalf("expected audit reason 'purge single', got '%s'", auditReason)
	}
}

