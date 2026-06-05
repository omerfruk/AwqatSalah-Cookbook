package awqat

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// tokenCacheFile; login sonrası alınan token'ın saklandığı dosya adı (kök .env yanında).
const tokenCacheFile = ".awqat-token.json"

// Config; örneğin ihtiyaç duyduğu tüm ayarları tutar.
// Kaynak önceliği: (1) işletim sistemi ortam değişkenleri, (2) kök .env dosyası.
type Config struct {
	BaseURL  string
	Email    string
	Password string

	// Baz konum (isimden bulunur). Varsayılan: Türkiye / Isparta.
	Country string
	State   string

	// Opsiyonel ID override (isim aramasını atlamak / kotadan tasarruf için).
	// 0 ise yok sayılır ve isimden bulunur.
	CountryID int64
	StateID   int64
	CityID    int64

	// Opsiyonel elle token (AWQAT_ACCESS_TOKEN / AWQAT_REFRESH_TOKEN). Verilirse
	// login atlanır; geçersizse 401 ile otomatik yeniden login olunur.
	AccessToken  string
	RefreshToken string

	// Token önbelleğinin yazılacağı/okunacağı tam yol (kök .env yanında).
	TokenCachePath string
}

// Load; ortam değişkenlerini ve kök .env dosyasını okuyup Config üretir.
// .env, çalışılan klasörden YUKARI doğru aranır (alt klasörden çalıştırınca da bulunsun diye).
func Load() (*Config, error) {
	dotenv, rootDir := loadDotEnvUpwards()

	// get: önce OS ortam değişkeni, sonra .env, sonra varsayılan.
	get := func(key, def string) string {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
		if v, ok := dotenv[key]; ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
		return def
	}

	cfg := &Config{
		BaseURL:        strings.TrimRight(get("AWQAT_BASE_URL", "https://awqatsalah.diyanet.gov.tr"), "/"),
		Email:          get("AWQAT_EMAIL", ""),
		Password:       get("AWQAT_PASSWORD", ""),
		Country:        get("AWQAT_COUNTRY", "Türkiye"),
		State:          get("AWQAT_STATE", "Isparta"),
		CountryID:      parseID(get("AWQAT_COUNTRY_ID", "")),
		StateID:        parseID(get("AWQAT_STATE_ID", "")),
		CityID:         parseID(get("AWQAT_CITY_ID", "")),
		AccessToken:    get("AWQAT_ACCESS_TOKEN", ""),
		RefreshToken:   get("AWQAT_REFRESH_TOKEN", ""),
		TokenCachePath: filepath.Join(rootDir, tokenCacheFile),
	}

	if cfg.Email == "" || cfg.Password == "" {
		return nil, fmt.Errorf(
			"AWQAT_EMAIL ve AWQAT_PASSWORD gerekli — kök dizinde .env oluşturup doldurun:\n" +
				"    cp .env.example .env   (sonra e-posta/şifrenizi yazın)")
	}
	return cfg, nil
}

// loadDotEnvUpwards; CWD'den köke kadar her klasörde .env arar.
// Bulunan .env'in klasörünü (token önbelleği için kök) ve değerlerini döndürür.
// .env yoksa CWD kök kabul edilir.
func loadDotEnvUpwards() (map[string]string, string) {
	dir, err := os.Getwd()
	if err != nil {
		return map[string]string{}, "."
	}
	start := dir
	for {
		if data, err := os.ReadFile(filepath.Join(dir, ".env")); err == nil {
			return parseDotEnv(string(data)), dir
		}
		parent := filepath.Dir(dir)
		if parent == dir { // kök dizine ulaşıldı, .env yok
			return map[string]string{}, start
		}
		dir = parent
	}
}

// parseDotEnv; basit KEY=VALUE biçimini ayrıştırır (# yorum, boş satır, tırnak soyma).
func parseDotEnv(s string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.Trim(strings.TrimSpace(line[eq+1:]), `"'`)
		if key != "" {
			out[key] = val
		}
	}
	return out
}

func parseID(s string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n
}
