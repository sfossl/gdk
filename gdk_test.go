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

func TestFacadeMiddlewareExports(t *testing.T) {
	guard := Guard(WithMessage("Guard error"))
	cd := Cooldown([]Policy{UserCooldown("5s")}, WithCooldownAutoRespond(false))
	req := RequireUser(1<<3, WithPermissionAutoRespond(false))

	pipe := NewPipeline(guard, cd, req)
	if pipe == nil {
		t.Fatal("expected non-nil pipeline")
	}

	chained := Chain(guard, cd)
	if chained == nil {
		t.Fatal("expected non-nil chained middleware")
	}

	names := PermissionNames(1 << 3)
	if len(names) != 1 || names[0] != "Administrator" {
		t.Fatalf("unexpected permission names: %v", names)
	}

	store := NewCooldownStore()
	ok, _ := store.Take("test", UserCooldown(1))
	if !ok {
		t.Fatal("expected initial cooldown token to be granted")
	}
}

func TestFacadeAppExports(t *testing.T) {
	cfg := DefaultConfig("token", 123)
	if cfg.Token != "token" || cfg.AppID != 123 {
		t.Fatalf("unexpected DefaultConfig: %v", cfg)
	}

	app, err := NewApp("token", 123,
		WithMode(ModeWebhook),
		WithAutoSync(false),
		WithAutoDefer(true),
		WithDevMode(true, 456),
	)
	if err != nil {
		t.Fatalf("unexpected error creating app via facade: %v", err)
	}
	if app.Config().Mode != ModeWebhook {
		t.Fatalf("expected ModeWebhook, got %v", app.Config().Mode)
	}
	if !app.Config().DevMode || app.Config().DevGuildID != 456 {
		t.Fatalf("expected dev mode enabled")
	}

	if ErrTokenRequired == nil || ErrAppIDRequired == nil || ErrCommandNotFound == nil {
		t.Fatal("expected non-nil error sentinels")
	}
}


