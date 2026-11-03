package context

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

type MessageAction int

const (
	MsgPin MessageAction = iota + 1
	MsgUnpin
	MsgDelete
	MsgCrosspost
	MsgReact
	MsgUnreact
	MsgClearReactions
	MsgEdit
	MsgPurge
)

var (
	ErrInvalidMessageID     = errors.New("invalid message id")
	ErrEmojiRequired        = errors.New("reaction emoji is required")
	ErrInvalidMessageAction = errors.New("invalid message action type")
	ErrInvalidPurgeCount    = errors.New("purge count must be between 2 and 100")
	ErrMessagesTooOld       = errors.New("messages are older than 14 days and cannot be purged")
)

type MessageOptions struct {
	Type       MessageAction
	MessageID  types.Snowflake
	ChannelID  types.Snowflake
	Reason     string
	Emojis     []string
	Content    string
	Count      int
	Embeds     []types.Embed
	Components []types.Component
	Files      []types.File
}

func cleanEmoji(e string) string {
	e = strings.TrimSpace(e)
	if strings.HasPrefix(e, "<a:") && strings.HasSuffix(e, ">") {
		return strings.TrimSuffix(strings.TrimPrefix(e, "<a:"), ">")
	}
	if strings.HasPrefix(e, "<:") && strings.HasSuffix(e, ">") {
		return strings.TrimSuffix(strings.TrimPrefix(e, "<:"), ">")
	}
	return e
}

func parseEmojis(v any) []string {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		if cleaned := cleanEmoji(val); cleaned != "" {
			return []string{cleaned}
		}
	case []string:
		var res []string
		for _, s := range val {
			if cleaned := cleanEmoji(s); cleaned != "" {
				res = append(res, cleaned)
			}
		}
		return res
	}
	return nil
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
		case "crosspost", "publish":
			return MsgCrosspost
		case "react", "reaction", "addreaction":
			return MsgReact
		case "unreact", "removereaction":
			return MsgUnreact
		case "clearreactions", "clearreaction":
			return MsgClearReactions
		case "edit":
			return MsgEdit
		case "purge":
			return MsgPurge
		}
	}
	return 0
}

func messageOptionsFromOptions(v Options, c *Context) MessageOptions {
	opts := MessageOptions{
		Type:       parseMessageAction(v.Type),
		MessageID:  v.ResolveMessage(c),
		ChannelID:  v.ResolveChannel(c),
		Reason:     v.Reason,
		Emojis:     parseEmojis(v.Emoji),
		Content:    v.Content,
		Count:      v.Count,
		Embeds:     v.Embeds,
		Components: v.Components,
		Files:      v.Files,
	}
	return opts
}

