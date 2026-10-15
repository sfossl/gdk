package gdk

import (
	"testing"
)

func TestFacadeExports(t *testing.T) {
	embed, err := NewEmbed().
		SetTitle("Facade Test").
		SetColor("blurple").
		AddField("F", "V").
		Build()
	if err != nil {
		t.Fatalf("unexpected error building embed from facade: %v", err)
	}
	if embed.Title != "Facade Test" || embed.Color != ColorBlurple {
		t.Fatalf("unexpected embed %v", embed)
	}

	rgb := RGB(10, 20, 30)
	if rgb != (10<<16)|(20<<8)|30 {
		t.Fatalf("unexpected RGB result %d", rgb)
	}

	hexVal, err := Hex("#5865F2")
	if err != nil || hexVal != ColorBlurple {
		t.Fatalf("unexpected Hex result %d err %v", hexVal, err)
	}

	btn := Primary("btn_id", "Label", WithButtonDisabled(true))
	if btn.CustomID != "btn_id" || !btn.Disabled {
		t.Fatalf("unexpected button %v", btn)
	}

	sel := StringSelect("sel_id", "Placeholder", SelectOption("L", "V"))
	if sel.CustomID != "sel_id" || len(sel.Options) != 1 {
		t.Fatalf("unexpected select %v", sel)
	}

	rows := PackRows(btn, sel)
	if len(rows) != 2 {
		t.Fatalf("expected 2 packed rows, got %d", len(rows))
	}
	if err := ValidateRows(rows); err != nil {
		t.Fatalf("expected valid rows, got %v", err)
	}

	inp := TextInputShort("inp_id", "Label", WithRequired(true))
	if inp.CustomID != "inp_id" || inp.Required == nil || !*inp.Required {
		t.Fatalf("unexpected text input %v", inp)
	}
}

