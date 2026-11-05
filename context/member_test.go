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

func TestMember_UnifiedOptions(t *testing.T) {
	var lastMethod, lastPath, lastReason string
	var lastPayload map[string]any

	guildID := types.Snowflake(100)
	targetID := types.Snowflake(200)
	botID := types.Snowflake(999)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path
		lastReason = r.Header.Get("X-Audit-Log-Reason")
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&lastPayload)
		}

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/100":
			_ = json.NewEncoder(w).Encode(types.Guild{
				ID:      guildID,
				OwnerID: 555,
				Roles: []types.Role{
					{ID: 10, Position: 10},
					{ID: 20, Position: 20},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/100/members/999":
			_ = json.NewEncoder(w).Encode(types.Member{
				Roles: []types.Snowflake{20},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/100/members/200":
			_ = json.NewEncoder(w).Encode(types.Member{
				Roles: []types.Snowflake{10},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/guilds/100/bans/200":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/100/bans/200":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/100/members/200":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/100/members/200":
			_ = json.NewEncoder(w).Encode(types.Member{})
		case r.Method == http.MethodPut && r.URL.Path == "/guilds/100/members/200/roles/10":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/100/members/200/roles/10":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	c := &Context{
		ctx:     context.Background(),
		REST:    client,
		GuildID: &guildID,
		AppID:   botID,
		Author:  &types.User{ID: 555},
	}

	err := c.Member(Options{
		Target:   targetID,
		Type:     "ban",
		Reason:   "bad behavior",
		Duration: "1d",
	})
	if err != nil {
		t.Fatalf("unexpected ban error: %v", err)
	}
	if lastMethod != http.MethodPut || lastPath != "/guilds/100/bans/200" || lastReason != "bad behavior" {
		t.Fatalf("unexpected ban request: %s %s reason: %s", lastMethod, lastPath, lastReason)
	}

	err = c.Member(Options{
		Target: targetID,
		Type:   "unban",
		Reason: "appealed",
	})
	if err != nil {
		t.Fatalf("unexpected unban error: %v", err)
	}
	if lastMethod != http.MethodDelete || lastPath != "/guilds/100/bans/200" || lastReason != "appealed" {
		t.Fatalf("unexpected unban request: %s %s reason: %s", lastMethod, lastPath, lastReason)
	}

	err = c.Member(Options{
		Target:   targetID,
		Type:     "timeout",
		Duration: 10 * time.Minute,
		Reason:   "cooling off",
	})
	if err != nil {
		t.Fatalf("unexpected timeout error: %v", err)
	}
	if lastMethod != http.MethodPatch || lastPath != "/guilds/100/members/200" {
		t.Fatalf("unexpected timeout request: %s %s", lastMethod, lastPath)
	}

	err = c.Member(Options{
		Target: targetID,
		Type:   "kick",
		Reason: "inactivity",
	})
	if err != nil {
		t.Fatalf("unexpected kick error: %v", err)
	}
	if lastMethod != http.MethodDelete || lastPath != "/guilds/100/members/200" || lastReason != "inactivity" {
		t.Fatalf("unexpected kick request: %s %s reason: %s", lastMethod, lastPath, lastReason)
	}

	newNick := "CleanNick"
	err = c.Member(Options{
		Target:   targetID,
		Nickname: &newNick,
		Reason:   "name fix",
	})
	if err != nil {
		t.Fatalf("unexpected nick error: %v", err)
	}
	if lastMethod != http.MethodPatch || lastReason != "name fix" {
		t.Fatalf("unexpected nick request: %s %s", lastMethod, lastReason)
	}

	err = c.Member(Options{
		Target: targetID,
		Type:   "addrole",
		Role:   10,
	})
	if err != nil {
		t.Fatalf("unexpected addrole error: %v", err)
	}
	if lastMethod != http.MethodPut || lastPath != "/guilds/100/members/200/roles/10" {
		t.Fatalf("unexpected addrole request: %s %s", lastMethod, lastPath)
	}

	err = c.Member(Options{
		Target: targetID,
		Type:   "removerole",
		Role:   10,
	})
	if err != nil {
		t.Fatalf("unexpected removerole error: %v", err)
	}
	if lastMethod != http.MethodDelete || lastPath != "/guilds/100/members/200/roles/10" {
		t.Fatalf("unexpected removerole request: %s %s", lastMethod, lastPath)
	}
}

func TestMember_PositionalOneLiners(t *testing.T) {
	var lastMethod, lastPath string
	guildID := types.Snowflake(100)
	targetID := types.Snowflake(200)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/100":
			_ = json.NewEncoder(w).Encode(types.Guild{
				ID:      guildID,
				OwnerID: 555,
				Roles:   []types.Role{{ID: 10, Position: 10}, {ID: 20, Position: 20}},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/100/members/999":
			_ = json.NewEncoder(w).Encode(types.Member{Roles: []types.Snowflake{20}})
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/100/members/200":
			_ = json.NewEncoder(w).Encode(types.Member{Roles: []types.Snowflake{10}})
		case r.Method == http.MethodPut && r.URL.Path == "/guilds/100/bans/200":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/100/members/200":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/100/bans/200":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/100/members/200":
			_ = json.NewEncoder(w).Encode(types.Member{})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	c := &Context{
		ctx:     context.Background(),
		REST:    client,
		GuildID: &guildID,
		AppID:   999,
		Author:  &types.User{ID: 555},
	}

	if err := c.Member(targetID, "ban", "rule 1", "1d"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastMethod != http.MethodPut || lastPath != "/guilds/100/bans/200" {
		t.Fatalf("unexpected ban call: %s %s", lastMethod, lastPath)
	}

	if err := c.Member(targetID, "kick", "afk"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastMethod != http.MethodDelete || lastPath != "/guilds/100/members/200" {
		t.Fatalf("unexpected kick call: %s %s", lastMethod, lastPath)
	}

	if err := c.Member(targetID, "unban"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastMethod != http.MethodDelete || lastPath != "/guilds/100/bans/200" {
		t.Fatalf("unexpected unban call: %s %s", lastMethod, lastPath)
	}

	if err := c.Member(targetID, "timeout", 15*time.Minute, "chill"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastMethod != http.MethodPatch || lastPath != "/guilds/100/members/200" {
		t.Fatalf("unexpected timeout call: %s %s", lastMethod, lastPath)
	}
}

func TestMember_SuiteMethods(t *testing.T) {
	var lastMethod, lastPath string
	guildID := types.Snowflake(100)
	targetID := types.Snowflake(200)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/100":
			_ = json.NewEncoder(w).Encode(types.Guild{
				ID:      guildID,
				OwnerID: 555,
				Roles:   []types.Role{{ID: 10, Position: 10}, {ID: 20, Position: 20}},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/100/members/999":
			_ = json.NewEncoder(w).Encode(types.Member{Roles: []types.Snowflake{20}})
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/100/members/200":
			_ = json.NewEncoder(w).Encode(types.Member{User: &types.User{ID: targetID}, Roles: []types.Snowflake{10}})
		case r.Method == http.MethodPut && r.URL.Path == "/guilds/100/bans/200":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/100/members/200":
			_ = json.NewEncoder(w).Encode(types.Member{})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	c := &Context{
		ctx:     context.Background(),
		REST:    client,
		GuildID: &guildID,
		AppID:   999,
		Author:  &types.User{ID: 555},
	}

	suite := c.MemberSuite(targetID)
	if suite.ID() != targetID {
		t.Fatalf("expected suite target ID %d, got %d", targetID, suite.ID())
	}

	mem, err := suite.Get()
	if err != nil {
		t.Fatalf("unexpected Get error: %v", err)
	}
	if mem.User.ID != targetID {
		t.Fatalf("expected member user ID %d, got %d", targetID, mem.User.ID)
	}

	if err := suite.Ban("1d", "spam suite"); err != nil {
		t.Fatalf("unexpected ban error: %v", err)
	}
	if lastMethod != http.MethodPut || lastPath != "/guilds/100/bans/200" {
		t.Fatalf("unexpected suite ban call: %s %s", lastMethod, lastPath)
	}

	if err := suite.SetNick("CoolNick"); err != nil {
		t.Fatalf("unexpected setnick error: %v", err)
	}
	if lastMethod != http.MethodPatch || lastPath != "/guilds/100/members/200" {
		t.Fatalf("unexpected suite setnick call: %s %s", lastMethod, lastPath)
	}

	if err := suite.Mute("voice mute"); err != nil {
		t.Fatalf("unexpected mute error: %v", err)
	}
	if lastMethod != http.MethodPatch || lastPath != "/guilds/100/members/200" {
		t.Fatalf("unexpected suite mute call: %s %s", lastMethod, lastPath)
	}
}

func TestMember_EmptyValidation(t *testing.T) {
	guildID := types.Snowflake(100)
	c := &Context{
		ctx:     context.Background(),
		REST:    &rest.Client{},
		GuildID: &guildID,
	}

	err := c.Member(Options{Target: 200})
	if !errors.Is(err, ErrEmptyMemberUpdate) {
		t.Fatalf("expected ErrEmptyMemberUpdate, got %v", err)
	}
}
