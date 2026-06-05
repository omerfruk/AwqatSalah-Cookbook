# AwqatSalah Cookbook — Go

Diyanet AwqatSalah (Namaz Vakti) API'sini **sıfır bağımlılıkla** (sadece standart kütüphane)
kullanan, çalışır Go örneği. `go run .` ile **resmi dokümandaki tüm endpoint'leri** sırayla çağırır
(Isparta baz alınarak).

> 🤖 Bir AI'a bu örneği oluşturtmak/genişletmek mi istiyorsunuz? → [`AI.md`](./AI.md)
> API'nin tam sözleşmesi → [`../SPEC.md`](../SPEC.md)

---

## Gereksinimler
- Go 1.23+
- Bir AwqatSalah hesabı (<https://awqatsalah.diyanet.gov.tr>)

## Kurulum
Kimlik bilgileri **repo kökündeki tek `.env`** dosyasından okunur (tüm diller aynı dosyayı paylaşır):

```bash
# repo kökünde
cp .env.example .env
# .env içine AWQAT_EMAIL ve AWQAT_PASSWORD yazın
```

## Çalıştırma
```bash
cd go
go run .
```

### Örnek çıktı (gerçek API)
```
====================================================
  AwqatSalah Cookbook — Go (tüm endpoint turu)
  Base: https://awqatsalah.diyanet.gov.tr
  Not: her endpoint 1 kez çağrılır · kota: path başına 5/gün
====================================================

[1] Login
     /api/Auth/Login
   ✓ token alındı: eyJhbGciOiJIUzI1…

[2] Günlük İçerik (Ayet/Hadis/Dua)
     /api/DailyContent
   Ayet : "İnanan ve salih ameller işleyenler…"
   Hadis: "Kim 'Allah'tan başka ilâh yoktur' der ve…"
   Dua  : "Allah'ım! Kötü bir ömür sürmekten Sana sığınırım…"

[3] Ülkeler            /api/Place/Countries          → 208 ülke · TÜRKİYE (id=2)
[4] Tüm iller          /api/Place/States             → 364 il
[5] Ülkeye göre iller  /api/Place/States/{id}        → 81 il · ISPARTA (id=538)
[6] Tüm ilçeler        /api/Place/Cities             → 8606 ilçe
[7] İle göre ilçeler   /api/Place/Cities/{id}        → 13 ilçe · ISPARTA (id=9528)
[8] İlçe detay         /api/Place/CityDetail/{id}    → kıble 151° · Kâbe 2023 km

[9] Günlük namaz vakitleri  /api/PrayerTime/Daily/{id}
   🕌 ISPARTA — 05.06.2026 (Hicri: 19.12.1447)
      İmsak  (Fajr)      03:44
      Güneş  (Sunrise)   05:29
      Öğle   (Dhuhr)     13:01
      İkindi (Asr)       16:54
      Akşam  (Maghrib)   20:23
      Yatsı  (Isha)      22:01

[10] Haftalık  → 8 gün     [11] Aylık → 32 gün
[12] Bayram    → Ramazan B. 20 Mart 2026 · Kurban B. 27 Mayıs 2026
[13] Ramazan imsakiyesi → 30 gün

✅ Tüm endpoint'ler çağrıldı.
```

---

## Kapsanan endpoint'ler
`/api/Auth/Login` · `/api/Auth/RefreshToken/{rt}` (otomatik) · `/api/DailyContent` ·
`/api/Place/Countries` · `/api/Place/States` · `/api/Place/States/{countryId}` ·
`/api/Place/Cities` · `/api/Place/Cities/{stateId}` · `/api/Place/CityDetail/{cityId}` ·
`/api/PrayerTime/Daily|Weekly|Monthly|Eid|Ramadan/{cityId}`

## Dosya düzeni
```
go/
├── main.go            # Demo: tüm endpoint turu (login → il/ilçe → vakitler → diğerleri)
└── awqat/             # Yeniden kullanılabilir istemci paketi
    ├── config.go      # .env + ortam değişkeni yükleme (yukarı doğru .env arar)
    ├── models.go      # DTO'lar (zarf, token, place, dailyContent, cityDetail, prayer, eid)
    ├── client.go      # HTTP: login(/api/Auth), refresh, Bearer, 401 retry, kota sayacı
    ├── token.go       # Token önbelleği (.awqat-token.json) oku/yaz
    ├── place.go       # Countries / (All)States / (All)Cities / CityDetail + FindByName
    ├── content.go     # DailyContent (günün ayet/hadis/dua)
    ├── prayer.go      # Daily / Weekly / Monthly / Ramadan / Eid
    └── client_test.go # Mock sunucuya karşı uçtan uca + fallback + token-reuse testleri
```

## Konfigürasyon
| Değişken | Zorunlu | Varsayılan |
|----------|---------|------------|
| `AWQAT_EMAIL` | ✅ | — |
| `AWQAT_PASSWORD` | ✅ | — |
| `AWQAT_BASE_URL` | ✖ | `https://awqatsalah.diyanet.gov.tr` |
| `AWQAT_COUNTRY` | ✖ | `Türkiye` |
| `AWQAT_STATE` | ✖ | `Isparta` |
| `AWQAT_COUNTRY_ID` / `AWQAT_STATE_ID` / `AWQAT_CITY_ID` | ✖ | (boşsa isimden bulunur) |

> Başka bir il için: `.env`'de `AWQAT_STATE=Ankara` yapın. ID biliyorsanız
> `AWQAT_STATE_ID`/`AWQAT_CITY_ID` vererek isim aramasını atlayıp kotadan tasarruf edebilirsiniz.

## Paketi kendi kodunuzda kullanma
```go
cfg, _ := awqat.Load()
c := awqat.New(cfg)
times, _ := c.DailyPrayerTimes(ctx, 9528) // Isparta merkez ilçe id
```

## Token yeniden kullanımı (login kotası koruması)
Login günde **5 istekle** sınırlıdır. Program login sonrası token'ı otomatik olarak kök dizindeki
gizli **`.awqat-token.json`** (izin `0600`, `.gitignore`'lu) dosyasına kaydeder ve sonraki
çalıştırmalarda **geçerliyse yeniden login ATMAZ**:
```bash
go run . -login   # SADECE token al/yenile, endpoint turunu atla (kotayı korur)
go run . -login   # 2. kez: "login ATILMADI (kota korundu)" — cache'ten okur
```
Süresi dolunca otomatik refresh, o da olmazsa yeniden login olunur. İstersen `.env`'e elle
`AWQAT_ACCESS_TOKEN` da yazabilirsin.

## Notlar
- **Kota:** Standart rolde her endpoint path'i için **günde 5 istek** (Developer rol: 100).
  Bir tam tur her path'ten 1 düşer. Token önbelleği sayesinde tekrar çalıştırmalar login harcamaz.
- **Token:** access ~30 dk; süresi dolmadan otomatik refresh, gerekirse yeniden login.
- **Test:** `go test ./...` — gerçek API'ye gerek yok (mock sunucu + fallback + token-reuse testi).
