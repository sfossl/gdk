package context

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

func setupRoleTestContext(handler http.HandlerFunc) (*Context, func()) {
	ts := httptest.NewServer(handler)
	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))

	guildID := types.Snowflake(100)
	botAppID := types.Snowflake(1)
	invokerID := types.Snowflake(2)

	c := &Context{
		ctx:     context.Background(),
		GuildID: &guildID,
		AppID:   botAppID,
		REST:    client,
		Author:  &types.User{ID: invokerID},
		InvokerMember:  &types.Member{Roles: []types.Snowflake{types.Snowflake(30)}},
	}
	return c, ts.Close
}

func standardRoleGuildHandler(extraHandler func(w http.ResponseWriter, r *http.Request) bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if extraHandler != nil && extraHandler(w, r) {
			return
		}
		switch r.URL.Path {
		case "/guilds/100":
			_ = json.NewEncoder(w).Encode(types.Guild{
				ID:      types.Snowflake(100),
				OwnerID: types.Snowflake(999),
				Roles: []types.Role{
					{ID: types.Snowflake(100), Position: 0, Name: "@everyone"},
					{ID: types.Snowflake(10), Position: 1, Name: "Muted"},
					{ID: types.Snowflake(20), Position: 5, Name: "Member"},
					{ID: types.Snowflake(30), Position: 10, Name: "Mod"},
					{ID: types.Snowflake(40), Position: 15, Name: "Admin"},
				},
			})
		case "/guilds/100/members/1":
			_ = json.NewEncoder(w).Encode(types.Member{
				Roles: []types.Snowflake{types.Snowflake(40)},
			})
		case "/guilds/100/members/2":
			_ = json.NewEncoder(w).Encode(types.Member{
				Roles: []types.Snowflake{types.Snowflake(30)},
			})
		case "/guilds/100/roles":
			_ = json.NewEncoder(w).Encode([]types.Role{
				{ID: types.Snowflake(100), Position: 0, Name: "@everyone"},
				{ID: types.Snowflake(10), Position: 1, Name: "Muted"},
				{ID: types.Snowflake(20), Position: 5, Name: "Member"},
				{ID: types.Snowflake(30), Position: 10, Name: "Mod"},
				{ID: types.Snowflake(40), Position: 15, Name: "Admin"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func TestRoleDefaultsAndTargeting(t *testing.T) {
	c := &Context{}
	act := c.RoleSuite()
	if act.ID() != 0 {
		t.Fatalf("expected role ID 0, got %v", act.ID())
	}

	actTarget := c.RoleSuite(types.Snowflake(20))
	if actTarget.ID() != types.Snowflake(20) {
		t.Fatalf("expected role ID 20, got %v", actTarget.ID())
	}
}

func TestRoleValidationErrors(t *testing.T) {
	cNoGuild := &Context{}
	if ok, err := cNoGuild.CanManageRole(types.Snowflake(20)); ok || !errors.Is(err, ErrNotInGuild) {
		t.Fatalf("expected ErrNotInGuild, got %v", err)
	}

	guildID := types.Snowflake(100)
	cNoREST := &Context{GuildID: &guildID}
	if ok, err := cNoREST.CanManageRole(types.Snowflake(20)); ok || !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}

	cInvalidID := &Context{GuildID: &guildID}
	if ok, err := cInvalidID.CanManageRole(0); ok || !errors.Is(err, ErrInvalidRoleID) {
		t.Fatalf("expected ErrInvalidRoleID, got %v", err)
	}

	if ok, err := cInvalidID.CanManageRole(types.Snowflake(100)); ok || !errors.Is(err, ErrCannotModifyEveryoneRole) {
		t.Fatalf("expected ErrCannotModifyEveryoneRole, got %v", err)
	}
}

func TestRoleHierarchyCheck(t *testing.T) {
	c, cleanup := setupRoleTestContext(standardRoleGuildHandler(nil))
	defer cleanup()

	ok, err := c.CanManageRole(types.Snowflake(20))
	if !ok || err != nil {
		t.Fatalf("expected success for role 20, got ok %v err %v", ok, err)
	}

	ok, err = c.CanManageRole(types.Snowflake(30))
	if ok || !errors.Is(err, ErrRoleOutranksUser) {
		t.Fatalf("expected ErrRoleOutranksUser for role 30, got ok %v err %v", ok, err)
	}

	ok, err = c.CanManageRole(types.Snowflake(40))
	if ok || !errors.Is(err, ErrRoleOutranksBot) {
		t.Fatalf("expected ErrRoleOutranksBot for role 40, got ok %v err %v", ok, err)
	}

	ok, err = c.CanManageRole(types.Snowflake(9999))
	if ok || !errors.Is(err, ErrRoleNotFound) {
		t.Fatalf("expected ErrRoleNotFound for nonexistent role, got ok %v err %v", ok, err)
	}

	c.Author = &types.User{ID: types.Snowflake(999)}
	c.InvokerMember = &types.Member{Roles: nil}
	ok, err = c.CanManageRole(types.Snowflake(30))
	if !ok || err != nil {
		t.Fatalf("expected owner bypass for role 30, got ok %v err %v", ok, err)
	}
}

func TestRoleGet(t *testing.T) {
	c, cleanup := setupRoleTestContext(standardRoleGuildHandler(nil))
	defer cleanup()

	role, err := c.RoleSuite(types.Snowflake(20)).Get()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if role.Name != "Member" {
		t.Fatalf("expected Member, got %s", role.Name)
	}

	if _, err := c.RoleSuite(0).Get(); !errors.Is(err, ErrInvalidRoleID) {
		t.Fatalf("expected ErrInvalidRoleID, got %v", err)
	}

	if _, err := c.RoleSuite(types.Snowflake(888)).Get(); !errors.Is(err, ErrRoleNotFound) {
		t.Fatalf("expected ErrRoleNotFound, got %v", err)
	}
}

func TestRoleCreate(t *testing.T) {
	name := "VIP"
	c, cleanup := setupRoleTestContext(standardRoleGuildHandler(func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == http.MethodPost && r.URL.Path == "/guilds/100/roles" {
			if r.Header.Get("X-Audit-Log-Reason") != "new perk" {
				t.Fatalf("unexpected audit header %s", r.Header.Get("X-Audit-Log-Reason"))
			}
			_ = json.NewEncoder(w).Encode(types.Role{
				ID:   types.Snowflake(50),
				Name: "VIP",
			})
			return true
		}
		return false
	}))
	defer cleanup()

	role, err := c.RoleSuite().Create(rest.RolePayload{Name: &name}, "new", "perk")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if role.ID != types.Snowflake(50) {
		t.Fatalf("expected role ID 50, got %v", role.ID)
	}
}

func TestRoleEditAndHelpers(t *testing.T) {
	c, cleanup := setupRoleTestContext(standardRoleGuildHandler(func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == http.MethodPatch && r.URL.Path == "/guilds/100/roles/20" {
			if r.Header.Get("X-Audit-Log-Reason") != "role tweak" {
				t.Fatalf("unexpected audit header %s", r.Header.Get("X-Audit-Log-Reason"))
			}
			var p rest.RolePayload
			_ = json.NewDecoder(r.Body).Decode(&p)
			role := types.Role{ID: types.Snowflake(20)}
			if p.Name != nil {
				role.Name = *p.Name
			}
			if p.Color != nil {
				role.Color = *p.Color
			}
			if p.Permissions != nil {
				role.Permissions = *p.Permissions
			}
			if p.Hoist != nil {
				role.Hoist = *p.Hoist
			}
			if p.Mentionable != nil {
				role.Mentionable = *p.Mentionable
			}
			_ = json.NewEncoder(w).Encode(role)
			return true
		}
		return false
	}))
	defer cleanup()

	act := c.RoleSuite(types.Snowflake(20))

	rName, err := act.SetName("SuperMember", "role", "tweak")
	if err != nil || rName.Name != "SuperMember" {
		t.Fatalf("expected SuperMember, got %v err %v", rName, err)
	}

	rColor, err := act.SetColor(0xFF0000, "role", "tweak")
	if err != nil || rColor.Color != 0xFF0000 {
		t.Fatalf("expected red color, got %v err %v", rColor, err)
	}

	rPerms, err := act.SetPermissions(types.PermissionSendMessages, "role", "tweak")
	if err != nil || rPerms.Permissions != types.PermissionSendMessages {
		t.Fatalf("expected send messages, got %v err %v", rPerms, err)
	}

	rHoist, err := act.SetHoist(true, "role", "tweak")
	if err != nil || !rHoist.Hoist {
		t.Fatalf("expected hoist true, got %v err %v", rHoist, err)
	}

	rMent, err := act.SetMentionable(true, "role", "tweak")
	if err != nil || !rMent.Mentionable {
		t.Fatalf("expected mentionable true, got %v err %v", rMent, err)
	}
}

func TestRoleDelete(t *testing.T) {
	c, cleanup := setupRoleTestContext(standardRoleGuildHandler(func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == http.MethodDelete && r.URL.Path == "/guilds/100/roles/20" {
			if r.Header.Get("X-Audit-Log-Reason") != "cleanup role" {
				t.Fatalf("unexpected audit header %s", r.Header.Get("X-Audit-Log-Reason"))
			}
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		return false
	}))
	defer cleanup()

	err := c.RoleSuite(types.Snowflake(20)).Delete("cleanup", "role")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestRoleAssignAndRemove(t *testing.T) {
	c, cleanup := setupRoleTestContext(standardRoleGuildHandler(func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == http.MethodPut && r.URL.Path == "/guilds/100/members/500/roles/20" {
			if r.Header.Get("X-Audit-Log-Reason") != "promoted" {
				t.Fatalf("unexpected audit header %s", r.Header.Get("X-Audit-Log-Reason"))
			}
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		if r.Method == http.MethodDelete && r.URL.Path == "/guilds/100/members/500/roles/20" {
			if r.Header.Get("X-Audit-Log-Reason") != "demoted" {
				t.Fatalf("unexpected audit header %s", r.Header.Get("X-Audit-Log-Reason"))
			}
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		return false
	}))
	defer cleanup()

	act := c.RoleSuite(types.Snowflake(20))

	if err := act.Assign(0); !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("expected ErrInvalidTarget, got %v", err)
	}
	if err := act.Remove(0); !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("expected ErrInvalidTarget, got %v", err)
	}

	if err := act.Assign(types.Snowflake(500), "promoted"); err != nil {
		t.Fatalf("unexpected error on assign: %v", err)
	}
	if err := act.Remove(types.Snowflake(500), "demoted"); err != nil {
		t.Fatalf("unexpected error on remove: %v", err)
	}
}

func TestManageRole_Unified(t *testing.T) {
	var created, edited, deleted bool
	var auditReason string

	c, cleanup := setupRoleTestContext(standardRoleGuildHandler(func(w http.ResponseWriter, r *http.Request) bool {
		auditReason = r.Header.Get("X-Audit-Log-Reason")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/guilds/100/roles":
			created = true
			_ = json.NewEncoder(w).Encode(types.Role{ID: types.Snowflake(50), Name: "CreatedRole"})
			return true
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/100/roles/20":
			edited = true
			_ = json.NewEncoder(w).Encode(types.Role{ID: types.Snowflake(20), Name: "EditedRole"})
			return true
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/100/roles/20":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
			return true
		default:
			return false
		}
	}))
	defer cleanup()

	hoist := true
	role, err := c.ManageRole(RoleOptions{
		Type:   RoleCreate,
		Name:   "CreatedRole",
		Color:  0x3498DB,
		Hoist:  &hoist,
		Reason: "create role",
	})
	if err != nil || !created || role == nil || role.Name != "CreatedRole" || auditReason != "create role" {
		t.Fatalf("RoleCreate failed: err=%v, created=%v, role=%v", err, created, role)
	}

	role, err = c.ManageRole(RoleOptions{
		Type:   RoleEdit,
		RoleID: types.Snowflake(20),
		Name:   "EditedRole",
		Reason: "edit role",
	})
	if err != nil || !edited || role == nil || role.Name != "EditedRole" || auditReason != "edit role" {
		t.Fatalf("RoleEdit failed: err=%v, edited=%v, role=%v", err, edited, role)
	}

	role, err = c.ManageRole(RoleOptions{
		Type:   RoleDelete,
		RoleID: types.Snowflake(20),
		Reason: "delete role",
	})
	if err != nil || !deleted || role != nil || auditReason != "delete role" {
		t.Fatalf("RoleDelete failed: err=%v, deleted=%v", err, deleted)
	}

	_, err = c.ManageRole(RoleOptions{
		Type:   RoleEdit,
		RoleID: 0,
	})
	if !errors.Is(err, ErrInvalidRoleID) {
		t.Fatalf("expected ErrInvalidRoleID, got %v", err)
	}

	_, err = c.ManageRole(RoleOptions{
		Type: RoleAction(999),
	})
	if !errors.Is(err, ErrInvalidRoleAction) {
		t.Fatalf("expected ErrInvalidRoleAction, got %v", err)
	}
}

func TestManageRole_OptionsAndShortcuts(t *testing.T) {
	var deleted bool
	var auditReason string
	c, cleanup := setupRoleTestContext(standardRoleGuildHandler(func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == "DELETE" && r.URL.Path == "/guilds/100/roles/20" {
			deleted = true
			auditReason = r.Header.Get("X-Audit-Log-Reason")
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		return false
	}))
	defer cleanup()

	_, err := c.ManageRole("delete", types.Snowflake(20), "cleanup role")
	if err != nil {
		t.Fatalf("unexpected role delete shortcut error: %v", err)
	}
	if !deleted || auditReason != "cleanup role" {
		t.Fatalf("expected role delete with reason, got deleted=%v reason=%s", deleted, auditReason)
	}

	deleted = false
	_, err = c.ManageRole(Options{
		Type:   "delete",
		Role:   types.Snowflake(20),
		Reason: "options delete",
	})
	if err != nil {
		t.Fatalf("unexpected role delete options error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected deleted true")
	}

	deleted = false
	_, err = c.Role(Options{
		Type:   "delete",
		Role:   types.Snowflake(20),
		Reason: "c.Role options delete",
	})
	if err != nil {
		t.Fatalf("unexpected c.Role delete options error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected deleted true")
	}

	deleted = false
	_, err = c.Role("delete", types.Snowflake(20), "c.Role direct delete")
	if err != nil {
		t.Fatalf("unexpected c.Role positional delete error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected deleted true")
	}
}

