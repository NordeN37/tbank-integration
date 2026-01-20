package tbank

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

func generateToken(data map[string]interface{}, password string) string {
	// Добавляем Password как в документации
	data["Password"] = password
	defer func() {
		delete(data, "Password")
	}()

	// Сортируем ключи
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Конкатенируем значения
	var sb strings.Builder
	for _, k := range keys {
		// Преобразуем все значения к строкам
		var val string
		switch v := data[k].(type) {
		case string:
			val = v
		case int:
			val = fmt.Sprintf("%d", v)
		case int64:
			val = fmt.Sprintf("%d", v)
		case float64:
			val = fmt.Sprintf("%.0f", v) // Без десятичных для Amount
		default:
			// Пропускаем нестроковые типы или преобразуем
			val = fmt.Sprintf("%v", v)
		}
		sb.WriteString(val)
	}

	sbVal := sb.String()
	hash := sha256.Sum256([]byte(sbVal))
	return hex.EncodeToString(hash[:])
}
