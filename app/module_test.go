package app

import (
	"errors"
	"testing"

	"github.com/sfossl/gdk/command"
	"github.com/sfossl/gdk/context"
)

type testModule struct {
	called bool
	err    error
}

func (m *testModule) Register(app *App) error {
	m.called = true
	if m.err != nil {
		return m.err
	}
	return app.RegisterCommand(command.Command{
		Type:        command.CommandTypeChatInput,
		Name:        "modcmd",
		Description: "from module",
	}, func(c *context.Context) error {
		return nil
	})
}

func TestApp_RegisterModule(t *testing.T) {
	app, err := New("token", 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := app.RegisterModule(nil); !errors.Is(err, ErrNilModule) {
		t.Fatalf("expected ErrNilModule, got %v", err)
	}

	mod := &testModule{}
	if err := app.RegisterModule(mod); err != nil {
		t.Fatalf("unexpected error registering module: %v", err)
	}
	if !mod.called {
		t.Fatal("expected module Register to be called")
	}

	if app.lookupCommand(int(command.CommandTypeChatInput), "modcmd") == nil {
		t.Fatal("expected command from module to be registered in app")
	}

	app2, _ := New("token", 12345)
	errMod := errors.New("module failed")
	app2.Module(&testModule{err: errMod})
	if !errors.Is(app2.initErr, errMod) {
		t.Fatalf("expected initErr from failed module, got %v", app2.initErr)
	}

	app2.Module(&testModule{})
}
