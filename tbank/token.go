package tbank

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

func generateToken(data map[string]string, password string) string {
	data["Password"] = password

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(data[k])
	}

	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}
