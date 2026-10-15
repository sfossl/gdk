package middleware

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/context"
)

var (
	ErrMissingPermissions = errors.New("missing required permissions")
	ErrNotInGuild          = errors.New("interaction not in a guild")
	ErrNoMemberPermissions = errors.New("member permissions unavailable")
	ErrNoBotPermissions    = errors.New("bot application permissions unavailable")
)

type PermissionTarget int

const (
	TargetUser PermissionTarget = iota
	TargetBot
)

func (t PermissionTarget) String() string {
	switch t {
	case TargetUser:
		return "user"
	case TargetBot:
		return "bot"
	default:
		return "unknown"
	}
}

type PermissionError struct {
	Target  PermissionTarget
	Missing types.Permission
	Names   []string
}

func (e *PermissionError) Error() string {
	return "missing required permissions"
}

func (e *PermissionError) Is(target error) bool {
	return target == ErrMissingPermissions
}

type PermissionConfig struct {
	AutoRespond bool
	MessageFn   func(target PermissionTarget, missing []string) string
}

type PermissionOption func(*PermissionConfig)

func WithPermissionAutoRespond(autoRespond bool) PermissionOption {
	return func(c *PermissionConfig) {
		c.AutoRespond = autoRespond
	}
}

func WithPermissionMessage(fn func(target PermissionTarget, missing []string) string) PermissionOption {
	return func(c *PermissionConfig) {
		c.MessageFn = fn
	}
}

func defaultPermissionMessage(target PermissionTarget, missing []string) string {
	joined := strings.Join(missing, ", ")
	if target == TargetBot {
		return fmt.Sprintf("The bot is missing the required permissions to execute this command (%s).", joined)
	}
	return fmt.Sprintf("You do not have the required permissions to execute this command (%s).", joined)
}

func RequireUser(perms types.Permission, opts ...PermissionOption) Middleware {
	return RequirePermissions(perms, 0, opts...)
}

func RequireBot(perms types.Permission, opts ...PermissionOption) Middleware {
	return RequirePermissions(0, perms, opts...)
}

