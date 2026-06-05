package awqat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

// TestEndToEnd; sahte (mock) sunucuya karşı tüm akışı doğrular:
// login → günlük içerik → il/ilçe çözümleme → ilçe detay → günlük namaz vakti.
// Gerçek API'ye ihtiyaç duymaz.
func TestEndToEnd(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/Auth/Login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("login POST bekleniyordu, gelen: %s", r.Method)
		}
		writeJSON(w, `{"success":true,"message":null,"data":{"accessToken":"ACCESS","refreshToken":"REFRESH"}}`)
	})
	mux.HandleFunc("/api/DailyContent", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"success":true,"data":{"id":333,"dayOfYear":333,"verse":"V","verseSource":"VS","hadith":"H","hadithSource":"HS","pray":"P","praySource":null}}`)
	})
	mux.HandleFunc("/api/Place/Countries", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer ACCESS" {
			t.Errorf("Authorization header yanlış: %q", got)
		}
		writeJSON(w, `{"success":true,"data":[{"id":1,"code":"NORTH CYPRUS","name":"KUZEY KIBRIS"},{"id":2,"code":"TURKEY","name":"TÜRKİYE"}]}`)
	})
	mux.HandleFunc("/api/Place/States/2", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"success":true,"data":[{"id":540,"code":"ISPARTA","name":"ISPARTA"}]}`)
	})
	mux.HandleFunc("/api/Place/Cities/540", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"success":true,"data":[{"id":9341,"code":"ISPARTA","name":"ISPARTA"}]}`)
	})
	mux.HandleFunc("/api/Place/CityDetail/9341", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"success":true,"data":{"id":"9341","name":"ISPARTA","code":null,"geographicQiblaAngle":"154","distanceToKaaba":"2100","qiblaAngle":"150","city":"ISPARTA","cityEn":null,"country":"TÜRKİYE","countryEn":"TÜRKİYE"}}`)
	})
	mux.HandleFunc("/api/PrayerTime/Daily/9341", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"success":true,"data":[{"fajr":"04:00","sunrise":"05:30","dhuhr":"13:00","asr":"16:45","maghrib":"20:15","isha":"21:40","gregorianDateShort":"05.06.2026","hijriDateShort":"19.11.1447","greenwichMeanTimezone":3}]}`)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := New(&Config{BaseURL: srv.URL, Email: "a@b.c", Password: "x", Country: "Türkiye", State: "Isparta"})
	ctx := context.Background()

	// Login (/api/Auth/Login)
	if err := c.EnsureAuth(ctx); err != nil {
		t.Fatalf("login: %v", err)
	}
	if c.AccessToken() != "ACCESS" {
		t.Fatalf("access token = %q, beklenen ACCESS", c.AccessToken())
	}

	// DailyContent
	dc, err := c.DailyContent(ctx)
	if err != nil || dc.Verse != "V" || dc.DayOfYear != 333 {
		t.Fatalf("dailyContent yanlış: %+v err=%v", dc, err)
	}

	// İl/İlçe: "Türkiye" araması "TÜRKİYE" ile eşleşmeli (Türkçe duyarsız)
	countries, err := c.Countries(ctx)
	if err != nil {
		t.Fatalf("countries: %v", err)
	}
	country, ok := FindByName(countries, "Türkiye")
	if !ok || country.ID != 2 {
		t.Fatalf("ülke bulunamadı / yanlış: %+v ok=%v", country, ok)
	}

	states, err := c.States(ctx, country.ID)
	if err != nil {
		t.Fatalf("states: %v", err)
	}
	state, ok := FindByName(states, "Isparta")
	if !ok || state.ID != 540 {
		t.Fatalf("il bulunamadı / yanlış: %+v ok=%v", state, ok)
	}

	cities, err := c.Cities(ctx, state.ID)
	if err != nil || len(cities) != 1 || cities[0].ID != 9341 {
		t.Fatalf("ilçe yanlış: %+v err=%v", cities, err)
	}

	// CityDetail (id STRING olmalı)
	detail, err := c.CityDetail(ctx, cities[0].ID)
	if err != nil || detail.ID != "9341" || detail.QiblaAngle != "150" {
		t.Fatalf("cityDetail yanlış: %+v err=%v", detail, err)
	}

	// Namaz vakitleri
	times, err := c.DailyPrayerTimes(ctx, cities[0].ID)
	if err != nil {
		t.Fatalf("prayer: %v", err)
	}
	if len(times) != 1 || times[0].Fajr != "04:00" || times[0].Isha != "21:40" {
		t.Fatalf("vakitler yanlış: %+v", times)
	}
}

// TestAuthPrefixFallback; /api/Auth/Login 404 dönerse istemci /Auth/Login'a düşmeli.
func TestAuthPrefixFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/Auth/Login", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r) // 404 → fallback tetiklensin
	})
	mux.HandleFunc("/Auth/Login", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"success":true,"data":{"accessToken":"OK","refreshToken":"R"}}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := New(&Config{BaseURL: srv.URL, Email: "a@b.c", Password: "x"})
	if err := c.EnsureAuth(context.Background()); err != nil {
		t.Fatalf("fallback login: %v", err)
	}
	if c.AccessToken() != "OK" {
		t.Fatalf("fallback sonrası token = %q, beklenen OK", c.AccessToken())
	}
}

// TestUnwrapFailure; success=false olduğunda hata fırlatılmalı.
func TestUnwrapFailure(t *testing.T) {
	_, err := unwrap[[]Place]([]byte(`{"success":false,"message":"yetkisiz","data":null}`))
	if err == nil {
		t.Fatal("success=false için hata bekleniyordu")
	}
}

// TestValidTokenSkipsLogin; geçerli bir token verildiğinde login İSTEĞİ ATILMAMALI.
func TestValidTokenSkipsLogin(t *testing.T) {
	loginCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/api/Auth/Login", func(w http.ResponseWriter, r *http.Request) {
		loginCalls++
		writeJSON(w, `{"success":true,"data":{"accessToken":"NEW","refreshToken":"R"}}`)
	})
	mux.HandleFunc("/api/Place/Countries", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer SEED" {
			t.Errorf("token yanlış: %q (SEED bekleniyordu)", got)
		}
		writeJSON(w, `{"success":true,"data":[]}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// AWQAT_ACCESS_TOKEN ile geçerli token verildi → login atlanmalı.
	c := New(&Config{BaseURL: srv.URL, Email: "a", Password: "b", AccessToken: "SEED", RefreshToken: "R"})
	if err := c.EnsureAuth(context.Background()); err != nil {
		t.Fatalf("ensureAuth: %v", err)
	}
	if _, err := c.Countries(context.Background()); err != nil {
		t.Fatalf("countries: %v", err)
	}
	if loginCalls != 0 {
		t.Fatalf("login %d kez atıldı, 0 bekleniyordu", loginCalls)
	}
	if c.TokenSource() != "env" {
		t.Fatalf("token kaynağı = %q, env bekleniyordu", c.TokenSource())
	}
}

// TestTokenCacheRoundTrip; kaydedilen token bir sonraki New()'de yüklenmeli.
func TestTokenCacheRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".awqat-token.json")
	saveTokenCache(path, tokenCache{AccessToken: "A", RefreshToken: "R", Expiry: time.Now().Add(time.Hour)})

	c := New(&Config{BaseURL: "http://x", Email: "a", Password: "b", TokenCachePath: path})
	if c.AccessToken() != "A" {
		t.Fatalf("önbellekten token yüklenmedi: %q", c.AccessToken())
	}
	if c.TokenSource() != "cache" {
		t.Fatalf("token kaynağı = %q, cache bekleniyordu", c.TokenSource())
	}
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}
