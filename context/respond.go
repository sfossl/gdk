package context

import (
	"errors"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/component"
)

var (
	ErrAlreadyResponded        = errors.New("interaction already acknowledged")
	ErrModalAfterDefer         = errors.New("cannot send modal after deferring")
	ErrNoInteraction           = errors.New("interaction payload missing")
	ErrEmptyMessage            = errors.New("cannot send an empty message: content, embeds, components, or files required")
	ErrInvalidModalCustomID    = errors.New("modal custom id must be between 1 and 100 characters")
	ErrInvalidModalTitle       = errors.New("modal title must be between 1 and 45 characters")
	ErrEmptyModalComponents    = errors.New("modal must contain at least 1 component")
	ErrInvalidModalInteraction = errors.New("cannot respond with a modal to ping, autocomplete, or modal submit interactions")
)

type Modal struct {
	CustomID   string
	Title      string
	Components []types.Component
}

type Response = Options

func (c *Context) Defer(ephemeral ...bool) error {
	if !c.state.CompareAndSwap(stateInitial, stateDeferred) {
		if c.state.Load() == stateDeferred {
			return nil
		}
		return ErrAlreadyResponded
	}

	isEphemeral := len(ephemeral) > 0 && ephemeral[0]
	resp := types.InteractionResponse{
		Type: types.CallbackTypeDeferredChannelMessageWithSource,
		Data: &types.InteractionResponseData{
			Flags: flagsFromEphemeral(isEphemeral),
		},
	}
	return c.sendInteractionResponse(resp)
}

