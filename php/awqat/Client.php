<?php

declare(strict_types=1);

namespace Awqat;

/**
 * AwqatSalah API istemcisi — sıfır bağımlılık (yerleşik curl).
 *
 * Kimlik doğrulama akışı (resmi PDF'e göre):
 *  1. POST /api/Auth/Login (email+password) -> accessToken + refreshToken
 *  2. Her istekte Authorization: Bearer <accessToken>
 *  3. Süre dolmadan GET /api/Auth/RefreshToken/{refreshToken} ile yenilenir
 *  4. Refresh de başarısızsa tekrar login olunur
 */
class Client
{
    private const TOKEN_LIFETIME = 30 * 60; // access token ömrü (PDF §3: 30 dk)
    private const REFRESH_SAFETY = 5 * 60;  // bitmeden bu kadar önce yenile
    private const MAX_PER_PATH_PER_DAY = 5; // Standart rol kotası: path başına / gün
    private const TIMEOUT = 30;

    public string $baseUrl;
    private Config $cfg;
    private string $accessToken = '';
    private string $refreshToken = '';
    private int $expiry = 0; // epoch saniye
    private string $authPrefix = '/api'; // 404 alınırsa '' (eski örnekler) ile denenir
    private string $source = '';         // 'env' | 'cache' | 'login' | 'refresh' | ''
    /** @var array<string,array{n:int,date:string}> */
    private array $rate = [];
    /** @var callable|null Test için enjekte edilebilir gönderici. */
    private $sender;

    public function __construct(Config $cfg, ?callable $sender = null)
    {
        $this->cfg = $cfg;
        $this->baseUrl = $cfg->baseUrl;
        $this->sender = $sender;

        // Mevcut token varsa yükle → geçerliyse login atılmaz (kota korunur).
        if ($cfg->accessToken !== '') {
            $this->accessToken = $cfg->accessToken;
            $this->refreshToken = $cfg->refreshToken;
            $this->expiry = time() + (self::TOKEN_LIFETIME - self::REFRESH_SAFETY);
            $this->source = 'env';
        } else {
            $tc = TokenCache::load($cfg->tokenCachePath);
            if ($tc !== null) {
                $this->accessToken = $tc['accessToken'];
                $this->refreshToken = $tc['refreshToken'] ?? '';
                $this->expiry = isset($tc['expiry']) ? (strtotime($tc['expiry']) ?: 0) : 0;
                $this->source = 'cache';
            }
        }
    }

    public function accessToken(): string
    {
        return $this->accessToken;
    }

    public function tokenSource(): string
    {
        return $this->source;
    }

    // ---- Kimlik doğrulama --------------------------------------------------

    public function ensureAuth(): void
    {
        if ($this->accessToken !== '' && time() < $this->expiry) {
            return; // token hâlâ geçerli
        }
        if ($this->refreshToken !== '') {
            try {
                $this->refresh();
                return;
            } catch (\Throwable $e) {
                // refresh başarısız → login'e düş
            }
        }
        $this->login();
    }

    private function login(): void
    {
        $payload = (string) json_encode(['email' => $this->cfg->email, 'password' => $this->cfg->password]);
        ['status' => $status, 'text' => $text] = $this->send('POST', $this->authPrefix . '/Auth/Login', $payload, '');
        if ($status === 404 && $this->authPrefix !== '') {
            $this->authPrefix = ''; // eski yola geç ve hatırla
            ['status' => $status, 'text' => $text] = $this->send('POST', '/Auth/Login', $payload, '');
        }
        if ($status !== 200) {
            throw new \RuntimeException("login başarısız (HTTP $status): $text");
        }
        $this->storeToken($text, 'login');
    }

    private function refresh(): void
    {
        if ($this->refreshToken === '') {
            throw new \RuntimeException('refresh token yok');
        }
        $path = $this->authPrefix . '/Auth/RefreshToken/' . $this->refreshToken;
        ['status' => $status, 'text' => $text] = $this->send('GET', $path, null, '');
        if ($status !== 200) {
            throw new \RuntimeException("refresh başarısız (HTTP $status)");
        }
        $this->storeToken($text, 'refresh');
    }

