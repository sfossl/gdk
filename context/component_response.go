package context

import (
	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/component"
)

func (c *Context) DeferUpdate() error {
	if !c.state.CompareAndSwap(stateInitial, stateDeferred) {
		return ErrAlreadyResponded
	}

	resp := types.InteractionResponse{
		Type: types.CallbackTypeDeferredUpdateMessage,
	}
	return c.sendInteractionResponse(resp)
}

func (c *Context) UpdateMessage(args ...any) (*types.Message, error) {
	res := parseResponseOptions(args...)
	if res.Deferred {
		return nil, c.DeferUpdate()
	}

	if err := c.validateResponseOptions(res); err != nil {
		return nil, err
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
		return c.editOriginal(res, flags)
	}

	if !c.state.CompareAndSwap(stateInitial, stateResponded) {
		return c.UpdateMessage(res)
	}

	resp := types.InteractionResponse{
		Type: types.CallbackTypeUpdateMessage,
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
