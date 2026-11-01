package context

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/defaults"
	"github.com/sfossl/gdk/util/time"
)

type ModType int

const (
	ModKick ModType = iota + 1
	ModBan
	ModUnban
	ModTimeout
	ModRemoveTimeout
	ModAddRole
	ModRemoveRole
	ModNick
	ModResetNick
	ModMute
	ModDeaf
	ModMove
	ModDisconnect
)

var (
	ErrInvalidModType        = errors.New("invalid moderation action type")
	ErrTimeoutTooLong        = errors.New("timeout duration exceeds maximum 28 days")
	ErrTimeoutTooShort       = errors.New("timeout duration must be at least 1 second")
	ErrBanDeleteTooLong      = errors.New("ban delete message duration exceeds maximum 7 days")
	ErrUnsupportedType       = errors.New("unsupported moderation parameter type")
	ErrEmptyMemberUpdate     = errors.New("cannot execute member update with empty options: specify nickname, roles, voice, or moderation action")
)

const (
	MaxTimeoutDuration = defaults.MaxTimeoutDuration
	MaxBanDeleteSec    = defaults.MaxBanDeleteSec
)

type MemberOptions struct {
	Type           ModType
	Target         types.Snowflake
	Duration       any
	DeleteMessages any
	Reason         string
	Nickname       *string
	AddRoles       []types.Snowflake
	RemoveRoles    []types.Snowflake
	VoiceMute      *bool
	VoiceDeaf      *bool
	VoiceChannel   *types.Snowflake
}

type ModOptions = MemberOptions
type UpdateOptions = MemberOptions

func parseMemberActionType(v any) ModType {
	switch val := v.(type) {
	case ModType:
		return val
	case string:
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "kick":
			return ModKick
		case "ban":
			return ModBan
		case "unban":
			return ModUnban
		case "timeout":
			return ModTimeout
		case "removetimeout", "untimeout":
			return ModRemoveTimeout
		case "addrole":
			return ModAddRole
		case "removerole":
			return ModRemoveRole
		case "nick", "nickname":
			return ModNick
		case "resetnick":
			return ModResetNick
		case "mute":
			return ModMute
		case "deaf":
			return ModDeaf
		case "move":
			return ModMove
		case "disconnect":
			return ModDisconnect
		}
	}
	return 0
}

func memberOptionsFromOptions(v Options, c *Context) MemberOptions {
	opts := MemberOptions{
		Type:           parseMemberActionType(v.Type),
		Target:         v.ResolveTarget(c),
		Duration:       v.Duration,
		DeleteMessages: v.Duration,
		Reason:         v.Reason,
		Nickname:       v.Nickname,
		AddRoles:       v.AddRoles,
		RemoveRoles:    v.RemoveRoles,
		VoiceMute:      v.VoiceMute,
		VoiceDeaf:      v.VoiceDeaf,
		VoiceChannel:   v.VoiceChannel,
	}
	if v.Role != 0 {
		if opts.Type == ModRemoveRole {
			opts.RemoveRoles = append(opts.RemoveRoles, v.Role)
		} else {
			opts.AddRoles = append(opts.AddRoles, v.Role)
		}
	}
	return opts
}

