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
	var kicked bool
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

	err := c.Mod(ModOptions{
		Type:   ModKick,
		Target: types.Snowflake(3),
		Reason: "breaking rules",
	})
	if err != nil {
		t.Fatalf("unexpected kick error: %v", err)
	}
	if !kicked || auditReason != "breaking rules" {
		t.Fatalf("expected kick with reason, got kicked=%v, reason=%s", kicked, auditReason)
	}

	err = c.Mod(ModOptions{
		Type:   ModKick,
		Target: types.Snowflake(999),
		Reason: "cannot kick owner",
	})
	if !errors.Is(err, ErrCannotModerateOwner) {
		t.Fatalf("expected ErrCannotModerateOwner, got: %v", err)
	}

	err = c.ModAction(types.Snowflake(3)).Kick("helper kick")
	if err != nil {
		t.Fatalf("unexpected mod action kick error: %v", err)
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

	err := c.Mod(ModOptions{
		Type:           ModBan,
		Target:         types.Snowflake(3),
		DeleteMessages: "1d",
		Reason:         "spam raid",
	})
	if err != nil {
		t.Fatalf("unexpected ban error: %v", err)
	}
	if deleteSeconds != 86400 || auditReason != "spam raid" {
		t.Fatalf("expected 86400s delete, got %d, reason %s", deleteSeconds, auditReason)
	}

	err = c.Mod(ModOptions{
		Type:           ModBan,
		Target:         types.Snowflake(3),
		DeleteMessages: 3600,
	})
	if err != nil {
		t.Fatalf("unexpected int ban error: %v", err)
	}
	if deleteSeconds != 3600 {
		t.Fatalf("expected 3600s delete, got %d", deleteSeconds)
	}

	err = c.Mod(ModOptions{
		Type:           ModBan,
		Target:         types.Snowflake(3),
		DeleteMessages: "8d",
	})
	if !errors.Is(err, ErrBanDeleteTooLong) {
		t.Fatalf("expected ErrBanDeleteTooLong, got: %v", err)
	}

	err = c.ModAction(types.Snowflake(3)).Ban("1h", "via modaction")
	if err != nil {
		t.Fatalf("unexpected modaction ban error: %v", err)
	}
}

func TestMod_Unban(t *testing.T) {
	var unbanned bool
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

	err := c.Mod(ModOptions{
		Type:   ModUnban,
		Target: types.Snowflake(444),
		Reason: "appealed",
	})
	if err != nil {
		t.Fatalf("unexpected unban error: %v", err)
	}
	if !unbanned || auditReason != "appealed" {
		t.Fatalf("expected unban with reason, got unbanned=%v, reason=%s", unbanned, auditReason)
	}

	err = c.Mod(ModOptions{
		Type:   ModUnban,
		Target: types.Snowflake(0),
	})
	if !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("expected ErrInvalidTarget, got: %v", err)
	}

	err = c.ModAction(types.Snowflake(444)).Unban("appealed action")
	if err != nil {
		t.Fatalf("unexpected modaction unban error: %v", err)
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

	err := c.Mod(ModOptions{
		Type:     ModTimeout,
		Target:   types.Snowflake(3),
		Duration: "15m",
		Reason:   "cooldown",
	})
	if err != nil {
		t.Fatalf("unexpected timeout error: %v", err)
	}
	if disabledUntil == nil || auditReason != "cooldown" {
		t.Fatalf("expected timeout until set with reason, got %v, reason %s", disabledUntil, auditReason)
	}

	err = c.Mod(ModOptions{
		Type:     ModTimeout,
		Target:   types.Snowflake(3),
		Duration: 30 * time.Minute,
	})
	if err != nil {
		t.Fatalf("unexpected duration timeout error: %v", err)
	}

	err = c.Mod(ModOptions{
		Type:     ModTimeout,
		Target:   types.Snowflake(3),
		Duration: "30d",
	})
	if !errors.Is(err, ErrTimeoutTooLong) {
		t.Fatalf("expected ErrTimeoutTooLong, got: %v", err)
	}

	err = c.Mod(ModOptions{
		Type:     ModTimeout,
		Target:   types.Snowflake(3),
		Duration: 500 * time.Millisecond,
	})
	if !errors.Is(err, ErrTimeoutTooShort) {
		t.Fatalf("expected ErrTimeoutTooShort, got: %v", err)
	}

	err = c.Mod(ModOptions{
		Type:   ModRemoveTimeout,
		Target: types.Snowflake(3),
		Reason: "forgiven",
	})
	if err != nil {
		t.Fatalf("unexpected remove timeout error: %v", err)
	}
	if disabledUntil != nil || auditReason != "forgiven" {
		t.Fatalf("expected nil disabledUntil, got: %v", disabledUntil)
	}

	err = c.ModAction(types.Snowflake(3)).Timeout(10*time.Minute, "action timeout")
	if err != nil {
		t.Fatalf("unexpected modaction timeout error: %v", err)
	}

	err = c.ModAction(types.Snowflake(3)).RemoveTimeout("action remove")
	if err != nil {
		t.Fatalf("unexpected modaction remove timeout error: %v", err)
	}

	err = c.Mod(ModOptions{
		Type:   ModType(999),
		Target: types.Snowflake(3),
	})
	if !errors.Is(err, ErrInvalidModType) {
		t.Fatalf("expected ErrInvalidModType, got: %v", err)
	}
}

