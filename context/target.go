package context

import (
	"github.com/sfossl/gda/types"
)

func (c *Context) TargetID() types.Snowflake {
	if c.Interaction == nil || c.Interaction.Data == nil {
		return 0
	}
	if c.Interaction.Data.TargetID != nil && *c.Interaction.Data.TargetID != 0 {
		return *c.Interaction.Data.TargetID
	}
	if sf := c.Opt("user").Snowflake(); sf != 0 {
		return sf
	}
	if sf := c.Opt("target").Snowflake(); sf != 0 {
		return sf
	}
	for _, name := range findUserOptionNames(c.Interaction.Data.Options) {
		if sf := c.Opt(name).Snowflake(); sf != 0 {
			return sf
		}
	}
	return 0
}

func (c *Context) TargetUser() *types.User {
	if c.Interaction == nil || c.Interaction.Data == nil {
		return nil
	}
	if c.Interaction.Data.TargetID != nil && c.Interaction.Data.Resolved != nil {
		if u, ok := c.Interaction.Data.Resolved.Users[*c.Interaction.Data.TargetID]; ok {
			return &u
		}
	}
	if u := c.Opt("user").User(); u != nil {
		return u
	}
	if u := c.Opt("target").User(); u != nil {
		return u
	}
	for _, name := range findUserOptionNames(c.Interaction.Data.Options) {
		if u := c.Opt(name).User(); u != nil {
			return u
		}
	}
	return nil
}

func (c *Context) TargetMember() *types.Member {
	if c.Interaction == nil || c.Interaction.Data == nil {
		return nil
	}
	if c.Interaction.Data.TargetID != nil && c.Interaction.Data.Resolved != nil {
		if m, ok := c.Interaction.Data.Resolved.Members[*c.Interaction.Data.TargetID]; ok {
			return &m
		}
	}
	if m := c.Opt("user").Member(); m != nil {
		return m
	}
	if m := c.Opt("target").Member(); m != nil {
		return m
	}
	for _, name := range findUserOptionNames(c.Interaction.Data.Options) {
		if m := c.Opt(name).Member(); m != nil {
			return m
		}
	}
	return nil
}

func findUserOptionNames(opts []types.ApplicationCommandOption) []string {
	var names []string
	for _, opt := range opts {
		switch opt.Type {
		case int(types.OptionTypeSubcommand), int(types.OptionTypeSubcommandGroup):
			names = append(names, findUserOptionNames(opt.Options)...)
		case int(types.OptionTypeUser), int(types.OptionTypeMentionable):
			names = append(names, opt.Name)
		}
	}
	return names
}

func (c *Context) TargetUserID() types.Snowflake {
	if u := c.TargetUser(); u != nil {
		return u.ID
	}
	return c.TargetID()
}

func (c *Context) ResolvedMessage() *types.Message {
	if c.Interaction == nil || c.Interaction.Data == nil || c.Interaction.Data.TargetID == nil || c.Interaction.Data.Resolved == nil {
		return nil
	}
	if msg, ok := c.Interaction.Data.Resolved.Messages[*c.Interaction.Data.TargetID]; ok {
		return &msg
	}
	return nil
}