    private function storeToken(string $text, string $op): void
    {
        $data = self::unwrap($text);
        $this->accessToken = $data['accessToken'];
        $this->refreshToken = $data['refreshToken'] ?? '';
        $this->expiry = time() + (self::TOKEN_LIFETIME - self::REFRESH_SAFETY);
        $this->source = $op; // 'login' veya 'refresh'
        // Token'ı kalıcı yap → sonraki çalıştırmalar geçerliyse login atmaz.
        TokenCache::save($this->cfg->tokenCachePath, [
            'accessToken' => $this->accessToken,
            'refreshToken' => $this->refreshToken,
            'expiry' => gmdate('c', $this->expiry),
        ]);
    }

    // ---- İstek katmanı (kota + auth + 401 retry) ---------------------------

    /** Kimlik doğrulamalı GET yapıp zarfı açarak data döndürür. */
    public function getJson(string $path)
    {
        return self::unwrap($this->doGet($path));
    }

    public function doGet(string $path): string
    {
        return $this->doRequest('GET', $path, null);
    }

    public function doPost(string $path, $body): string
    {
        return $this->doRequest('POST', $path, (string) json_encode($body));
    }

    private function doRequest(string $method, string $path, ?string $payload): string
    {
        $this->checkRate($path);
        $this->ensureAuth();

        ['status' => $status, 'text' => $text] = $this->send($method, $path, $payload, $this->accessToken);

        // 401 → token geçersiz olmuş olabilir: sıfırla, yeniden doğrula, 1 kez tekrar dene.
        if ($status === 401) {
            $this->accessToken = '';
            $this->refreshToken = '';
            $this->ensureAuth();
            ['status' => $status, 'text' => $text] = $this->send($method, $path, $payload, $this->accessToken);
        }

        if ($status !== 200) {
            throw new \RuntimeException("$method $path başarısız (HTTP $status): $text");
        }
        return $text;
    }

    /** @return array{status:int,text:string} */
    private function send(string $method, string $path, ?string $payload, string $token): array
    {
        if ($this->sender !== null) {
            return ($this->sender)($method, $path, $payload, $token);
        }
        return $this->curlSend($method, $path, $payload, $token);
    }

    /** @return array{status:int,text:string} */
    private function curlSend(string $method, string $path, ?string $payload, string $token): array
    {
        $headers = ['Content-Type: application/json'];
        if ($token !== '') {
            $headers[] = 'Authorization: Bearer ' . $token;
        }

        // Ağ hatasında (reset / ölü bağlantı) kısa beklemeyle 3 kez dene.
        $lastErr = '';
        for ($attempt = 0; $attempt < 3; $attempt++) {
            $ch = curl_init($this->baseUrl . $path);
            curl_setopt_array($ch, [
                CURLOPT_CUSTOMREQUEST => $method,
                CURLOPT_RETURNTRANSFER => true,
                CURLOPT_HTTPHEADER => $headers,
                CURLOPT_TIMEOUT => self::TIMEOUT,
            ]);
            if ($payload !== null) {
                curl_setopt($ch, CURLOPT_POSTFIELDS, $payload);
            }
            $body = curl_exec($ch);
            if ($body === false) {
                $lastErr = curl_error($ch);
                usleep(200000 * ($attempt + 1));
                continue;
            }
            $status = (int) curl_getinfo($ch, CURLINFO_RESPONSE_CODE);
            return ['status' => $status, 'text' => (string) $body];
        }
        throw new \RuntimeException("$method $path: $lastErr");
    }

    /** Path için günlük kotayı (5/gün) kontrol eder. Sayaç süreç-içidir. */
    private function checkRate(string $path): void
    {
        $today = date('Y-m-d');
        $cnt = $this->rate[$path] ?? null;
        if ($cnt === null || $cnt['date'] !== $today) {
            $this->rate[$path] = ['n' => 1, 'date' => $today];
            return;
        }
        if ($cnt['n'] >= self::MAX_PER_PATH_PER_DAY) {
            throw new \RuntimeException(
                "kota: $path için bugün {$cnt['n']}/" . self::MAX_PER_PATH_PER_DAY . " istek kullanıldı"
            );
        }
        $this->rate[$path]['n']++;
    }

    /** Ortak zarfı açar, success kontrol eder, data döndürür. */
    public static function unwrap(string $text)
    {
        $resp = json_decode($text, true);
        if (!is_array($resp)) {
            throw new \RuntimeException('yanıt çözümlenemedi');
        }
        if (empty($resp['success'])) {
            $msg = $resp['message'] ?? 'bilinmeyen hata';
            throw new \RuntimeException("API başarısız: $msg");
        }
        return $resp['data'] ?? null;
    }
}