func (c *Context) sendInteractionResponse(resp types.InteractionResponse) error {
	if !c.gatewayMode && c.responseChan != nil {
		select {
		case c.responseChan <- resp:
			return nil
		default:
			c.state.Store(stateInitial)
			return ErrAlreadyResponded
		}
	}

	if c.REST == nil {
		c.state.Store(stateInitial)
		return ErrNoRESTClient
	}
	if c.Interaction == nil {
		c.state.Store(stateInitial)
		return ErrNoInteraction
	}
	err := c.REST.CreateInteractionResponse(c.Context(), c.ID, c.Interaction.Token, resp)
	if err != nil {
		c.state.Store(stateInitial)
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
	if override.Reference == nil && base.Reference != nil {
		override.Reference = base.Reference
	}
	if override.Reply == 0 && base.Reply != 0 {
		override.Reply = base.Reply
	}
	if override.AllowedMentions == nil && base.AllowedMentions != nil {
		override.AllowedMentions = base.AllowedMentions
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
		case types.MessageReference:
			ref := v
			res.Reference = &ref
		case *types.MessageReference:
			res.Reference = v
		case types.AllowedMentions:
			ment := v
			res.AllowedMentions = &ment
		case *types.AllowedMentions:
			res.AllowedMentions = v
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

	if err := c.validateResponseOptions(res); err != nil {
		return nil, err
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

		dmFlags := flagsFromEphemeral(res.Ephemeral)
		if component.HasLayoutComponents(res.Components) {
			dmFlags |= types.MessageFlagIsComponentsV2
		}

		payload := rest.MessagePayload{
			Content:          res.Content,
			Embeds:           res.Embeds,
			Components:       res.Components,
			Flags:            dmFlags,
			TTS:              res.TTS,
			Files:            res.Files,
			AllowedMentions:  res.ResolveAllowedMentions(c.DefaultMention),
			MessageReference: res.ResolveReference(),
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

		sendFlags := flagsFromEphemeral(res.Ephemeral)
		if component.HasLayoutComponents(res.Components) {
			sendFlags |= types.MessageFlagIsComponentsV2
		}

		msg, err := c.REST.SendMessage(c.Context(), channelID, c.buildMessagePayload(res, sendFlags))
		if err != nil {
			return nil, err
		}

		return msg, nil
	}

	flags := flagsFromEphemeral(res.Ephemeral)
	if component.HasLayoutComponents(res.Components) {
		flags |= types.MessageFlagIsComponentsV2
	}

	if c.state.CompareAndSwap(stateDeferred, stateResponded) {
		msg, err := c.editOriginal(res, flags)
		if err != nil {
			c.state.Store(stateDeferred)
			return nil, err
		}
		return msg, nil
	}

	if c.state.Load() == stateResponded {
		if c.Interaction != nil {
			switch c.Interaction.Type {
			case types.InteractionTypePing, types.InteractionTypeApplicationCommandAutocomplete:
				return nil, ErrAlreadyResponded
			}
		}
		if c.REST == nil {
			return nil, ErrNoRESTClient
		}
		if c.Interaction == nil {
			return nil, ErrNoInteraction
		}
		msg, err := c.REST.CreateFollowupMessage(c.Context(), c.AppID, c.Interaction.Token, c.buildMessagePayload(res, flags))
		if err != nil {
			return nil, err
		}
		c.scheduleLifecycleTimers(res, msg)
		return msg, nil
	}

	if !c.state.CompareAndSwap(stateInitial, stateResponded) {
		return c.Respond(res)
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
			AllowedMentions: res.ResolveAllowedMentions(c.DefaultMention),
		},
	}

	err := c.sendInteractionResponse(resp)
	if err != nil {
		return nil, err
	}
	c.scheduleLifecycleTimers(res, nil)
	return nil, nil
}

func (c *Context) respondReceipt(receipt string) error {
	if c.Interaction == nil {
		return nil
	}
	if c.state.CompareAndSwap(stateDeferred, stateResponded) {
		if c.REST == nil {
			c.state.Store(stateDeferred)
			return ErrNoRESTClient
		}
		_, err := c.REST.EditOriginalInteractionResponse(c.Context(), c.AppID, c.Interaction.Token, rest.MessagePayload{
			Content: receipt,
			Flags:   types.MessageFlagEphemeral,
		})
		if err != nil {
			c.state.Store(stateDeferred)
			return err
		}
		return nil
	}

	if !c.state.CompareAndSwap(stateInitial, stateResponded) {
		return nil
	}

	resp := types.InteractionResponse{
		Type: types.CallbackTypeChannelMessageWithSource,
		Data: &types.InteractionResponseData{
			Content: receipt,
			Flags:   types.MessageFlagEphemeral,
		},
	}
	return c.sendInteractionResponse(resp)
}

func flagsFromEphemeral(ephemeral bool) types.MessageFlag {
	if ephemeral {
		return types.MessageFlagEphemeral
	}
	return 0
}

func (c *Context) Pong() error {
	if !c.state.CompareAndSwap(stateInitial, stateResponded) {
		return ErrAlreadyResponded
	}

	resp := types.InteractionResponse{
		Type: types.CallbackTypePong,
	}
	return c.sendInteractionResponse(resp)
}

func (c *Context) Autocomplete(choices []types.Choice) error {
	if len(choices) > 25 {
		choices = choices[:25]
	}
	if !c.state.CompareAndSwap(stateInitial, stateResponded) {
		return ErrAlreadyResponded
	}

	resp := types.InteractionResponse{
		Type: types.CallbackTypeApplicationCommandAutocompleteResult,
		Data: &types.InteractionResponseData{
			Choices: choices,
		},
	}
	return c.sendInteractionResponse(resp)
}

func (c *Context) Modal(m Modal) error {
	if c.Interaction != nil {
		switch c.Interaction.Type {
		case types.InteractionTypePing,
			types.InteractionTypeApplicationCommandAutocomplete,
			types.InteractionTypeModalSubmit:
			return ErrInvalidModalInteraction
		}
	}
	if c.state.Load() == stateDeferred {
		return ErrModalAfterDefer
	}
	if !c.state.CompareAndSwap(stateInitial, stateResponded) {
		return ErrAlreadyResponded
	}
	if c.gatewayMode && c.REST == nil {
		c.state.Store(stateInitial)
		return ErrNoRESTClient
	}
	if len(m.CustomID) == 0 || len(m.CustomID) > 100 {
		c.state.Store(stateInitial)
		return ErrInvalidModalCustomID
	}
	if len(m.Title) == 0 || len(m.Title) > 45 {
		c.state.Store(stateInitial)
		return ErrInvalidModalTitle
	}
	if len(m.Components) == 0 {
		c.state.Store(stateInitial)
		return ErrEmptyModalComponents
	}

	resp := types.InteractionResponse{
		Type: types.CallbackTypeModal,
		Data: &types.InteractionResponseData{
			CustomID:   m.CustomID,
			Title:      m.Title,
			Components: m.Components,
		},
	}
	return c.sendInteractionResponse(resp)
}

func (c *Context) validateResponseOptions(res Options) error {
	if res.Content == "" && len(res.Embeds) == 0 && len(res.Components) == 0 && len(res.Files) == 0 {
		return ErrEmptyMessage
	}
	if len(res.Components) > 0 {
		if err := component.ValidateComponents(res.Components); err != nil {
			return err
		}
		if component.HasLayoutComponents(res.Components) {
			if len(res.Embeds) > 0 {
				return component.ErrEmbedsNotAllowedWithLayouts
			}
			if res.Content != "" {
				return component.ErrContentNotAllowedWithLayouts
			}
		}
	}
	return nil
}

func (c *Context) buildMessagePayload(res Options, flags types.MessageFlag) rest.MessagePayload {
	return rest.MessagePayload{
		Content:          res.Content,
		Embeds:           res.Embeds,
		Components:       res.Components,
		Flags:            flags,
		TTS:              res.TTS,
		Files:            res.Files,
		AllowedMentions:  res.ResolveAllowedMentions(c.DefaultMention),
		MessageReference: res.ResolveReference(),
	}
}

func (c *Context) editOriginal(res Options, flags types.MessageFlag) (*types.Message, error) {
	if c.REST == nil {
		return nil, ErrNoRESTClient
	}
	if c.Interaction == nil {
		return nil, ErrNoInteraction
	}
	msg, err := c.REST.EditOriginalInteractionResponse(c.Context(), c.AppID, c.Interaction.Token, c.buildMessagePayload(res, flags))
	if err != nil {
		return nil, err
	}
	c.scheduleLifecycleTimers(res, msg)
	return msg, nil
}
