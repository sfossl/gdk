package context

import (
	"errors"
	"strings"
	"time"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/util/time"
)

type ModType int

const (
	ModKick ModType = iota + 1
	ModBan
	ModUnban
	ModTimeout
	ModRemoveTimeout
)

var (
	ErrInvalidModType   = errors.New("invalid moderation action type")
	ErrTimeoutTooLong   = errors.New("timeout duration exceeds maximum 28 days")
	ErrTimeoutTooShort  = errors.New("timeout duration must be at least 1 second")
	ErrBanDeleteTooLong = errors.New("ban delete message duration exceeds maximum 7 days")
	ErrUnsupportedType  = errors.New("unsupported moderation parameter type")
)

const (
	MaxTimeoutDuration = 28 * 24 * time.Hour
	MaxBanDeleteSec    = 7 * 24 * 60 * 60
)

type ModOptions struct {
	Type           ModType
	Target         types.Snowflake
	Duration       any
	DeleteMessages any
	Reason         string
}

func parseModType(v any) ModType {
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
		case "removetimeout":
			return ModRemoveTimeout
		}
	}
	return 0
}

func (c *Context) Mod(args ...any) error {
	if c.GuildID == nil {
		return ErrNotInGuild
	}
	if c.REST == nil {
		return ErrNoRESTClient
	}
	if len(args) == 0 {
		return ErrInvalidModType
	}

	opts := ModOptions{}
	parseRest := false
	switch v := args[0].(type) {
	case ModOptions:
		opts = v
	case *ModOptions:
		if v != nil {
			opts = *v
		}
	case Options:
		opts = ModOptions{
			Type:           parseModType(v.Type),
			Target:         v.ResolveTarget(c),
			Duration:       v.Duration,
			DeleteMessages: v.Duration,
			Reason:         v.Reason,
		}
	case *Options:
		if v != nil {
			opts = ModOptions{
				Type:           parseModType(v.Type),
				Target:         v.ResolveTarget(c),
				Duration:       v.Duration,
				DeleteMessages: v.Duration,
				Reason:         v.Reason,
			}
		}
	case ModType:
		opts.Type = v
		parseRest = true
	case string:
		opts.Type = parseModType(v)
		parseRest = true
	default:
		return ErrInvalidModType
	}

	if parseRest {
		for _, arg := range args[1:] {
			switch a := arg.(type) {
			case types.Snowflake:
				opts.Target = a
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
				switch opts.Type {
				case ModTimeout:
					if opts.Duration == nil {
						opts.Duration = a
					} else {
						opts.Reason = a
					}
				case ModBan:
					if opts.DeleteMessages == nil && (strings.HasSuffix(a, "d") || strings.HasSuffix(a, "h") || strings.HasSuffix(a, "m") || strings.HasSuffix(a, "s")) {
						opts.DeleteMessages = a
					} else {
						opts.Reason = a
					}
				default:
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
	default:
		return ErrInvalidModType
	}
}

type ModAction struct {
	ctx      *Context
	targetID types.Snowflake
}

func (c *Context) ModAction(targetID types.Snowflake) ModAction {
	return ModAction{
		ctx:      c,
		targetID: targetID,
	}
}

func (m ModAction) Kick(reason ...string) error {
	return m.ctx.Mod(ModOptions{
		Type:   ModKick,
		Target: m.targetID,
		Reason: strings.Join(reason, " "),
	})
}

func (m ModAction) Ban(deleteHistory any, reason ...string) error {
	return m.ctx.Mod(ModOptions{
		Type:           ModBan,
		Target:         m.targetID,
		DeleteMessages: deleteHistory,
		Reason:         strings.Join(reason, " "),
	})
}

func (m ModAction) Unban(reason ...string) error {
	return m.ctx.Mod(ModOptions{
		Type:   ModUnban,
		Target: m.targetID,
		Reason: strings.Join(reason, " "),
	})
}

func (m ModAction) Timeout(duration any, reason ...string) error {
	return m.ctx.Mod(ModOptions{
		Type:     ModTimeout,
		Target:   m.targetID,
		Duration: duration,
		Reason:   strings.Join(reason, " "),
	})
}

func (m ModAction) RemoveTimeout(reason ...string) error {
	return m.ctx.Mod(ModOptions{
		Type:   ModRemoveTimeout,
		Target: m.targetID,
		Reason: strings.Join(reason, " "),
	})
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
