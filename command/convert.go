package command

import (
	"github.com/sfossl/gda/types"
)

func ToApplicationCommand(cmd Command) types.ApplicationCommand {
	cmdType := int(cmd.Type)
	if cmdType == 0 {
		cmdType = int(CommandTypeChatInput)
	}

	appCmd := types.ApplicationCommand{
		Name:                     cmd.Name,
		Description:              cmd.Description,
		Type:                     cmdType,
		DefaultMemberPermissions: cmd.DefaultMemberPermissions,
		DMPermission:             cmd.DMPermission,
	}

	if cmd.GuildID != 0 {
		appCmd.GuildID = &cmd.GuildID
	}

	if len(cmd.SubcommandGroups) > 0 {
		for _, grp := range cmd.SubcommandGroups {
			appCmd.Options = append(appCmd.Options, convertSubcommandGroup(grp))
		}
	} else if len(cmd.Subcommands) > 0 {
		for _, sub := range cmd.Subcommands {
			appCmd.Options = append(appCmd.Options, convertSubcommand(sub))
		}
	} else if len(cmd.Options) > 0 {
		for _, opt := range cmd.Options {
			appCmd.Options = append(appCmd.Options, convertOption(opt))
		}
	}

	return appCmd
}

func convertSubcommandGroup(grp SubcommandGroup) types.ApplicationCommandOption {
	opt := types.ApplicationCommandOption{
		Name:        grp.Name,
		Description: grp.Description,
		Type:        int(OptionTypeSubcommandGroup),
	}
	for _, sub := range grp.Subcommands {
		opt.Options = append(opt.Options, convertSubcommand(sub))
	}
	return opt
}

func convertSubcommand(sub Subcommand) types.ApplicationCommandOption {
	opt := types.ApplicationCommandOption{
		Name:        sub.Name,
		Description: sub.Description,
		Type:        int(OptionTypeSubcommand),
	}
	for _, subOpt := range sub.Options {
		opt.Options = append(opt.Options, convertOption(subOpt))
	}
	return opt
}

func convertOption(opt Option) types.ApplicationCommandOption {
	appOpt := types.ApplicationCommandOption{
		Name:         opt.Name,
		Description:  opt.Description,
		Type:         int(opt.Type),
		Required:     opt.Required,
		ChannelTypes: opt.ChannelTypes,
		MinValue:     opt.MinValue,
		MaxValue:     opt.MaxValue,
		MinLength:    opt.MinLength,
		MaxLength:    opt.MaxLength,
		Autocomplete: opt.Autocomplete,
	}

	for _, c := range opt.Choices {
		appOpt.Choices = append(appOpt.Choices, types.Choice{
			Name:  c.Name,
			Value: c.Value,
		})
	}

	return appOpt
}
