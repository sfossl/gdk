package middleware

import (
	stdctx "context"
	"errors"
	"testing"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/context"
)

func TestPermission_TargetString(t *testing.T) {
	if TargetUser.String() != "user" || TargetBot.String() != "bot" || PermissionTarget(99).String() != "unknown" {
		t.Fatal("unexpected target string representation")
	}
}

func TestRequireUser_AllowedAndAdminBypass(t *testing.T) {
	mw := RequireUser(types.PermissionBanMembers|types.PermissionKickMembers, WithPermissionAutoRespond(false))
	called := false
	handler := mw(func(ctx *context.Context) error {
		called = true
		return nil
	})

	inValid := &types.Interaction{
		GuildID: ptr(types.Snowflake(100)),
		Member: &types.Member{
			Permissions: types.PermissionBanMembers | types.PermissionKickMembers | types.PermissionSendMessages,
		},
	}
	ctxValid := context.New(stdctx.Background(), inValid, nil)
	if err := handler(ctxValid); err != nil {
		t.Fatalf("expected allowed user, got %v", err)
	}
	if !called {
		t.Fatal("expected handler to be invoked")
	}

	called = false
	inAdmin := &types.Interaction{
		GuildID: ptr(types.Snowflake(100)),
		Member: &types.Member{
			Permissions: types.PermissionAdministrator,
		},
	}
	ctxAdmin := context.New(stdctx.Background(), inAdmin, nil)
	if err := handler(ctxAdmin); err != nil {
		t.Fatalf("expected admin bypass, got %v", err)
	}
	if !called {
		t.Fatal("expected handler to be invoked for admin")
	}
}

func TestRequireUser_MissingPermissionsAndNames(t *testing.T) {
	mw := RequireUser(types.PermissionManageMessages|types.PermissionManageRoles, WithPermissionAutoRespond(false))
	handler := mw(func(ctx *context.Context) error {
		return nil
	})

	in := &types.Interaction{
		GuildID: ptr(types.Snowflake(100)),
		Member: &types.Member{
			Permissions: types.PermissionSendMessages,
		},
	}
	ctx := context.New(stdctx.Background(), in, nil)

	err := handler(ctx)
	if err == nil {
		t.Fatal("expected missing permission error")
	}
	if !errors.Is(err, ErrMissingPermissions) {
		t.Fatalf("expected ErrMissingPermissions, got %v", err)
	}

	var permErr *PermissionError
	if !errors.As(err, &permErr) {
		t.Fatalf("expected *PermissionError, got %T", err)
	}
	if permErr.Target != TargetUser {
		t.Fatalf("expected TargetUser, got %v", permErr.Target)
	}
	if permErr.Missing != (types.PermissionManageMessages | types.PermissionManageRoles) {
		t.Fatalf("unexpected missing bitmask %v", permErr.Missing)
	}
	if len(permErr.Names) != 2 {
		t.Fatalf("expected 2 missing names, got %v", permErr.Names)
	}
}

func TestRequireBot_PermissionsAndBypass(t *testing.T) {
	mw := RequireBot(types.PermissionManageRoles, WithPermissionAutoRespond(false))
	handler := mw(func(ctx *context.Context) error {
		return nil
	})

	botAdmin := types.PermissionAdministrator
	inAdmin := &types.Interaction{
		GuildID:        ptr(types.Snowflake(100)),
		AppPermissions: &botAdmin,
	}
	ctxAdmin := context.New(stdctx.Background(), inAdmin, nil)
	if err := handler(ctxAdmin); err != nil {
		t.Fatalf("expected bot admin bypass, got %v", err)
	}

	botPerms := types.PermissionSendMessages
	inMissing := &types.Interaction{
		GuildID:        ptr(types.Snowflake(100)),
		AppPermissions: &botPerms,
	}
	ctxMissing := context.New(stdctx.Background(), inMissing, nil)
	err := handler(ctxMissing)
	if !errors.Is(err, ErrMissingPermissions) {
		t.Fatalf("expected ErrMissingPermissions for bot, got %v", err)
	}

	var permErr *PermissionError
	if !errors.As(err, &permErr) || permErr.Target != TargetBot {
		t.Fatalf("expected TargetBot permission error, got %v", err)
	}
}

func TestRequirePermissions_BothAndAutoRespond(t *testing.T) {
	mw := RequirePermissions(types.PermissionKickMembers, types.PermissionKickMembers, WithPermissionAutoRespond(true))
	handler := mw(func(ctx *context.Context) error {
		return nil
	})

	in := &types.Interaction{
		ID:      types.Snowflake(999),
		Token:   "perm_token",
		GuildID: ptr(types.Snowflake(100)),
		Member: &types.Member{
			Permissions: types.PermissionSendMessages,
		},
	}
	ctx := context.New(stdctx.Background(), in, nil)

	err := handler(ctx)
	if !errors.Is(err, ErrMissingPermissions) {
		t.Fatalf("expected ErrMissingPermissions, got %v", err)
	}

	select {
	case resp := <-ctx.ResponseChannel():
		if resp.Type != types.CallbackTypeChannelMessageWithSource {
			t.Fatalf("expected message callback type, got %v", resp.Type)
		}
		if resp.Data == nil || resp.Data.Content == "" {
			t.Fatal("expected non-empty permission error notice")
		}
		if resp.Data.Flags != types.MessageFlagEphemeral {
			t.Fatalf("expected ephemeral response flag, got %v", resp.Data.Flags)
		}
	default:
		t.Fatal("expected auto-respond message on channel")
	}
}

func TestRequirePermissions_Errors(t *testing.T) {
	mwUser := RequireUser(types.PermissionBanMembers)
	handlerUser := mwUser(func(ctx *context.Context) error { return nil })

	ctxDM := context.New(stdctx.Background(), &types.Interaction{}, nil)
	if err := handlerUser(ctxDM); !errors.Is(err, ErrNotInGuild) {
		t.Fatalf("expected ErrNotInGuild, got %v", err)
	}

	inNoMember := &types.Interaction{GuildID: ptr(types.Snowflake(100))}
	ctxNoMember := context.New(stdctx.Background(), inNoMember, nil)
	if err := handlerUser(ctxNoMember); !errors.Is(err, ErrNoMemberPermissions) {
		t.Fatalf("expected ErrNoMemberPermissions, got %v", err)
	}

	mwBot := RequireBot(types.PermissionManageChannels)
	handlerBot := mwBot(func(ctx *context.Context) error { return nil })
	if err := handlerBot(ctxNoMember); !errors.Is(err, ErrNoBotPermissions) {
		t.Fatalf("expected ErrNoBotPermissions, got %v", err)
	}
}

func TestPermissionNames_Decoding(t *testing.T) {
	combo := types.PermissionAdministrator | types.PermissionManageChannels | types.PermissionBanMembers
	names := PermissionNames(combo)
	if len(names) != 3 {
		t.Fatalf("expected 3 decoded names, got %d: %v", len(names), names)
	}
}

