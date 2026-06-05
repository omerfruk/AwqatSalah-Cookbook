// Package awqat; Diyanet İşleri Başkanlığı AwqatSalah (Namaz Vakti) API'si için
// sade, bağımlılıksız bir Go istemcisidir.
//
// Kimlik doğrulama akışı (resmi PDF'e göre):
//  1. POST /api/Auth/Login (email+password) → accessToken + refreshToken
//  2. Her istekte Authorization: Bearer <accessToken>
//  3. Süre dolmadan GET /api/Auth/RefreshToken/{refreshToken} ile yenilenir
//  4. Refresh de başarısızsa tekrar login olunur
package awqat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	// Resmi PDF §3: access token 30 dk, refresh penceresi +15 dk.
	tokenLifetime    = 30 * time.Minute
	refreshSafety    = 5 * time.Minute // bitmeden bu kadar önce yenile (güvenlik marjı)
	maxPerPathPerDay = 5               // Standart rol kotası: path başına / gün (Developer rol: 100)
)

// Client; AwqatSalah API'si için yeniden kullanılabilir HTTP istemcisidir.
type Client struct {
	cfg  *Config
	http *http.Client

	mu           sync.Mutex
	accessToken  string
	refreshToken string
	expiry       time.Time
	authPrefix   string // "/api" (resmi PDF) — 404 alınırsa "" (eski örnekler) ile otomatik denenir
	source       string // token kaynağı: "env" | "cache" | "login" | "refresh" | "" (yok)

	rate map[string]*counter // path başına günlük istek sayacı (kota koruması)
}

type counter struct {
	n    int
	date string // YYYY-MM-DD
}

// New; verilen konfigürasyondan yeni bir istemci oluşturur.
// Mevcut bir token varsa (env override veya önbellek dosyası) onu yükler;
// böylece geçerliyse yeniden login atılmaz (kota korunur).
func New(cfg *Config) *Client {
	c := &Client{
		cfg:        cfg,
		http:       &http.Client{Timeout: 30 * time.Second},
		authPrefix: "/api",
		rate:       map[string]*counter{},
	}

	if cfg.AccessToken != "" {
		// Elle verilen token (AWQAT_ACCESS_TOKEN). Süresi bilinmiyor → iyimser kabul;
		// gerçekten geçersizse ilk istekte 401 alınır ve otomatik yeniden login olunur.
		c.accessToken = cfg.AccessToken
		c.refreshToken = cfg.RefreshToken
		c.expiry = time.Now().Add(tokenLifetime - refreshSafety)
		c.source = "env"
	} else if tc, ok := loadTokenCache(cfg.TokenCachePath); ok {
		c.accessToken = tc.AccessToken
		c.refreshToken = tc.RefreshToken
		c.expiry = tc.Expiry
		c.source = "cache"
	}
	return c
}

// TokenSource; mevcut token'ın nereden geldiğini döndürür
// ("env" | "cache" | "login" | "refresh" | "").
func (c *Client) TokenSource() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.source
}

// ---------------------------------------------------------------------------
// Kimlik doğrulama
// ---------------------------------------------------------------------------

// EnsureAuth; geçerli bir access token olduğundan emin olur (gerekirse login/refresh).
func (c *Client) EnsureAuth(ctx context.Context) error { return c.ensureAuth(ctx) }

// AccessToken; mevcut access token'ı döndürür (demo/inceleme amaçlı).
func (c *Client) AccessToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.accessToken
}

func (c *Client) ensureAuth(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.expiry) {
		return nil // token hâlâ geçerli
	}
	if c.refreshToken != "" {
		if err := c.refresh(ctx); err == nil {
			return nil // refresh ile yenilendi
		}
		// refresh başarısız → login'e düş
	}
	return c.login(ctx)
}

