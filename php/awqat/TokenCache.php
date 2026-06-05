<?php

declare(strict_types=1);

namespace Awqat;

/**
 * Token önbelleği: login/refresh sonrası token'ı diske kalıcı yazar.
 * Sonraki çalıştırmalarda geçerliyse yeniden login ATILMAZ (kota korunur).
 * Biçim Go/JS/Python ile UYUMLU: {"accessToken","refreshToken","expiry" (ISO 8601)}.
 */
class TokenCache
{
    /** @return array{accessToken:string,refreshToken?:string,expiry?:string}|null */
    public static function load(string $path): ?array
    {
        if ($path === '' || !is_file($path)) {
            return null;
        }
        $data = json_decode((string) file_get_contents($path), true);
        if (!is_array($data) || empty($data['accessToken'])) {
            return null;
        }
        return $data;
    }

    /** Token'ı yalnızca sahibin okuyabileceği (0600) dosyaya yazar. Hata yutulur. */
    public static function save(string $path, array $tc): void
    {
        if ($path === '') {
            return;
        }
        $json = json_encode($tc, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE);
        if ($json === false) {
            return;
        }
        if (@file_put_contents($path, $json) !== false) {
            @chmod($path, 0600);
        }
    }
}
