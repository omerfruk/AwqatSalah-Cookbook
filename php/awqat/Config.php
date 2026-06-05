<?php

declare(strict_types=1);

namespace Awqat;

/**
 * Yapılandırma: ortam değişkenleri + kök .env (yukarı doğru aranır).
 */
class Config
{
    public string $baseUrl;
    public string $email;
    public string $password;
    public string $country;
    public string $state;
    public int $countryId;
    public int $stateId;
    public int $cityId;
    public string $accessToken;
    public string $refreshToken;
    public string $tokenCachePath;

    // Token önbelleği dosya adı (kök .env yanında). Tüm diller AYNI dosyayı paylaşır.
    public const TOKEN_CACHE_FILE = '.awqat-token.json';

    public static function load(): Config
    {
        [$dotenv, $rootDir] = self::findDotEnv();

        // get: önce OS ortam değişkeni, sonra .env, sonra varsayılan.
        $get = static function (string $key, string $def = '') use ($dotenv): string {
            $env = getenv($key);
            if ($env !== false && trim($env) !== '') {
                return trim($env);
            }
            if (isset($dotenv[$key]) && trim($dotenv[$key]) !== '') {
                return trim($dotenv[$key]);
            }
            return $def;
        };

        $c = new self();
        $c->baseUrl = rtrim($get('AWQAT_BASE_URL', 'https://awqatsalah.diyanet.gov.tr'), '/');
        $c->email = $get('AWQAT_EMAIL');
        $c->password = $get('AWQAT_PASSWORD');
        $c->country = $get('AWQAT_COUNTRY', 'Türkiye');
        $c->state = $get('AWQAT_STATE', 'Isparta');
        $c->countryId = (int) $get('AWQAT_COUNTRY_ID');
        $c->stateId = (int) $get('AWQAT_STATE_ID');
        $c->cityId = (int) $get('AWQAT_CITY_ID');
        $c->accessToken = $get('AWQAT_ACCESS_TOKEN');
        $c->refreshToken = $get('AWQAT_REFRESH_TOKEN');
        $c->tokenCachePath = $rootDir . DIRECTORY_SEPARATOR . self::TOKEN_CACHE_FILE;

        if ($c->email === '' || $c->password === '') {
            throw new \RuntimeException(
                "AWQAT_EMAIL ve AWQAT_PASSWORD gerekli — kök dizinde .env oluşturup doldurun:\n" .
                "    cp .env.example .env   (sonra e-posta/şifrenizi yazın)"
            );
        }
        return $c;
    }

    /**
     * CWD'den köke kadar .env arar.
     * @return array{0: array<string,string>, 1: string} [değerler, kök_klasör]
     */
    private static function findDotEnv(): array
    {
        $dir = getcwd() ?: '.';
        $start = $dir;
        while (true) {
            $file = $dir . DIRECTORY_SEPARATOR . '.env';
            if (is_file($file)) {
                return [self::parseDotEnv((string) file_get_contents($file)), $dir];
            }
            $parent = \dirname($dir);
            if ($parent === $dir) {
                return [[], $start]; // .env yok → CWD kök
            }
            $dir = $parent;
        }
    }

    /**
     * Basit KEY=VALUE ayrıştırıcı (# yorum, boş satır, tırnak soyma).
     * @return array<string,string>
     */
    private static function parseDotEnv(string $text): array
    {
        $out = [];
        foreach (preg_split('/\r\n|\r|\n/', $text) ?: [] as $line) {
            $line = trim($line);
            if ($line === '' || $line[0] === '#') {
                continue;
            }
            if (str_starts_with($line, 'export ')) {
                $line = substr($line, 7);
            }
            $eq = strpos($line, '=');
            if ($eq === false) {
                continue;
            }
            $key = trim(substr($line, 0, $eq));
            $val = trim(substr($line, $eq + 1));
            $len = strlen($val);
            if ($len >= 2) {
                $first = $val[0];
                $last = $val[$len - 1];
                if (($first === '"' && $last === '"') || ($first === "'" && $last === "'")) {
                    $val = substr($val, 1, -1);
                }
            }
            if ($key !== '') {
                $out[$key] = $val;
            }
        }
        return $out;
    }
}