// login; POST {prefix}/Auth/Login. PDF /api/Auth/Login der; 404'te /Auth/Login'a düşer.
// mu kilitliyken çağrılır.
func (c *Client) login(ctx context.Context) error {
	payload, _ := json.Marshal(map[string]string{"email": c.cfg.Email, "password": c.cfg.Password})

	body, status, err := c.send(ctx, http.MethodPost, c.authPrefix+"/Auth/Login", payload, "")
	if err == nil && status == http.StatusNotFound && c.authPrefix != "" {
		c.authPrefix = "" // eski yola geç ve hatırla
		body, status, err = c.send(ctx, http.MethodPost, "/Auth/Login", payload, "")
	}
	if err != nil {
		return fmt.Errorf("login isteği: %w", err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("login başarısız (HTTP %d): %s", status, string(body))
	}
	return c.storeToken(body, "login")
}

// refresh; GET {prefix}/Auth/RefreshToken/{refreshToken}. mu kilitliyken çağrılır.
func (c *Client) refresh(ctx context.Context) error {
	if c.refreshToken == "" {
		return fmt.Errorf("refresh token yok")
	}
	body, status, err := c.send(ctx, http.MethodGet, c.authPrefix+"/Auth/RefreshToken/"+c.refreshToken, nil, "")
	if err != nil {
		return fmt.Errorf("refresh isteği: %w", err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("refresh başarısız (HTTP %d)", status)
	}
	return c.storeToken(body, "refresh")
}

func (c *Client) storeToken(body []byte, op string) error {
	tok, err := unwrap[Token](body)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	c.accessToken = tok.AccessToken
	c.refreshToken = tok.RefreshToken
	c.expiry = time.Now().Add(tokenLifetime - refreshSafety)
	c.source = op // "login" veya "refresh"
	// Token'ı kalıcı yap → sonraki çalıştırmalar geçerliyse login atmaz.
	saveTokenCache(c.cfg.TokenCachePath, tokenCache{
		AccessToken:  c.accessToken,
		RefreshToken: c.refreshToken,
		Expiry:       c.expiry,
	})
	return nil
}

// ---------------------------------------------------------------------------
// İstek katmanı (kota + auth + 401 retry)
// ---------------------------------------------------------------------------

// DoGet; kimlik doğrulamalı GET isteği yapar, ham gövdeyi döndürür.
func (c *Client) DoGet(ctx context.Context, path string) ([]byte, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

// DoPost; kimlik doğrulamalı POST isteği yapar.
func (c *Client) DoPost(ctx context.Context, path string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, http.MethodPost, path, raw)
}

func (c *Client) do(ctx context.Context, method, path string, payload []byte) ([]byte, error) {
	if err := c.checkRate(path); err != nil {
		return nil, err
	}
	if err := c.ensureAuth(ctx); err != nil {
		return nil, fmt.Errorf("kimlik doğrulama: %w", err)
	}

	body, status, err := c.authedSend(ctx, method, path, payload)
	if err != nil {
		return nil, err
	}

	// 401 → token geçersiz olmuş olabilir: sıfırla, yeniden doğrula, 1 kez tekrar dene.
	if status == http.StatusUnauthorized {
		c.mu.Lock()
		c.accessToken, c.refreshToken = "", ""
		c.mu.Unlock()
		if err := c.ensureAuth(ctx); err != nil {
			return nil, fmt.Errorf("401 sonrası yeniden doğrulama: %w", err)
		}
		body, status, err = c.authedSend(ctx, method, path, payload)
		if err != nil {
			return nil, err
		}
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("%s %s başarısız (HTTP %d): %s", method, path, status, string(body))
	}
	return body, nil
}

// authedSend; geçerli access token ile istek gönderir.
func (c *Client) authedSend(ctx context.Context, method, path string, payload []byte) ([]byte, int, error) {
	c.mu.Lock()
	token := c.accessToken
	c.mu.Unlock()
	return c.send(ctx, method, path, payload, token)
}

// send; düşük seviye HTTP gönderimi. token boşsa Authorization eklenmez (auth istekleri).
func (c *Client) send(ctx context.Context, method, path string, payload []byte, token string) ([]byte, int, error) {
	var rdr io.Reader
	if payload != nil {
		rdr = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, rdr)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return data, resp.StatusCode, nil
}

// checkRate; path için günlük kotayı (5/gün) kontrol eder ve sayacı artırır.
// Not: sayaç süreç-içidir; her `go run` taze başlar (gerçek kota sunucu tarafındadır).
func (c *Client) checkRate(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	cnt := c.rate[path]
	if cnt == nil || cnt.date != today {
		c.rate[path] = &counter{n: 1, date: today}
		return nil
	}
	if cnt.n >= maxPerPathPerDay {
		return fmt.Errorf("kota: %s için bugün %d/%d istek kullanıldı", path, cnt.n, maxPerPathPerDay)
	}
	cnt.n++
	return nil
}

// ---------------------------------------------------------------------------
// Generic yardımcılar
// ---------------------------------------------------------------------------

// unwrap; ortak zarfı açar, success kontrol eder, data'yı döndürür.
func unwrap[T any](body []byte) (T, error) {
	var zero T
	var resp APIResponse[T]
	if err := json.Unmarshal(body, &resp); err != nil {
		return zero, fmt.Errorf("yanıt çözümlenemedi: %w", err)
	}
	if !resp.Success {
		msg := "bilinmeyen hata"
		if resp.Message != nil {
			msg = *resp.Message
		}
		return zero, fmt.Errorf("API başarısız: %s", msg)
	}
	return resp.Data, nil
}

// getJSON; GET yapıp zarfı açarak tipli sonuç döndürür.
func getJSON[T any](ctx context.Context, c *Client, path string) (T, error) {
	var zero T
	body, err := c.DoGet(ctx, path)
	if err != nil {
		return zero, err
	}
	return unwrap[T](body)
}
