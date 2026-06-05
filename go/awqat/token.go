package awqat

import (
	"encoding/json"
	"os"
	"time"
)

// tokenCache; login/refresh sonrası token'ın diske yazılan kalıcı hâli.
// Sonraki çalıştırmalarda geçerliyse yeniden login ATILMAZ (kota korunur).
type tokenCache struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	Expiry       time.Time `json:"expiry"`
}

// loadTokenCache; önbellek dosyasını okur. Yoksa/bozuksa/boşsa ok=false döner.
func loadTokenCache(path string) (tokenCache, bool) {
	if path == "" {
		return tokenCache{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return tokenCache{}, false
	}
	var tc tokenCache
	if err := json.Unmarshal(data, &tc); err != nil || tc.AccessToken == "" {
		return tokenCache{}, false
	}
	return tc, true
}

// saveTokenCache; token'ı yalnızca sahibin okuyabileceği (0600) bir dosyaya yazar.
// Hata sessizce yutulur — önbellek bir kolaylıktır, kritik değildir.
func saveTokenCache(path string, tc tokenCache) {
	if path == "" {
		return
	}
	if data, err := json.MarshalIndent(tc, "", "  "); err == nil {
		_ = os.WriteFile(path, data, 0o600)
	}
}
