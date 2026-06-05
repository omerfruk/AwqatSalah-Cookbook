# AI.md — PHP implementasyonu (AI promptu)

> Bu dosya bir AI'a **doğrudan verilebilir**. Amacı: AwqatSalah istemcisini **PHP** ile
> sıfırdan üretmek (veya bu örneği genişletmek). Tek doğruluk kaynağı: [`../SPEC.md`](../SPEC.md).
> Referans desen: kardeş [`../go`](../go), [`../js`](../js), [`../python`](../python) implementasyonları.

---

## Görev
AwqatSalah (Diyanet Namaz Vakti) API'si için PHP'de, **sıfır Composer bağımlılığıyla** (yerleşik curl),
tek komutla çalışan bir örnek üret. `php main.php` çalışınca **SPEC.md §4'teki tüm endpoint'leri**
sırayla birer kez çağırmalı ve okunaklı yazdırmalı. `php main.php --login` ise sadece kimlik adımını
çalıştırmalı (kotayı korumak için).

## Ön koşul: ÖNCE SPEC.md'yi oku
`../SPEC.md` API'nin tamamını içerir: base URL, auth, **tüm endpoint'ler**, DTO alanları, ortak zarf,
kota, token yeniden kullanımı (§2.5) ve config konvansiyonu (§6). **Oradan sapma.**

## Üretilecek dosyalar (namespace `Awqat`)
```
php/
├── main.php               # tüm endpoint turu + --login bayrağı, okunaklı çıktı
├── awqat/
│   ├── Config.php         # Config::load(): env + kök .env (yukarı doğru arama), AWQAT_* sözlüğü
│   ├── TokenCache.php     # .awqat-token.json oku/yaz (login kotası koruması, 0600)
│   ├── Client.php         # Client: ensureAuth, login(/api/Auth), refresh, getJson/doGet/doPost, 401 retry, kota, curl ağ-retry; unwrap; enjekte edilebilir $sender
│   ├── Place.php          # countries, allStates, states(id), allCities, cities(id), cityDetail; findByName/fold (Türkçe-duyarsız)
│   ├── Content.php        # dailyContent
│   └── Prayer.php         # daily/weekly/monthly/ramadan/eid
└── tests/ClientTest.php   # enjekte gönderici ile (uçtan uca + auth fallback + token-reuse), PHPUnit YOK
```
> Endpoint metotları istemciyi ilk argüman alan statik metotlardır: `Prayer::daily($client, $cityId)`.

## Kritik kurallar (SPEC.md §8 ile aynı)
1. **Tüm yollar `/api/...`** — Auth dâhil: `POST /api/Auth/Login`, `GET /api/Auth/RefreshToken/{rt}`.
   (Sağlamlık: `/api/Auth/Login` 404 dönerse `/Auth/Login`'a düş.)
2. Auth dışı her istekte `Authorization: Bearer <accessToken>`.
3. Yanıt zarfını aç (`Client::unwrap`): `success` kontrol et, değilse `message` ile hata fırlat, yoksa `data` dön.
4. **401** → token'ı sıfırla, yeniden kimlik doğrula, isteği **1 kez** tekrar et.
5. Token süresini izle (~30 dk), dolmadan refresh; refresh başarısızsa login.
6. Config: **getenv → kök `.env`** (yukarı doğru aranır). Yeni env adı uydurma; `AWQAT_*` kullan.
7. Kota: aynı path'e günde 5'ten fazla isteği engelleyen süreç-içi sayaç.
8. İsim eşleştirme Türkçe + büyük/küçük harf **duyarsız** (`strtr` fold + `strtolower`).
9. **`cityDetail['id']` STRING**, `place['id']` NUMBER — karıştırma.
10. **Token'ı yeniden kullan:** login/refresh sonrası `{accessToken,refreshToken,expiry}`'yi
    `.awqat-token.json`'a (ISO tarih `gmdate('c')`, 0600, gitignore'lu) yaz; başlangıçta yükle; geçerliyse login ATMA (§2.5).
    `AWQAT_ACCESS_TOKEN` env override'ını destekle. `--login` bayrağı sun.

## PHP'ye özel notlar
- `declare(strict_types=1);` kullan. Namespace `Awqat`, dosyalar manuel `require` ile yüklenir (Composer yok).
- HTTP: yerleşik `curl`. `curl_exec` false dönerse ağ hatası → birkaç kez tekrar dene. PHP 8.0+'da `curl_close()` GEREKMEZ (8.5'te deprecated) — çağırma.
- Çift tırnaklı string'de değişkenden sonra çok-baytlı karakter gelirse `{$var}…` ile sınırla
  (yoksa `$var…` tek değişken adı sanılır).
- Test edilebilirlik: `Client` kurucusu opsiyonel `?callable $sender` alsın; testler curl yerine sahte gönderici enjekte etsin.
- Token önbelleği biçimi Go/JS/Python ile **uyumlu** (`expiry` ISO 8601) — diller aynı dosyayı paylaşır.

## Bitti kabul kriterleri
- [ ] `php -l` tüm dosyalarda temiz; `php tests/ClientTest.php` 0 başarısız (uyarı/deprecated yok).
- [ ] `.env` yokken nazik bir hata (creds gerekli mesajı).
- [ ] `.env` doluyken SPEC.md §4'teki tüm endpoint'leri çağırıp sonuçları yazdırıyor.
- [ ] İkinci çalıştırmada cache'ten token kullanıp login ATMIYOR.
- [ ] Testler uçtan uca akışı + fallback + token-reuse'u doğruluyor (gerçek API'siz).

---

## Üst seviye reçete
Mimari ve "yeni dil ekleme" adımları: [`../CLAUDE.md`](../CLAUDE.md).
