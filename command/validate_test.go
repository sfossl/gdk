package command

import (
	"errors"
	"strings"
	"testing"
)

func TestValidate_ValidChatInputCommand(t *testing.T) {
	cmd := Command{
		Name:        "ban",
		Description: "Ban a user from the server",
		Options: []Option{
			UserOption("target", "Target member to ban").WithRequired(),
			StringOption("reason", "Audit log reason").
				WithChoices(
					Choice{Name: "Spam", Value: "spam"},
					Choice{Name: "Abuse", Value: "abuse"},
				),
			IntOption("days", "Days of messages to delete").
				WithMinMax(0, 7),
		},
	}

	if err := Validate(cmd); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestValidate_ValidContextMenuCommands(t *testing.T) {
	userCmd := Command{
		Type: CommandTypeUser,
		Name: "View User Profile",
	}
	if err := Validate(userCmd); err != nil {
		t.Fatalf("unexpected user context menu error: %v", err)
	}

	msgCmd := Command{
		Type: CommandTypeMessage,
		Name: "Bookmark Message",
	}
	if err := Validate(msgCmd); err != nil {
		t.Fatalf("unexpected message context menu error: %v", err)
	}
}

func TestValidate_ContextMenuRules(t *testing.T) {
	cmd := Command{
		Type:        CommandTypeUser,
		Name:        "View Profile",
		Description: "Should not exist",
	}
	if err := Validate(cmd); !errors.Is(err, ErrContextMenuDescription) {
		t.Fatalf("expected ErrContextMenuDescription, got: %v", err)
	}
	cmdOpts := Command{
		Type:    CommandTypeUser,
		Name:    "View Profile",
		Options: []Option{StringOption("reason", "Reason")},
	}
	if err := Validate(cmdOpts); !errors.Is(err, ErrContextMenuOptions) {
		t.Fatalf("expected ErrContextMenuOptions, got: %v", err)
	}
	cmdSubs := Command{
		Type:        CommandTypeMessage,
		Name:        "Bookmark",
		Subcommands: []Subcommand{{Name: "sub", Description: "desc"}},
	}
	if err := Validate(cmdSubs); !errors.Is(err, ErrContextMenuOptions) {
		t.Fatalf("expected ErrContextMenuOptions, got: %v", err)
	}
}

func TestValidate_NameRules(t *testing.T) {
	cases := []struct {
		name    string
		wantErr error
	}{
		{"", ErrEmptyName},
		{"Invalid Name", ErrInvalidName},
		{"UPPERCASE", ErrInvalidName},
		{"special@char", ErrInvalidName},
		{strings.Repeat("a", 33), ErrNameTooLong},
		{"valid_name-123", nil},
	}

	for _, tc := range cases {
		cmd := Command{
			Name:        tc.name,
			Description: "Valid description",
		}
		err := Validate(cmd)
		if tc.wantErr != nil {
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("for name %q expected error %v, got %v", tc.name, tc.wantErr, err)
			}
		} else if err != nil {
			t.Errorf("for name %q expected no error, got %v", tc.name, err)
		}
	}
}

func TestValidate_DescriptionRules(t *testing.T) {
	cases := []struct {
		desc    string
		wantErr error
	}{
		{"", ErrEmptyDescription},
		{strings.Repeat("a", 101), ErrDescriptionTooLong},
		{"Valid description", nil},
	}

	for _, tc := range cases {
		cmd := Command{
			Name:        "test",
			Description: tc.desc,
		}
		err := Validate(cmd)
		if tc.wantErr != nil {
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("for description length %d expected error %v, got %v", len(tc.desc), tc.wantErr, err)
			}
		} else if err != nil {
			t.Errorf("for description length %d expected no error, got %v", len(tc.desc), err)
		}
	}
}

