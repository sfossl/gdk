package defaults

import "time"

const (
	DMReceipt        = "Delivered to direct messages."
	PanicMessage     = "An unexpected error occurred while executing this command."
	CooldownMessage  = "You are on cooldown. Please wait %s before trying again."
	UserMissingPerms = "You do not have the required permissions to execute this command (%s)."
	BotMissingPerms  = "The bot is missing the required permissions to execute this command (%s)."

	AutoDeferDelay = 2500 * time.Millisecond
	Timeout        = 14*time.Minute + 50*time.Second

	MaxTimeoutDuration = 28 * 24 * time.Hour
	MaxBanDeleteSec    = 604800
	MaxSlowmodeSeconds = 21600
)

