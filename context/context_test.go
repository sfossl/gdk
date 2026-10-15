package context

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/sfossl/gda/types"
)

func TestContext_Initialization(t *testing.T) {
	guildID := types.Snowflake(100)
	chanID := types.Snowflake(200)
	userID := types.Snowflake(300)

	in := &types.Interaction{
		ID:        types.Snowflake(1),
		AppID:     types.Snowflake(2),
		GuildID:   &guildID,
		ChannelID: &chanID,
		User: &types.User{
			ID:       userID,
			Username: "testuser",
		},
	}

	ctx := New(context.Background(), in, nil)
	if ctx.ID != 1 || ctx.AppID != 2 || *ctx.GuildID != 100 || ctx.ChannelID != 200 {
		t.Fatalf("context fields mismatch")
	}
	if ctx.Author == nil || ctx.Author.Username != "testuser" {
		t.Fatalf("author resolution mismatch")
	}
	if ctx.IsAcknowledged() || ctx.IsDeferred() {
		t.Fatalf("initial state must be unacknowledged")
	}

	memberIn := &types.Interaction{
		Member: &types.Member{
			User: &types.User{
				ID:       userID,
				Username: "memberuser",
			},
		},
	}
	mCtx := NewGatewayContext(context.Background(), memberIn, nil)
	if mCtx.Author == nil || mCtx.Author.Username != "memberuser" {
		t.Fatalf("member user fallback failed")
	}
}

func TestContext_StoreConcurrency(t *testing.T) {
	ctx := New(context.Background(), nil, nil)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ctx.Set("key", n)
			_ = ctx.Get("key")
		}(i)
	}

	wg.Wait()
}

func TestOpt_ExtractionAndFallbacks(t *testing.T) {
	in := &types.Interaction{
		Data: &types.InteractionData{
			Options: []types.ApplicationCommandOption{
				{Name: "str_val", Type: 3, Value: "hello"},
				{Name: "int_val", Type: 4, Value: 42},
				{Name: "int64_val", Type: 4, Value: int64(100)},
				{Name: "float_val", Type: 10, Value: 3.14},
				{Name: "bool_val", Type: 5, Value: true},
				{Name: "snow_val", Type: 6, Value: "123456789012345678"},
				{Name: "json_num", Type: 4, Value: json.Number("999")},
			},
		},
	}

	ctx := New(context.Background(), in, nil)

	if v := ctx.Opt("str_val").String(); v != "hello" {
		t.Errorf("expected hello, got %s", v)
	}
	if v := ctx.Opt("missing").String("default"); v != "default" {
		t.Errorf("expected default, got %s", v)
	}

	if v := ctx.Opt("int_val").Int(); v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
	if v := ctx.Opt("missing").Int(7); v != 7 {
		t.Errorf("expected 7, got %d", v)
	}

	if v := ctx.Opt("int64_val").Int64(); v != 100 {
		t.Errorf("expected 100, got %d", v)
	}
	if v := ctx.Opt("json_num").Int(); v != 999 {
		t.Errorf("expected 999, got %d", v)
	}

	if v := ctx.Opt("float_val").Float(); v != 3.14 {
		t.Errorf("expected 3.14, got %f", v)
	}
	if v := ctx.Opt("missing").Float(2.5); v != 2.5 {
		t.Errorf("expected 2.5, got %f", v)
	}

	if v := ctx.Opt("bool_val").Bool(); !v {
		t.Errorf("expected true, got %v", v)
	}
	if v := ctx.Opt("missing").Bool(true); !v {
		t.Errorf("expected true fallback, got %v", v)
	}

	if v := ctx.Opt("snow_val").Snowflake(); v != 123456789012345678 {
		t.Errorf("expected snowflake 123456789012345678, got %d", v)
	}
	if v := ctx.Opt("missing").Snowflake(999); v != 999 {
		t.Errorf("expected snowflake 999, got %d", v)
	}

	if ctx.Opt("missing").Exists() {
		t.Errorf("missing option should not report exists")
	}
	if !ctx.Opt("str_val").Exists() {
		t.Errorf("existing option should report exists")
	}
}

func TestOpt_ResolvedEntities(t *testing.T) {
	targetID := types.Snowflake(555)

	chanName := "general"

	in := &types.Interaction{
		Data: &types.InteractionData{
			Options: []types.ApplicationCommandOption{
				{Name: "user_opt", Type: 6, Value: "555"},
				{Name: "role_opt", Type: 8, Value: "555"},
				{Name: "chan_opt", Type: 7, Value: "555"},
			},
			Resolved: &types.ResolvedData{
				Users: map[types.Snowflake]types.User{
					targetID: {ID: targetID, Username: "resolved_user"},
				},
				Members: map[types.Snowflake]types.Member{
					targetID: {Roles: []types.Snowflake{10}},
				},
				Roles: map[types.Snowflake]types.Role{
					targetID: {ID: targetID, Name: "Admin"},
				},
				Channels: map[types.Snowflake]types.Channel{
					targetID: {ID: targetID, Name: &chanName},
				},
			},
		},
	}

	ctx := New(context.Background(), in, nil)

	u := ctx.Opt("user_opt").User()
	if u == nil || u.Username != "resolved_user" {
		t.Fatalf("user resolution failed")
	}

	m := ctx.Opt("user_opt").Member()
	if m == nil || len(m.Roles) != 1 {
		t.Fatalf("member resolution failed")
	}

	r := ctx.Opt("role_opt").Role()
	if r == nil || r.Name != "Admin" {
		t.Fatalf("role resolution failed")
	}

	ch := ctx.Opt("chan_opt").Channel()
	if ch == nil || ch.Name == nil || *ch.Name != "general" {
		t.Fatalf("channel resolution failed")
	}
}

func TestOpt_NestedSubcommands(t *testing.T) {
	in := &types.Interaction{
		Data: &types.InteractionData{
			Options: []types.ApplicationCommandOption{
				{
					Name: "group",
					Type: 2,
					Options: []types.ApplicationCommandOption{
						{
							Name: "sub",
							Type: 1,
							Options: []types.ApplicationCommandOption{
								{Name: "nested_arg", Type: 3, Value: "deep_value"},
							},
						},
					},
				},
			},
		},
	}

	ctx := New(context.Background(), in, nil)
	if v := ctx.Opt("nested_arg").String(); v != "deep_value" {
		t.Fatalf("expected deep_value, got %s", v)
	}
}
