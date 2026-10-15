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

func TestCanModerate_NotInGuild(t *testing.T) {
	c := &Context{}
	ok, err := c.CanModerate(types.Snowflake(123))
	if ok || !errors.Is(err, ErrNotInGuild) {
		t.Fatalf("expected ErrNotInGuild, got ok=%v, err=%v", ok, err)
	}
}

func TestCanModerate_NoRESTClient(t *testing.T) {
	guildID := types.Snowflake(100)
	c := &Context{
		GuildID: &guildID,
	}
	ok, err := c.CanModerate(types.Snowflake(123))
	if ok || !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got ok=%v, err=%v", ok, err)
	}
}

func TestCanModerate_SelfAndOwner(t *testing.T) {
	guildID := types.Snowflake(100)
	ownerID := types.Snowflake(999)
	authorID := types.Snowflake(555)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/guilds/100" {
			_ = json.NewEncoder(w).Encode(types.Guild{
				ID:      guildID,
				OwnerID: ownerID,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	c := &Context{
		ctx:     context.Background(),
		GuildID: &guildID,
		REST:    client,
		Author: &types.User{
			ID: authorID,
		},
	}

	ok, err := c.CanModerate(ownerID)
	if ok || !errors.Is(err, ErrCannotModerateOwner) {
		t.Fatalf("expected ErrCannotModerateOwner, got ok=%v, err=%v", ok, err)
	}

	ok, err = c.CanModerate(authorID)
	if ok || !errors.Is(err, ErrCannotModerateSelf) {
		t.Fatalf("expected ErrCannotModerateSelf, got ok=%v, err=%v", ok, err)
	}
}

func TestCanModerate_HierarchyComparisons(t *testing.T) {
	guildID := types.Snowflake(100)
	ownerID := types.Snowflake(999)
	botAppID := types.Snowflake(1)
	invokerID := types.Snowflake(2)
	targetID := types.Snowflake(3)

	roleLow := types.Role{ID: types.Snowflake(10), Position: 1}
	roleMid := types.Role{ID: types.Snowflake(20), Position: 5}
	roleHigh := types.Role{ID: types.Snowflake(30), Position: 10}

	type memberSpec struct {
		roles []types.Snowflake
	}

	specs := map[types.Snowflake]memberSpec{
		botAppID:  {roles: []types.Snowflake{roleHigh.ID}},
		invokerID: {roles: []types.Snowflake{roleMid.ID}},
		targetID:  {roles: []types.Snowflake{roleLow.ID}},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/guilds/100":
			_ = json.NewEncoder(w).Encode(types.Guild{
				ID:      guildID,
				OwnerID: ownerID,
				Roles:   []types.Role{roleLow, roleMid, roleHigh},
			})
		case "/guilds/100/members/1":
			_ = json.NewEncoder(w).Encode(types.Member{Roles: specs[botAppID].roles})
		case "/guilds/100/members/2":
			_ = json.NewEncoder(w).Encode(types.Member{Roles: specs[invokerID].roles})
		case "/guilds/100/members/3":
			_ = json.NewEncoder(w).Encode(types.Member{Roles: specs[targetID].roles})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))

	t.Run("success standard hierarchy", func(t *testing.T) {
		c := &Context{
			ctx:     context.Background(),
			GuildID: &guildID,
			AppID:   botAppID,
			REST:    client,
			Author:  &types.User{ID: invokerID},
			Member:  &types.Member{Roles: specs[invokerID].roles},
		}

		ok, err := c.CanModerate(targetID)
		if !ok || err != nil {
			t.Fatalf("expected success, got ok=%v, err=%v", ok, err)
		}
	})

	t.Run("bot outranked", func(t *testing.T) {
		specs[botAppID] = memberSpec{roles: []types.Snowflake{roleLow.ID}}
		specs[targetID] = memberSpec{roles: []types.Snowflake{roleMid.ID}}

		c := &Context{
			ctx:     context.Background(),
			GuildID: &guildID,
			AppID:   botAppID,
			REST:    client,
			Author:  &types.User{ID: invokerID},
			Member:  &types.Member{Roles: []types.Snowflake{roleHigh.ID}},
		}

		ok, err := c.CanModerate(targetID)
		if ok || !errors.Is(err, ErrTargetOutranksBot) {
			t.Fatalf("expected ErrTargetOutranksBot, got ok=%v, err=%v", ok, err)
		}
	})

	t.Run("user outranked", func(t *testing.T) {
		specs[botAppID] = memberSpec{roles: []types.Snowflake{roleHigh.ID}}
		specs[invokerID] = memberSpec{roles: []types.Snowflake{roleLow.ID}}
		specs[targetID] = memberSpec{roles: []types.Snowflake{roleMid.ID}}

		c := &Context{
			ctx:     context.Background(),
			GuildID: &guildID,
			AppID:   botAppID,
			REST:    client,
			Author:  &types.User{ID: invokerID},
			Member:  &types.Member{Roles: specs[invokerID].roles},
		}

		ok, err := c.CanModerate(targetID)
		if ok || !errors.Is(err, ErrTargetOutranksUser) {
			t.Fatalf("expected ErrTargetOutranksUser, got ok=%v, err=%v", ok, err)
		}
	})

	t.Run("owner invoker bypasses role position", func(t *testing.T) {
		specs[botAppID] = memberSpec{roles: []types.Snowflake{roleHigh.ID}}
		specs[targetID] = memberSpec{roles: []types.Snowflake{roleMid.ID}}

		c := &Context{
			ctx:     context.Background(),
			GuildID: &guildID,
			AppID:   botAppID,
			REST:    client,
			Author:  &types.User{ID: ownerID},
			Member:  &types.Member{Roles: []types.Snowflake{}},
		}

		ok, err := c.CanModerate(targetID)
		if !ok || err != nil {
			t.Fatalf("expected owner bypass success, got ok=%v, err=%v", ok, err)
		}
	})

	t.Run("target resolved in interaction", func(t *testing.T) {
		specs[botAppID] = memberSpec{roles: []types.Snowflake{roleHigh.ID}}

		resolvedID := types.Snowflake(777)
		c := &Context{
			ctx:     context.Background(),
			GuildID: &guildID,
			AppID:   botAppID,
			REST:    client,
			Author:  &types.User{ID: invokerID},
			Member:  &types.Member{Roles: []types.Snowflake{roleMid.ID}},
			Interaction: &types.Interaction{
				Data: &types.InteractionData{
					Resolved: &types.ResolvedData{
						Members: map[types.Snowflake]types.Member{
							resolvedID: {Roles: []types.Snowflake{roleLow.ID}},
						},
					},
				},
			},
		}

		ok, err := c.CanModerate(resolvedID)
		if !ok || err != nil {
			t.Fatalf("expected resolved target success, got ok=%v, err=%v", ok, err)
		}
	})

	t.Run("target resolved with nil roles skips REST", func(t *testing.T) {
		specs[botAppID] = memberSpec{roles: []types.Snowflake{roleHigh.ID}}

		rolelessID := types.Snowflake(888)
		c := &Context{
			ctx:     context.Background(),
			GuildID: &guildID,
			AppID:   botAppID,
			REST:    client,
			Author:  &types.User{ID: invokerID},
			Member:  &types.Member{Roles: []types.Snowflake{roleMid.ID}},
			Interaction: &types.Interaction{
				Data: &types.InteractionData{
					Resolved: &types.ResolvedData{
						Members: map[types.Snowflake]types.Member{
							rolelessID: {Roles: nil},
						},
					},
				},
			},
		}

		ok, err := c.CanModerate(rolelessID)
		if !ok || err != nil {
			t.Fatalf("expected roleless target moderation success without REST error, got ok=%v, err=%v", ok, err)
		}
	})
}

