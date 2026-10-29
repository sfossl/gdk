package command

import (
	"errors"
	"regexp"
	"unicode/utf8"
)

var (
	chatInputNamePattern = regexp.MustCompile(`^[a-z0-9_-]{1,32}$`)

	ErrInvalidName                 = errors.New("command name must match discord regex rules")
	ErrNameTooLong                 = errors.New("command name exceeds 32 runes")
	ErrEmptyName                   = errors.New("command name cannot be empty")
	ErrEmptyDescription            = errors.New("chat input command requires description")
	ErrDescriptionTooLong          = errors.New("command description exceeds 100 runes")
	ErrSubcommandTopOptions        = errors.New("commands with subcommands cannot have top level options")
	ErrTooManyOptions              = errors.New("command exceeds maximum of 25 options")
	ErrTooManyChoices              = errors.New("option exceeds maximum of 25 choices")
	ErrAutocompleteWithChoices     = errors.New("autocomplete options cannot define static choices")
	ErrEmptySubcommandGroup        = errors.New("subcommand group must contain at least one subcommand")
	ErrEmptyChoiceName             = errors.New("choice name cannot be empty")
	ErrChoiceNameTooLong           = errors.New("choice name exceeds 100 runes")
	ErrInvalidChoiceValue          = errors.New("choice value does not match option type")
	ErrRequiredOptionAfterOptional = errors.New("required options must precede optional options")
	ErrDuplicateOptionName         = errors.New("option names must be unique within a command")
	ErrDuplicateChoiceName         = errors.New("choice names must be unique within an option")
	ErrContextMenuHasDescription   = errors.New("context menu command cannot have description")
	ErrContextMenuHasOptions       = errors.New("context menu command cannot have options")
)

func Validate(cmd Command) error {
	nameLen := utf8.RuneCountInString(cmd.Name)
	if nameLen == 0 {
		return ErrEmptyName
	}
	if nameLen > 32 {
		return ErrNameTooLong
	}
	if cmd.Type == CommandTypeUser || cmd.Type == CommandTypeMessage {
		if cmd.Description != "" {
			return ErrContextMenuHasDescription
		}
		if len(cmd.Options) > 0 || len(cmd.Subcommands) > 0 || len(cmd.SubcommandGroups) > 0 {
			return ErrContextMenuHasOptions
		}
		return nil
	}
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

	subCount := len(cmd.Subcommands) + len(cmd.SubcommandGroups)
	if subCount > 0 && len(cmd.Options) > 0 {
		return ErrSubcommandTopOptions
	}

	if subCount > 25 {
		return ErrTooManyOptions
	}

	if err := validateOptions(cmd.Options); err != nil {
		return err
	}

	topLevelNames := make(map[string]struct{}, subCount)
	for _, sub := range cmd.Subcommands {
		if _, exists := topLevelNames[sub.Name]; exists {
			return ErrDuplicateOptionName
		}
		topLevelNames[sub.Name] = struct{}{}
		if err := validateSubcommand(sub); err != nil {
			return err
		}
	}

	for _, grp := range cmd.SubcommandGroups {
		if _, exists := topLevelNames[grp.Name]; exists {
			return ErrDuplicateOptionName
		}
		topLevelNames[grp.Name] = struct{}{}
		if err := validateSubcommandGroup(grp); err != nil {
			return err
		}
	}

	return nil
}

func validateNameAndDesc(name, desc string) error {
	nameLen := utf8.RuneCountInString(name)
	if nameLen == 0 {
		return ErrEmptyName
	}
	if nameLen > 32 || !chatInputNamePattern.MatchString(name) {
		return ErrInvalidName
	}

	descLen := utf8.RuneCountInString(desc)
	if descLen == 0 {
		return ErrEmptyDescription
	}
	if descLen > 100 {
		return ErrDescriptionTooLong
	}

	return nil
}

func validateOptions(opts []Option) error {
	if len(opts) > 25 {
		return ErrTooManyOptions
	}

	seenOptional := false
	names := make(map[string]struct{}, len(opts))

	for _, opt := range opts {
		if _, exists := names[opt.Name]; exists {
			return ErrDuplicateOptionName
		}
		names[opt.Name] = struct{}{}

		if opt.Required {
			if seenOptional {
				return ErrRequiredOptionAfterOptional
			}
		} else {
			seenOptional = true
		}

		if err := validateOption(opt); err != nil {
			return err
		}
	}

	return nil
}

func validateSubcommand(sub Subcommand) error {
	if err := validateNameAndDesc(sub.Name, sub.Description); err != nil {
		return err
	}

	return validateOptions(sub.Options)
}

func validateSubcommandGroup(grp SubcommandGroup) error {
	if err := validateNameAndDesc(grp.Name, grp.Description); err != nil {
		return err
	}

	if len(grp.Subcommands) == 0 {
		return ErrEmptySubcommandGroup
	}

	if len(grp.Subcommands) > 25 {
		return ErrTooManyOptions
	}

	names := make(map[string]struct{}, len(grp.Subcommands))
	for _, sub := range grp.Subcommands {
		if _, exists := names[sub.Name]; exists {
			return ErrDuplicateOptionName
		}
		names[sub.Name] = struct{}{}

		if err := validateSubcommand(sub); err != nil {
			return err
		}
	}

	return nil
}

func validateOption(opt Option) error {
	if err := validateNameAndDesc(opt.Name, opt.Description); err != nil {
		return err
	}

	if len(opt.Choices) > 25 {
		return ErrTooManyChoices
	}

	if opt.Autocomplete && len(opt.Choices) > 0 {
		return ErrAutocompleteWithChoices
	}

	choiceNames := make(map[string]struct{}, len(opt.Choices))
	for _, choice := range opt.Choices {
		if _, exists := choiceNames[choice.Name]; exists {
			return ErrDuplicateChoiceName
		}
		choiceNames[choice.Name] = struct{}{}

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
		if !ok || s == "" || utf8.RuneCountInString(s) > 100 {
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