func (c *Context) Message(args ...any) error {
	if len(args) == 0 {
		return ErrInvalidMessageAction
	}

	opts := MessageOptions{}
	parseRest := false

	switch v := args[0].(type) {
	case MessageOptions:
		opts = v
	case *MessageOptions:
		if v != nil {
			opts = *v
		}
	case Options:
		opts = messageOptionsFromOptions(v, c)
	case *Options:
		if v != nil {
			opts = messageOptionsFromOptions(*v, c)
		}
	case MessageAction:
		opts.Type = v
		parseRest = true
	case types.Snowflake:
		opts.MessageID = v
		parseRest = true
	case int:
		opts.MessageID = types.Snowflake(v)
		parseRest = true
	case int64:
		opts.MessageID = types.Snowflake(v)
		parseRest = true
	case string:
		action := parseMessageAction(v)
		if action != 0 {
			opts.Type = action
		} else {
			opts.Reason = v
		}
		parseRest = true
	default:
		return ErrInvalidMessageAction
	}

	if parseRest {
		for _, arg := range args[1:] {
			switch a := arg.(type) {
			case types.Snowflake:
				if opts.MessageID == 0 {
					opts.MessageID = a
				} else {
					opts.ChannelID = a
				}
			case int:
				if opts.Type == MsgPurge && opts.Count == 0 {
					opts.Count = a
				} else if opts.MessageID == 0 {
					opts.MessageID = types.Snowflake(a)
				} else {
					opts.ChannelID = types.Snowflake(a)
				}
			case int64:
				if opts.Type == MsgPurge && opts.Count == 0 {
					opts.Count = int(a)
				} else if opts.MessageID == 0 {
					opts.MessageID = types.Snowflake(a)
				} else {
					opts.ChannelID = types.Snowflake(a)
				}
			case MessageAction:
				opts.Type = a
			case string:
				action := parseMessageAction(a)
				if action != 0 && opts.Type == 0 {
					opts.Type = action
				} else if opts.Type == MsgReact || opts.Type == MsgUnreact || opts.Type == MsgClearReactions {
					if cleaned := cleanEmoji(a); cleaned != "" {
						opts.Emojis = append(opts.Emojis, cleaned)
					}
				} else if opts.Type == MsgEdit && opts.Content == "" {
					opts.Content = a
				} else {
					opts.Reason = a
				}
			case []string:
				if opts.Type == MsgReact || opts.Type == MsgUnreact || opts.Type == MsgClearReactions {
					opts.Emojis = append(opts.Emojis, parseEmojis(a)...)
				}
			case types.Embed:
				opts.Embeds = append(opts.Embeds, a)
			case []types.Embed:
				opts.Embeds = append(opts.Embeds, a...)
			case types.Component:
				opts.Components = append(opts.Components, a)
			case []types.Component:
				opts.Components = append(opts.Components, a...)
			}
		}
	}

	if opts.Type != MsgPurge {
		if opts.MessageID == 0 && c.Interaction != nil && c.Interaction.Data != nil && c.Interaction.Data.TargetID != nil {
			opts.MessageID = *c.Interaction.Data.TargetID
		}
		if opts.MessageID == 0 {
			return ErrInvalidMessageID
		}
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
	case MsgReact:
		if len(opts.Emojis) == 0 {
			return ErrEmojiRequired
		}
		for _, emoji := range opts.Emojis {
			if err := c.REST.CreateReaction(c.Context(), chID, opts.MessageID, emoji); err != nil {
				return err
			}
		}
		return nil
	case MsgUnreact:
		if len(opts.Emojis) == 0 {
			return ErrEmojiRequired
		}
		for _, emoji := range opts.Emojis {
			if err := c.REST.DeleteOwnReaction(c.Context(), chID, opts.MessageID, emoji); err != nil {
				return err
			}
		}
		return nil
	case MsgClearReactions:
		if len(opts.Emojis) > 0 {
			for _, emoji := range opts.Emojis {
				endpoint := rest.Endpoint{
					Method: "DELETE",
					Path:   fmt.Sprintf("/channels/%s/messages/%s/reactions/%s", chID, opts.MessageID, url.PathEscape(emoji)),
					Key:    fmt.Sprintf("DELETE:/channels/%s/messages/reactions", chID),
				}
				if _, err := c.REST.Do(c.Context(), rest.Request{Endpoint: endpoint}); err != nil {
					return err
				}
			}
			return nil
		}
		endpoint := rest.Endpoint{
			Method: "DELETE",
			Path:   fmt.Sprintf("/channels/%s/messages/%s/reactions", chID, opts.MessageID),
			Key:    fmt.Sprintf("DELETE:/channels/%s/messages/reactions", chID),
		}
		_, err := c.REST.Do(c.Context(), rest.Request{Endpoint: endpoint})
		return err
	case MsgEdit:
		payload := rest.MessagePayload{
			Content:    opts.Content,
			Embeds:     opts.Embeds,
			Components: opts.Components,
			Files:      opts.Files,
		}
		_, err := c.REST.EditMessage(c.Context(), chID, opts.MessageID, payload)
		return err
	case MsgPurge:
		_, err := c.purgeMessages(chID, opts.Count, opts.Reason)
		return err
	default:
		return ErrInvalidMessageAction
	}
}

func (c *Context) Purge(count int, reason ...string) (int, error) {
	var r string
	if len(reason) > 0 {
		r = reason[0]
	}
	return c.purgeMessages(c.ChannelID, count, r)
}

func (c *Context) purgeMessages(chID types.Snowflake, count int, reason string) (int, error) {
	if count < 2 || count > 100 {
		return 0, ErrInvalidPurgeCount
	}
	if chID == 0 {
		chID = c.ChannelID
	}
	if chID == 0 {
		return 0, ErrInvalidChannelID
	}
	if c.REST == nil {
		return 0, ErrNoRESTClient
	}
	endpoint := rest.Endpoint{
		Method: "GET",
		Path:   fmt.Sprintf("/channels/%s/messages?limit=%d", chID, count),
		Key:    fmt.Sprintf("GET:/channels/%s/messages", chID),
	}
	data, err := c.REST.Do(c.Context(), rest.Request{Endpoint: endpoint})
	if err != nil {
		return 0, err
	}
	var msgs []types.Message
	if err := json.Unmarshal(data, &msgs); err != nil {
		return 0, err
	}
	if len(msgs) == 0 {
		return 0, nil
	}
	cutoff := time.Now().Add(-14 * 24 * time.Hour)
	var validIDs []types.Snowflake
	for _, m := range msgs {
		if m.ID>>22 == 0 || m.ID.Time().After(cutoff) {
			validIDs = append(validIDs, m.ID)
		}
	}
	if len(validIDs) == 0 {
		return 0, ErrMessagesTooOld
	}
	if len(validIDs) == 1 {
		if err := c.REST.DeleteMessage(c.Context(), chID, validIDs[0], reason); err != nil {
			return 0, err
		}
		return 1, nil
	}
	bulkEndpoint := rest.Endpoint{
		Method: "POST",
		Path:   fmt.Sprintf("/channels/%s/messages/bulk-delete", chID),
		Key:    fmt.Sprintf("POST:/channels/%s/messages/bulk-delete", chID),
	}
	bulkPayload := struct {
		Messages []types.Snowflake `json:"messages"`
	}{
		Messages: validIDs,
	}
	req := rest.Request{
		Endpoint: bulkEndpoint,
		Body:     bulkPayload,
	}
	if reason != "" {
		req.Headers = http.Header{"X-Audit-Log-Reason": []string{reason}}
	}
	if _, err := c.REST.Do(c.Context(), req); err != nil {
		return 0, err
	}
	return len(validIDs), nil
}

func (c *Context) ManageMessage(args ...any) error {
	return c.Message(args...)
}
