package app

import (
	stdctx "context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/command"
	"github.com/sfossl/gdk/context"
)

func TestSync_DevGuildRequired(t *testing.T) {
	app, err := New("token", 12345, WithDevMode(true))
	if err != nil {
		t.Fatalf("unexpected error creating app: %v", err)
	}

	err = app.Sync(stdctx.Background())
	if !errors.Is(err, ErrDevGuildRequired) {
		t.Fatalf("expected ErrDevGuildRequired, got %v", err)
	}
}

func TestSync_NoRESTClient(t *testing.T) {
	app, err := New("token", 12345)
	if err != nil {
		t.Fatalf("unexpected error creating app: %v", err)
	}

	app.REST = nil
	err = app.Sync(stdctx.Background())
	if !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}
}

func TestSync_GlobalAndGuildCommands_FullWorkflow(t *testing.T) {
	var globalCalls atomic.Int32
	var guildCalls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if strings.Contains(r.URL.Path, "/guilds/999/commands") {
			guildCalls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]types.ApplicationCommand{})
			return
		}

		if strings.Contains(r.URL.Path, "/commands") {
			globalCalls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]types.ApplicationCommand{})
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	app, err := New("token", 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	app.REST = rest.NewClient("token", rest.WithBaseURL(server.URL))

	globalCmd := command.Command{
		Type:        command.CommandTypeChatInput,
		Name:        "ping",
		Description: "ping description",
	}
	guildCmd := command.Command{
		Type:        command.CommandTypeChatInput,
		Name:        "admin",
		Description: "admin description",
		GuildID:     999,
	}

	_ = app.RegisterCommand(globalCmd, nil)
	_ = app.RegisterCommand(guildCmd, nil)

	ctx := stdctx.Background()

	res1, err := app.SyncWithResult(ctx)
	if err != nil {
		t.Fatalf("unexpected sync error: %v", err)
	}
	if !res1.GlobalSynced {
		t.Fatal("expected global commands to be synced on first run")
	}
	if !res1.GuildsSynced[999] {
		t.Fatal("expected guild 999 commands to be synced on first run")
	}
	if res1.Updated != 2 || res1.Skipped != 0 {
		t.Fatalf("expected updated 2, skipped 0, got updated %d, skipped %d", res1.Updated, res1.Skipped)
	}
	if globalCalls.Load() != 1 || guildCalls.Load() != 1 {
		t.Fatalf("expected 1 global and 1 guild call, got global %d, guild %d", globalCalls.Load(), guildCalls.Load())
	}

	res2, err := app.SyncWithResult(ctx)
	if err != nil {
		t.Fatalf("unexpected error on second sync: %v", err)
	}
	if res2.Updated != 0 || res2.Skipped != 2 {
		t.Fatalf("expected cached sync to skip all, got updated %d, skipped %d", res2.Updated, res2.Skipped)
	}
	if globalCalls.Load() != 1 || guildCalls.Load() != 1 {
		t.Fatal("no network calls should be made when hashes match")
	}

	res3, err := app.SyncWithResult(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error on force sync: %v", err)
	}
	if res3.Updated != 2 || res3.Skipped != 0 {
		t.Fatalf("expected force sync to update all, got updated %d, skipped %d", res3.Updated, res3.Skipped)
	}
	if globalCalls.Load() != 2 || guildCalls.Load() != 2 {
		t.Fatalf("expected 2 global and 2 guild calls after force, got global %d, guild %d", globalCalls.Load(), guildCalls.Load())
	}
}

func TestSync_DevMode_RoutesAllToGuild(t *testing.T) {
	var globalCalls atomic.Int32
	var guildCalls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/guilds/777/commands") {
			guildCalls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]types.ApplicationCommand{})
			return
		}
		if strings.Contains(r.URL.Path, "/commands") {
			globalCalls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]types.ApplicationCommand{})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	app, err := New("token", 12345, WithDevMode(true, 777))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	app.REST = rest.NewClient("token", rest.WithBaseURL(server.URL))

	c1 := command.Command{Type: command.CommandTypeChatInput, Name: "one", Description: "desc one"}
	c2 := command.Command{Type: command.CommandTypeChatInput, Name: "two", Description: "desc two", GuildID: 999}

	_ = app.RegisterCommand(c1, nil)
	_ = app.RegisterCommand(c2, nil)

	res, err := app.SyncWithResult(stdctx.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.GlobalSynced {
		t.Fatal("dev mode must not sync globally")
	}
	if !res.GuildsSynced[777] {
		t.Fatal("dev mode must sync all commands to dev guild 777")
	}
	if globalCalls.Load() != 0 {
		t.Fatal("global endpoint should not have been called in dev mode")
	}
	if guildCalls.Load() != 1 {
		t.Fatalf("expected 1 dev guild call, got %d", guildCalls.Load())
	}
}

func TestSync_AutoSync_Run(t *testing.T) {
	var syncCalled atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		syncCalled.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]types.ApplicationCommand{})
	}))
	defer server.Close()

	app, err := New("token", 12345, WithAutoSync(true), WithWebhook(":0", "/hook", "deadbeef"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	app.REST = rest.NewClient("token", rest.WithBaseURL(server.URL))

	c := command.Command{Type: command.CommandTypeChatInput, Name: "autosync", Description: "autosync test"}
	_ = app.RegisterCommand(c, nil)

	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), 50*time.Millisecond)
	defer cancel()

	_ = app.Run(ctx)
	if syncCalled.Load() == 0 {
		t.Fatal("expected AutoSync to trigger Sync during Run")
	}
}

func TestSync_NonBlockingConcurrentHandle(t *testing.T) {
	syncStarted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(syncStarted)
		time.Sleep(150 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]types.ApplicationCommand{})
	}))
	defer server.Close()

	app, err := New("token", 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	app.REST = rest.NewClient("token", rest.WithBaseURL(server.URL))

	handled := make(chan struct{})
	cmd := command.Command{
		Type:        command.CommandTypeChatInput,
		Name:        "fastcmd",
		Description: "non-blocking test",
	}
	_ = app.RegisterCommand(cmd, func(c *context.Context) error {
		close(handled)
		return nil
	})

	go func() {
		_ = app.Sync(stdctx.Background(), true)
	}()

	select {
	case <-syncStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for sync to begin")
	}

	start := time.Now()
	in := &types.Interaction{
		Type: types.InteractionTypeApplicationCommand,
		Data: &types.InteractionData{
			Name: "fastcmd",
			Type: int(command.CommandTypeChatInput),
		},
	}
	gdkCtx := context.New(stdctx.Background(), in, nil)
	if err := app.Handle(gdkCtx); err != nil {
		t.Fatalf("unexpected handle error: %v", err)
	}

	select {
	case <-handled:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Handle was blocked by in-flight sync network call")
	}

	if time.Since(start) >= 100*time.Millisecond {
		t.Fatalf("Handle took too long (%v), lock contention detected", time.Since(start))
	}
}
