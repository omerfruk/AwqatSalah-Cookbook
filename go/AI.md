# AI.md — Go implementasyonu (AI promptu)

> Bu dosya bir AI'a **doğrudan verilebilir**. Amacı: AwqatSalah istemcisini **Go** dilinde
> sıfırdan üretmek (veya bu örneği genişletmek). Tek doğruluk kaynağı: [`../SPEC.md`](../SPEC.md).
> Bu, diğer dillerin `AI.md` dosyaları için de **şablondur** (dili/komutu değiştir, deseni koru).

---

## Görev
AwqatSalah (Diyanet Namaz Vakti) API'si için Go'da, **sıfır harici bağımlılıkla** (sadece stdlib),
tek komutla çalışan bir örnek üret. `go run .` çalışınca **SPEC.md §4'teki tüm endpoint'leri**
sırayla birer kez çağırmalı ve okunaklı yazdırmalı:

- **A) Login** → token.
- **B) İl/İlçe**: Türkiye → Isparta → ilçe (DailyContent, AllStates, AllCities, CityDetail dâhil).
- **C) Namaz vakitleri**: Daily + Weekly + Monthly + Eid + Ramadan.

## Ön koşul: ÖNCE SPEC.md'yi oku
`../SPEC.md` API'nin tamamını içerir: base URL, auth akışı, **tüm endpoint'ler**, DTO alanları,
ortak yanıt zarfı, kota, config konvansiyonu ve akışlar. **Oradan sapma.**

## Üretilecek dosyalar
```
go/
├── go.mod                 # module .../go ; go 1.23
├── main.go                # package main: tüm endpoint turu, okunaklı çıktı
└── awqat/                 # yeniden kullanılabilir istemci paketi
    ├── config.go          # Load(): env + kök .env (yukarı doğru arama), AWQAT_* sözlüğü
    ├── models.go          # APIResponse[T], Token, Place, DailyContent, CityDetail, PrayerTime, EidPrayerTime
    ├── client.go          # New(token tohumla), EnsureAuth, login(/api/Auth), refresh, DoGet/DoPost, 401 retry, kota; generic unwrap/getJSON
    ├── token.go           # .awqat-token.json oku/yaz (login kotası koruması)
    ├── place.go           # Countries, AllStates, States(id), AllCities, Cities(id), CityDetail; FindByName (Türkçe-duyarsız)
    ├── content.go         # DailyContent()
    ├── prayer.go          # Daily/Weekly/Monthly/Ramadan/Eid PrayerTimes
    └── client_test.go     # httptest mock'a karşı uçtan uca + auth fallback + token-reuse testi
```

## Kritik kurallar (SPEC.md §8 ile aynı)
1. **Tüm yollar `/api/...`** — Auth dâhil: `POST /api/Auth/Login`, `GET /api/Auth/RefreshToken/{rt}`.
   (Sağlamlık: `/api/Auth/Login` 404 dönerse `/Auth/Login`'a düş.)
2. Auth dışı her istekte `Authorization: Bearer <accessToken>`.
3. Yanıt zarfını aç: `success` kontrol et, değilse `message` ile hata fırlat, yoksa `data` dön.
4. **401** → token'ı sıfırla, yeniden kimlik doğrula, isteği **1 kez** tekrar et.
5. Token süresini izle (~30 dk), dolmadan refresh et; refresh başarısızsa login.
6. Config: **OS ortam değişkenleri → kök `.env`** (yukarı doğru aranır). Yeni env adı uydurma; `AWQAT_*` kullan.
7. Kota: aynı path'e günde 5'ten fazla isteği engelleyen süreç-içi sayaç ekle.
8. İsim eşleştirme Türkçe + büyük/küçük harf **duyarsız** ("TÜRKİYE"↔"Türkiye", "ISPARTA"↔"Isparta").
9. **`CityDetail.id` STRING**, `Place.id` NUMBER — karıştırma.
10. **Token'ı yeniden kullan:** login/refresh sonrası `{accessToken,refreshToken,expiry}`'yi
    `.awqat-token.json`'a (0600, gitignore'lu) yaz; başlangıçta yükle; geçerliyse login ATMA (SPEC §2.5).
    `AWQAT_ACCESS_TOKEN` env override'ını da destekle. Tek-kimlik-adımı bayrağı sun (Go: `-login`).

## Go'ya özel notlar
- Generic'ler: `APIResponse[T]`, `unwrap[T]`, `getJSON[T]` (metotlar tip parametresi alamaz → paket düzeyi fonksiyon kullan).
- Sadece stdlib: `net/http`, `encoding/json`, `context`, `sync`, `time`.
- `Place.id` `int64`; `greenwichMeanTimezone` için `json.Number`; `CityDetail.id` için `string`.
- `gofmt` temiz olmalı, `go vet ./...` ve `go test ./...` geçmeli.

## Bitti kabul kriterleri
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` geçiyor; `gofmt -l .` boş.
- [ ] `.env` yokken nazik bir hata veriyor (creds gerekli mesajı).
- [ ] `.env` doluyken SPEC.md §4'teki tüm endpoint'leri çağırıp sonuçları yazdırıyor.
- [ ] `client_test.go` mock sunucuyla tüm akışı + `/api/Auth` fallback'i doğruluyor (gerçek API'siz).

---

## Başka bir dile uyarlarken (örn. JS/Python)
Bu dosyayı şablon al; sadece şunları değiştir: **dil, çalıştırma komutu** (`node index.js`,
`python main.py`...), **dosya isimleri/idiyomlar** ve dilin paket yöneticisi. Akış (tüm endpoint turu),
`AWQAT_*` config sözlüğü, kök `.env` okuma, `/api/...` yolları, zarf açma, 401 retry ve
Türkçe-duyarsız eşleştirme **aynı kalır**. Üst seviye reçete: [`../CLAUDE.md`](../CLAUDE.md).
