# AwqatSalah Cookbook — PHP

Diyanet AwqatSalah (Namaz Vakti) API'sini **sıfır bağımlılıkla** (yerleşik `curl`, Composer yok)
kullanan, çalışır PHP örneği. `php main.php` ile **resmi dokümandaki tüm endpoint'leri** sırayla çağırır.

> 🤖 Bir AI'a bu örneği oluşturtmak/genişletmek mi istiyorsunuz? → [`AI.md`](./AI.md)
> API'nin tam sözleşmesi → [`../SPEC.md`](../SPEC.md)

---

## Gereksinimler
- PHP 8.1+ (geliştirilen sürüm: 8.5) — `curl` ve `mbstring` uzantıları (genelde varsayılan).
- Bir AwqatSalah hesabı (<https://awqatsalah.diyanet.gov.tr>)
- **Composer / paket yok.**

## Kurulum
Kimlik bilgileri **repo kökündeki tek `.env`** dosyasından okunur (tüm diller aynı dosyayı paylaşır):

```bash
# repo kökünde
cp .env.example .env
# .env içine AWQAT_EMAIL ve AWQAT_PASSWORD yazın
```

## Çalıştırma
```bash
cd php
php main.php            # tüm endpoint turu
php main.php --login    # SADECE token al/yenile (endpoint turunu atla, kotayı korur)
php tests/ClientTest.php   # testler (sahte gönderici, gerçek API'siz)
```

### Örnek çıktı (gerçek API)
```
[1] Kimlik (token)        /api/Auth/Login
   ✓ kayıtlı token kullanıldı, login ATILMADI (kota korundu): eyJhbGciOiJIUzI1…
[2] Günlük İçerik         /api/DailyContent          → Ayet / Hadis / Dua
[3] Ülkeler               /api/Place/Countries        → 208 ülke · TÜRKİYE (id=2)
[5] Ülkeye göre iller     /api/Place/States/{id}      → 81 il · ISPARTA (id=538)
[7] İle göre ilçeler      /api/Place/Cities/{id}      → 13 ilçe · ISPARTA (id=9528)
[8] İlçe detay            /api/Place/CityDetail/{id}  → kıble 151° · Kâbe 2023 km
[9] Günlük vakitler       /api/PrayerTime/Daily/{id}
   🕌 ISPARTA — 05.06.2026
      İmsak 03:44 · Güneş 05:29 · Öğle 13:01 · İkindi 16:54 · Akşam 20:23 · Yatsı 22:01
[10-13] Haftalık/Aylık/Bayram/Ramazan …
✅ Tüm endpoint'ler çağrıldı.
```

## Token yeniden kullanımı (login kotası koruması)
Login günde **5 istekle** sınırlıdır. Program login sonrası token'ı otomatik olarak kök dizindeki
gizli **`.awqat-token.json`** (`0600`, `.gitignore`'lu) dosyasına kaydeder ve sonraki çalıştırmalarda
**geçerliyse yeniden login ATMAZ**. Bu dosya **tüm dillerle ortaktır** — Go/JS/Python ile login olunca
PHP da aynı token'ı kullanır. İstersen `.env`'e elle `AWQAT_ACCESS_TOKEN` da yazabilirsin.

## Dosya düzeni
```
php/
├── main.php               # Demo: tüm endpoint turu (--login ile sadece kimlik)
├── awqat/
│   ├── Config.php         # .env + ortam değişkeni yükleme (yukarı doğru .env arar)
│   ├── TokenCache.php     # Token önbelleği (.awqat-token.json) oku/yaz
│   ├── Client.php         # AwqatClient: login(/api/Auth), refresh, Bearer, 401 retry, kota, ağ-retry
│   ├── Place.php          # countries / (all)States / (all)Cities / cityDetail + findByName
│   ├── Content.php        # dailyContent
│   └── Prayer.php         # daily / weekly / monthly / ramadan / eid
└── tests/ClientTest.php   # Enjekte gönderici ile (uçtan uca + fallback + token-reuse)
```

## Kütüphane olarak kullanma
```php
require 'awqat/Config.php'; require 'awqat/TokenCache.php';
require 'awqat/Client.php'; require 'awqat/Prayer.php';

$c = new Awqat\Client(Awqat\Config::load());
$times = Awqat\Prayer::daily($c, 9528); // Isparta merkez ilçe id
```

## Notlar
- **Kota:** her endpoint path'i için **günde 5 istek** (Developer rol: 100). Token önbelleği login'i korur.
- **Token:** access ~30 dk; süresi dolmadan otomatik refresh, gerekirse yeniden login.
- **Ağ:** geçici bağlantı resetlerine karşı istekler 3 kez denenir.
