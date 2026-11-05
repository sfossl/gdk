package component

import (
	"testing"

	"github.com/sfossl/gda/types"
)

func TestButtons(t *testing.T) {
	p := Button("primary", "btn_1", "Primary", WithButtonDisabled(true))
	if p.Type != types.ComponentTypeButton || p.Style != ButtonStylePrimary || p.CustomID != "btn_1" || p.Label != "Primary" || !p.Disabled {
		t.Fatalf("unexpected primary button %v", p)
	}

	sec := Button("secondary", "btn_2", "Secondary")
	if sec.Style != ButtonStyleSecondary || sec.CustomID != "btn_2" {
		t.Fatalf("unexpected secondary button %v", sec)
	}

	succ := Button("success", "btn_3", "Success")
	if succ.Style != ButtonStyleSuccess || succ.CustomID != "btn_3" {
		t.Fatalf("unexpected success button %v", succ)
	}

	d := Button("danger", "btn_4", "Danger")
	if d.Style != ButtonStyleDanger || d.CustomID != "btn_4" {
		t.Fatalf("unexpected danger button %v", d)
	}

	emojiID := types.Snowflake(123456789)
	link := Button("link", "https://example.com", "Visit", WithButtonEmoji("custom", emojiID))
	if link.Style != ButtonStyleLink || link.URL != "https://example.com" || link.Label != "Visit" {
		t.Fatalf("unexpected link button %v", link)
	}
	if link.Emoji == nil || link.Emoji.Name != "custom" || link.Emoji.ID == nil || *link.Emoji.ID != emojiID {
		t.Fatalf("unexpected button emoji %v", link.Emoji)
	}

	uni := Button("primary", "btn_5", "Unicode", WithButtonEmoji("👍"))
	if uni.Emoji == nil || uni.Emoji.Name != "👍" || uni.Emoji.ID != nil {
		t.Fatalf("unexpected unicode emoji %v", uni.Emoji)
	}
}

func TestButton_Unified(t *testing.T) {
	b1 := Button("primary", "btn_1", "Click Me")
	if b1.Style != ButtonStylePrimary || b1.CustomID != "btn_1" || b1.Label != "Click Me" {
		t.Fatalf("unexpected b1: %v", b1)
	}

	b2 := Button("danger", "btn_2", "Delete", true)
	if b2.Style != ButtonStyleDanger || b2.CustomID != "btn_2" || !b2.Disabled {
		t.Fatalf("unexpected b2: %v", b2)
	}

	b3 := Button("link", "https://example.com", "Visit")
	if b3.Style != ButtonStyleLink || b3.URL != "https://example.com" || b3.Label != "Visit" {
		t.Fatalf("unexpected b3: %v", b3)
	}

	bOpts := Button(ButtonOptions{
		Style:    "success",
		CustomID: "btn_opts",
		Label:    "Confirm",
		Emoji:    "✅",
		Disabled: false,
	})
	if bOpts.Style != ButtonStyleSuccess || bOpts.CustomID != "btn_opts" || bOpts.Emoji == nil || bOpts.Emoji.Name != "✅" {
		t.Fatalf("unexpected bOpts: %v", bOpts)
	}
}

func TestSelectMenus(t *testing.T) {
	opts := []types.SelectOption{
		Option("Choice 1", "val_1", "First description"),
		WithOptionDefault(Option("Choice 2", "val_2"), true),
		WithOptionEmoji(Option("Choice 3", "val_3"), "star"),
	}

	strSel := StringSelect("str_sel", "Select an item", opts...)
	if strSel.Type != types.ComponentTypeStringSelect || strSel.CustomID != "str_sel" || strSel.Placeholder != "Select an item" {
		t.Fatalf("unexpected string select %v", strSel)
	}
	if len(strSel.Options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(strSel.Options))
	}
	if strSel.Options[0].Description != "First description" {
		t.Fatalf("unexpected description %s", strSel.Options[0].Description)
	}
	if !strSel.Options[1].Default {
		t.Fatalf("expected option 1 default true")
	}
	if strSel.Options[2].Emoji == nil || strSel.Options[2].Emoji.Name != "star" {
		t.Fatalf("expected star emoji")
	}

	uSel := UserSelect("usr_sel", "Select users", WithRange(1, 3), WithSelectDisabled(true))
	if uSel.Type != types.ComponentTypeUserSelect || uSel.CustomID != "usr_sel" || !uSel.Disabled {
		t.Fatalf("unexpected user select %v", uSel)
	}
	if uSel.MinValues == nil || *uSel.MinValues != 1 || uSel.MaxValues == nil || *uSel.MaxValues != 3 {
		t.Fatalf("unexpected values range %v, %v", uSel.MinValues, uSel.MaxValues)
	}

	rSel := RoleSelect("role_sel", "Select role")
	if rSel.Type != types.ComponentTypeRoleSelect {
		t.Fatalf("unexpected role select type")
	}

	mSel := MentionableSelect("ment_sel", "Select mentionable")
	if mSel.Type != types.ComponentTypeMentionableSelect {
		t.Fatalf("unexpected mentionable select type")
	}

	chSel := ChannelSelect("ch_sel", "Select channel", WithChannelTypes(0, 2))
	if chSel.Type != types.ComponentTypeChannelSelect {
		t.Fatalf("unexpected channel select type")
	}
	if len(chSel.ChannelTypes) != 2 || chSel.ChannelTypes[0] != 0 || chSel.ChannelTypes[1] != 2 {
		t.Fatalf("unexpected channel types %v", chSel.ChannelTypes)
	}
}

func TestTextInputs(t *testing.T) {
	short := TextInputShort("inp_short", "Your Name",
		WithPlaceholder("Enter name..."),
		WithValue("John Doe"),
		WithMinLength(2),
		WithMaxLength(32),
		WithRequired(true),
	)

	if short.Type != types.ComponentTypeTextInput || short.Style != TextInputStyleShort {
		t.Fatalf("unexpected text input short %v", short)
	}
	if short.CustomID != "inp_short" || short.Label != "Your Name" || short.Placeholder != "Enter name..." || short.Value != "John Doe" {
		t.Fatalf("unexpected fields in short input %v", short)
	}
	if short.MinLength == nil || *short.MinLength != 2 || short.MaxLength == nil || *short.MaxLength != 32 {
		t.Fatalf("unexpected min/max length")
	}
	if short.Required == nil || !*short.Required {
		t.Fatalf("expected required true")
	}

	para := TextInputParagraph("inp_para", "Bio")
	if para.Type != types.ComponentTypeTextInput || para.Style != TextInputStyleParagraph {
		t.Fatalf("unexpected paragraph style %v", para)
	}
}
