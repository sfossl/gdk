package context

import (
	"net/http"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

type UpdateOptions struct {
	Target       types.Snowflake
	Nickname     *string
	AddRoles     []types.Snowflake
	RemoveRoles  []types.Snowflake
	VoiceMute    *bool
	VoiceDeaf    *bool
	VoiceChannel *types.Snowflake
	Reason       string
}

func (c *Context) Update(args ...any) error {
	if c.GuildID == nil || *c.GuildID == 0 {
		return ErrNotInGuild
	}
	if c.REST == nil {
		return ErrNoRESTClient
	}
	if len(args) == 0 {
		return nil
	}

	opts := UpdateOptions{}
	switch v := args[0].(type) {
	case UpdateOptions:
		opts = v
	case *UpdateOptions:
		if v != nil {
			opts = *v
		}
	case Options:
		opts = UpdateOptions{
			Target:       v.ResolveTarget(c),
			Nickname:     v.Nickname,
			AddRoles:     v.AddRoles,
			RemoveRoles:  v.RemoveRoles,
			VoiceMute:    v.VoiceMute,
			VoiceDeaf:    v.VoiceDeaf,
			VoiceChannel: v.VoiceChannel,
			Reason:       v.Reason,
		}
	case *Options:
		if v != nil {
			opts = UpdateOptions{
				Target:       v.ResolveTarget(c),
				Nickname:     v.Nickname,
				AddRoles:     v.AddRoles,
				RemoveRoles:  v.RemoveRoles,
				VoiceMute:    v.VoiceMute,
				VoiceDeaf:    v.VoiceDeaf,
				VoiceChannel: v.VoiceChannel,
				Reason:       v.Reason,
			}
		}
	case types.Snowflake:
		opts.Target = v
		for _, arg := range args[1:] {
			switch a := arg.(type) {
			case string:
				opts.Nickname = &a
			case []types.Snowflake:
				opts.AddRoles = a
			case types.Snowflake:
				opts.AddRoles = append(opts.AddRoles, a)
			}
		}
	}

	target := opts.Target
	if target == 0 {
		target = c.TargetID()
	}
	if target == 0 {
		return ErrInvalidTarget
	}

	if opts.Nickname == nil && len(opts.AddRoles) == 0 && len(opts.RemoveRoles) == 0 &&
		opts.VoiceMute == nil && opts.VoiceDeaf == nil && opts.VoiceChannel == nil {
		return nil
	}

	payload := rest.ModifyMemberPayload{
		Nick:      opts.Nickname,
		Mute:      opts.VoiceMute,
		Deaf:      opts.VoiceDeaf,
		ChannelID: opts.VoiceChannel,
	}

	if len(opts.AddRoles) > 0 || len(opts.RemoveRoles) > 0 {
		var currentRoles []types.Snowflake
		found := false
		if c.Interaction != nil && c.Interaction.Data != nil && c.Interaction.Data.Resolved != nil && c.Interaction.Data.Resolved.Members != nil {
			if tm, ok := c.Interaction.Data.Resolved.Members[target]; ok {
				currentRoles = tm.Roles
				found = true
			}
		}
		if !found && c.Member != nil && c.Author != nil && target == c.Author.ID {
			currentRoles = c.Member.Roles
			found = true
		}
		if !found {
			m, err := c.REST.GetGuildMember(c.Context(), *c.GuildID, target)
			if err != nil {
				return err
			}
			currentRoles = m.Roles
		}

		roleMap := make(map[types.Snowflake]bool, len(currentRoles)+len(opts.AddRoles))
		for _, r := range currentRoles {
			roleMap[r] = true
		}
		for _, r := range opts.RemoveRoles {
			delete(roleMap, r)
		}
		for _, r := range opts.AddRoles {
			roleMap[r] = true
		}
		newRoles := make([]types.Snowflake, 0, len(roleMap))
		for r := range roleMap {
			newRoles = append(newRoles, r)
		}
		payload.Roles = &newRoles
	}

	var headers http.Header
	if opts.Reason != "" {
		headers = http.Header{"X-Audit-Log-Reason": []string{opts.Reason}}
	}

	_, err := c.REST.Do(c.Context(), rest.Request{
		Endpoint: rest.EndpointModifyGuildMember(*c.GuildID, target),
		Body:     payload,
		Headers:  headers,
	})
	return err
}

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

func (c *Context) UpdateMessage(res Response) (*types.Message, error) {
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
		return msg, nil
	}

	if c.responded.Load() {
		if c.REST == nil {
			return nil, ErrNoRESTClient
		}
		if c.Interaction == nil {
			return nil, ErrNoInteraction
		}
		return c.REST.EditOriginalInteractionResponse(c.Context(), c.AppID, c.Interaction.Token, rest.MessagePayload{
			Content:    res.Content,
			Embeds:     res.Embeds,
			Components: res.Components,
			Flags:      flags,
			TTS:        res.TTS,
			Files:      res.Files,
		})
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

	return nil, nil
}
