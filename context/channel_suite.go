package context

import (
	"strings"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/component"
)

type ChannelSuite struct {
	ctx       *Context
	channelID types.Snowflake
}

func (c *Context) ChannelSuite(channelID ...types.Snowflake) ChannelSuite {
	chID := c.ChannelID
	if len(channelID) > 0 && channelID[0] != 0 {
		chID = channelID[0]
	}
	return ChannelSuite{
		ctx:       c,
		channelID: chID,
	}
}

func (a ChannelSuite) ID() types.Snowflake {
	return a.channelID
}

func (a ChannelSuite) Get() (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	return a.ctx.REST.GetChannel(a.ctx.Context(), a.channelID)
}

func (a ChannelSuite) Delete(reason ...string) error {
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return ErrNoRESTClient
	}
	r := joinReason(reason)
	return a.ctx.REST.DeleteChannel(a.ctx.Context(), a.channelID, r)
}

func (a ChannelSuite) SetName(name string, reason ...string) (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	if strings.TrimSpace(name) == "" {
		return nil, ErrChannelNameRequired
	}
	r := joinReason(reason)
	return a.ctx.REST.ModifyChannel(a.ctx.Context(), a.channelID, rest.ModifyChannelPayload{
		Name: &name,
	}, r)
}

func (a ChannelSuite) SetTopic(topic string, reason ...string) (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	r := joinReason(reason)
	return a.ctx.REST.ModifyChannel(a.ctx.Context(), a.channelID, rest.ModifyChannelPayload{
		Topic: &topic,
	}, r)
}

func (a ChannelSuite) SetSlowmode(duration any, reason ...string) (*types.Channel, error) {
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
	r := joinReason(reason)
	return a.ctx.REST.ModifyChannel(a.ctx.Context(), a.channelID, rest.ModifyChannelPayload{
		RateLimitPerUser: &sec,
	}, r)
}

func (a ChannelSuite) Edit(payload rest.ModifyChannelPayload, reason ...string) (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	r := joinReason(reason)
	return a.ctx.REST.ModifyChannel(a.ctx.Context(), a.channelID, payload, r)
}

func (a ChannelSuite) Create(payload rest.CreateChannelPayload, reason ...string) (*types.Channel, error) {
	if a.ctx.GuildID == nil {
		return nil, ErrNotInGuild
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	if strings.TrimSpace(payload.Name) == "" {
		return nil, ErrChannelNameRequired
	}
	r := joinReason(reason)
	return a.ctx.REST.CreateGuildChannel(a.ctx.Context(), *a.ctx.GuildID, payload, r)
}

func (a ChannelSuite) Send(args ...any) (*types.Message, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	res := parseResponseOptions(args...)
	if res.Content == "" && len(res.Embeds) == 0 && len(res.Components) == 0 && len(res.Files) == 0 {
		return nil, ErrEmptyMessage
	}

	var flags types.MessageFlag
	if len(res.Components) > 0 {
		if err := component.ValidateComponents(res.Components); err != nil {
			return nil, err
		}
		if component.HasLayoutComponents(res.Components) {
			if len(res.Embeds) > 0 {
				return nil, component.ErrEmbedsNotAllowedWithLayouts
			}
			if res.Content != "" {
				return nil, component.ErrContentNotAllowedWithLayouts
			}
			flags |= types.MessageFlagIsComponentsV2
		}
	}

	payload := rest.MessagePayload{
		Content:          res.Content,
		Embeds:           res.Embeds,
		Components:       res.Components,
		Flags:            flags,
		TTS:              res.TTS,
		Files:            res.Files,
		AllowedMentions:  res.ResolveAllowedMentions(a.ctx.DefaultMention),
		MessageReference: res.ResolveReference(),
	}
	return a.ctx.REST.SendMessage(a.ctx.Context(), a.channelID, payload)
}

func (a ChannelSuite) Lock(reason ...string) error {
	return a.ctx.Channel(ChannelOptions{
		Type:      ChannelLock,
		ChannelID: a.channelID,
		Locked:    true,
		Reason:    joinReason(reason),
	})
}

func (a ChannelSuite) Unlock(reason ...string) error {
	return a.ctx.Channel(ChannelOptions{
		Type:      ChannelLock,
		ChannelID: a.channelID,
		Locked:    false,
		Reason:    joinReason(reason),
	})
}

func (a ChannelSuite) Purge(count int, reason ...string) (int, error) {
	return a.ctx.purgeMessages(a.channelID, count, joinReason(reason))
}

func (a ChannelSuite) Typing() error {
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return ErrNoRESTClient
	}
	return a.ctx.REST.TriggerTyping(a.ctx.Context(), a.channelID)
}

func (c *Context) Typing() error {
	return c.ChannelSuite().Typing()
}
