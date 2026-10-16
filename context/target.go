package context

import (
	"github.com/sfossl/gda/types"
)

func (c *Context) CommandType() int {
	if c.Interaction == nil || c.Interaction.Data == nil {
		return 0
	}
	if c.Interaction.Data.Type == 0 {
		return 1
	}
	return c.Interaction.Data.Type
}

func (c *Context) IsSlash() bool {
	return c.CommandType() == 1
}

func (c *Context) IsUserContext() bool {
	return c.CommandType() == 2
}

func (c *Context) IsMessageContext() bool {
	return c.CommandType() == 3
}

func (c *Context) TargetID() types.Snowflake {
	if c.Interaction == nil || c.Interaction.Data == nil {
		return 0
	}
	if c.Interaction.Data.TargetID != nil {
		return *c.Interaction.Data.TargetID
	}
	if c.IsSlash() {
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
	}
	return 0
}

func (c *Context) TargetUser() *types.User {
	if c.Interaction == nil || c.Interaction.Data == nil {
		return nil
	}

	data := c.Interaction.Data
	if c.IsUserContext() && data.TargetID != nil && data.Resolved != nil && data.Resolved.Users != nil {
		if u, ok := data.Resolved.Users[*data.TargetID]; ok {
			return &u
		}
	}

	if c.IsMessageContext() && data.TargetID != nil && data.Resolved != nil && data.Resolved.Messages != nil {
		if msg, ok := data.Resolved.Messages[*data.TargetID]; ok {
			return &msg.Author
		}
	}

	if c.IsSlash() {
		if u := c.Opt("user").User(); u != nil {
			return u
		}
		if u := c.Opt("target").User(); u != nil {
			return u
		}
		for _, opt := range data.Options {
			if opt.Type == 6 {
				if u := c.Opt(opt.Name).User(); u != nil {
					return u
				}
			}
		}
	}

	return nil
}

func (c *Context) TargetMember() *types.Member {
	if c.Interaction == nil || c.Interaction.Data == nil {
		return nil
	}

	data := c.Interaction.Data
	if c.IsUserContext() && data.TargetID != nil && data.Resolved != nil && data.Resolved.Members != nil {
		if m, ok := data.Resolved.Members[*data.TargetID]; ok {
			if m.User == nil && data.Resolved.Users != nil {
				if u, okU := data.Resolved.Users[*data.TargetID]; okU {
					m.User = &u
				}
			}
			return &m
		}
	}

	if c.IsMessageContext() && data.TargetID != nil && data.Resolved != nil && data.Resolved.Messages != nil {
		if msg, ok := data.Resolved.Messages[*data.TargetID]; ok {
			authorID := msg.Author.ID
			if data.Resolved.Members != nil {
				if m, ok := data.Resolved.Members[authorID]; ok {
					if m.User == nil {
						m.User = &msg.Author
					}
					return &m
				}
			}
		}
	}

	if c.IsSlash() {
		if m := c.Opt("user").Member(); m != nil {
			return m
		}
		if m := c.Opt("target").Member(); m != nil {
			return m
		}
		for _, opt := range data.Options {
			if opt.Type == 6 {
				if m := c.Opt(opt.Name).Member(); m != nil {
					return m
				}
			}
		}
	}

	return nil
}

func (c *Context) TargetUserID() types.Snowflake {
	if u := c.TargetUser(); u != nil {
		return u.ID
	}
	if c.IsUserContext() && c.Interaction != nil && c.Interaction.Data != nil && c.Interaction.Data.TargetID != nil {
		return *c.Interaction.Data.TargetID
	}
	return 0
}

