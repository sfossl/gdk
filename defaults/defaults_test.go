package defaults

import (
	"testing"
	"time"
)

func TestDefaultsValues(t *testing.T) {
	if DMReceipt != "Delivered to direct messages." {
		t.Fatalf("unexpected DMReceipt: %s", DMReceipt)
	}
	if PanicMessage != "An unexpected error occurred while executing this command." {
		t.Fatalf("unexpected PanicMessage: %s", PanicMessage)
	}
	if CooldownMessage != "You are on cooldown. Please wait %s before trying again." {
		t.Fatalf("unexpected CooldownMessage: %s", CooldownMessage)
	}
	if UserMissingPerms != "You do not have the required permissions to execute this command (%s)." {
		t.Fatalf("unexpected UserMissingPerms: %s", UserMissingPerms)
	}
	if BotMissingPerms != "The bot is missing the required permissions to execute this command (%s)." {
		t.Fatalf("unexpected BotMissingPerms: %s", BotMissingPerms)
	}
	if AutoDeferDelay != 2500*time.Millisecond {
		t.Fatalf("unexpected AutoDeferDelay: %v", AutoDeferDelay)
	}
	if Timeout != 14*time.Minute+50*time.Second {
		t.Fatalf("unexpected Timeout: %v", Timeout)
	}
}

