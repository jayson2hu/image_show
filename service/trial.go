package service

import (
	"crypto/sha256"
	"encoding/hex"
)

func TrialAnonymousID(ip, fingerprint string) string {
	sum := sha256.Sum256([]byte(ip + "|" + fingerprint))
	return hex.EncodeToString(sum[:])
}

func ResetTrialStoreForTest() {
}
