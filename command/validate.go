package command

import (
	"errors"
	"regexp"
	"unicode/utf8"
)

var (
	chatInputNamePattern = regexp.MustCompile(`^[a-z0-9_-]{1,32}$`)

	ErrInvalidName             = errors.New("command name must match discord regex rules")
	ErrNameTooLong             = errors.New("command name exceeds 32 runes")
	ErrEmptyName               = errors.New("command name cannot be empty")
	ErrEmptyDescription        = errors.New("chat input command requires description")
	ErrDescriptionTooLong      = errors.New("command description exceeds 100 runes")
	ErrContextMenuDescription  = errors.New("context menu commands cannot have a description")
	ErrSubcommandTopOptions    = errors.New("commands with subcommands cannot have top level options")
	ErrTooManyOptions          = errors.New("command exceeds maximum of 25 options")
	ErrTooManyChoices          = errors.New("option exceeds maximum of 25 choices")
	ErrAutocompleteWithChoices = errors.New("autocomplete options cannot define static choices")
	ErrEmptySubcommandGroup    = errors.New("subcommand group must contain at least one subcommand")
	ErrEmptyChoiceName         = errors.New("choice name cannot be empty")
	ErrChoiceNameTooLong       = errors.New("choice name exceeds 100 runes")
	ErrInvalidChoiceValue      = errors.New("choice value does not match option type")
)

func Validate(cmd Command) error {
	cmdType := cmd.Type
	if cmdType == 0 {
		cmdType = CommandTypeChatInput
	}

	nameLen := utf8.RuneCountInString(cmd.Name)
	if nameLen == 0 {
		return ErrEmptyName
	}
	if nameLen > 32 {
		return ErrNameTooLong
	}

	switch cmdType {
	case CommandTypeChatInput:
		if !chatInputNamePattern.MatchString(cmd.Name) {
			return ErrInvalidName
		}
		descLen := utf8.RuneCountInString(cmd.Description)
		if descLen == 0 {
			return ErrEmptyDescription
		}
		if descLen > 100 {
			return ErrDescriptionTooLong
		}

	case CommandTypeUser, CommandTypeMessage:
		if cmd.Description != "" {
			return ErrContextMenuDescription
		}
		return nil
	}

	hasSubcommands := len(cmd.Subcommands) > 0 || len(cmd.SubcommandGroups) > 0
	if hasSubcommands && len(cmd.Options) > 0 {
		return ErrSubcommandTopOptions
	}

	if len(cmd.Options) > 25 {
		return ErrTooManyOptions
	}

	for _, opt := range cmd.Options {
		if err := validateOption(opt); err != nil {
			return err
		}
	}

	for _, sub := range cmd.Subcommands {
		if err := validateSubcommand(sub); err != nil {
			return err
		}
	}

	for _, grp := range cmd.SubcommandGroups {
		if err := validateSubcommandGroup(grp); err != nil {
			return err
		}
	}

	return nil
}

func validateSubcommand(sub Subcommand) error {
	nameLen := utf8.RuneCountInString(sub.Name)
	if nameLen == 0 {
		return ErrEmptyName
	}
	if nameLen > 32 || !chatInputNamePattern.MatchString(sub.Name) {
		return ErrInvalidName
	}

	descLen := utf8.RuneCountInString(sub.Description)
	if descLen == 0 {
		return ErrEmptyDescription
	}
	if descLen > 100 {
		return ErrDescriptionTooLong
	}

	if len(sub.Options) > 25 {
		return ErrTooManyOptions
	}

	for _, opt := range sub.Options {
		if err := validateOption(opt); err != nil {
			return err
		}
	}

	return nil
}

func validateSubcommandGroup(grp SubcommandGroup) error {
	nameLen := utf8.RuneCountInString(grp.Name)
	if nameLen == 0 {
		return ErrEmptyName
	}
	if nameLen > 32 || !chatInputNamePattern.MatchString(grp.Name) {
		return ErrInvalidName
	}

	descLen := utf8.RuneCountInString(grp.Description)
	if descLen == 0 {
		return ErrEmptyDescription
	}
	if descLen > 100 {
		return ErrDescriptionTooLong
	}

	if len(grp.Subcommands) == 0 {
		return ErrEmptySubcommandGroup
	}

	for _, sub := range grp.Subcommands {
		if err := validateSubcommand(sub); err != nil {
			return err
		}
	}

	return nil
}

func validateOption(opt Option) error {
	nameLen := utf8.RuneCountInString(opt.Name)
	if nameLen == 0 {
		return ErrEmptyName
	}
	if nameLen > 32 || !chatInputNamePattern.MatchString(opt.Name) {
		return ErrInvalidName
	}

	descLen := utf8.RuneCountInString(opt.Description)
	if descLen == 0 {
		return ErrEmptyDescription
	}
	if descLen > 100 {
		return ErrDescriptionTooLong
	}

	if len(opt.Choices) > 25 {
		return ErrTooManyChoices
	}

	if opt.Autocomplete && len(opt.Choices) > 0 {
		return ErrAutocompleteWithChoices
	}

	for _, choice := range opt.Choices {
		if err := validateChoice(opt.Type, choice); err != nil {
			return err
		}
	}

	return nil
}

func validateChoice(optType OptionType, choice Choice) error {
	nameLen := utf8.RuneCountInString(choice.Name)
	if nameLen == 0 {
		return ErrEmptyChoiceName
	}
	if nameLen > 100 {
		return ErrChoiceNameTooLong
	}

	switch optType {
	case OptionTypeString:
		s, ok := choice.Value.(string)
		if !ok || utf8.RuneCountInString(s) > 100 {
			return ErrInvalidChoiceValue
		}

	case OptionTypeInteger:
		switch choice.Value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		default:
			return ErrInvalidChoiceValue
		}

	case OptionTypeNumber:
		switch choice.Value.(type) {
		case float32, float64, int, int64:
		default:
			return ErrInvalidChoiceValue
		}

	default:
		return ErrInvalidChoiceValue
	}

	return nil
}
