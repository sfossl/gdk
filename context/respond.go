package context

import (
	"errors"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

var (
	ErrAlreadyResponded = errors.New("interaction already acknowledged")
	ErrNoInteraction    = errors.New("interaction payload missing")
)

type Response struct {
	Content    string
	Embeds     []types.Embed
	Components []types.Component
	Files      []types.File
	Ephemeral  bool
	TTS        bool
	Deferred   bool
}

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

	if c.REST != nil && c.Interaction != nil {
		err := c.REST.CreateInteractionResponse(c.Context(), c.ID, c.Interaction.Token, resp)
		if err != nil {
			c.deferred.Store(false)
			c.responded.Store(false)
			return err
		}
		return nil
	}
	return nil
}

func (c *Context) Respond(res Response) (*types.Message, error) {
	if res.Deferred {
		err := c.Defer(res.Ephemeral)
		return nil, err
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
		return msg, nil
	}

	if c.responded.Load() {
		if c.REST == nil {
			return nil, ErrNoRESTClient
		}
		if c.Interaction == nil {
			return nil, ErrNoInteraction
		}
		return c.REST.CreateFollowupMessage(c.Context(), c.AppID, c.Interaction.Token, rest.MessagePayload{
			Content:    res.Content,
			Embeds:     res.Embeds,
			Components: res.Components,
			Flags:      flags,
			TTS:        res.TTS,
			Files:      res.Files,
		})
	}

	if !c.responded.CompareAndSwap(false, true) {
		return c.Respond(res)
	}

	resp := types.InteractionResponse{
		Type: types.CallbackTypeChannelMessageWithSource,
		Data: &types.InteractionResponseData{
			Content:    res.Content,
			Embeds:     res.Embeds,
			Components: res.Components,
			Flags:      flags,
			TTS:        res.TTS,
			Files:      res.Files,
		},
	}

	if !c.gatewayMode && c.responseChan != nil {
		select {
		case c.responseChan <- resp:
			return nil, nil
		default:
		}
	}

	if c.REST != nil && c.Interaction != nil {
		err := c.REST.CreateInteractionResponse(c.Context(), c.ID, c.Interaction.Token, resp)
		if err != nil {
			c.responded.Store(false)
			return nil, err
		}
	}

	return nil, nil
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

	if c.REST != nil && c.Interaction != nil {
		err := c.REST.CreateInteractionResponse(c.Context(), c.ID, c.Interaction.Token, resp)
		if err != nil {
			c.responded.Store(false)
			return err
		}
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

	if c.REST != nil && c.Interaction != nil {
		err := c.REST.CreateInteractionResponse(c.Context(), c.ID, c.Interaction.Token, resp)
		if err != nil {
			c.responded.Store(false)
			return err
		}
	}

	return nil
}

