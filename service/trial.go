package service

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

type trialEntry struct {
	UsedAt time.Time
}

var trialStore = struct {
	sync.Mutex
	items map[string]trialEntry
}{items: make(map[string]trialEntry)}

func TrialAnonymousID(ip, fingerprint string) string {
	sum := sha256.Sum256([]byte(ip + "|" + fingerprint))
	return hex.EncodeToString(sum[:])
}

func CheckTrialEligible(ip, fingerprint string) (string, bool) {
	anonymousID := TrialAnonymousID(ip, fingerprint)
	trialStore.Lock()
	defer trialStore.Unlock()

	entry, ok := trialStore.items[anonymousID]
	if ok && time.Since(entry.UsedAt) < 30*24*time.Hour {
		return anonymousID, false
	}
	return anonymousID, true
}

func MarkTrialUsed(anonymousID string) {
	trialStore.Lock()
	trialStore.items[anonymousID] = trialEntry{UsedAt: time.Now()}
	trialStore.Unlock()
}

func ResetTrialStoreForTest() {
	trialStore.Lock()
	trialStore.items = make(map[string]trialEntry)
	trialStore.Unlock()
}
