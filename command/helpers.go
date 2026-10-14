package command

func StringOption(name, description string) Option {
	return Option{
		Type:        OptionTypeString,
		Name:        name,
		Description: description,
	}
}

func IntOption(name, description string) Option {
	return Option{
		Type:        OptionTypeInteger,
		Name:        name,
		Description: description,
	}
}

func NumberOption(name, description string) Option {
	return Option{
		Type:        OptionTypeNumber,
		Name:        name,
		Description: description,
	}
}

func BoolOption(name, description string) Option {
	return Option{
		Type:        OptionTypeBoolean,
		Name:        name,
		Description: description,
	}
}

func UserOption(name, description string) Option {
	return Option{
		Type:        OptionTypeUser,
		Name:        name,
		Description: description,
	}
}

func ChannelOption(name, description string, channelTypes ...int) Option {
	return Option{
		Type:         OptionTypeChannel,
		Name:         name,
		Description:  description,
		ChannelTypes: channelTypes,
	}
}

func RoleOption(name, description string) Option {
	return Option{
		Type:        OptionTypeRole,
		Name:        name,
		Description: description,
	}
}

func MentionableOption(name, description string) Option {
	return Option{
		Type:        OptionTypeMentionable,
		Name:        name,
		Description: description,
	}
}

func AttachmentOption(name, description string) Option {
	return Option{
		Type:        OptionTypeAttachment,
		Name:        name,
		Description: description,
	}
}

func (o Option) WithRequired() Option {
	o.Required = true
	return o
}

func (o Option) WithChoices(choices ...Choice) Option {
	o.Choices = append(o.Choices, choices...)
	return o
}

func (o Option) WithChoice(name string, value any) Option {
	o.Choices = append(o.Choices, Choice{Name: name, Value: value})
	return o
}

func (o Option) WithAutocomplete() Option {
	o.Autocomplete = true
	return o
}

func (o Option) WithMinMax(min, max float64) Option {
	o.MinValue = &min
	o.MaxValue = &max
	return o
}

func (o Option) WithLength(min, max int) Option {
	o.MinLength = &min
	o.MaxLength = &max
	return o
}
