package context

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

type MessageAction int

const (
	MsgPin MessageAction = iota + 1
	MsgUnpin
	MsgDelete
	MsgCrosspost
)

var (
	ErrInvalidMessageID     = errors.New("invalid message id")
	ErrEmojiRequired        = errors.New("reaction emoji is required")
	ErrInvalidMessageAction = errors.New("invalid message action type")
)

type MessageOptions struct {
	Type      MessageAction
	MessageID types.Snowflake
	ChannelID types.Snowflake
	Reason    string
}

func parseMessageAction(v any) MessageAction {
	switch val := v.(type) {
	case MessageAction:
		return val
	case string:
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "pin":
			return MsgPin
		case "unpin":
			return MsgUnpin
		case "delete":
			return MsgDelete
		case "crosspost":
			return MsgCrosspost
		}
	}
	return 0
}

func (c *Context) ManageMessage(args ...any) error {
	if len(args) == 0 {
		return ErrInvalidMessageAction
	}

	opts := MessageOptions{}
	switch v := args[0].(type) {
	case MessageOptions:
		opts = v
	case *MessageOptions:
		if v != nil {
			opts = *v
		}
	case Options:
		opts = MessageOptions{
			Type:      parseMessageAction(v.Type),
			MessageID: v.ResolveMessage(c),
			ChannelID: v.ResolveChannel(c),
			Reason:    v.Reason,
		}
	case *Options:
		if v != nil {
			opts = MessageOptions{
				Type:      parseMessageAction(v.Type),
				MessageID: v.ResolveMessage(c),
				ChannelID: v.ResolveChannel(c),
				Reason:    v.Reason,
			}
		}
	case MessageAction:
		opts.Type = v
		for _, arg := range args[1:] {
			switch a := arg.(type) {
			case types.Snowflake:
				if opts.MessageID == 0 {
					opts.MessageID = a
				} else {
					opts.ChannelID = a
				}
			case string:
				opts.Reason = a
			}
		}
	case string:
		opts.Type = parseMessageAction(v)
		for _, arg := range args[1:] {
			switch a := arg.(type) {
			case types.Snowflake:
				if opts.MessageID == 0 {
					opts.MessageID = a
				} else {
					opts.ChannelID = a
				}
			case string:
				opts.Reason = a
			}
		}
	default:
		return ErrInvalidMessageAction
	}

	if opts.MessageID == 0 && c.Interaction != nil && c.Interaction.Data != nil && c.Interaction.Data.TargetID != nil {
		opts.MessageID = *c.Interaction.Data.TargetID
	}
	if opts.MessageID == 0 {
		return ErrInvalidMessageID
	}
	chID := opts.ChannelID
	if chID == 0 {
		chID = c.ChannelID
	}
	if chID == 0 {
		return ErrInvalidChannelID
	}
	if c.REST == nil {
		return ErrNoRESTClient
	}

	switch opts.Type {
	case MsgPin:
		return c.REST.PinMessage(c.Context(), chID, opts.MessageID, opts.Reason)
	case MsgUnpin:
		return c.REST.UnpinMessage(c.Context(), chID, opts.MessageID, opts.Reason)
	case MsgDelete:
		return c.REST.DeleteMessage(c.Context(), chID, opts.MessageID, opts.Reason)
	case MsgCrosspost:
		endpoint := rest.Endpoint{
			Method: "POST",
			Path:   fmt.Sprintf("/channels/%s/messages/%s/crosspost", chID, opts.MessageID),
			Key:    fmt.Sprintf("POST:/channels/%s/messages/crosspost", chID),
		}
		_, err := c.REST.Do(c.Context(), rest.Request{Endpoint: endpoint})
		return err
	default:
		return ErrInvalidMessageAction
	}
}

type MessageSuite struct {
	ctx       *Context
	messageID types.Snowflake
	channelID types.Snowflake
}

func (c *Context) Message(messageID types.Snowflake, channelID ...types.Snowflake) MessageSuite {
	chID := c.ChannelID
	if len(channelID) > 0 && channelID[0] != 0 {
		chID = channelID[0]
	}
	return MessageSuite{
		ctx:       c,
		messageID: messageID,
		channelID: chID,
	}
}

func (c *Context) TargetMessage() MessageSuite {
	return c.Message(c.TargetID())
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
	return a.ctx.ManageMessage(MessageOptions{
		Type:      MsgDelete,
		MessageID: a.messageID,
		ChannelID: a.channelID,
		Reason:    strings.Join(reason, " "),
	})
}

func (a MessageSuite) Pin(reason ...string) error {
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	return a.ctx.ManageMessage(MessageOptions{
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
	return a.ctx.ManageMessage(MessageOptions{
		Type:      MsgUnpin,
		MessageID: a.messageID,
		ChannelID: a.channelID,
		Reason:    strings.Join(reason, " "),
	})
}

func (a MessageSuite) React(emoji string) error {
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

func (a MessageSuite) RemoveReaction(emoji string, userID ...types.Snowflake) error {
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
