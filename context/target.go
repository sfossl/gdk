package context

import (
	"github.com/sfossl/gda/types"
)

func (c *Context) TargetID() types.Snowflake {
	if c.Interaction == nil || c.Interaction.Data == nil {
		return 0
	}
	if sf := c.Opt("user").Snowflake(); sf != 0 {
		return sf
	}
	if sf := c.Opt("target").Snowflake(); sf != 0 {
		return sf
	}
	for _, opt := range c.Interaction.Data.Options {
		if opt.Type == 6 {
			if sf := c.Opt(opt.Name).Snowflake(); sf != 0 {
				return sf
			}
		}
	}
	return 0
}

func (c *Context) TargetUser() *types.User {
	if c.Interaction == nil || c.Interaction.Data == nil {
		return nil
	}
	if u := c.Opt("user").User(); u != nil {
		return u
	}
	if u := c.Opt("target").User(); u != nil {
		return u
	}
	for _, opt := range c.Interaction.Data.Options {
		if opt.Type == 6 {
			if u := c.Opt(opt.Name).User(); u != nil {
				return u
			}
		}
	}
	return nil
}

func (c *Context) TargetMember() *types.Member {
	if c.Interaction == nil || c.Interaction.Data == nil {
		return nil
	}
	if m := c.Opt("user").Member(); m != nil {
		return m
	}
	if m := c.Opt("target").Member(); m != nil {
		return m
	}
	for _, opt := range c.Interaction.Data.Options {
		if opt.Type == 6 {
			if m := c.Opt(opt.Name).Member(); m != nil {
				return m
			}
		}
	}
	return nil
}

func (c *Context) TargetUserID() types.Snowflake {
	if u := c.TargetUser(); u != nil {
		return u.ID
	}
	return c.TargetID()
}
