package middleware

import (
	stdctx "context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/context"
)

func TestCooldown_ScopesAndPolicyParsing(t *testing.T) {
	if ScopeUser.String() != "user" || ScopeChannel.String() != "channel" ||
		ScopeGuild.String() != "guild" || ScopeGlobal.String() != "global" ||
		Scope(99).String() != "unknown" {
		t.Fatal("unexpected scope string representation")
	}

	pUser := UserCooldown(5 * time.Second)
	if pUser.Scope != ScopeUser || pUser.Duration != 5*time.Second || pUser.Burst != 1 {
		t.Fatalf("unexpected user policy: %+v", pUser)
	}

	pChan := ChannelCooldown(10, 3)
	if pChan.Scope != ScopeChannel || pChan.Duration != 10*time.Second || pChan.Burst != 3 {
		t.Fatalf("unexpected channel policy: %+v", pChan)
	}

	pGuild := GuildCooldown("30s")
	if pGuild.Scope != ScopeGuild || pGuild.Duration != 30*time.Second || pGuild.Burst != 1 {
		t.Fatalf("unexpected guild policy: %+v", pGuild)
	}

	pGlobal := GlobalCooldown(int64(60))
	if pGlobal.Scope != ScopeGlobal || pGlobal.Duration != 60*time.Second {
		t.Fatalf("unexpected global policy: %+v", pGlobal)
	}

	pNil := UserCooldown(nil)
	if pNil.Duration != time.Second {
		t.Fatalf("expected 1s fallback for nil duration, got %v", pNil.Duration)
	}
}

func TestCooldown_EnforcementAndBurst(t *testing.T) {
	policy := UserCooldown(100*time.Millisecond, 2)
	mw := Cooldown([]Policy{policy}, WithCooldownAutoRespond(false))

	called := 0
	handler := mw(func(ctx *context.Context) error {
		called++
		return nil
	})

	in := &types.Interaction{
		User: &types.User{ID: types.Snowflake(1001)},
	}
	ctx := context.New(stdctx.Background(), in, nil)

	if err := handler(ctx); err != nil {
		t.Fatalf("1st call should succeed, got %v", err)
	}
	if err := handler(ctx); err != nil {
		t.Fatalf("2nd call should succeed with burst 2, got %v", err)
	}

	err := handler(ctx)
	if err == nil {
		t.Fatal("3rd call should trigger cooldown")
	}
	if !errors.Is(err, ErrOnCooldown) {
		t.Fatalf("expected ErrOnCooldown, got %v", err)
	}

	var cdErr *CooldownError
	if !errors.As(err, &cdErr) {
		t.Fatalf("expected *CooldownError type, got %T", err)
	}
	if cdErr.Scope != ScopeUser {
		t.Fatalf("expected ScopeUser, got %v", cdErr.Scope)
	}
	if cdErr.Remaining <= 0 {
		t.Fatalf("expected positive remaining time, got %v", cdErr.Remaining)
	}

	time.Sleep(120 * time.Millisecond)

	if err := handler(ctx); err != nil {
		t.Fatalf("call after refill should succeed, got %v", err)
	}
	if called != 3 {
		t.Fatalf("expected 3 successful calls, got %d", called)
	}
}

