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
