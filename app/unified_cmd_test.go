package app_test

import (
	stdctx "context"
	"fmt"
	"testing"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/app"
	"github.com/sfossl/gdk/command"
	"github.com/sfossl/gdk/context"
)

func TestApp_UnifiedCommand_AllThree(t *testing.T) {
	bot, err := app.New("token", 123456)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	var executedType int
	var executedTarget types.Snowflake

	cmd := command.Command{
		Name:            "ban",
		Description:     "Ban user",
		UserContext:     true,
		UserContextName: "Ban Member",
		MsgContext:      true,
		MsgContextName:  "Ban Message Author",
		Handler: func(ctx *context.Context) error {
			executedType = ctx.CommandType()
			executedTarget = ctx.TargetUserID()
			return nil
		},
	}

	if err := bot.RegisterCommand(cmd, cmd.Handler.(func(*context.Context) error)); err != nil {
		t.Fatalf("failed to register unified command: %v", err)
	}

	slashIn := &types.Interaction{
		Type: 2,
		Data: &types.InteractionData{
			Type: 1,
			Name: "ban",
			Options: []types.ApplicationCommandOption{
				{Name: "user", Type: 6, Value: "1001"},
			},
			Resolved: &types.ResolvedData{
				Users: map[types.Snowflake]types.User{
					1001: {ID: 1001},
				},
			},
		},
	}
	ctx1 := context.New(stdctx.Background(), slashIn, nil)
	if err := bot.Handle(ctx1); err != nil {
		t.Fatalf("slash handle failed: %v", err)
	}
	if executedType != 1 || executedTarget != 1001 {
		t.Fatalf("expected slash type 1 target 1001, got type %d target %d", executedType, executedTarget)
	}

	targetUser := types.Snowflake(2002)
	userIn := &types.Interaction{
		Type: 2,
		Data: &types.InteractionData{
			Type:     2,
			Name:     "Ban Member",
			TargetID: &targetUser,
			Resolved: &types.ResolvedData{
				Users: map[types.Snowflake]types.User{
					2002: {ID: 2002},
				},
			},
		},
	}
	ctx2 := context.New(stdctx.Background(), userIn, nil)
	if err := bot.Handle(ctx2); err != nil {
		t.Fatalf("user context handle failed: %v", err)
	}
	if executedType != 2 || executedTarget != 2002 {
		t.Fatalf("expected user context type 2 target 2002, got type %d target %d", executedType, executedTarget)
	}

	targetMsg := types.Snowflake(3003)
	msgAuthor := types.Snowflake(4004)
	msgIn := &types.Interaction{
		Type: 2,
		Data: &types.InteractionData{
			Type:     3,
			Name:     "Ban Message Author",
			TargetID: &targetMsg,
			Resolved: &types.ResolvedData{
				Messages: map[types.Snowflake]types.Message{
					3003: {
						ID:     3003,
						Author: types.User{ID: msgAuthor},
					},
				},
			},
		},
	}
	ctx3 := context.New(stdctx.Background(), msgIn, nil)
	if err := bot.Handle(ctx3); err != nil {
		t.Fatalf("msg context handle failed: %v", err)
	}
	if executedType != 3 || executedTarget != msgAuthor {
		t.Fatalf("expected msg context type 3 target %d, got type %d target %d", msgAuthor, executedType, executedTarget)
	}
}

func TestApp_UnifiedCommand_DefaultFallbacks(t *testing.T) {
	bot, err := app.New("token", 123456)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	cmd := command.Command{
		Name:        "kick",
		Description: "Kick user",
		UserContext: true,
		MsgContext:  true,
		Handler: func(ctx *context.Context) error {
			return nil
		},
	}

	if err := bot.RegisterCommand(cmd, cmd.Handler.(func(*context.Context) error)); err != nil {
		t.Fatalf("failed to register command: %v", err)
	}

	targetUser := types.Snowflake(5005)
	userIn := &types.Interaction{
		Type: 2,
		Data: &types.InteractionData{
			Type:     2,
			Name:     "kick (User)",
			TargetID: &targetUser,
		},
	}
	if err := bot.Handle(context.New(stdctx.Background(), userIn, nil)); err != nil {
		t.Fatalf("expected fallback user context to succeed: %v", err)
	}

	targetMsg := types.Snowflake(6006)
	msgIn := &types.Interaction{
		Type: 2,
		Data: &types.InteractionData{
			Type:     3,
			Name:     "kick (Message)",
			TargetID: &targetMsg,
		},
	}
	if err := bot.Handle(context.New(stdctx.Background(), msgIn, nil)); err != nil {
		t.Fatalf("expected fallback msg context to succeed: %v", err)
	}
}

func TestApp_UnifiedCommand_ContextLimits(t *testing.T) {
	bot, err := app.New("token", 123456, func(c *app.Config) {
		c.DevMode = true
		c.DevGuildID = 99999
	})
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	for i := 0; i < 6; i++ {
		cmd := command.Command{
			Name:            fmt.Sprintf("cmd%d", i),
			Description:     "test",
			UserContext:     true,
			UserContextName: fmt.Sprintf("User Cmd %d", i),
			Handler: func(ctx *context.Context) error {
				return nil
			},
		}
		if err := bot.RegisterCommand(cmd, cmd.Handler.(func(*context.Context) error)); err != nil {
			t.Fatalf("register failed: %v", err)
		}
	}

	_, err = bot.SyncWithResult(stdctx.Background(), true)
	if err != app.ErrMaxUserContextCommands {
		t.Fatalf("expected ErrMaxUserContextCommands, got %v", err)
	}
}

