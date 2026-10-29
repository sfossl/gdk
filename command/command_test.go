package command

import (
	"errors"
	"testing"

	"github.com/sfossl/gda/types"
)

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

func TestToApplicationCommand_MixedSubcommandsAndGroups(t *testing.T) {
	cmd := Command{
		Name:        "project",
		Description: "Project management",
		Subcommands: []Subcommand{
			{Name: "status", Description: "Show status"},
		},
		SubcommandGroups: []SubcommandGroup{
			{
				Name:        "team",
				Description: "Team management",
				Subcommands: []Subcommand{
					{Name: "list", Description: "List members"},
				},
			},
		},
	}

	appCmd := ToApplicationCommand(cmd)
	if len(appCmd.Options) != 2 {
		t.Fatalf("expected 2 options (1 group + 1 subcommand), got %d", len(appCmd.Options))
	}
}

func TestToApplicationCommand_NSFW(t *testing.T) {
	cmd := Command{
		Name:        "adult",
		Description: "NSFW content",
		NSFW:        true,
	}

	appCmd := ToApplicationCommand(cmd)
	if appCmd.NSFW == nil || !*appCmd.NSFW {
		t.Fatalf("expected NSFW true, got %v", appCmd.NSFW)
	}
}

func TestValidate_SubcommandLimits(t *testing.T) {
	subs := make([]Subcommand, 26)
	for i := range subs {
		subs[i] = Subcommand{Name: "sub" + string(rune('a'+i)), Description: "desc"}
	}

	cmd := Command{
		Name:        "toomanysubs",
		Description: "Exceeds subcommand limit",
		Subcommands: subs,
	}
	if err := Validate(cmd); !errors.Is(err, ErrTooManyOptions) {
		t.Fatalf("expected ErrTooManyOptions for subcommands, got %v", err)
	}

	grp := SubcommandGroup{
		Name:        "group",
		Description: "Group with too many subcommands",
		Subcommands: subs,
	}
	cmdGroup := Command{
		Name:             "badgroup",
		Description:      "Exceeds group subcommands",
		SubcommandGroups: []SubcommandGroup{grp},
	}
	if err := Validate(cmdGroup); !errors.Is(err, ErrTooManyOptions) {
		t.Fatalf("expected ErrTooManyOptions for group subcommands, got %v", err)
	}
}

func TestValidate_RequiredOptionAfterOptional(t *testing.T) {
	cmd := Command{
		Name:        "order",
		Description: "Option ordering test",
		Options: []Option{
			StringOption("first", "Optional first"),
			StringOption("second", "Required second").WithRequired(),
		},
	}
	if err := Validate(cmd); !errors.Is(err, ErrRequiredOptionAfterOptional) {
		t.Fatalf("expected ErrRequiredOptionAfterOptional, got %v", err)
	}
}

func TestValidate_EmptyStringChoiceValue(t *testing.T) {
	cmd := Command{
		Name:        "emptychoice",
		Description: "Empty choice test",
		Options: []Option{
			StringOption("reason", "Reason").WithChoice("Empty", ""),
		},
	}
	if err := Validate(cmd); !errors.Is(err, ErrInvalidChoiceValue) {
		t.Fatalf("expected ErrInvalidChoiceValue for empty string choice, got %v", err)
	}
}

func TestValidate_DuplicateOptionNames(t *testing.T) {
	cmd := Command{
		Name:        "dupopts",
		Description: "Duplicate option test",
		Options: []Option{
			StringOption("target", "First target"),
			UserOption("target", "Second target"),
		},
	}
	if err := Validate(cmd); !errors.Is(err, ErrDuplicateOptionName) {
		t.Fatalf("expected ErrDuplicateOptionName, got %v", err)
	}
}

func TestValidate_DuplicateChoiceNames(t *testing.T) {
	cmd := Command{
		Name:        "dupchoices",
		Description: "Duplicate choice test",
		Options: []Option{
			StringOption("type", "Type").
				WithChoice("Same", "v1").
				WithChoice("Same", "v2"),
		},
	}
	if err := Validate(cmd); !errors.Is(err, ErrDuplicateChoiceName) {
		t.Fatalf("expected ErrDuplicateChoiceName, got %v", err)
	}
}

func TestValidate_DuplicateSubcommandNames(t *testing.T) {
	cmd := Command{
		Name:        "dupsubs",
		Description: "Duplicate subcommand test",
		SubcommandGroups: []SubcommandGroup{
			{
				Name:        "group",
				Description: "Group test",
				Subcommands: []Subcommand{
					{Name: "run", Description: "First run"},
					{Name: "run", Description: "Second run"},
				},
			},
		},
	}
	if err := Validate(cmd); !errors.Is(err, ErrDuplicateOptionName) {
		t.Fatalf("expected ErrDuplicateOptionName for subcommands, got %v", err)
	}
}

func TestValidate_ContextMenuCommands(t *testing.T) {
	userCmd := Command{
		Type: CommandTypeUser,
		Name: "Get Avatar",
	}
	if err := Validate(userCmd); err != nil {
		t.Fatalf("expected valid user command, got %v", err)
	}

	appCmd := ToApplicationCommand(userCmd)
	if appCmd.Type != int(CommandTypeUser) || appCmd.Description != "" {
		t.Fatalf("unexpected appCmd for user command: %+v", appCmd)
	}

	msgCmd := Command{
		Type: CommandTypeMessage,
		Name: "Pin Message",
	}
	if err := Validate(msgCmd); err != nil {
		t.Fatalf("expected valid message command, got %v", err)
	}

	appMsgCmd := ToApplicationCommand(msgCmd)
	if appMsgCmd.Type != int(CommandTypeMessage) || appMsgCmd.Description != "" {
		t.Fatalf("unexpected appMsgCmd for message command: %+v", appMsgCmd)
	}

	invalidUserCmd := Command{
		Type:        CommandTypeUser,
		Name:        "Get Avatar",
		Description: "should not have description",
	}
	if err := Validate(invalidUserCmd); !errors.Is(err, ErrContextMenuHasDescription) {
		t.Fatalf("expected ErrContextMenuHasDescription, got %v", err)
	}

	invalidOptionsCmd := Command{
		Type:    CommandTypeMessage,
		Name:    "Pin",
		Options: []Option{StringOption("opt", "invalid")},
	}
	if err := Validate(invalidOptionsCmd); !errors.Is(err, ErrContextMenuHasOptions) {
		t.Fatalf("expected ErrContextMenuHasOptions, got %v", err)
	}
}


