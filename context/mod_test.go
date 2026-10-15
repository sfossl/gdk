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

func setupModTestContext(handler http.HandlerFunc) (*Context, func()) {
	ts := httptest.NewServer(handler)
	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))

	guildID := types.Snowflake(100)
	botAppID := types.Snowflake(1)
	invokerID := types.Snowflake(2)

	c := &Context{
		ctx:     context.Background(),
		GuildID: &guildID,
		AppID:   botAppID,
		REST:    client,
		Author:  &types.User{ID: invokerID},
		Member:  &types.Member{Roles: []types.Snowflake{types.Snowflake(30)}},
	}
	return c, ts.Close
}

func standardGuildHandler(extraHandler func(w http.ResponseWriter, r *http.Request) bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if extraHandler != nil && extraHandler(w, r) {
			return
		}
		switch r.URL.Path {
		case "/guilds/100":
			_ = json.NewEncoder(w).Encode(types.Guild{
				ID:      types.Snowflake(100),
				OwnerID: types.Snowflake(999),
				Roles: []types.Role{
					{ID: types.Snowflake(10), Position: 1},
					{ID: types.Snowflake(20), Position: 5},
					{ID: types.Snowflake(30), Position: 10},
				},
			})
		case "/guilds/100/members/1":
			_ = json.NewEncoder(w).Encode(types.Member{Roles: []types.Snowflake{types.Snowflake(30)}})
		case "/guilds/100/members/2":
			_ = json.NewEncoder(w).Encode(types.Member{Roles: []types.Snowflake{types.Snowflake(20)}})
		case "/guilds/100/members/3":
			_ = json.NewEncoder(w).Encode(types.Member{Roles: []types.Snowflake{types.Snowflake(10)}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func TestMod_Kick(t *testing.T) {
	kicked := false
	var auditReason string

	c, cleanup := setupModTestContext(standardGuildHandler(func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == "DELETE" && r.URL.Path == "/guilds/100/members/3" {
			kicked = true
			auditReason = r.Header.Get("X-Audit-Log-Reason")
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		return false
	}))
	defer cleanup()

	err := c.Mod(types.Snowflake(3)).Kick("breaking rules")
	if err != nil {
		t.Fatalf("unexpected kick error: %v", err)
	}
	if !kicked || auditReason != "breaking rules" {
		t.Fatalf("expected kick with reason, got kicked=%v, reason=%s", kicked, auditReason)
	}

	err = c.Mod(types.Snowflake(999)).Kick("cannot kick owner")
	if !errors.Is(err, ErrCannotModerateOwner) {
		t.Fatalf("expected ErrCannotModerateOwner, got: %v", err)
	}
}

func TestMod_Ban(t *testing.T) {
	var deleteSeconds int
	var auditReason string

	c, cleanup := setupModTestContext(standardGuildHandler(func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == "PUT" && r.URL.Path == "/guilds/100/bans/3" {
			auditReason = r.Header.Get("X-Audit-Log-Reason")
			var body struct {
				DeleteMessageSeconds int `json:"delete_message_seconds"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			deleteSeconds = body.DeleteMessageSeconds
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		return false
	}))
	defer cleanup()

	err := c.Mod(types.Snowflake(3)).Ban("1d", "spam raid")
	if err != nil {
		t.Fatalf("unexpected ban error: %v", err)
	}
	if deleteSeconds != 86400 || auditReason != "spam raid" {
		t.Fatalf("expected 86400s delete, got %d, reason %s", deleteSeconds, auditReason)
	}

	err = c.Mod(types.Snowflake(3)).Ban(3600)
	if err != nil {
		t.Fatalf("unexpected int ban error: %v", err)
	}
	if deleteSeconds != 3600 {
		t.Fatalf("expected 3600s delete, got %d", deleteSeconds)
	}

	err = c.Mod(types.Snowflake(3)).Ban("8d")
	if !errors.Is(err, ErrBanDeleteTooLong) {
		t.Fatalf("expected ErrBanDeleteTooLong, got: %v", err)
	}
}

func TestMod_Unban(t *testing.T) {
	unbanned := false
	var auditReason string

	c, cleanup := setupModTestContext(standardGuildHandler(func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == "DELETE" && r.URL.Path == "/guilds/100/bans/444" {
			unbanned = true
			auditReason = r.Header.Get("X-Audit-Log-Reason")
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		return false
	}))
	defer cleanup()

	err := c.Mod(types.Snowflake(444)).Unban("appealed")
	if err != nil {
		t.Fatalf("unexpected unban error: %v", err)
	}
	if !unbanned || auditReason != "appealed" {
		t.Fatalf("expected unban with reason, got unbanned=%v, reason=%s", unbanned, auditReason)
	}

	err = c.Mod(types.Snowflake(0)).Unban()
	if !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("expected ErrInvalidTarget, got: %v", err)
	}
}

func TestMod_Timeout(t *testing.T) {
	var disabledUntil *time.Time
	var auditReason string

	c, cleanup := setupModTestContext(standardGuildHandler(func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == "PATCH" && r.URL.Path == "/guilds/100/members/3" {
			auditReason = r.Header.Get("X-Audit-Log-Reason")
			var body struct {
				CommunicationDisabledUntil *time.Time `json:"communication_disabled_until"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			disabledUntil = body.CommunicationDisabledUntil
			w.WriteHeader(http.StatusOK)
			return true
		}
		return false
	}))
	defer cleanup()

	err := c.Mod(types.Snowflake(3)).Timeout("15m", "cooldown")
	if err != nil {
		t.Fatalf("unexpected timeout error: %v", err)
	}
	if disabledUntil == nil || auditReason != "cooldown" {
		t.Fatalf("expected timeout until set with reason, got %v, reason %s", disabledUntil, auditReason)
	}

	err = c.Mod(types.Snowflake(3)).Timeout(30 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected duration timeout error: %v", err)
	}

	err = c.Mod(types.Snowflake(3)).Timeout("30d")
	if !errors.Is(err, ErrTimeoutTooLong) {
		t.Fatalf("expected ErrTimeoutTooLong, got: %v", err)
	}

	err = c.Mod(types.Snowflake(3)).Timeout(500 * time.Millisecond)
	if !errors.Is(err, ErrTimeoutTooShort) {
		t.Fatalf("expected ErrTimeoutTooShort, got: %v", err)
	}

	err = c.Mod(types.Snowflake(3)).RemoveTimeout("forgiven")
	if err != nil {
		t.Fatalf("unexpected remove timeout error: %v", err)
	}
	if disabledUntil != nil || auditReason != "forgiven" {
		t.Fatalf("expected nil disabledUntil, got: %v", disabledUntil)
	}
}