func TestCooldown_CompoundPolicies(t *testing.T) {
	policies := []Policy{
		UserCooldown(50*time.Millisecond, 1),
		ChannelCooldown(200*time.Millisecond, 2),
	}
	mw := Cooldown(policies, WithCooldownAutoRespond(false))

	handler := mw(func(ctx *context.Context) error {
		return nil
	})

	inUser1 := &types.Interaction{
		ChannelID: ptr(types.Snowflake(9001)),
		User:      &types.User{ID: types.Snowflake(101)},
	}
	ctx1 := context.New(stdctx.Background(), inUser1, nil)

	inUser2 := &types.Interaction{
		ChannelID: ptr(types.Snowflake(9001)),
		User:      &types.User{ID: types.Snowflake(102)},
	}
	ctx2 := context.New(stdctx.Background(), inUser2, nil)

	inUser3 := &types.Interaction{
		ChannelID: ptr(types.Snowflake(9001)),
		User:      &types.User{ID: types.Snowflake(103)},
	}
	ctx3 := context.New(stdctx.Background(), inUser3, nil)

	if err := handler(ctx1); err != nil {
		t.Fatalf("user 1 call 1 should succeed, got %v", err)
	}
	if err := handler(ctx1); !errors.Is(err, ErrOnCooldown) {
		t.Fatalf("user 1 call 2 should fail on user cooldown, got %v", err)
	}

	if err := handler(ctx2); err != nil {
		t.Fatalf("user 2 call 1 should succeed on channel token 2, got %v", err)
	}

	err := handler(ctx3)
	if err == nil {
		t.Fatal("user 3 call 1 should fail on channel capacity 2")
	}
	var cdErr *CooldownError
	if !errors.As(err, &cdErr) || cdErr.Scope != ScopeChannel {
		t.Fatalf("expected channel cooldown, got %v", err)
	}
}

func TestCooldown_AutoRespond(t *testing.T) {
	policy := GlobalCooldown(50 * time.Millisecond)
	mw := Cooldown([]Policy{policy}, WithCooldownAutoRespond(true))

	handler := mw(func(ctx *context.Context) error {
		return nil
	})

	in := &types.Interaction{
		ID:    types.Snowflake(1),
		Token: "tok",
	}
	ctx1 := context.New(stdctx.Background(), in, nil)
	ctx2 := context.New(stdctx.Background(), in, nil)

	if err := handler(ctx1); err != nil {
		t.Fatalf("first call should succeed, got %v", err)
	}

	err := handler(ctx2)
	if !errors.Is(err, ErrOnCooldown) {
		t.Fatalf("expected ErrOnCooldown, got %v", err)
	}

	select {
	case resp := <-ctx2.ResponseChannel():
		if resp.Type != types.CallbackTypeChannelMessageWithSource {
			t.Fatalf("expected message callback type, got %v", resp.Type)
		}
		if resp.Data == nil || resp.Data.Content == "" {
			t.Fatal("expected cooldown response message content")
		}
		if resp.Data.Flags != types.MessageFlagEphemeral {
			t.Fatalf("expected ephemeral response, got flags %v", resp.Data.Flags)
		}
	default:
		t.Fatal("expected auto-respond message on channel")
	}
}

func TestCooldown_StoreSweepAndConcurrency(t *testing.T) {
	store := NewCooldownStore()
	policy := UserCooldown(20 * time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := "test/u/" + string(rune('a'+id))
			ok, _ := store.Take(key, policy)
			if !ok {
				t.Errorf("expected initial token to be granted for key %s", key)
			}
		}(i)
	}
	wg.Wait()

	time.Sleep(30 * time.Millisecond)
	store.Sweep()

	totalRemaining := 0
	for _, sh := range store.shards {
		sh.mu.Lock()
		totalRemaining += len(sh.buckets)
		sh.mu.Unlock()
	}
	if totalRemaining != 0 {
		t.Fatalf("expected 0 buckets after sweep, got %d", totalRemaining)
	}
}

func ptr[T any](v T) *T {
	return &v
}

func TestCooldown_AutoPruneOnTake(t *testing.T) {
	store := NewCooldownStore()
	policy := UserCooldown(5 * time.Millisecond)

	for i := 0; i < 3000; i++ {
		_, _ = store.Take("key_"+strconv.Itoa(i), policy)
	}

	time.Sleep(15 * time.Millisecond)

	for i := 0; i < 3000; i++ {
		_, _ = store.Take("new_key_"+strconv.Itoa(i), policy)
	}

	total := 0
	for _, sh := range store.shards {
		sh.mu.Lock()
		total += len(sh.buckets)
		sh.mu.Unlock()
	}
	if total >= 6000 {
		t.Fatalf("expected old expired buckets to be pruned below 6000, got %d", total)
	}
}

