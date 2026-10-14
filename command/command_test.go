package command

import (
	"errors"
	"strings"
	"testing"

	"github.com/sfossl/gda/types"
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

func TestValidate_ContextMenuWithDescription(t *testing.T) {
	cmd := Command{
		Type:        CommandTypeUser,
		Name:        "View Profile",
		Description: "Should not exist",
	}
	if err := Validate(cmd); !errors.Is(err, ErrContextMenuDescription) {
		t.Fatalf("expected ErrContextMenuDescription, got: %v", err)
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

func TestHelpers_FluentConstruction(t *testing.T) {
	numOpt := NumberOption("ratio", "Aspect ratio").
		WithMinMax(1.0, 16.0)
	if *numOpt.MinValue != 1.0 || *numOpt.MaxValue != 16.0 {
		t.Errorf("MinMax failed on NumberOption")
	}

	strOpt := StringOption("code", "Postal code").
		WithLength(3, 10).
		WithRequired()
	if !strOpt.Required || *strOpt.MinLength != 3 || *strOpt.MaxLength != 10 {
		t.Errorf("Length/Required failed on StringOption")
	}

	chanOpt := ChannelOption("target", "Target channel", int(ChannelTypeGuildText), int(ChannelTypeGuildVoice))
	if len(chanOpt.ChannelTypes) != 2 {
		t.Errorf("ChannelTypes not recorded correctly")
	}

	boolOpt := BoolOption("notify", "Send alert")
	if boolOpt.Type != OptionTypeBoolean {
		t.Errorf("unexpected type for BoolOption: %v", boolOpt.Type)
	}

	roleOpt := RoleOption("role", "Target role")
	if roleOpt.Type != OptionTypeRole {
		t.Errorf("unexpected type for RoleOption: %v", roleOpt.Type)
	}

	mentOpt := MentionableOption("subject", "Target mentionable")
	if mentOpt.Type != OptionTypeMentionable {
		t.Errorf("unexpected type for MentionableOption: %v", mentOpt.Type)
	}

	attOpt := AttachmentOption("file", "Uploaded file")
	if attOpt.Type != OptionTypeAttachment {
		t.Errorf("unexpected type for AttachmentOption: %v", attOpt.Type)
	}
}

func TestToApplicationCommand_Conversion(t *testing.T) {
	guildID := types.Snowflake(1234567890)
	cmd := Command{
		Name:        "kick",
		Description: "Kick a user",
		GuildID:     guildID,
		Options: []Option{
			UserOption("member", "Member to kick").WithRequired(),
			StringOption("reason", "Reason").WithChoice("Inactivity", "inactive"),
		},
	}

	appCmd := ToApplicationCommand(cmd)

	if appCmd.Name != "kick" {
		t.Errorf("expected name kick, got %s", appCmd.Name)
	}
	if appCmd.Description != "kick a user" && appCmd.Description != "Kick a user" {
		t.Errorf("expected description Kick a user, got %s", appCmd.Description)
	}
	if appCmd.GuildID == nil || *appCmd.GuildID != guildID {
		t.Errorf("expected guild ID %d, got %v", guildID, appCmd.GuildID)
	}
	if appCmd.Type != int(CommandTypeChatInput) {
		t.Errorf("expected type 1, got %d", appCmd.Type)
	}
	if len(appCmd.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(appCmd.Options))
	}
	if appCmd.Options[0].Name != "member" || !appCmd.Options[0].Required {
		t.Errorf("first option conversion mismatch")
	}
	if len(appCmd.Options[1].Choices) != 1 || appCmd.Options[1].Choices[0].Name != "Inactivity" {
		t.Errorf("choice conversion mismatch")
	}
}

func TestToApplicationCommand_NestedSubcommands(t *testing.T) {
	cmd := Command{
		Name:        "git",
		Description: "Git commands",
		SubcommandGroups: []SubcommandGroup{
			{
				Name:        "remote",
				Description: "Remote subcommands",
				Subcommands: []Subcommand{
					{
						Name:        "add",
						Description: "Add remote",
						Options: []Option{
							StringOption("name", "Remote name").WithRequired(),
							StringOption("url", "Remote URL").WithRequired(),
						},
					},
				},
			},
		},
	}

	appCmd := ToApplicationCommand(cmd)
	if len(appCmd.Options) != 1 {
		t.Fatalf("expected 1 group option, got %d", len(appCmd.Options))
	}

	groupOpt := appCmd.Options[0]
	if groupOpt.Type != int(OptionTypeSubcommandGroup) || groupOpt.Name != "remote" {
		t.Errorf("subcommand group conversion mismatch")
	}

	if len(groupOpt.Options) != 1 {
		t.Fatalf("expected 1 subcommand in group, got %d", len(groupOpt.Options))
	}

	subOpt := groupOpt.Options[0]
	if subOpt.Type != int(OptionTypeSubcommand) || subOpt.Name != "add" {
		t.Errorf("subcommand conversion mismatch")
	}

	if len(subOpt.Options) != 2 {
		t.Fatalf("expected 2 options in subcommand, got %d", len(subOpt.Options))
	}
}
