package middleware

import (
	stdctx "context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/context"
)

func TestStress_CooldownExactBurstLimit(t *testing.T) {
	store := NewCooldownStore()
	policy := Policy{
		Scope:    ScopeUser,
		Duration: 5 * time.Second,
		Burst:    7,
	}

	const goroutines = 1000
	var successCount int64
	var failureCount int64

	var wg sync.WaitGroup
	wg.Add(goroutines)

	start := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start
			ok, remaining := store.Take("stress_user_1", policy)
			if ok {
				atomic.AddInt64(&successCount, 1)
			} else {
				if remaining <= 0 {
					t.Errorf("expected positive remaining cooldown, got %v", remaining)
				}
				atomic.AddInt64(&failureCount, 1)
			}
		}()
	}

	close(start)
	wg.Wait()

	if successCount != 7 {
		t.Fatalf("expected exactly 7 successful burst takes, got %d", successCount)
	}
	if failureCount != goroutines-7 {
		t.Fatalf("expected %d cooldown rejections, got %d", goroutines-7, failureCount)
	}
}

func TestStress_CompoundCooldownConcurrent(t *testing.T) {
	store := NewCooldownStore()
	policies := []Policy{
		UserCooldown(10*time.Second, 2),
		ChannelCooldown(10*time.Second, 10),
		GlobalCooldown(10*time.Second, 50),
	}

	mw := Cooldown(policies, WithCooldownStore(store), WithCooldownAutoRespond(false))
	handler := mw(func(ctx *context.Context) error {
		return nil
	})

	const users = 20
	const callsPerUser = 5
	var allowed int64
	var userBlocked int64
	var chanBlocked int64
	var globalBlocked int64

	var wg sync.WaitGroup
	start := make(chan struct{})

	for u := 0; u < users; u++ {
		for c := 0; c < callsPerUser; c++ {
			wg.Add(1)
			go func(userID int) {
				defer wg.Done()
				<-start

				chID := types.Snowflake(555)
				in := &types.Interaction{
					ChannelID: &chID,
					User:      &types.User{ID: types.Snowflake(1000 + userID)},
				}
				ctx := context.New(stdctx.Background(), in, nil)

				err := handler(ctx)
				if err == nil {
					atomic.AddInt64(&allowed, 1)
					return
				}

				var cdErr *CooldownError
				if errors.As(err, &cdErr) {
					switch cdErr.Scope {
					case ScopeUser:
						atomic.AddInt64(&userBlocked, 1)
					case ScopeChannel:
						atomic.AddInt64(&chanBlocked, 1)
					case ScopeGlobal:
						atomic.AddInt64(&globalBlocked, 1)
					}
				}
			}(u)
		}
	}

	close(start)
	wg.Wait()

	if allowed > 10 {
		t.Fatalf("channel burst limit of 10 was exceeded, got %d", allowed)
	}
	totalAttempts := int64(users * callsPerUser)
	if allowed+userBlocked+chanBlocked+globalBlocked != totalAttempts {
		t.Fatalf("count mismatch: total %d != %d", totalAttempts, allowed+userBlocked+chanBlocked+globalBlocked)
	}
}

func TestStress_NilStoreAndInvalidPolicy(t *testing.T) {
	var store *CooldownStore
	ok, rem := store.Take("key", Policy{Duration: time.Second, Burst: 1})
	if !ok || rem != 0 {
		t.Fatalf("expected nil store to allow take cleanly, got ok=%v, rem=%v", ok, rem)
	}
	store.Sweep()

	var pipe *Pipeline
	pipe = pipe.Use(nil)
	if pipe != nil {
		t.Fatalf("expected nil pipeline from nil Use, got %v", pipe)
	}
	cloned := pipe.Clone()
	if cloned == nil {
		t.Fatal("expected cloned nil pipeline to return new empty pipeline")
	}
	h := pipe.Wrap(nil)
	if err := h(nil); err != nil {
		t.Fatalf("expected nil pipeline wrap to run cleanly, got %v", err)
	}
	if err := pipe.Execute(nil, nil); err != nil {
		t.Fatalf("expected nil pipeline execute to run cleanly, got %v", err)
	}

	mw := Cooldown([]Policy{{Scope: ScopeUser, Duration: -1, Burst: 1}})
	err := mw(nil)(context.New(stdctx.Background(), &types.Interaction{User: &types.User{ID: 1}}, nil))
	if !errors.Is(err, ErrInvalidCooldown) {
		t.Fatalf("expected ErrInvalidCooldown for negative duration, got %v", err)
	}

	mwUnknown := Cooldown([]Policy{{Scope: Scope(99), Duration: time.Second, Burst: 1}})
	errUnknown := mwUnknown(nil)(context.New(stdctx.Background(), &types.Interaction{User: &types.User{ID: 1}}, nil))
	if !errors.Is(errUnknown, ErrInvalidCooldown) {
		t.Fatalf("expected ErrInvalidCooldown for unknown scope, got %v", errUnknown)
	}
}

func TestStress_PermissionsAdministratorBypass(t *testing.T) {
	allPerms := []types.Permission{
		types.PermissionManageGuild,
		types.PermissionManageRoles,
		types.PermissionManageChannels,
		types.PermissionKickMembers,
		types.PermissionBanMembers,
		types.PermissionManageWebhooks,
		types.PermissionModerateMembers,
	}

	for _, reqPerm := range allPerms {
		mw := RequirePermissions(reqPerm, reqPerm, WithPermissionAutoRespond(false))
		handler := mw(func(ctx *context.Context) error {
			return nil
		})

		gid := types.Snowflake(100)
		appAdmin := types.PermissionAdministrator
		in := &types.Interaction{
			GuildID:        &gid,
			AppPermissions: &appAdmin,
			Member: &types.Member{
				Permissions: types.PermissionAdministrator,
			},
		}
		ctx := context.New(stdctx.Background(), in, nil)

		if err := handler(ctx); err != nil {
			t.Fatalf("administrator failed to bypass perm requirement %d: %v", reqPerm, err)
		}
	}
}

func TestStress_UnknownPermissionHexFormatting(t *testing.T) {
	unknown := types.Permission(1 << 62)
	names := PermissionNames(unknown)
	if len(names) != 1 {
		t.Fatalf("expected 1 formatted name, got %d", len(names))
	}
	expected := fmt.Sprintf("Unknown Permission (0x%X)", uint64(unknown))
	if names[0] != expected {
		t.Fatalf("expected %s, got %s", expected, names[0])
	}
}

func TestStress_GuardPanicStorm(t *testing.T) {
	var panicCount int64
	guard := Guard(
		WithPanicHandler(func(ctx *context.Context, recovered any, stack []byte) {
			atomic.AddInt64(&panicCount, 1)
			if atomic.LoadInt64(&panicCount)%2 == 0 {
				panic("nested telemetry panic")
			}
		}),
		WithEphemeral(true),
	)

	handler := guard(func(ctx *context.Context) error {
		panic("simulated disaster")
	})

	const goroutines = 500
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ctx := context.New(stdctx.Background(), &types.Interaction{ID: 1, Token: "t"}, nil)
			err := handler(ctx)
			if !errors.Is(err, ErrPanicRecovered) {
				t.Errorf("expected ErrPanicRecovered, got %v", err)
			}
		}()
	}

	wg.Wait()

	if panicCount != goroutines {
		t.Fatalf("expected %d panic recoveries, got %d", goroutines, panicCount)
	}
}

