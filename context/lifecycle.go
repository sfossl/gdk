package context

import (
	"context"
	"strings"
	"time"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

func (c *Context) scheduleLifecycleTimers(res Options, msg *types.Message) {
	if c.REST == nil || c.Interaction == nil {
		return
	}

	compTimeout := res.ResolveComponentTimeout()
	deleteAfter := res.ResolveDeleteAfter()

	if compTimeout <= 0 && deleteAfter <= 0 {
		return
	}

	c.lifecycleMu.Lock()
	defer c.lifecycleMu.Unlock()

	if len(res.Components) > 0 {
		c.activeComponents = res.Components
	}
	if msg != nil && msg.ID != 0 {
		c.msgID = msg.ID
	}

	if deleteAfter > 0 && !res.Ephemeral {
		if c.deleteAfterTimer != nil {
			c.deleteAfterTimer.Stop()
		}
		c.deleteAfterTimer = time.AfterFunc(deleteAfter, func() {
			c.executeDeleteAfter()
		})
	}

	if compTimeout > 0 && len(c.activeComponents) > 0 {
		if c.compTimeoutTimer != nil {
			c.compTimeoutTimer.Stop()
		}
		c.compTimeoutTimer = time.AfterFunc(compTimeout, func() {
			c.executeComponentTimeout()
		})
	}
}

func (c *Context) executeDeleteAfter() {
	c.lifecycleMu.Lock()
	if c.isDeleted {
		c.lifecycleMu.Unlock()
		return
	}
	c.isDeleted = true
	if c.compTimeoutTimer != nil {
		c.compTimeoutTimer.Stop()
	}
	c.lifecycleMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if c.Interaction != nil && c.Interaction.Token != "" {
		err := c.REST.DeleteOriginalInteractionResponse(ctx, c.AppID, c.Interaction.Token)
		if err == nil || isIgnorableNotFound(err) {
			return
		}
	}

	if c.ChannelID != 0 && c.msgID != 0 {
		_ = c.REST.DeleteMessage(ctx, c.ChannelID, c.msgID)
	}
}

func (c *Context) executeComponentTimeout() {
	c.lifecycleMu.Lock()
	if c.isDeleted || len(c.activeComponents) == 0 {
		c.lifecycleMu.Unlock()
		return
	}
	disabled := DisableAllComponents(c.activeComponents)
	c.activeComponents = disabled
	c.lifecycleMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if c.Interaction != nil && c.Interaction.Token != "" {
		payload := rest.MessagePayload{
			Components: disabled,
		}
		_, err := c.REST.EditOriginalInteractionResponse(ctx, c.AppID, c.Interaction.Token, payload)
		if err == nil || isIgnorableNotFound(err) {
			return
		}
	}

	if c.ChannelID != 0 && c.msgID != 0 {
		payload := rest.MessagePayload{
			Components: disabled,
		}
		_, _ = c.REST.EditMessage(ctx, c.ChannelID, c.msgID, payload)
	}
}

func (c *Context) DisableComponents() error {
	var rows []types.Component
	if c.Interaction != nil && c.Interaction.Message != nil && len(c.Interaction.Message.Components) > 0 {
		rows = c.Interaction.Message.Components
	} else {
		c.lifecycleMu.Lock()
		rows = c.activeComponents
		c.lifecycleMu.Unlock()
	}

	if len(rows) == 0 {
		return nil
	}

	disabled := DisableAllComponents(rows)
	c.lifecycleMu.Lock()
	c.activeComponents = disabled
	c.lifecycleMu.Unlock()

	_, err := c.UpdateMessage(Options{
		Components: disabled,
	})
	return err
}

func DisableAllComponents(rows []types.Component) []types.Component {
	out := make([]types.Component, len(rows))
	for i, row := range rows {
		outRow := row
		if len(row.Components) > 0 {
			inner := make([]types.Component, len(row.Components))
			for j, comp := range row.Components {
				c := comp
				c.Disabled = true
				inner[j] = c
			}
			outRow.Components = inner
		} else {
			outRow.Disabled = true
		}
		out[i] = outRow
	}
	return out
}

func isIgnorableNotFound(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "404") || strings.Contains(s, "10008") || strings.Contains(s, "unknown message") || strings.Contains(s, "unknown interaction")
}