func (c *Context) Member(args ...any) error {
	if c.GuildID == nil || *c.GuildID == 0 {
		return ErrNotInGuild
	}
	if c.REST == nil {
		return ErrNoRESTClient
	}
	if len(args) == 0 {
		return ErrEmptyMemberUpdate
	}

	opts := MemberOptions{}
	parseRest := false

	switch v := args[0].(type) {
	case MemberOptions:
		opts = v
	case *MemberOptions:
		if v != nil {
			opts = *v
		}
	case Options:
		opts = memberOptionsFromOptions(v, c)
	case *Options:
		if v != nil {
			opts = memberOptionsFromOptions(*v, c)
		}
	case ModType:
		opts.Type = v
		parseRest = true
	case types.Snowflake:
		opts.Target = v
		parseRest = true
	case string:
		action := parseMemberActionType(v)
		if action != 0 {
			opts.Type = action
		} else {
			opts.Reason = v
		}
		parseRest = true
	default:
		return ErrInvalidModType
	}

	if parseRest {
		for _, arg := range args[1:] {
			switch a := arg.(type) {
			case types.Snowflake:
				if opts.Target == 0 {
					opts.Target = a
				} else {
					if opts.Type == ModRemoveRole {
						opts.RemoveRoles = append(opts.RemoveRoles, a)
					} else {
						opts.AddRoles = append(opts.AddRoles, a)
					}
				}
			case []types.Snowflake:
				if opts.Type == ModRemoveRole {
					opts.RemoveRoles = append(opts.RemoveRoles, a...)
				} else {
					opts.AddRoles = append(opts.AddRoles, a...)
				}
			case ModType:
				opts.Type = a
			case time.Duration:
				opts.Duration = a
			case int:
				switch opts.Type {
				case ModBan:
					opts.DeleteMessages = a
				case ModTimeout:
					opts.Duration = a
				}
			case string:
				action := parseMemberActionType(a)
				if action != 0 && opts.Type == 0 {
					opts.Type = action
				} else if opts.Type == ModTimeout && opts.Duration == nil {
					opts.Duration = a
				} else if opts.Type == ModBan && opts.DeleteMessages == nil && (strings.HasSuffix(a, "d") || strings.HasSuffix(a, "h") || strings.HasSuffix(a, "m") || strings.HasSuffix(a, "s")) {
					opts.DeleteMessages = a
				} else if opts.Type == ModNick && opts.Nickname == nil {
					opts.Nickname = &a
				} else {
					opts.Reason = a
				}
			}
		}
	}

	target := opts.Target
	if target == 0 {
		target = c.TargetID()
	}
	if target == 0 {
		return ErrInvalidTarget
	}
	opts.Target = target

	if opts.Type != 0 {
		switch opts.Type {
		case ModKick:
			if ok, err := c.CanModerate(opts.Target); !ok {
				return err
			}
			return c.REST.KickMember(c.Context(), *c.GuildID, opts.Target, opts.Reason)
		case ModBan:
			if ok, err := c.CanModerate(opts.Target); !ok {
				return err
			}
			sec, err := parseDeleteHistorySeconds(opts.DeleteMessages)
			if err != nil {
				return err
			}
			return c.REST.BanMember(c.Context(), *c.GuildID, opts.Target, sec, opts.Reason)
		case ModUnban:
			return c.REST.UnbanMember(c.Context(), *c.GuildID, opts.Target, opts.Reason)
		case ModTimeout:
			if ok, err := c.CanModerate(opts.Target); !ok {
				return err
			}
			until, err := parseTimeoutUntil(opts.Duration)
			if err != nil {
				return err
			}
			return c.REST.TimeoutMember(c.Context(), *c.GuildID, opts.Target, until, opts.Reason)
		case ModRemoveTimeout:
			if ok, err := c.CanModerate(opts.Target); !ok {
				return err
			}
			return c.REST.TimeoutMember(c.Context(), *c.GuildID, opts.Target, nil, opts.Reason)
		case ModAddRole:
			for _, r := range opts.AddRoles {
				if ok, err := c.CanManageRole(r); !ok {
					return err
				}
				if err := c.REST.AddGuildMemberRole(c.Context(), *c.GuildID, opts.Target, r, opts.Reason); err != nil {
					return err
				}
			}
			return nil
		case ModRemoveRole:
			for _, r := range opts.RemoveRoles {
				if ok, err := c.CanManageRole(r); !ok {
					return err
				}
				if err := c.REST.RemoveGuildMemberRole(c.Context(), *c.GuildID, opts.Target, r, opts.Reason); err != nil {
					return err
				}
			}
			return nil
		case ModResetNick:
			empty := ""
			opts.Nickname = &empty
		case ModNick, ModMute, ModDeaf, ModMove, ModDisconnect:
		default:
			return ErrInvalidModType
		}
	}

	if opts.Nickname == nil && len(opts.AddRoles) == 0 && len(opts.RemoveRoles) == 0 &&
		opts.VoiceMute == nil && opts.VoiceDeaf == nil && opts.VoiceChannel == nil {
		if opts.Type == 0 {
			return ErrEmptyMemberUpdate
		}
		return nil
	}

	payload := rest.ModifyMemberPayload{
		Nick:      opts.Nickname,
		Mute:      opts.VoiceMute,
		Deaf:      opts.VoiceDeaf,
		ChannelID: opts.VoiceChannel,
	}

	if len(opts.AddRoles) > 0 || len(opts.RemoveRoles) > 0 {
		var currentRoles []types.Snowflake
		found := false
		if c.Interaction != nil && c.Interaction.Data != nil && c.Interaction.Data.Resolved != nil && c.Interaction.Data.Resolved.Members != nil {
			if tm, ok := c.Interaction.Data.Resolved.Members[target]; ok {
				currentRoles = tm.Roles
				found = true
			}
		}
		if !found && c.InvokerMember != nil && c.Author != nil && target == c.Author.ID {
			currentRoles = c.InvokerMember.Roles
			found = true
		}
		if !found {
			m, err := c.REST.GetGuildMember(c.Context(), *c.GuildID, target)
			if err != nil {
				return err
			}
			currentRoles = m.Roles
		}

		roleMap := make(map[types.Snowflake]bool, len(currentRoles)+len(opts.AddRoles))
		for _, r := range currentRoles {
			roleMap[r] = true
		}
		for _, r := range opts.RemoveRoles {
			delete(roleMap, r)
		}
		for _, r := range opts.AddRoles {
			roleMap[r] = true
		}
		newRoles := make([]types.Snowflake, 0, len(roleMap))
		for r := range roleMap {
			newRoles = append(newRoles, r)
		}
		payload.Roles = &newRoles
	}

	return c.modifyGuildMember(*c.GuildID, target, payload, opts.Reason)
}

