package context

import (
	"errors"
	"strings"
	"time"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/util/time"
)

var (
	ErrTimeoutTooLong   = errors.New("timeout duration exceeds maximum 28 days")
	ErrTimeoutTooShort  = errors.New("timeout duration must be at least 1 second")
	ErrBanDeleteTooLong = errors.New("ban delete message duration exceeds maximum 7 days")
	ErrUnsupportedType  = errors.New("unsupported moderation parameter type")
)

const (
	MaxTimeoutDuration = 28 * 24 * time.Hour
	MaxBanDeleteSec    = 7 * 24 * 60 * 60
)

type ModAction struct {
	ctx      *Context
	targetID types.Snowflake
}

func (c *Context) Mod(targetID types.Snowflake) ModAction {
	return ModAction{
		ctx:      c,
		targetID: targetID,
	}
}

func (m ModAction) Kick(reason ...string) error {
	if ok, err := m.ctx.CanModerate(m.targetID); !ok {
		return err
	}
	r := strings.Join(reason, " ")
	return m.ctx.REST.KickMember(m.ctx.Context(), *m.ctx.GuildID, m.targetID, r)
}

func (m ModAction) Ban(deleteHistory any, reason ...string) error {
	if ok, err := m.ctx.CanModerate(m.targetID); !ok {
		return err
	}

	sec, err := parseDeleteHistorySeconds(deleteHistory)
	if err != nil {
		return err
	}

	r := strings.Join(reason, " ")
	return m.ctx.REST.BanMember(m.ctx.Context(), *m.ctx.GuildID, m.targetID, sec, r)
}

func (m ModAction) Unban(reason ...string) error {
	if m.ctx.GuildID == nil {
		return ErrNotInGuild
	}
	if m.ctx.REST == nil {
		return ErrNoRESTClient
	}
	if m.targetID == 0 {
		return ErrInvalidTarget
	}

	r := strings.Join(reason, " ")
	return m.ctx.REST.UnbanMember(m.ctx.Context(), *m.ctx.GuildID, m.targetID, r)
}

func (m ModAction) Timeout(duration any, reason ...string) error {
	if ok, err := m.ctx.CanModerate(m.targetID); !ok {
		return err
	}

	until, err := parseTimeoutUntil(duration)
	if err != nil {
		return err
	}

	r := strings.Join(reason, " ")
	return m.ctx.REST.TimeoutMember(m.ctx.Context(), *m.ctx.GuildID, m.targetID, until, r)
}

func (m ModAction) RemoveTimeout(reason ...string) error {
	if ok, err := m.ctx.CanModerate(m.targetID); !ok {
		return err
	}

	r := strings.Join(reason, " ")
	return m.ctx.REST.TimeoutMember(m.ctx.Context(), *m.ctx.GuildID, m.targetID, nil, r)
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
