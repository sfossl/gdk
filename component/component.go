package component

import (
	"github.com/sfossl/gda/types"
)

const (
	ButtonStylePrimary   = 1
	ButtonStyleSecondary = 2
	ButtonStyleSuccess   = 3
	ButtonStyleDanger    = 4
	ButtonStyleLink      = 5

	TextInputStyleShort     = 1
	TextInputStyleParagraph = 2
)

type ButtonOpt func(*types.Component)
type SelectOpt func(*types.Component)
type TextInputOpt func(*types.Component)

func WithButtonEmoji(name string, id ...types.Snowflake) ButtonOpt {
	return func(c *types.Component) {
		var sfID *types.Snowflake
		if len(id) > 0 && id[0] != 0 {
			val := id[0]
			sfID = &val
		}
		c.Emoji = &types.ComponentEmoji{
			Name: name,
			ID:   sfID,
		}
	}
}

func WithButtonDisabled(disabled bool) ButtonOpt {
	return func(c *types.Component) {
		c.Disabled = disabled
	}
}

func Primary(customID, label string, opts ...ButtonOpt) types.Component {
	return newButton(types.ComponentTypeButton, ButtonStylePrimary, customID, label, "", opts...)
}

func Secondary(customID, label string, opts ...ButtonOpt) types.Component {
	return newButton(types.ComponentTypeButton, ButtonStyleSecondary, customID, label, "", opts...)
}

func Success(customID, label string, opts ...ButtonOpt) types.Component {
	return newButton(types.ComponentTypeButton, ButtonStyleSuccess, customID, label, "", opts...)
}

func Danger(customID, label string, opts ...ButtonOpt) types.Component {
	return newButton(types.ComponentTypeButton, ButtonStyleDanger, customID, label, "", opts...)
}

func Link(url, label string, opts ...ButtonOpt) types.Component {
	return newButton(types.ComponentTypeButton, ButtonStyleLink, "", label, url, opts...)
}

func newButton(typ types.ComponentType, style int, customID, label, url string, opts ...ButtonOpt) types.Component {
	c := types.Component{
		Type:     typ,
		Style:    style,
		CustomID: customID,
		Label:    label,
		URL:      url,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&c)
		}
	}
	return c
}

func WithRange(min, max int) SelectOpt {
	return func(c *types.Component) {
		minVal := min
		maxVal := max
		c.MinValues = &minVal
		c.MaxValues = &maxVal
	}
}

func WithSelectDisabled(disabled bool) SelectOpt {
	return func(c *types.Component) {
		c.Disabled = disabled
	}
}

func WithChannelTypes(channelTypes ...int) SelectOpt {
	return func(c *types.Component) {
		c.ChannelTypes = channelTypes
	}
}

func StringSelect(customID, placeholder string, options ...types.SelectOption) types.Component {
	return types.Component{
		Type:        types.ComponentTypeStringSelect,
		CustomID:    customID,
		Placeholder: placeholder,
		Options:     options,
	}
}

func UserSelect(customID, placeholder string, opts ...SelectOpt) types.Component {
	return newSelect(types.ComponentTypeUserSelect, customID, placeholder, opts...)
}

func RoleSelect(customID, placeholder string, opts ...SelectOpt) types.Component {
	return newSelect(types.ComponentTypeRoleSelect, customID, placeholder, opts...)
}

func MentionableSelect(customID, placeholder string, opts ...SelectOpt) types.Component {
	return newSelect(types.ComponentTypeMentionableSelect, customID, placeholder, opts...)
}

func ChannelSelect(customID, placeholder string, opts ...SelectOpt) types.Component {
	return newSelect(types.ComponentTypeChannelSelect, customID, placeholder, opts...)
}

func newSelect(typ types.ComponentType, customID, placeholder string, opts ...SelectOpt) types.Component {
	c := types.Component{
		Type:        typ,
		CustomID:    customID,
		Placeholder: placeholder,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&c)
		}
	}
	return c
}

func Option(label, value string, description ...string) types.SelectOption {
	var desc string
	if len(description) > 0 {
		desc = description[0]
	}
	return types.SelectOption{
		Label:       label,
		Value:       value,
		Description: desc,
	}
}

func WithOptionEmoji(opt types.SelectOption, name string, id ...types.Snowflake) types.SelectOption {
	var sfID *types.Snowflake
	if len(id) > 0 && id[0] != 0 {
		val := id[0]
		sfID = &val
	}
	opt.Emoji = &types.ComponentEmoji{
		Name: name,
		ID:   sfID,
	}
	return opt
}

func WithOptionDefault(opt types.SelectOption, isDefault bool) types.SelectOption {
	opt.Default = isDefault
	return opt
}

func WithPlaceholder(placeholder string) TextInputOpt {
	return func(c *types.Component) {
		c.Placeholder = placeholder
	}
}

func WithValue(value string) TextInputOpt {
	return func(c *types.Component) {
		c.Value = value
	}
}

func WithMinLength(min int) TextInputOpt {
	return func(c *types.Component) {
		m := min
		c.MinLength = &m
	}
}

func WithMaxLength(max int) TextInputOpt {
	return func(c *types.Component) {
		m := max
		c.MaxLength = &m
	}
}

func WithRequired(required bool) TextInputOpt {
	return func(c *types.Component) {
		r := required
		c.Required = &r
	}
}

func TextInputShort(customID, label string, opts ...TextInputOpt) types.Component {
	return newTextInput(customID, label, TextInputStyleShort, opts...)
}

func TextInputParagraph(customID, label string, opts ...TextInputOpt) types.Component {
	return newTextInput(customID, label, TextInputStyleParagraph, opts...)
}

func newTextInput(customID, label string, style int, opts ...TextInputOpt) types.Component {
	c := types.Component{
		Type:     types.ComponentTypeTextInput,
		CustomID: customID,
		Label:    label,
		Style:    style,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&c)
		}
	}
	return c
}

