<?php

declare(strict_types=1);

/**
 * AwqatSalah Cookbook — PHP örneği.
 *
 * Çalıştırma:  cd php && php main.php            (tüm endpoint turu)
 *              cd php && php main.php --login     (sadece token al/yenile, kotayı korur)
 *
 * Resmi PDF'teki TÜM endpoint'leri sırayla çağırır (Isparta baz alınarak).
 * Tüm ayarlar kök .env / ortam değişkenlerinden okunur (bkz. ../.env.example).
 */

require __DIR__ . '/awqat/Config.php';
require __DIR__ . '/awqat/TokenCache.php';
require __DIR__ . '/awqat/Client.php';
require __DIR__ . '/awqat/Place.php';
require __DIR__ . '/awqat/Content.php';
require __DIR__ . '/awqat/Prayer.php';

use Awqat\Client;
use Awqat\Config;
use Awqat\Content;
use Awqat\Place;
use Awqat\Prayer;

function sec(int $n, string $title, string $path): void
{
    echo "\n[$n] $title\n     $path\n";
}

function warn(string $label, \Throwable $e): void
{
    echo "   ⚠ $label atlandı: {$e->getMessage()}\n";
}

function shortTok(string $s): string
{
    return mb_substr($s, 0, 16);
}

function truncateStr(?string $s, int $n): string
{
    $s = trim((string) $s);
    return mb_strlen($s) <= $n ? $s : mb_substr($s, 0, $n) . '…';
}

function firstDate(array $t): string
{
    return $t ? (string) $t[0]['gregorianDateShort'] : '-';
}

function lastDate(array $t): string
{
    return $t ? (string) $t[count($t) - 1]['gregorianDateShort'] : '-';
}

function pickCity(array $items, Config $cfg): array
{
    if ($cfg->cityId) {
        foreach ($items as $c) {
            if ((int) $c['id'] === $cfg->cityId) {
                return $c;
            }
        }
        return ['id' => $cfg->cityId, 'name' => '(AWQAT_CITY_ID)'];
    }
    $f = Place::findByName($items, $cfg->state);
    if ($f !== null) {
        return $f;
    }
    return $items[0] ?? ['id' => 0, 'name' => ''];
}

function printPrayer(array $city, array $t): void
{
    printf("   🕌 %s — %s (Hicri: %s)\n", $city['name'], $t['gregorianDateShort'] ?? '', $t['hijriDateShort'] ?? '');
    $rows = [
        ['İmsak  (Fajr)', $t['fajr'] ?? ''],
        ['Güneş  (Sunrise)', $t['sunrise'] ?? ''],
        ['Öğle   (Dhuhr)', $t['dhuhr'] ?? ''],
        ['İkindi (Asr)', $t['asr'] ?? ''],
        ['Akşam  (Maghrib)', $t['maghrib'] ?? ''],
        ['Yatsı  (Isha)', $t['isha'] ?? ''],
    ];
    foreach ($rows as [$label, $val]) {
        printf("      %-18s %s\n", $label, $val);
    }
}

