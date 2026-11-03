package context

import (
	"strings"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

type MessageSuite struct {
	ctx       *Context
	messageID types.Snowflake
	channelID types.Snowflake
}

func (c *Context) MessageSuite(target ...types.Snowflake) MessageSuite {
	var msgID types.Snowflake
	chID := c.ChannelID
	if len(target) > 0 {
		msgID = target[0]
	}
	if len(target) > 1 && target[1] != 0 {
		chID = target[1]
	}
	return MessageSuite{
		ctx:       c,
		messageID: msgID,
		channelID: chID,
	}
}

func (c *Context) TargetMessage() MessageSuite {
	return c.MessageSuite(c.TargetID())
}

func (a MessageSuite) ID() types.Snowflake {
	return a.messageID
}

func (a MessageSuite) ChannelID() types.Snowflake {
	return a.channelID
}

func (a MessageSuite) Get() (*types.Message, error) {
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

func (a MessageSuite) Edit(payload rest.MessagePayload) (*types.Message, error) {
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

func (a MessageSuite) SetContent(content string) (*types.Message, error) {
	return a.Edit(rest.MessagePayload{Content: content})
}

func (a MessageSuite) SetEmbeds(embeds ...types.Embed) (*types.Message, error) {
	return a.Edit(rest.MessagePayload{Embeds: embeds})
}

func (a MessageSuite) SetComponents(components ...types.Component) (*types.Message, error) {
	return a.Edit(rest.MessagePayload{Components: components})
}

func (a MessageSuite) Delete(reason ...string) error {
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	return a.ctx.Message(MessageOptions{
		Type:      MsgDelete,
		MessageID: a.messageID,
		ChannelID: a.channelID,
		Reason:    strings.Join(reason, " "),
	})
}

func (a MessageSuite) Purge(count int, reason ...string) (int, error) {
	chID := a.channelID
	if chID == 0 {
		return 0, ErrInvalidChannelID
	}
	return a.ctx.purgeMessages(chID, count, strings.Join(reason, " "))
}

func (a MessageSuite) Pin(reason ...string) error {
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	return a.ctx.Message(MessageOptions{
		Type:      MsgPin,
		MessageID: a.messageID,
		ChannelID: a.channelID,
		Reason:    strings.Join(reason, " "),
	})
}

func (a MessageSuite) Unpin(reason ...string) error {
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	return a.ctx.Message(MessageOptions{
		Type:      MsgUnpin,
		MessageID: a.messageID,
		ChannelID: a.channelID,
		Reason:    strings.Join(reason, " "),
	})
}

func (a MessageSuite) Crosspost() error {
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	return a.ctx.Message(MessageOptions{
		Type:      MsgCrosspost,
		MessageID: a.messageID,
		ChannelID: a.channelID,
	})
}

func (a MessageSuite) React(emojis ...string) error {
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	var cleaned []string
	for _, e := range emojis {
		if c := cleanEmoji(e); c != "" {
			cleaned = append(cleaned, c)
		}
	}
	if len(cleaned) == 0 {
		return ErrEmojiRequired
	}
	return a.ctx.Message(MessageOptions{
		Type:      MsgReact,
		MessageID: a.messageID,
		ChannelID: a.channelID,
		Emojis:    cleaned,
	})
}

func (a MessageSuite) Unreact(emojis ...string) error {
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	var cleaned []string
	for _, e := range emojis {
		if c := cleanEmoji(e); c != "" {
			cleaned = append(cleaned, c)
		}
	}
	return a.ctx.Message(MessageOptions{
		Type:      MsgUnreact,
		MessageID: a.messageID,
		ChannelID: a.channelID,
		Emojis:    cleaned,
	})
}

func (a MessageSuite) RemoveReaction(emoji string, userID ...types.Snowflake) error {
	trimmed := cleanEmoji(emoji)
	if trimmed == "" {
		return ErrEmojiRequired
	}
	if a.messageID == 0 {
		return ErrInvalidMessageID
	}
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return ErrNoRESTClient
	}
	if len(userID) > 0 && userID[0] != 0 {
		return a.ctx.REST.DeleteUserReaction(a.ctx.Context(), a.channelID, a.messageID, trimmed, userID[0])
	}
	return a.Unreact(trimmed)
}

func (a MessageSuite) ClearReactions(emojis ...string) error {
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	var cleaned []string
	for _, e := range emojis {
		if c := cleanEmoji(e); c != "" {
			cleaned = append(cleaned, c)
		}
	}
	return a.ctx.Message(MessageOptions{
		Type:      MsgClearReactions,
		MessageID: a.messageID,
		ChannelID: a.channelID,
		Emojis:    cleaned,
	})
}
