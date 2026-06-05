<?php

declare(strict_types=1);

/**
 * Sahte (enjekte) gönderici ile testler — gerçek API'ye gerek yok, PHPUnit gerekmez.
 * Çalıştırma:  php tests/ClientTest.php
 */

require __DIR__ . '/../awqat/Config.php';
require __DIR__ . '/../awqat/TokenCache.php';
require __DIR__ . '/../awqat/Client.php';
require __DIR__ . '/../awqat/Place.php';
require __DIR__ . '/../awqat/Content.php';
require __DIR__ . '/../awqat/Prayer.php';

use Awqat\Client;
use Awqat\Config;
use Awqat\Content;
use Awqat\Place;
use Awqat\Prayer;

$total = 0;
$failed = 0;

function check(string $name, bool $cond): void
{
    global $total, $failed;
    $total++;
    if ($cond) {
        echo "  ok   $name\n";
    } else {
        echo "  FAIL $name\n";
        $failed++;
    }
}

function makeCfg(array $over = []): Config
{
    $c = new Config();
    $c->baseUrl = 'http://mock';
    $c->email = 'a';
    $c->password = 'b';
    $c->country = 'Türkiye';
    $c->state = 'Isparta';
    $c->countryId = 0;
    $c->stateId = 0;
    $c->cityId = 0;
    $c->accessToken = '';
    $c->refreshToken = '';
    $c->tokenCachePath = '';
    foreach ($over as $k => $v) {
        $c->$k = $v;
    }
    return $c;
}

/**
 * Sahte gönderici: path -> [status, body]. İstek sayılarını ve son token'ı kaydeder.
 * @param array<string,array{0:int,1:string}> $routes
 */
function makeSender(array $routes, array &$calls, string &$lastToken): callable
{
    return function (string $method, string $path, ?string $payload, string $token) use ($routes, &$calls, &$lastToken): array {
        $calls[$path] = ($calls[$path] ?? 0) + 1;
        $lastToken = $token;
        if (isset($routes[$path])) {
            return ['status' => $routes[$path][0], 'text' => $routes[$path][1]];
        }
        return ['status' => 404, 'text' => 'not found'];
    };
}

// --- Test 1: uçtan uca ----------------------------------------------------
$calls = [];
$lastToken = '';
$routes = [
    '/api/Auth/Login' => [200, '{"success":true,"data":{"accessToken":"ACCESS","refreshToken":"R"}}'],
    '/api/DailyContent' => [200, '{"success":true,"data":{"verse":"V","hadith":"H","pray":"P"}}'],
    '/api/Place/Countries' => [200, '{"success":true,"data":[{"id":1,"name":"KUZEY KIBRIS"},{"id":2,"name":"TÜRKİYE"}]}'],
    '/api/Place/States/2' => [200, '{"success":true,"data":[{"id":538,"name":"ISPARTA"}]}'],
    '/api/Place/Cities/538' => [200, '{"success":true,"data":[{"id":9528,"name":"ISPARTA"}]}'],
    '/api/Place/CityDetail/9528' => [200, '{"success":true,"data":{"id":"9528","qiblaAngle":"151","distanceToKaaba":"2023","city":"ISPARTA","country":"TÜRKİYE"}}'],
    '/api/PrayerTime/Daily/9528' => [200, '{"success":true,"data":[{"fajr":"03:44","isha":"22:01"}]}'],
];
$c = new Client(makeCfg(), makeSender($routes, $calls, $lastToken));
$c->ensureAuth();
check('login token', $c->accessToken() === 'ACCESS');

$dc = Content::dailyContent($c);
check('dailyContent', $dc['verse'] === 'V');

$country = Place::findByName(Place::countries($c), 'Türkiye'); // "TÜRKİYE" ile eşleşmeli
check('ülke (Türkçe duyarsız)', $country !== null && $country['id'] === 2);
check('Bearer header', $lastToken === 'ACCESS');

$state = Place::findByName(Place::states($c, $country['id']), 'Isparta');
check('il', $state !== null && $state['id'] === 538);

$cities = Place::cities($c, $state['id']);
check('ilçe', $cities[0]['id'] === 9528);

$d = Place::cityDetail($c, $cities[0]['id']);
check('cityDetail id STRING', $d['id'] === '9528');
check('cityDetail qibla', $d['qiblaAngle'] === '151');

$times = Prayer::daily($c, $cities[0]['id']);
check('namaz vakti', $times[0]['fajr'] === '03:44');

// --- Test 2: geçerli token → login atılmaz --------------------------------
$calls = [];
$lastToken = '';
$routes2 = [
    '/api/Auth/Login' => [200, '{"success":true,"data":{"accessToken":"NEW"}}'],
    '/api/Place/Countries' => [200, '{"success":true,"data":[]}'],
];
$c2 = new Client(makeCfg(['accessToken' => 'SEED']), makeSender($routes2, $calls, $lastToken));
$c2->ensureAuth();
Place::countries($c2);
check('login İSTEĞİ atılmadı', ($calls['/api/Auth/Login'] ?? 0) === 0);
check('source=env', $c2->tokenSource() === 'env');
check('seed token kullanıldı', $lastToken === 'SEED');

// --- Test 3: /api/Auth/Login 404 → /Auth/Login fallback -------------------
$calls = [];
$lastToken = '';
$routes3 = [
    '/api/Auth/Login' => [404, 'nf'],
    '/Auth/Login' => [200, '{"success":true,"data":{"accessToken":"OK"}}'],
];
$c3 = new Client(makeCfg(), makeSender($routes3, $calls, $lastToken));
$c3->ensureAuth();
check('auth prefix fallback', $c3->accessToken() === 'OK');

// --- Test 4: unwrap success=false → hata ----------------------------------
try {
    Client::unwrap('{"success":false,"message":"yetkisiz"}');
    check('unwrap success=false hata', false);
} catch (\Throwable $e) {
    check('unwrap success=false hata', true);
}

// --- Test 5: fold Türkçe-duyarsız -----------------------------------------
check('fold TÜRKİYE', Place::fold('TÜRKİYE') === 'turkiye');
check('fold Isparta', Place::fold('Isparta') === 'isparta');

echo "\n$total test, $failed başarısız\n";
exit($failed === 0 ? 0 : 1);