function runDemo(bool $loginOnly): void
{
    $cfg = Config::load();
    $c = new Client($cfg);

    echo "====================================================\n";
    echo "  AwqatSalah Cookbook — PHP (tüm endpoint turu)\n";
    echo "  Base: {$cfg->baseUrl}\n";
    echo "  Not: her endpoint 1 kez çağrılır · kota: path başına 5/gün\n";
    echo "====================================================\n";

    // [1] Kimlik (token)
    sec(1, 'Kimlik (token)', '/api/Auth/Login');
    $c->ensureAuth();
    $tok = shortTok($c->accessToken());
    switch ($c->tokenSource()) {
        case 'cache':
            echo "   ✓ kayıtlı token kullanıldı, login ATILMADI (kota korundu): {$tok}…\n";
            break;
        case 'env':
            echo "   ✓ .env token'ı kullanıldı, login ATILMADI: {$tok}…\n";
            break;
        case 'refresh':
            echo "   ✓ token yenilendi (refresh): {$tok}…\n";
            break;
        default:
            echo "   ✓ login yapıldı, token .awqat-token.json'a kaydedildi: {$tok}…\n";
    }
    if ($loginOnly) {
        echo "\n(--login) Sadece kimlik adımı çalıştırıldı; endpoint turu atlandı.\n";
        return;
    }

    // [2] Günün Ayet/Hadis/Dua
    sec(2, 'Günlük İçerik (Ayet/Hadis/Dua)', '/api/DailyContent');
    try {
        $dc = Content::dailyContent($c);
        echo '   Ayet : ' . truncateStr($dc['verse'] ?? '', 90) . "\n";
        echo '   Hadis: ' . truncateStr($dc['hadith'] ?? '', 90) . "\n";
        echo '   Dua  : ' . truncateStr($dc['pray'] ?? '', 90) . "\n";
    } catch (\Throwable $e) {
        warn('dailyContent', $e);
    }

    // [3] Ülkeler
    sec(3, 'Ülkeler', '/api/Place/Countries');
    $country = ['id' => $cfg->countryId, 'name' => $cfg->country];
    try {
        $list = Place::countries($c);
        echo '   ' . count($list) . " ülke alındı.\n";
        $found = Place::findByName($list, $cfg->country);
        if ($found !== null) {
            $country = $found;
        }
    } catch (\Throwable $e) {
        warn('countries', $e);
    }
    echo "   → seçilen ülke: {$country['name']} (id={$country['id']})\n";
    if (!$country['id']) {
        throw new \RuntimeException("ülke çözümlenemedi (AWQAT_COUNTRY={$cfg->country})");
    }

    // [4] Tüm iller
    sec(4, 'Tüm iller (parametresiz)', '/api/Place/States');
    try {
        $all = Place::allStates($c);
        echo '   ' . count($all) . " il (tüm ülkeler) alındı.\n";
    } catch (\Throwable $e) {
        warn('allStates', $e);
    }

    // [5] Ülkeye göre iller
    sec(5, 'Ülkeye göre iller', '/api/Place/States/{countryId}');
    $state = ['id' => $cfg->stateId, 'name' => $cfg->state];
    try {
        $list = Place::states($c, (int) $country['id']);
        echo "   {$country['name']} için " . count($list) . " il alındı.\n";
        $found = Place::findByName($list, $cfg->state);
        if ($found !== null) {
            $state = $found;
        }
    } catch (\Throwable $e) {
        warn('states', $e);
    }
    echo "   → seçilen il: {$state['name']} (id={$state['id']})\n";
    if (!$state['id']) {
        throw new \RuntimeException("il çözümlenemedi (AWQAT_STATE={$cfg->state})");
    }

    // [6] Tüm ilçeler
    sec(6, 'Tüm ilçeler (parametresiz)', '/api/Place/Cities');
    try {
        $all = Place::allCities($c);
        echo '   ' . count($all) . " ilçe (tüm iller) alındı.\n";
    } catch (\Throwable $e) {
        warn('allCities', $e);
    }

    // [7] İle göre ilçeler
    sec(7, 'İle göre ilçeler', '/api/Place/Cities/{stateId}');
    $city = ['id' => $cfg->cityId, 'name' => '(AWQAT_CITY_ID)'];
    try {
        $list = Place::cities($c, (int) $state['id']);
        echo "   {$state['name']} için " . count($list) . " ilçe alındı.\n";
        $city = pickCity($list, $cfg);
    } catch (\Throwable $e) {
        warn('cities', $e);
    }
    echo "   → seçilen ilçe: {$city['name']} (id={$city['id']})\n";
    if (!$city['id']) {
        throw new \RuntimeException('ilçe çözümlenemedi');
    }

    // [8] İlçe detay (kıble)
    sec(8, 'İlçe detay (kıble açısı)', '/api/Place/CityDetail/{cityId}');
    try {
        $d = Place::cityDetail($c, (int) $city['id']);
        printf(
            "   %s / %s · kıble açısı: %s° · Kâbe'ye uzaklık: %s km\n",
            $d['city'] ?? '',
            $d['country'] ?? '',
            $d['qiblaAngle'] ?? '',
            $d['distanceToKaaba'] ?? ''
        );
    } catch (\Throwable $e) {
        warn('cityDetail', $e);
    }

    // [9] Günlük namaz vakitleri
    sec(9, 'Günlük namaz vakitleri', '/api/PrayerTime/Daily/{cityId}');
    try {
        $times = Prayer::daily($c, (int) $city['id']);
        if ($times) {
            printPrayer($city, $times[0]);
        }
    } catch (\Throwable $e) {
        warn('daily', $e);
    }

    // [10] Haftalık
    sec(10, 'Haftalık namaz vakitleri', '/api/PrayerTime/Weekly/{cityId}');
    try {
        $t = Prayer::weekly($c, (int) $city['id']);
        echo '   ' . count($t) . ' günlük veri (' . firstDate($t) . ' … ' . lastDate($t) . ")\n";
    } catch (\Throwable $e) {
        warn('weekly', $e);
    }

    // [11] Aylık
    sec(11, 'Aylık namaz vakitleri', '/api/PrayerTime/Monthly/{cityId}');
    try {
        $t = Prayer::monthly($c, (int) $city['id']);
        echo '   ' . count($t) . ' günlük veri (' . firstDate($t) . ' … ' . lastDate($t) . ")\n";
    } catch (\Throwable $e) {
        warn('monthly', $e);
    }

    // [12] Bayram namazı
    sec(12, 'Bayram namazı', '/api/PrayerTime/Eid/{cityId}');
    try {
        $ev = Prayer::eid($c, (int) $city['id']);
        printf(
            "   Ramazan B.: %s %s · Kurban B.: %s %s\n",
            $ev['eidAlFitrDate'] ?? '',
            $ev['eidAlFitrTime'] ?? '',
            $ev['eidAlAdhaDate'] ?? '',
            $ev['eidAlAdhaTime'] ?? ''
        );
    } catch (\Throwable $e) {
        warn('eid', $e);
    }

    // [13] Ramazan imsakiyesi
    sec(13, 'Ramazan imsakiyesi', '/api/PrayerTime/Ramadan/{cityId}');
    try {
        $t = Prayer::ramadan($c, (int) $city['id']);
        echo '   ' . count($t) . " günlük imsakiye verisi\n";
    } catch (\Throwable $e) {
        warn('ramadan', $e);
    }

    echo "\n====================================================\n";
    echo "✅ Tüm endpoint'ler çağrıldı.\n";
    echo "====================================================\n";
}

$loginOnly = in_array('--login', $argv, true) || in_array('-login', $argv, true);

try {
    runDemo($loginOnly);
} catch (\Throwable $e) {
    fwrite(STDERR, "\n❌ Hata: {$e->getMessage()}\n");
    exit(1);
}
