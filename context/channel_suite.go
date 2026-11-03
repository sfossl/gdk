package context

import (
	"strings"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

type ChannelSuite struct {
	ctx       *Context
	channelID types.Snowflake
}

type ChannelActionSuite = ChannelSuite

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

func (c *Context) ChannelAction(channelID ...types.Snowflake) ChannelSuite {
	return c.ChannelSuite(channelID...)
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
	r := strings.Join(reason, " ")
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
	r := strings.Join(reason, " ")
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
	r := strings.Join(reason, " ")
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
	r := strings.Join(reason, " ")
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
	r := strings.Join(reason, " ")
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
	r := strings.Join(reason, " ")
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
	payload := rest.MessagePayload{
		Content:    res.Content,
		Embeds:     res.Embeds,
		Components: res.Components,
		TTS:        res.TTS,
		Files:      res.Files,
	}
	return a.ctx.REST.SendMessage(a.ctx.Context(), a.channelID, payload)
}

func (a ChannelSuite) Lock(reason ...string) error {
	return a.ctx.Channel(ChannelOptions{
		Type:      ChannelLock,
		ChannelID: a.channelID,
		Locked:    true,
		Reason:    strings.Join(reason, " "),
	})
}

func (a ChannelSuite) Unlock(reason ...string) error {
	return a.ctx.Channel(ChannelOptions{
		Type:      ChannelLock,
		ChannelID: a.channelID,
		Locked:    false,
		Reason:    strings.Join(reason, " "),
	})
}

func (a ChannelSuite) Purge(count int, reason ...string) (int, error) {
	return a.ctx.purgeMessages(a.channelID, count, strings.Join(reason, " "))
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
