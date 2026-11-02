package context

import (
	"errors"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

var (
	ErrAlreadyResponded = errors.New("interaction already acknowledged")
	ErrModalAfterDefer  = errors.New("cannot send modal after deferring")
	ErrNoInteraction    = errors.New("interaction payload missing")
	ErrEmptyMessage     = errors.New("cannot send an empty message: content, embeds, components, or files required")
)

type Modal struct {
	CustomID   string
	Title      string
	Components []types.Component
}

type Response = Options

func (c *Context) Defer(ephemeral ...bool) error {
	if c.responded.Load() {
		return ErrAlreadyResponded
	}
	if !c.responded.CompareAndSwap(false, true) {
		return ErrAlreadyResponded
	}
	c.deferred.Store(true)

	isEphemeral := len(ephemeral) > 0 && ephemeral[0]
	resp := types.InteractionResponse{
		Type: types.CallbackTypeDeferredChannelMessageWithSource,
		Data: &types.InteractionResponseData{
			Flags: flagsFromEphemeral(isEphemeral),
		},
	}

	if !c.gatewayMode && c.responseChan != nil {
		select {
		case c.responseChan <- resp:
			return nil
		default:
		}
	}

	if c.REST == nil {
		c.deferred.Store(false)
		c.responded.Store(false)
		return ErrNoRESTClient
	}
	if c.Interaction == nil {
		c.deferred.Store(false)
		c.responded.Store(false)
		return ErrNoInteraction
	}
	err := c.REST.CreateInteractionResponse(c.Context(), c.ID, c.Interaction.Token, resp)
	if err != nil {
		c.deferred.Store(false)
		c.responded.Store(false)
		return err
	}
	return nil
}

func mergeOptions(base, override Options) Options {
	if override.Content == "" {
		override.Content = base.Content
	}
	if !override.Ephemeral {
		override.Ephemeral = base.Ephemeral
	}
	if len(override.Embeds) == 0 {
		override.Embeds = base.Embeds
	}
	if len(override.Components) == 0 {
		override.Components = base.Components
	}
	if len(override.Files) == 0 {
		override.Files = base.Files
	}
	if override.ComponentTimeout == nil && base.ComponentTimeout != nil {
		override.ComponentTimeout = base.ComponentTimeout
	}
	if override.DeleteAfter == nil && base.DeleteAfter != nil {
		override.DeleteAfter = base.DeleteAfter
	}
	return override
}

func parseResponseOptions(args ...any) Options {
	res := Options{}
	for _, arg := range args {
		switch v := arg.(type) {
		case Options:
			res = mergeOptions(res, v)
		case *Options:
			if v != nil {
				res = mergeOptions(res, *v)
			}
		case string:
			res.Content = v
		case bool:
			res.Ephemeral = v
		case error:
			res.Content = v.Error()
			res.Ephemeral = true
		case types.Embed:
			res.Embeds = append(res.Embeds, v)
		case *types.Embed:
			if v != nil {
				res.Embeds = append(res.Embeds, *v)
			}
		case []types.Embed:
			res.Embeds = append(res.Embeds, v...)
		case types.Component:
			res.Components = append(res.Components, v)
		case []types.Component:
			res.Components = append(res.Components, v...)
		case types.File:
			res.Files = append(res.Files, v)
		case []types.File:
			res.Files = append(res.Files, v...)
		}
	}
	if res.Embed != nil {
		switch e := res.Embed.(type) {
		case types.Embed:
			res.Embeds = append(res.Embeds, e)
		case *types.Embed:
			if e != nil {
				res.Embeds = append(res.Embeds, *e)
			}
		}
	}
	return res
}

func (c *Context) Respond(args ...any) (*types.Message, error) {
	res := parseResponseOptions(args...)
	if res.Deferred {
		err := c.Defer(res.Ephemeral)
		return nil, err
	}

	if res.Content == "" && len(res.Embeds) == 0 && len(res.Components) == 0 && len(res.Files) == 0 {
		return nil, ErrEmptyMessage
	}

	if res.Typing && c.REST != nil {
		targetChan := c.ChannelID
		if res.Channel != 0 {
			targetChan = res.Channel
		}
		if targetChan != 0 {
			_ = c.REST.TriggerTyping(c.Context(), targetChan)
		}
	}

	if res.ResolveScope() == "dm" {
		target := res.Target
		if target == 0 {
			target = c.TargetID()
		}
		if target == 0 && c.Author != nil {
			target = c.Author.ID
		}
		if target == 0 {
			return nil, ErrInvalidTarget
		}
		if c.REST == nil {
			return nil, ErrNoRESTClient
		}

		dmChan, err := c.REST.CreateDM(c.Context(), target)
		if err != nil {
			return nil, err
		}

		payload := rest.MessagePayload{
			Content:    res.Content,
			Embeds:     res.Embeds,
			Components: res.Components,
			TTS:        res.TTS,
			Files:      res.Files,
		}
		msg, err := c.REST.SendMessage(c.Context(), dmChan.ID, payload)
		if err != nil {
			return nil, err
		}

		_ = c.respondReceipt(DefaultDMReceipt)
		return msg, nil
	}

	if res.ResolveType() == "send" || (res.Channel != 0 && res.Channel != c.ChannelID) {
		channelID := res.Channel
		if channelID == 0 {
			channelID = c.ChannelID
		}
		if channelID == 0 {
			return nil, ErrInvalidChannelID
		}
		if c.REST == nil {
			return nil, ErrNoRESTClient
		}

		payload := rest.MessagePayload{
			Content:    res.Content,
			Embeds:     res.Embeds,
			Components: res.Components,
			TTS:        res.TTS,
			Files:      res.Files,
		}
		msg, err := c.REST.SendMessage(c.Context(), channelID, payload)
		if err != nil {
			return nil, err
		}

		return msg, nil
	}

	flags := flagsFromEphemeral(res.Ephemeral)

	if c.deferred.CompareAndSwap(true, false) {
		if c.REST == nil {
			c.deferred.Store(true)
			return nil, ErrNoRESTClient
		}
		if c.Interaction == nil {
			c.deferred.Store(true)
			return nil, ErrNoInteraction
		}
		msg, err := c.REST.EditOriginalInteractionResponse(c.Context(), c.AppID, c.Interaction.Token, rest.MessagePayload{
			Content:    res.Content,
			Embeds:     res.Embeds,
			Components: res.Components,
			Flags:      flags,
			TTS:        res.TTS,
			Files:      res.Files,
		})
		if err != nil {
			c.deferred.Store(true)
			return nil, err
		}
		c.scheduleLifecycleTimers(res, msg)
		return msg, nil
	}

	if c.responded.Load() {
		if c.REST == nil {
			return nil, ErrNoRESTClient
		}
		if c.Interaction == nil {
			return nil, ErrNoInteraction
		}
		msg, err := c.REST.CreateFollowupMessage(c.Context(), c.AppID, c.Interaction.Token, rest.MessagePayload{
			Content:    res.Content,
			Embeds:     res.Embeds,
			Components: res.Components,
			Flags:      flags,
			TTS:        res.TTS,
			Files:      res.Files,
		})
		if err != nil {
			return nil, err
		}
		c.scheduleLifecycleTimers(res, msg)
		return msg, nil
	}

	if !c.responded.CompareAndSwap(false, true) {
		return c.Respond(res)
	}

	var allowedMentions any
	if res.Mention != nil {
		allowedMentions = map[string]any{
			"replied_user": *res.Mention,
		}
	}

	resp := types.InteractionResponse{
		Type: types.CallbackTypeChannelMessageWithSource,
		Data: &types.InteractionResponseData{
			Content:         res.Content,
			Embeds:          res.Embeds,
			Components:      res.Components,
			Flags:           flags,
			TTS:             res.TTS,
			Files:           res.Files,
			AllowedMentions: allowedMentions,
		},
	}

	if !c.gatewayMode && c.responseChan != nil {
		select {
		case c.responseChan <- resp:
			c.scheduleLifecycleTimers(res, nil)
			return nil, nil
		default:
		}
	}

	if c.REST == nil {
		c.responded.Store(false)
		return nil, ErrNoRESTClient
	}
	if c.Interaction == nil {
		c.responded.Store(false)
		return nil, ErrNoInteraction
	}
	err := c.REST.CreateInteractionResponse(c.Context(), c.ID, c.Interaction.Token, resp)
	if err != nil {
		c.responded.Store(false)
		return nil, err
	}
	c.scheduleLifecycleTimers(res, nil)
	return nil, nil
}

func (c *Context) respondReceipt(receipt string) error {
	if c.Interaction == nil {
		return nil
	}
	if c.deferred.CompareAndSwap(true, false) {
		if c.REST == nil {
			c.deferred.Store(true)
			return ErrNoRESTClient
		}
		_, err := c.REST.EditOriginalInteractionResponse(c.Context(), c.AppID, c.Interaction.Token, rest.MessagePayload{
			Content: receipt,
			Flags:   types.MessageFlagEphemeral,
		})
		if err != nil {
			c.deferred.Store(true)
			return err
		}
		return nil
	}

	if c.responded.Load() {
		return nil
	}

	if !c.responded.CompareAndSwap(false, true) {
		return nil
	}

	resp := types.InteractionResponse{
		Type: types.CallbackTypeChannelMessageWithSource,
		Data: &types.InteractionResponseData{
			Content: receipt,
			Flags:   types.MessageFlagEphemeral,
		},
	}

	if !c.gatewayMode && c.responseChan != nil {
		select {
		case c.responseChan <- resp:
			return nil
		default:
		}
	}

	if c.REST == nil {
		c.responded.Store(false)
		return ErrNoRESTClient
	}
	err := c.REST.CreateInteractionResponse(c.Context(), c.ID, c.Interaction.Token, resp)
	if err != nil {
		c.responded.Store(false)
		return err
	}
	return nil
}

func flagsFromEphemeral(ephemeral bool) types.MessageFlag {
	if ephemeral {
		return types.MessageFlagEphemeral
	}
	return 0
}

func (c *Context) Pong() error {
	if c.responded.Load() {
		return ErrAlreadyResponded
	}
	if !c.responded.CompareAndSwap(false, true) {
		return ErrAlreadyResponded
	}

	resp := types.InteractionResponse{
		Type: types.CallbackTypePong,
	}

	if !c.gatewayMode && c.responseChan != nil {
		select {
		case c.responseChan <- resp:
			return nil
		default:
		}
	}

	if c.REST == nil {
		c.responded.Store(false)
		return ErrNoRESTClient
	}
	if c.Interaction == nil {
		c.responded.Store(false)
		return ErrNoInteraction
	}
	err := c.REST.CreateInteractionResponse(c.Context(), c.ID, c.Interaction.Token, resp)
	if err != nil {
		c.responded.Store(false)
		return err
	}
	return nil
}

func (c *Context) Autocomplete(choices []types.Choice) error {
	if c.responded.Load() {
		return ErrAlreadyResponded
	}
	if !c.responded.CompareAndSwap(false, true) {
		return ErrAlreadyResponded
	}

	resp := types.InteractionResponse{
		Type: types.CallbackTypeApplicationCommandAutocompleteResult,
		Data: &types.InteractionResponseData{
			Choices: choices,
		},
	}

	if !c.gatewayMode && c.responseChan != nil {
		select {
		case c.responseChan <- resp:
			return nil
		default:
		}
	}

	if c.REST == nil {
		c.responded.Store(false)
		return ErrNoRESTClient
	}
	if c.Interaction == nil {
		c.responded.Store(false)
		return ErrNoInteraction
	}
	err := c.REST.CreateInteractionResponse(c.Context(), c.ID, c.Interaction.Token, resp)
	if err != nil {
		c.responded.Store(false)
		return err
	}
	return nil
}

func (c *Context) Modal(m Modal) error {
	if c.deferred.Load() {
		return ErrModalAfterDefer
	}
	if c.responded.Load() {
		return ErrAlreadyResponded
	}
	if !c.responded.CompareAndSwap(false, true) {
		return ErrAlreadyResponded
	}

	resp := types.InteractionResponse{
		Type: types.CallbackTypeModal,
		Data: &types.InteractionResponseData{
			CustomID:   m.CustomID,
			Title:      m.Title,
			Components: m.Components,
		},
	}

	if !c.gatewayMode && c.responseChan != nil {
		select {
		case c.responseChan <- resp:
			return nil
		default:
		}
	}

	if c.REST == nil {
		c.responded.Store(false)
		return ErrNoRESTClient
	}
	if c.Interaction == nil {
		c.responded.Store(false)
		return ErrNoInteraction
	}
	err := c.REST.CreateInteractionResponse(c.Context(), c.ID, c.Interaction.Token, resp)
	if err != nil {
		c.responded.Store(false)
		return err
	}
	return nil
}
