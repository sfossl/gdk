package context_test

import (
	stdctx "context"
	"testing"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/context"
)

func TestContext_TargetUser_SlashCommand(t *testing.T) {
	targetID := types.Snowflake(555666777)
	in := &types.Interaction{
		Type: 2,
		Data: &types.InteractionData{
			Type: 1,
			Options: []types.ApplicationCommandOption{
				{
					Name:  "user",
					Type:  6,
					Value: targetID.String(),
				},
			},
			Resolved: &types.ResolvedData{
				Users: map[types.Snowflake]types.User{
					targetID: {ID: targetID, Username: "slashuser"},
				},
			},
		},
	}

	ctx := context.New(stdctx.Background(), in, nil)

	if ctx.TargetID() != targetID {
		t.Fatalf("expected TargetID %v, got %v", targetID, ctx.TargetID())
	}
	if ctx.TargetUserID() != targetID {
		t.Fatalf("expected TargetUserID %v, got %v", targetID, ctx.TargetUserID())
	}

	user := ctx.TargetUser()
	if user == nil || user.Username != "slashuser" {
		t.Fatalf("expected user slashuser, got %v", user)
	}
}

func TestContext_TargetUser_ContextMenu(t *testing.T) {
	targetID := types.Snowflake(111222333)
	nick := "contextnick"
	in := &types.Interaction{
		Type: 2,
		Data: &types.InteractionData{
			Type:     2,
			TargetID: &targetID,
			Resolved: &types.ResolvedData{
				Users: map[types.Snowflake]types.User{
					targetID: {ID: targetID, Username: "contextuser"},
				},
				Members: map[types.Snowflake]types.Member{
					targetID: {Nick: &nick},
				},
			},
		},
	}

	ctx := context.New(stdctx.Background(), in, nil)

	if ctx.TargetID() != targetID {
		t.Fatalf("expected TargetID %v, got %v", targetID, ctx.TargetID())
	}
	if ctx.TargetUserID() != targetID {
		t.Fatalf("expected TargetUserID %v, got %v", targetID, ctx.TargetUserID())
	}

	user := ctx.TargetUser()
	if user == nil || user.Username != "contextuser" {
		t.Fatalf("expected user contextuser, got %v", user)
	}

	member := ctx.TargetMember()
	if member == nil || member.Nick == nil || *member.Nick != "contextnick" {
		t.Fatalf("expected member contextnick, got %v", member)
	}
}

func TestContext_TargetMessage_ContextMenu(t *testing.T) {
	targetID := types.Snowflake(999888777)
	in := &types.Interaction{
		Type: 2,
		Data: &types.InteractionData{
			Type:     3,
			TargetID: &targetID,
			Resolved: &types.ResolvedData{
				Messages: map[types.Snowflake]types.Message{
					targetID: {ID: targetID, Content: "pinned message"},
				},
			},
		},
	}

	ctx := context.New(stdctx.Background(), in, nil)

	if ctx.TargetID() != targetID {
		t.Fatalf("expected TargetID %v, got %v", targetID, ctx.TargetID())
	}

	msg := ctx.ResolvedMessage()
	if msg == nil || msg.Content != "pinned message" {
		t.Fatalf("expected message 'pinned message', got %v", msg)
	}
}

func TestContext_TargetUser_NestedSubcommand(t *testing.T) {
	targetID := types.Snowflake(777888999)
	nick := "nestednick"
	in := &types.Interaction{
		Type: types.InteractionTypeApplicationCommand,
		Data: &types.InteractionData{
			Type: 1,
			Options: []types.ApplicationCommandOption{
				{
					Name: "mod",
					Type: int(types.OptionTypeSubcommandGroup),
					Options: []types.ApplicationCommandOption{
						{
							Name: "ban",
							Type: int(types.OptionTypeSubcommand),
							Options: []types.ApplicationCommandOption{
								{
									Name:  "member",
									Type:  int(types.OptionTypeUser),
									Value: targetID.String(),
								},
							},
						},
					},
				},
			},
			Resolved: &types.ResolvedData{
				Users: map[types.Snowflake]types.User{
					targetID: {ID: targetID, Username: "nesteduser"},
				},
				Members: map[types.Snowflake]types.Member{
					targetID: {Nick: &nick},
				},
			},
		},
	}

	ctx := context.New(stdctx.Background(), in, nil)

	if ctx.TargetID() != targetID {
		t.Fatalf("expected TargetID %v, got %v", targetID, ctx.TargetID())
	}
	if ctx.TargetUserID() != targetID {
		t.Fatalf("expected TargetUserID %v, got %v", targetID, ctx.TargetUserID())
	}

	user := ctx.TargetUser()
	if user == nil || user.Username != "nesteduser" {
		t.Fatalf("expected user nesteduser, got %v", user)
	}

	member := ctx.TargetMember()
	if member == nil || member.Nick == nil || *member.Nick != "nestednick" {
		t.Fatalf("expected member nestednick, got %v", member)
	}
}

func TestContext_TargetUser_Mentionable(t *testing.T) {
	targetID := types.Snowflake(333444555)
	in := &types.Interaction{
		Type: types.InteractionTypeApplicationCommand,
		Data: &types.InteractionData{
			Type: 1,
			Options: []types.ApplicationCommandOption{
				{
					Name:  "entity",
					Type:  int(types.OptionTypeMentionable),
					Value: targetID.String(),
				},
			},
			Resolved: &types.ResolvedData{
				Users: map[types.Snowflake]types.User{
					targetID: {ID: targetID, Username: "mentionuser"},
				},
			},
		},
	}

	ctx := context.New(stdctx.Background(), in, nil)

	if ctx.TargetID() != targetID {
		t.Fatalf("expected TargetID %v, got %v", targetID, ctx.TargetID())
	}
	user := ctx.TargetUser()
	if user == nil || user.Username != "mentionuser" {
		t.Fatalf("expected user mentionuser, got %v", user)
	}
}

func TestContext_TargetMember_AvatarURL(t *testing.T) {
	targetID := types.Snowflake(11223344)
	guildID := types.Snowflake(99887766)
	uHash := "useravatarhash"
	mHash := "memberavatarhash"

	in := &types.Interaction{
		Type:    2,
		GuildID: &guildID,
		Data: &types.InteractionData{
			Type:     2,
			TargetID: &targetID,
			Resolved: &types.ResolvedData{
				Users: map[types.Snowflake]types.User{
					targetID: {ID: targetID, Username: "avataruser", Avatar: &uHash},
				},
				Members: map[types.Snowflake]types.Member{
					targetID: {Avatar: &mHash},
				},
			},
		},
	}

	ctx := context.New(stdctx.Background(), in, nil)
	member := ctx.TargetMember()
	if member == nil {
		t.Fatalf("expected member")
	}

	guildAvatar := member.AvatarURL()
	if guildAvatar != "https://cdn.discordapp.com/guilds/99887766/users/11223344/avatars/memberavatarhash.png?size=512" {
		t.Fatalf("unexpected guild avatar: %s", guildAvatar)
	}

	globalAvatar := member.AvatarURL(true, 256)
	if globalAvatar != "https://cdn.discordapp.com/avatars/11223344/useravatarhash.png?size=256" {
		t.Fatalf("unexpected global avatar: %s", globalAvatar)
	}

	member.Avatar = nil
	fallbackGlobal := member.AvatarURL(false, 128)
	if fallbackGlobal != "https://cdn.discordapp.com/avatars/11223344/useravatarhash.png?size=128" {
		t.Fatalf("unexpected fallback global avatar: %s", fallbackGlobal)
	}
}
