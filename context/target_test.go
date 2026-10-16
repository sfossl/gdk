package context_test

import (
	stdctx "context"
	"testing"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/context"
)

func TestContext_TargetUser_UserContext(t *testing.T) {
	targetID := types.Snowflake(123456789)
	nick := "TargetNick"
	in := &types.Interaction{
		Type: 2,
		Data: &types.InteractionData{
			Type:     2,
			TargetID: &targetID,
			Resolved: &types.ResolvedData{
				Users: map[types.Snowflake]types.User{
					targetID: {ID: targetID, Username: "targetuser"},
				},
				Members: map[types.Snowflake]types.Member{
					targetID: {Nick: &nick},
				},
			},
		},
	}

	ctx := context.New(stdctx.Background(), in, nil)

	if !ctx.IsUserContext() {
		t.Fatal("expected IsUserContext to be true")
	}
	if ctx.IsSlash() {
		t.Fatal("expected IsSlash to be false")
	}
	if ctx.IsMessageContext() {
		t.Fatal("expected IsMessageContext to be false")
	}

	if ctx.TargetID() != targetID {
		t.Fatalf("expected TargetID %v, got %v", targetID, ctx.TargetID())
	}
	if ctx.TargetUserID() != targetID {
		t.Fatalf("expected TargetUserID %v, got %v", targetID, ctx.TargetUserID())
	}

	user := ctx.TargetUser()
	if user == nil || user.Username != "targetuser" {
		t.Fatalf("expected user targetuser, got %v", user)
	}

	member := ctx.TargetMember()
	if member == nil || member.Nick == nil || *member.Nick != "TargetNick" {
		t.Fatalf("expected member TargetNick, got %v", member)
	}
}

func TestContext_TargetUser_MessageContext(t *testing.T) {
	msgID := types.Snowflake(987654321)
	authorID := types.Snowflake(1122334455)
	authorNick := "AuthorNick"
	in := &types.Interaction{
		Type: 2,
		Data: &types.InteractionData{
			Type:     3,
			TargetID: &msgID,
			Resolved: &types.ResolvedData{
				Messages: map[types.Snowflake]types.Message{
					msgID: {
						ID: msgID,
						Author: types.User{
							ID:       authorID,
							Username: "msgauthor",
						},
					},
				},
				Members: map[types.Snowflake]types.Member{
					authorID: {Nick: &authorNick},
				},
			},
		},
	}

	ctx := context.New(stdctx.Background(), in, nil)

	if !ctx.IsMessageContext() {
		t.Fatal("expected IsMessageContext to be true")
	}
	if ctx.TargetID() != msgID {
		t.Fatalf("expected TargetID %v, got %v", msgID, ctx.TargetID())
	}
	if ctx.TargetUserID() != authorID {
		t.Fatalf("expected TargetUserID %v, got %v", authorID, ctx.TargetUserID())
	}

	user := ctx.TargetUser()
	if user == nil || user.Username != "msgauthor" {
		t.Fatalf("expected user msgauthor, got %v", user)
	}

	member := ctx.TargetMember()
	if member == nil || member.Nick == nil || *member.Nick != "AuthorNick" {
		t.Fatalf("expected member AuthorNick, got %v", member)
	}
}

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

	if !ctx.IsSlash() {
		t.Fatal("expected IsSlash to be true")
	}
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
