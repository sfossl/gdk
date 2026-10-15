package context

import (
	"errors"
	"strings"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

var (
	ErrInvalidMessageID = errors.New("invalid message id")
	ErrEmojiRequired    = errors.New("reaction emoji is required")
)

type MessageAction struct {
	ctx       *Context
	messageID types.Snowflake
	channelID types.Snowflake
}

func (c *Context) Message(messageID types.Snowflake, channelID ...types.Snowflake) MessageAction {
	chID := c.ChannelID
	if len(channelID) > 0 && channelID[0] != 0 {
		chID = channelID[0]
	}
	return MessageAction{
		ctx:       c,
		messageID: messageID,
		channelID: chID,
	}
}

func (c *Context) ManageMessage(messageID types.Snowflake, channelID ...types.Snowflake) MessageAction {
	return c.Message(messageID, channelID...)
}

func (c *Context) TargetMessage() MessageAction {
	var targetID types.Snowflake
	if c.Interaction != nil && c.Interaction.Data != nil && c.Interaction.Data.TargetID != nil {
		targetID = *c.Interaction.Data.TargetID
	}
	return c.Message(targetID)
}

func (a MessageAction) ID() types.Snowflake {
	return a.messageID
}

func (a MessageAction) ChannelID() types.Snowflake {
	return a.channelID
}

func (a MessageAction) Get() (*types.Message, error) {
	if a.messageID == 0 {
		return nil, ErrInvalidMessageID
	}
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	return a.ctx.REST.GetMessage(a.ctx.Context(), a.channelID, a.messageID)
}

func (a MessageAction) Edit(payload rest.MessagePayload) (*types.Message, error) {
	if a.messageID == 0 {
		return nil, ErrInvalidMessageID
	}
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	return a.ctx.REST.EditMessage(a.ctx.Context(), a.channelID, a.messageID, payload)
}

func (a MessageAction) SetContent(content string) (*types.Message, error) {
	return a.Edit(rest.MessagePayload{Content: content})
}

func (a MessageAction) SetEmbeds(embeds ...types.Embed) (*types.Message, error) {
	return a.Edit(rest.MessagePayload{Embeds: embeds})
}

func (a MessageAction) SetComponents(components ...types.Component) (*types.Message, error) {
	return a.Edit(rest.MessagePayload{Components: components})
}

func (a MessageAction) Delete(reason ...string) error {
	if a.messageID == 0 {
		return ErrInvalidMessageID
	}
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return ErrNoRESTClient
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.DeleteMessage(a.ctx.Context(), a.channelID, a.messageID, r)
}

func (a MessageAction) Pin(reason ...string) error {
	if a.messageID == 0 {
		return ErrInvalidMessageID
	}
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return ErrNoRESTClient
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.PinMessage(a.ctx.Context(), a.channelID, a.messageID, r)
}

func (a MessageAction) Unpin(reason ...string) error {
	if a.messageID == 0 {
		return ErrInvalidMessageID
	}
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return ErrNoRESTClient
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.UnpinMessage(a.ctx.Context(), a.channelID, a.messageID, r)
}

func (a MessageAction) React(emoji string) error {
	if a.messageID == 0 {
		return ErrInvalidMessageID
	}
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return ErrNoRESTClient
	}
	trimmed := strings.TrimSpace(emoji)
	if trimmed == "" {
		return ErrEmojiRequired
	}
	return a.ctx.REST.CreateReaction(a.ctx.Context(), a.channelID, a.messageID, trimmed)
}

func (a MessageAction) RemoveReaction(emoji string, userID ...types.Snowflake) error {
	if a.messageID == 0 {
		return ErrInvalidMessageID
	}
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return ErrNoRESTClient
	}
	trimmed := strings.TrimSpace(emoji)
	if trimmed == "" {
		return ErrEmojiRequired
	}
	if len(userID) > 0 && userID[0] != 0 {
		return a.ctx.REST.DeleteUserReaction(a.ctx.Context(), a.channelID, a.messageID, trimmed, userID[0])
	}
	return a.ctx.REST.DeleteOwnReaction(a.ctx.Context(), a.channelID, a.messageID, trimmed)
}