func TestValidate_SubcommandsHierarchy(t *testing.T) {
	cmdWithConflict := Command{
		Name:        "config",
		Description: "Configuration commands",
		Options: []Option{
			StringOption("key", "Setting key"),
		},
		Subcommands: []Subcommand{
			{Name: "get", Description: "Get config"},
		},
	}
	if err := Validate(cmdWithConflict); !errors.Is(err, ErrSubcommandTopOptions) {
		t.Fatalf("expected ErrSubcommandTopOptions, got: %v", err)
	}

	emptyGroup := Command{
		Name:        "admin",
		Description: "Admin tools",
		SubcommandGroups: []SubcommandGroup{
			{
				Name:        "roles",
				Description: "Role management",
				Subcommands: nil,
			},
		},
	}
	if err := Validate(emptyGroup); !errors.Is(err, ErrEmptySubcommandGroup) {
		t.Fatalf("expected ErrEmptySubcommandGroup, got: %v", err)
	}

	validGroup := Command{
		Name:        "admin",
		Description: "Admin tools",
		SubcommandGroups: []SubcommandGroup{
			{
				Name:        "roles",
				Description: "Role management",
				Subcommands: []Subcommand{
					{Name: "add", Description: "Add role"},
					{Name: "remove", Description: "Remove role"},
				},
			},
		},
	}
	if err := Validate(validGroup); err != nil {
		t.Fatalf("unexpected error for valid subcommand group: %v", err)
	}
}

func TestValidate_OptionLimits(t *testing.T) {
	opts := make([]Option, 26)
	for i := range opts {
		opts[i] = StringOption("opt"+string(rune('a'+i)), "desc")
	}

	cmd := Command{
		Name:        "flood",
		Description: "Too many options",
		Options:     opts,
	}
	if err := Validate(cmd); !errors.Is(err, ErrTooManyOptions) {
		t.Fatalf("expected ErrTooManyOptions, got: %v", err)
	}
}

func TestValidate_ChoiceLimitsAndTypes(t *testing.T) {
	choices := make([]Choice, 26)
	for i := range choices {
		choices[i] = Choice{Name: "c" + string(rune('a'+i)), Value: "val"}
	}

	cmd := Command{
		Name:        "choices",
		Description: "Too many choices",
		Options: []Option{
			StringOption("pick", "Pick one").WithChoices(choices...),
		},
	}
	if err := Validate(cmd); !errors.Is(err, ErrTooManyChoices) {
		t.Fatalf("expected ErrTooManyChoices, got: %v", err)
	}

	autoCmd := Command{
		Name:        "search",
		Description: "Search query",
		Options: []Option{
			StringOption("query", "Search term").
				WithAutocomplete().
				WithChoice("Invalid", "forbidden"),
		},
	}
	if err := Validate(autoCmd); !errors.Is(err, ErrAutocompleteWithChoices) {
		t.Fatalf("expected ErrAutocompleteWithChoices, got: %v", err)
	}

	typeMismatchCmd := Command{
		Name:        "badchoice",
		Description: "Type mismatch",
		Options: []Option{
			IntOption("number", "Enter integer").
				WithChoice("NotAnInt", "stringValue"),
		},
	}
	if err := Validate(typeMismatchCmd); !errors.Is(err, ErrInvalidChoiceValue) {
		t.Fatalf("expected ErrInvalidChoiceValue, got: %v", err)
	}
}

func TestValidate_TopLevelDuplicateSubcommands(t *testing.T) {
	cmdDupSubs := Command{
		Name:        "test",
		Description: "test command",
		Subcommands: []Subcommand{
			{Name: "view", Description: "view 1"},
			{Name: "view", Description: "view 2"},
		},
	}
	if err := Validate(cmdDupSubs); !errors.Is(err, ErrDuplicateOptionName) {
		t.Fatalf("expected ErrDuplicateOptionName for duplicate subcommands, got %v", err)
	}

	cmdDupSubAndGroup := Command{
		Name:        "test",
		Description: "test command",
		Subcommands: []Subcommand{
			{Name: "config", Description: "sub config"},
		},
		SubcommandGroups: []SubcommandGroup{
			{
				Name:        "config",
				Description: "group config",
				Subcommands: []Subcommand{
					{Name: "get", Description: "get"},
				},
			},
		},
	}
	if err := Validate(cmdDupSubAndGroup); !errors.Is(err, ErrDuplicateOptionName) {
		t.Fatalf("expected ErrDuplicateOptionName for collision between sub and group, got %v", err)
	}
}