func TestMod_UniversalOptionsAndShortcuts(t *testing.T) {
	var lastMethod, lastPath, lastReason string
	c, cleanup := setupModTestContext(standardGuildHandler(func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == "DELETE" && r.URL.Path == "/guilds/100/members/3" {
			lastMethod = r.Method
			lastPath = r.URL.Path
			lastReason = r.Header.Get("X-Audit-Log-Reason")
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		if r.Method == "PUT" && r.URL.Path == "/guilds/100/bans/3" {
			lastMethod = r.Method
			lastPath = r.URL.Path
			lastReason = r.Header.Get("X-Audit-Log-Reason")
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		return false
	}))
	defer cleanup()

	target := types.Snowflake(3)
	c.Interaction = &types.Interaction{
		Data: &types.InteractionData{
			TargetID: &target,
		},
	}

	err := c.Mod("kick", "spamming")
	if err != nil {
		t.Fatalf("unexpected kick shortcut error: %v", err)
	}
	if lastMethod != "DELETE" || lastPath != "/guilds/100/members/3" || lastReason != "spamming" {
		t.Fatalf("unexpected kick call: %s %s, reason: %s", lastMethod, lastPath, lastReason)
	}

	err = c.Mod(Options{
		Type:   "kick",
		Reason: "universal options kick",
	})
	if err != nil {
		t.Fatalf("unexpected universal options kick error: %v", err)
	}
	if lastMethod != "DELETE" || lastPath != "/guilds/100/members/3" || lastReason != "universal options kick" {
		t.Fatalf("unexpected kick options call: %s %s, reason: %s", lastMethod, lastPath, lastReason)
	}

	err = c.Mod("ban", "7d", "raiding")
	if err != nil {
		t.Fatalf("unexpected ban shortcut error: %v", err)
	}
	if lastMethod != "PUT" || lastPath != "/guilds/100/bans/3" || lastReason != "raiding" {
		t.Fatalf("unexpected ban call: %s %s, reason: %s", lastMethod, lastPath, lastReason)
	}

	err = c.Mod(Options{
		Type:     "ban",
		Duration: "7d",
		Reason:   "universal options ban",
	})
	if err != nil {
		t.Fatalf("unexpected universal options ban error: %v", err)
	}
	if lastMethod != "PUT" || lastPath != "/guilds/100/bans/3" || lastReason != "universal options ban" {
		t.Fatalf("unexpected ban options call: %s %s, reason: %s", lastMethod, lastPath, lastReason)
	}
}
