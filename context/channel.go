package context

import (
	"errors"
	"strings"
	"time"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/util/time"
)

var (
	ErrInvalidChannelID    = errors.New("invalid channel id")
	ErrChannelNameRequired = errors.New("channel name is required")
	ErrSlowmodeTooLong     = errors.New("slowmode duration exceeds maximum 6 hours")
	ErrSlowmodeNegative    = errors.New("slowmode duration cannot be negative")
)

const MaxSlowmodeSeconds = 21600

type ChannelAction struct {
	ctx       *Context
	channelID types.Snowflake
}

func (c *Context) Channel(channelID ...types.Snowflake) ChannelAction {
	chID := c.ChannelID
	if len(channelID) > 0 && channelID[0] != 0 {
		chID = channelID[0]
	}
	return ChannelAction{
		ctx:       c,
		channelID: chID,
	}
}

func (c *Context) ManageChannel(channelID ...types.Snowflake) ChannelAction {
	return c.Channel(channelID...)
}

func (a ChannelAction) ID() types.Snowflake {
	return a.channelID
}

func (a ChannelAction) Get() (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	return a.ctx.REST.GetChannel(a.ctx.Context(), a.channelID)
}

func (a ChannelAction) Delete(reason ...string) error {
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return ErrNoRESTClient
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.DeleteChannel(a.ctx.Context(), a.channelID, r)
}

func (a ChannelAction) SetName(name string, reason ...string) (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	if strings.TrimSpace(name) == "" {
		return nil, ErrChannelNameRequired
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.ModifyChannel(a.ctx.Context(), a.channelID, rest.ModifyChannelPayload{
		Name: &name,
	}, r)
}

func (a ChannelAction) SetTopic(topic string, reason ...string) (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.ModifyChannel(a.ctx.Context(), a.channelID, rest.ModifyChannelPayload{
		Topic: &topic,
	}, r)
}

func (a ChannelAction) SetSlowmode(duration any, reason ...string) (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	sec, err := parseSlowmodeSeconds(duration)
	if err != nil {
		return nil, err
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.ModifyChannel(a.ctx.Context(), a.channelID, rest.ModifyChannelPayload{
		RateLimitPerUser: &sec,
	}, r)
}

func (a ChannelAction) Edit(payload rest.ModifyChannelPayload, reason ...string) (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.ModifyChannel(a.ctx.Context(), a.channelID, payload, r)
}

func (a ChannelAction) Create(payload rest.CreateChannelPayload, reason ...string) (*types.Channel, error) {
	if a.ctx.GuildID == nil {
		return nil, ErrNotInGuild
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	if strings.TrimSpace(payload.Name) == "" {
		return nil, ErrChannelNameRequired
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.CreateGuildChannel(a.ctx.Context(), *a.ctx.GuildID, payload, r)
}

func parseSlowmodeSeconds(v any) (int, error) {
	if v == nil {
		return 0, nil
	}
	switch val := v.(type) {
	case int:
		if val < 0 {
			return 0, ErrSlowmodeNegative
		}
		if val > MaxSlowmodeSeconds {
			return 0, ErrSlowmodeTooLong
		}
		return val, nil
	case int64:
		if val < 0 {
			return 0, ErrSlowmodeNegative
		}
		if val > int64(MaxSlowmodeSeconds) {
			return 0, ErrSlowmodeTooLong
		}
		return int(val), nil
	case time.Duration:
		sec := int(val.Seconds())
		if sec < 0 {
			return 0, ErrSlowmodeNegative
		}
		if sec > MaxSlowmodeSeconds {
			return 0, ErrSlowmodeTooLong
		}
		return sec, nil
	case string:
		if val == "" || val == "0" || val == "0s" {
			return 0, nil
		}
		d, err := timeutil.ParseDuration(val)
		if err != nil {
			return 0, err
		}
		sec := int(d.Seconds())
		if sec < 0 {
			return 0, ErrSlowmodeNegative
		}
		if sec > MaxSlowmodeSeconds {
			return 0, ErrSlowmodeTooLong
		}
		return sec, nil
	default:
		return 0, ErrUnsupportedType
	}
}
