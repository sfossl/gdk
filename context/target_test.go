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