func RequirePermissions(userPerms, botPerms types.Permission, opts ...PermissionOption) Middleware {
	cfg := PermissionConfig{
		AutoRespond: true,
		MessageFn:   defaultPermissionMessage,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	return func(next Handler) Handler {
		return func(ctx *context.Context) error {
			if ctx == nil {
				if next != nil {
					return next(ctx)
				}
				return nil
			}


			if userPerms != 0 {
				if ctx.GuildID == nil {
					return ErrNotInGuild
				}
				if ctx.Member == nil {
					return ErrNoMemberPermissions
				}

				effective := ctx.Member.Permissions
				if (effective & types.PermissionAdministrator) != types.PermissionAdministrator {
					if (effective & userPerms) != userPerms {
						missing := userPerms &^ effective
						names := PermissionNames(missing)
						permErr := &PermissionError{
							Target:  TargetUser,
							Missing: missing,
							Names:   names,
						}
						if cfg.AutoRespond && (!ctx.IsAcknowledged() || ctx.IsDeferred()) {
							msg := cfg.MessageFn(TargetUser, names)
							_, _ = ctx.Respond(context.Response{
								Content:   msg,
								Ephemeral: true,
							})
						}
						return permErr
					}
				}
			}

			if botPerms != 0 {
				if ctx.GuildID == nil {
					return ErrNotInGuild
				}
				if ctx.Interaction == nil || ctx.Interaction.AppPermissions == nil {
					return ErrNoBotPermissions
				}

				botEffective := *ctx.Interaction.AppPermissions
				if (botEffective & types.PermissionAdministrator) != types.PermissionAdministrator {
					if (botEffective & botPerms) != botPerms {
						missing := botPerms &^ botEffective
						names := PermissionNames(missing)
						permErr := &PermissionError{
							Target:  TargetBot,
							Missing: missing,
							Names:   names,
						}
						if cfg.AutoRespond && (!ctx.IsAcknowledged() || ctx.IsDeferred()) {
							msg := cfg.MessageFn(TargetBot, names)
							_, _ = ctx.Respond(context.Response{
								Content:   msg,
								Ephemeral: true,
							})
						}
						return permErr
					}
				}
			}

			if next != nil {
				return next(ctx)
			}
			return nil
		}
	}
}

func PermissionNames(perms types.Permission) []string {
	var names []string

	checks := []struct {
		perm types.Permission
		name string
	}{
		{types.PermissionAdministrator, "Administrator"},
		{types.PermissionManageGuild, "Manage Server"},
		{types.PermissionManageRoles, "Manage Roles"},
		{types.PermissionManageChannels, "Manage Channels"},
		{types.PermissionKickMembers, "Kick Members"},
		{types.PermissionBanMembers, "Ban Members"},
		{types.PermissionCreateInstantInvite, "Create Invite"},
		{types.PermissionChangeNickname, "Change Nickname"},
		{types.PermissionManageNicknames, "Manage Nicknames"},
		{types.PermissionManageGuildExpressions, "Manage Expressions"},
		{types.PermissionManageWebhooks, "Manage Webhooks"},
		{types.PermissionViewChannel, "View Channel"},
		{types.PermissionViewAuditLog, "View Audit Log"},
		{types.PermissionViewGuildInsights, "View Server Insights"},
		{types.PermissionViewCreatorMonetizationAnalytics, "View Monetization Analytics"},
		{types.PermissionSendMessages, "Send Messages"},
		{types.PermissionSendTTSMessages, "Send TTS Messages"},
		{types.PermissionManageMessages, "Manage Messages"},
		{types.PermissionEmbedLinks, "Embed Links"},
		{types.PermissionAttachFiles, "Attach Files"},
		{types.PermissionReadMessageHistory, "Read Message History"},
		{types.PermissionMentionEveryone, "Mention Everyone"},
		{types.PermissionUseExternalEmojis, "Use External Emojis"},
		{types.PermissionUseExternalStickers, "Use External Stickers"},
		{types.PermissionAddReactions, "Add Reactions"},
		{types.PermissionConnect, "Connect"},
		{types.PermissionSpeak, "Speak"},
		{types.PermissionMuteMembers, "Mute Members"},
		{types.PermissionDeafenMembers, "Deafen Members"},
		{types.PermissionMoveMembers, "Move Members"},
		{types.PermissionUseVAD, "Use Voice Activity"},
		{types.PermissionPrioritySpeaker, "Priority Speaker"},
		{types.PermissionStream, "Video"},
		{types.PermissionUseEmbeddedActivities, "Use Activities"},
		{types.PermissionUseSoundboard, "Use Soundboard"},
		{types.PermissionCreateGuildExpressions, "Create Expressions"},
		{types.PermissionUseExternalSounds, "Use External Sounds"},
		{types.PermissionSendVoiceMessages, "Send Voice Messages"},
		{types.PermissionSendPolls, "Create Polls"},
		{types.PermissionUseApplicationCommands, "Use Application Commands"},
		{types.PermissionRequestToSpeak, "Request to Speak"},
		{types.PermissionManageEvents, "Manage Events"},
		{types.PermissionCreateEvents, "Create Events"},
		{types.PermissionManageThreads, "Manage Threads"},
		{types.PermissionCreatePublicThreads, "Create Public Threads"},
		{types.PermissionCreatePrivateThreads, "Create Private Threads"},
		{types.PermissionSendMessagesInThreads, "Send Messages in Threads"},
		{types.PermissionModerateMembers, "Timeout Members"},
	}

	for _, c := range checks {
		if (perms & c.perm) == c.perm {
			names = append(names, c.name)
		}
	}

	if len(names) == 0 && perms != 0 {
		names = append(names, fmt.Sprintf("Unknown Permission (0x%X)", uint64(perms)))
	}

	return names
}
