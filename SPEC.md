# SPEC — AwqatSalah API Sözleşmesi (Dilden Bağımsız)

> Bu dosya, AwqatSalah (Diyanet İşleri Başkanlığı Namaz Vakti) API'sinin **tek doğruluk kaynağıdır**.
> Herhangi bir programlama dilinde istemci yazmak için ihtiyacın olan her şey buradadır.
> Yeni bir dil eklerken: **bu dosyayı oku + `go/` referansını taklit et.** Bkz. [CLAUDE.md](./CLAUDE.md).
>
> Kaynak: Diyanet resmi "Namaz Vakitleri REST Servisi" kılavuzu. Tüm endpoint'ler canlı API'ye karşı doğrulanmıştır.

---

## 1. Genel Bakış

AwqatSalah; ülke/eyalet/şehir (il/ilçe) listeleri, namaz vakitleri ve günlük dini içerik sunan,
**JWT tabanlı kimlik doğrulamalı** bir REST API'dir.

- **Base URL:** `https://awqatsalah.diyanet.gov.tr`
- **Kimlik doğrulama:** Bearer token (login → access + refresh token)
- **Tüm yanıtlar** ortak bir zarf (envelope) ile döner (bkz. §3).
- **Kota:** **Standart Rol** → her endpoint için **günde 5 istek**. **Developer Rol** → günde 100.
  (Parametre alan endpoint'lerde her parametre kombinasyonu ayrı sayılır.)

---

## 2. Kimlik Doğrulama (Auth) — `/api/Auth/...`

### 2.1 Login
```
POST {BASE_URL}/api/Auth/Login
Content-Type: application/json

{ "email": "<AWQAT_EMAIL>", "password": "<AWQAT_PASSWORD>" }
```
Yanıt `data`: `{ "accessToken": "...", "refreshToken": "..." }`

### 2.2 Refresh Token
```
GET {BASE_URL}/api/Auth/RefreshToken/{refreshToken}
```
Yanıt `data`: `{ "accessToken": "...", "refreshToken": "..." }`

### 2.3 Token kullanımı
Auth dışındaki **her** istekte header gönderilir:
```
Authorization: Bearer <accessToken>
```

### 2.4 Token ömürleri ve akış
- **Access token:** ~30 dakika geçerli.
- **Refresh penceresi:** access bittikten sonra ~15 dakika daha refresh ile yenilenebilir.
- Pratik kural (5 dk güvenlik marjı): token'ı `now + 25dk`'da "yenilenmeli" say.
- **Akış:** token geçerli → kullan · süresi dolmuş → refresh dene · refresh de başarısız/yok → tekrar login.
- Bir istek **401** dönerse: token'ı geçersiz say, yeniden kimlik doğrula ve isteği **1 kez** tekrarla.

> Tarihsel not: bazı eski örneklerde login `/Auth/Login` (api'siz) kullanılır. Resmî ve doğrulanmış
> yol **`/api/Auth/Login`**'dir. Sağlamlık için 404 alınırsa `/Auth/...`'a düşülebilir.

### 2.5 Token'ı yeniden kullan (kota koruması) — ZORUNLU konvansiyon
Login günde **5 istekle** sınırlı olduğundan, her çalıştırmada yeniden login olmak kotayı tüketir.
Bu yüzden her dil token'ı **kalıcı saklamalı ve yeniden kullanmalı**:
1. Login/refresh başarılı olunca `{accessToken, refreshToken, expiry}`'yi gizli bir dosyaya yaz:
   **`.awqat-token.json`** (kök `.env` yanında, `.gitignore`'lu, izin `0600`).
2. Başlangıçta bu dosyayı (veya `AWQAT_ACCESS_TOKEN` ortam değişkenini) yükle.
   Token varsa ve `now < expiry` → **login ATMA**, doğrudan kullan.
3. Süresi dolduysa → refresh dene; o da olmazsa → login. 401 alınırsa token'ı sıfırla, yeniden login.
> Sonuç: token geçerli olduğu sürece (~30 dk) tekrar tekrar çalıştırmak login kotası harcamaz.

---

## 3. Ortak Yanıt Zarfı (IResult)

**Her** endpoint bu zarfı döndürür:
```json
{ "data": <T>, "success": true, "message": null }
```
- `success=false` ise `message` hata sebebini içerir; `data`'yı kullanma, hata fırlat.
- `<T>` endpoint'e göre nesne ya da dizi olur (aşağıda her biri belirtilmiştir).

---

## 4. Endpoint'ler (tamamı)

### 4.1 Auth — `/api/Auth/...`
| Amaç | Method & Path | `data` |
|------|---------------|--------|
| Giriş | `POST /api/Auth/Login` body `{email,password}` | `Token` |
| Token yenile | `GET /api/Auth/RefreshToken/{refreshToken}` | `Token` |

### 4.2 DailyContent — `/api/DailyContent`
| Amaç | Method & Path | `data` |
|------|---------------|--------|
| Günün Ayet/Hadis/Dua'sı | `GET /api/DailyContent` (parametresiz) | `DailyContent` |

### 4.3 Place (Ülke / İl / İlçe) — `/api/Place/...`
| Amaç | Method & Path | `data` |
|------|---------------|--------|
| Tüm ülkeler | `GET /api/Place/Countries` | `Place[]` |
| Tüm iller | `GET /api/Place/States` | `Place[]` |
| Ülkenin illeri | `GET /api/Place/States/{countryId}` | `Place[]` |
| Tüm ilçeler | `GET /api/Place/Cities` | `Place[]` |
| İlin ilçeleri | `GET /api/Place/Cities/{stateId}` | `Place[]` |
| İlçe detayı (kıble vb.) | `GET /api/Place/CityDetail/{cityId}` | `CityDetail` |

> Hiyerarşi: **Country → State (il) → City (ilçe)**. Namaz vakitleri **City (ilçe) id'si** ile çekilir.
> Doğrulanmış örnek: Türkiye `id=2` → Isparta `id=538` → Isparta merkez `id=9528`.
> (ID'ler değişebilir; isimden çözmek daha sağlamdır.)

### 4.4 PrayerTime (Namaz Vakitleri) — `/api/PrayerTime/...`
| Amaç | Method & Path | `data` |
|------|---------------|--------|
| Günlük | `GET /api/PrayerTime/Daily/{cityId}` | `PrayerTime[]` |
| Haftalık | `GET /api/PrayerTime/Weekly/{cityId}` | `PrayerTime[]` |
| Aylık | `GET /api/PrayerTime/Monthly/{cityId}` | `PrayerTime[]` |
| Bayram namazı | `GET /api/PrayerTime/Eid/{cityId}` | `EidPrayerTime` |
| Ramazan imsakiyesi | `GET /api/PrayerTime/Ramadan/{cityId}` | `PrayerTime[]` |

---

## 5. Veri Tipleri (DTO)

JSON alan adları **aynen** aşağıdaki gibidir (camelCase). Büyük/küçük harf önemlidir
(Go json case-insensitive'dir ama JS/Python değildir — birebir uy).

### Token
```
accessToken  : string
refreshToken : string
```

### Place (ülke / il / ilçe — ortak)
```
id   : number      // örn. ülke 2 (TÜRKİYE), il 538 (ISPARTA)
code : string      // örn. "TURKEY", "ADANA"
name : string      // örn. "TÜRKİYE", "ISPARTA"
```

### DailyContent
```
id           : number
dayOfYear    : number
verse        : string   // Ayet
verseSource  : string
hadith       : string   // Hadis
hadithSource : string
pray         : string   // Dua
praySource   : string|null
```

### CityDetail   (DİKKAT: id burada STRING'tir)
```
id                   : string   // "9528"  ← Place.id'den farklı, tırnaklı
name                 : string
code                 : string|null
geographicQiblaAngle : string
distanceToKaaba      : string
qiblaAngle           : string
city                 : string
cityEn               : string|null
country              : string
countryEn            : string
```

### PrayerTime
```
shapeMoonUrl              : string  // ay safhası görseli
fajr                      : string  // İmsak
sunrise                   : string  // Güneş
dhuhr                     : string  // Öğle
asr                       : string  // İkindi
maghrib                   : string  // Akşam
isha                      : string  // Yatsı
astronomicalSunset        : string
astronomicalSunrise       : string
hijriDateShort            : string
hijriDateShortIso8601     : string
hijriDateLong             : string
hijriDateLongIso8601      : string
qiblaTime                 : string
gregorianDateShort        : string  // "DD.MM.YYYY"
gregorianDateShortIso8601 : string  // "YYYY-MM-DD"
gregorianDateLong         : string
gregorianDateLongIso8601  : string
greenwichMeanTimezone     : number  // örn. 3 = GMT+3
```

### EidPrayerTime
```
eidAlAdhaHijri : string   // Kurban Bayramı (Hicri)
eidAlAdhaDate  : string
eidAlAdhaTime  : string
eidAlFitrHijri : string   // Ramazan Bayramı (Hicri)
eidAlFitrDate  : string
eidAlFitrTime  : string
```

---

## 6. Konfigürasyon Konvansiyonu (TÜM DİLLER İÇİN ZORUNLU)

Her dil örneği, ayarları şu sırayla okur (üstteki kazanır):
1. **İşletim sistemi ortam değişkenleri** (CI/Docker override için)
2. **Repo kökündeki `.env` dosyası** — örnek, çalıştığı klasörden yukarı doğru `.env` aranarak bulunur.

| Değişken | Zorunlu | Varsayılan |
|----------|---------|------------|
| `AWQAT_EMAIL` | ✅ | — |
| `AWQAT_PASSWORD` | ✅ | — |
| `AWQAT_BASE_URL` | ✖ | `https://awqatsalah.diyanet.gov.tr` |
| `AWQAT_COUNTRY` | ✖ | `Türkiye` |
| `AWQAT_STATE` | ✖ | `Isparta` |
| `AWQAT_COUNTRY_ID` / `AWQAT_STATE_ID` / `AWQAT_CITY_ID` | ✖ | (boşsa isimden bulunur) |
| `AWQAT_ACCESS_TOKEN` / `AWQAT_REFRESH_TOKEN` | ✖ | (elle token; bkz. §2.5) |

> Amaç: kullanıcı creds'i **bir kez** kök `.env`'e girsin, **her dil** çalışsın.
> Token önbelleği (`.awqat-token.json`) tüm diller için **ortaktır** — bir dilde login olunca diğeri de kullanabilir.

---

## 7. Üç Temel Akış (her dil en azından bunları göstermeli)

### Akış A — Login
```
token = POST /api/Auth/Login { email, password }   →  data.accessToken
```

### Akış B — İl/İlçe çek (Isparta örneği)
```
if AWQAT_STATE_ID yoksa:
    countries = GET /api/Place/Countries
    country   = countries içinde adı AWQAT_COUNTRY ile eşleşen   (örn. "Türkiye" → id 2)
    states    = GET /api/Place/States/{country.id}
    state     = states içinde adı AWQAT_STATE ile eşleşen          (örn. "Isparta" → id 538)
cities = GET /api/Place/Cities/{state.id}                          // Isparta'nın 13 ilçesi
city   = AWQAT_CITY_ID varsa o, yoksa ilçe adı il adıyla eşleşen (merkez 9528), yoksa ilk ilçe
```
> İsim eşleştirme Türkçe + büyük/küçük harf **duyarsız** olmalı ("TÜRKİYE"↔"türkiye", "ISPARTA"↔"isparta").

### Akış C — Namaz vakitlerini çek
```
times = GET /api/PrayerTime/Daily/{city.id}   →  bugünün satırı: fajr, sunrise, dhuhr, asr, maghrib, isha
```

> "Tüm API turu" isteniyorsa: A→B→C'ye ek olarak DailyContent, AllStates, AllCities, CityDetail,
> Weekly, Monthly, Eid, Ramadan da birer kez çağrılır (her biri kotadan 1 düşer).

---

## 8. Bir Dile İmplemente Ederken — Hata Önleme Kontrol Listesi

- [ ] Tüm yollar `/api/...` (Auth dâhil: `/api/Auth/Login`).
- [ ] Her istekte `Authorization: Bearer` (auth istekleri hariç).
- [ ] Yanıt zarfını aç: `success` kontrol et, `data`'yı dön.
- [ ] 401'de yeniden auth + 1 kez retry.
- [ ] Token süresini takip et (~30 dk), dolmadan refresh et.
- [ ] Token'ı `.awqat-token.json`'a kaydet & yeniden kullan — geçerliyse login ATMA (§2.5).
- [ ] Kök `.env` + ortam değişkenlerini oku (§6).
- [ ] `CityDetail.id` STRING; `Place.id` NUMBER — karıştırma.
- [ ] İsim eşleştirme Türkçe-duyarsız.
- [ ] Kota: aynı path'e günde 5'ten fazla vurma (istemci tarafı sayaç önerilir).
- [ ] Tek komutla (`go run .` / `node index.js` / `python main.py`) akış çalışıp sonucu yazdırsın.
