package app

import (
	stdctx "context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strconv"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/command"
)

var (
	ErrDevGuildRequired = errors.New("dev guild id required when dev mode is enabled")
	ErrNoRESTClient     = errors.New("rest client is not configured")
)

type SyncResult struct {
	GlobalSynced bool
	GuildsSynced map[types.Snowflake]bool
	Skipped      int
	Updated      int
}

func (a *App) Sync(ctx stdctx.Context, force ...bool) error {
	_, err := a.SyncWithResult(ctx, force...)
	return err
}

func (a *App) SyncWithResult(ctx stdctx.Context, force ...bool) (*SyncResult, error) {
	if a.REST == nil {
		return nil, ErrNoRESTClient
	}

	forceSync := len(force) > 0 && force[0]

	if a.config.DevMode && a.config.DevGuildID == 0 {
		return nil, ErrDevGuildRequired
	}

	a.syncMu.Lock()
	defer a.syncMu.Unlock()

	a.mu.Lock()
	seen := make(map[*commandEntry]bool)
	var entries []*commandEntry
	for _, entry := range a.commands {
		if !seen[entry] {
			seen[entry] = true
			entries = append(entries, entry)
		}
	}
	syncedHashes := make(map[string]string, len(a.syncedHashes))
	for k, v := range a.syncedHashes {
		syncedHashes[k] = v
	}
	a.mu.Unlock()

	result := &SyncResult{
		GuildsSynced: make(map[types.Snowflake]bool),
	}

	var globalList []types.ApplicationCommand
	guildMap := make(map[types.Snowflake][]types.ApplicationCommand)

	if a.config.DevMode {
		for _, entry := range entries {
			guildMap[a.config.DevGuildID] = append(guildMap[a.config.DevGuildID], command.ToApplicationCommand(entry.cmd))
		}
	} else {
		for _, entry := range entries {
			appCmd := command.ToApplicationCommand(entry.cmd)
			if entry.cmd.GuildID != 0 {
				guildMap[entry.cmd.GuildID] = append(guildMap[entry.cmd.GuildID], appCmd)
			} else {
				globalList = append(globalList, appCmd)
			}
		}
	}

	sortCommands(globalList)
	for _, list := range guildMap {
		sortCommands(list)
	}

	updatedHashes := make(map[string]string)

	if !a.config.DevMode {
		if len(globalList) > 0 || syncedHashes["global"] != "" || forceSync {
			hash, err := hashCommands(globalList)
			if err != nil {
				return nil, err
			}

			if !forceSync && syncedHashes["global"] == hash {
				result.Skipped++
			} else {
				_, err := a.REST.BulkOverwriteGlobalCommands(ctx, a.config.AppID, globalList)
				if err != nil {
					return nil, err
				}
				updatedHashes["global"] = hash
				result.GlobalSynced = true
				result.Updated++
			}
		}
	}

	var guildIDs []types.Snowflake
	for gid := range guildMap {
		guildIDs = append(guildIDs, gid)
	}
	sort.Slice(guildIDs, func(i, j int) bool {
		return guildIDs[i] < guildIDs[j]
	})

	for _, gid := range guildIDs {
		cmds := guildMap[gid]
		hash, err := hashCommands(cmds)
		if err != nil {
			return nil, err
		}

		key := "guild/" + strconv.FormatUint(uint64(gid), 10)
		if !forceSync && syncedHashes[key] == hash {
			result.Skipped++
		} else {
			_, err := a.REST.BulkOverwriteGuildCommands(ctx, a.config.AppID, gid, cmds)
			if err != nil {
				return nil, err
			}
			updatedHashes[key] = hash
			result.GuildsSynced[gid] = true
			result.Updated++
		}
	}

	if len(updatedHashes) > 0 {
		a.mu.Lock()
		for k, v := range updatedHashes {
			a.syncedHashes[k] = v
		}
		a.mu.Unlock()
	}

	return result, nil
}

func sortCommands(cmds []types.ApplicationCommand) {
	sort.Slice(cmds, func(i, j int) bool {
		if cmds[i].Name == cmds[j].Name {
			return cmds[i].Type < cmds[j].Type
		}
		return cmds[i].Name < cmds[j].Name
	})
}

func hashCommands(cmds []types.ApplicationCommand) (string, error) {
	if cmds == nil {
		cmds = []types.ApplicationCommand{}
	}
	data, err := json.Marshal(cmds)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}

