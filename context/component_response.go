package context

import (
	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

func (c *Context) DeferUpdate() error {
	if c.responded.Load() {
		return ErrAlreadyResponded
	}
	if !c.responded.CompareAndSwap(false, true) {
		return ErrAlreadyResponded
	}
	c.deferred.Store(true)

	resp := types.InteractionResponse{
		Type: types.CallbackTypeDeferredUpdateMessage,
	}
	return c.sendInteractionResponse(resp)
}

func (c *Context) UpdateMessage(args ...any) (*types.Message, error) {
	res := parseResponseOptions(args...)
	if res.Deferred {
		err := c.DeferUpdate()
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
		msg, err := c.REST.EditOriginalInteractionResponse(c.Context(), c.AppID, c.Interaction.Token, rest.MessagePayload{
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
		return c.UpdateMessage(res)
	}

	resp := types.InteractionResponse{
		Type: types.CallbackTypeUpdateMessage,
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