func (c *Context) modifyGuildMember(guildID, userID types.Snowflake, payload rest.ModifyMemberPayload, reason string) error {
	var headers http.Header
	if reason != "" {
		headers = http.Header{"X-Audit-Log-Reason": []string{reason}}
	}
	_, err := c.REST.Do(c.Context(), rest.Request{
		Endpoint: rest.EndpointModifyGuildMember(guildID, userID),
		Body:     payload,
		Headers:  headers,
	})
	return err
}

func (c *Context) Mod(args ...any) error {
	return c.Member(args...)
}

func (c *Context) Update(args ...any) error {
	if len(args) == 2 {
		if target, ok := args[0].(types.Snowflake); ok {
			if nick, ok := args[1].(string); ok {
				return c.Member(MemberOptions{
					Target:   target,
					Nickname: &nick,
				})
			}
		}
	}
	return c.Member(args...)
}

func parseDeleteHistorySeconds(v any) (int, error) {
	if v == nil {
		return 0, nil
	}
	var sec int
	switch val := v.(type) {
	case int:
		if val < 0 {
			return 0, nil
		}
		sec = val
	case int64:
		if val < 0 {
			return 0, nil
		}
		sec = int(val)
	case time.Duration:
		if val <= 0 {
			return 0, nil
		}
		sec = int(val.Seconds())
	case string:
		if val == "" {
			return 0, nil
		}
		d, err := timeutil.ParseDuration(val)
		if err != nil {
			return 0, err
		}
		sec = int(d.Seconds())
	default:
		return 0, ErrUnsupportedType
	}

	if sec > MaxBanDeleteSec {
		return 0, ErrBanDeleteTooLong
	}
	return sec, nil
}

func parseTimeoutUntil(v any) (*time.Time, error) {
	if v == nil {
		return nil, ErrTimeoutTooShort
	}

	now := time.Now().UTC()
	var until time.Time

	switch val := v.(type) {
	case time.Duration:
		if val < time.Second {
			return nil, ErrTimeoutTooShort
		}
		if val > MaxTimeoutDuration {
			return nil, ErrTimeoutTooLong
		}
		until = now.Add(val)
	case string:
		item := timeutil.New(val)
		if item.Error() != nil {
			return nil, item.Error()
		}
		d := item.Duration()
		if d < time.Second {
			return nil, ErrTimeoutTooShort
		}
		if d > MaxTimeoutDuration {
			return nil, ErrTimeoutTooLong
		}
		until = item.Time()
	case time.Time:
		until = val.UTC()
		diff := until.Sub(now)
		if diff < time.Second {
			return nil, ErrTimeoutTooShort
		}
		if diff > MaxTimeoutDuration {
			return nil, ErrTimeoutTooLong
		}
	default:
		return nil, ErrUnsupportedType
	}

	return &until, nil
}
