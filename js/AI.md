# AI.md — JavaScript (Node) implementasyonu (AI promptu)

> Bu dosya bir AI'a **doğrudan verilebilir**. Amacı: AwqatSalah istemcisini **Node.js** ile
> sıfırdan üretmek (veya bu örneği genişletmek). Tek doğruluk kaynağı: [`../SPEC.md`](../SPEC.md).
> Referans desen: kardeş [`../go`](../go) implementasyonu.

---

## Görev
AwqatSalah (Diyanet Namaz Vakti) API'si için Node'da, **sıfır npm bağımlılığıyla** (yerleşik `fetch`, ESM),
tek komutla çalışan bir örnek üret. `node index.js` çalışınca **SPEC.md §4'teki tüm endpoint'leri**
sırayla birer kez çağırmalı ve okunaklı yazdırmalı (login → DailyContent → il/ilçe → vakitler).
`node index.js --login` ise sadece kimlik adımını çalıştırmalı (kotayı korumak için).

## Ön koşul: ÖNCE SPEC.md'yi oku
`../SPEC.md` API'nin tamamını içerir: base URL, auth, **tüm endpoint'ler**, DTO alanları, ortak zarf,
kota, token yeniden kullanımı (§2.5) ve config konvansiyonu (§6). **Oradan sapma.**

## Üretilecek dosyalar
```
js/
├── package.json           # { "type": "module" } · bağımlılık YOK · scripts: start/login/test
├── index.js               # tüm endpoint turu + --login bayrağı, okunaklı çıktı
└── awqat/
    ├── config.js          # loadConfig(): env + kök .env (yukarı doğru arama), AWQAT_* sözlüğü
    ├── token.js           # .awqat-token.json oku/yaz (login kotası koruması, 0600)
    ├── client.js          # AwqatClient: ensureAuth, login(/api/Auth), refresh, getJSON/doGet/doPost, 401 retry, kota, ağ-retry; unwrap
    ├── place.js           # countries, allStates, states(id), allCities, cities(id), cityDetail; findByName/fold (Türkçe-duyarsız)
    ├── content.js         # dailyContent
    ├── prayer.js          # daily/weekly/monthly/ramadan/eid
    └── client.test.js     # node:test + node:http mock (uçtan uca + auth fallback + token-reuse)
```
> Endpoint fonksiyonları istemciyi ilk argüman alan serbest fonksiyonlardır: `daily(client, cityId)`.

## Kritik kurallar (SPEC.md §8 ile aynı)
1. **Tüm yollar `/api/...`** — Auth dâhil: `POST /api/Auth/Login`, `GET /api/Auth/RefreshToken/{rt}`.
   (Sağlamlık: `/api/Auth/Login` 404 dönerse `/Auth/Login`'a düş.)
2. Auth dışı her istekte `Authorization: Bearer <accessToken>`.
3. Yanıt zarfını aç (`unwrap`): `success` kontrol et, değilse `message` ile hata fırlat, yoksa `data` dön.
4. **401** → token'ı sıfırla, yeniden kimlik doğrula, isteği **1 kez** tekrar et.
5. Token süresini izle (~30 dk), dolmadan refresh; refresh başarısızsa login.
6. Config: **process.env → kök `.env`** (yukarı doğru aranır). Yeni env adı uydurma; `AWQAT_*` kullan.
7. Kota: aynı path'e günde 5'ten fazla isteği engelleyen süreç-içi sayaç.
8. İsim eşleştirme Türkçe + büyük/küçük harf **duyarsız** ("TÜRKİYE"↔"Türkiye").
9. **`cityDetail.id` STRING**, `place.id` NUMBER — karıştırma.
10. **Token'ı yeniden kullan:** login/refresh sonrası `{accessToken,refreshToken,expiry}`'yi
    `.awqat-token.json`'a (ISO tarih, 0600, gitignore'lu) yaz; başlangıçta yükle; geçerliyse login ATMA (§2.5).
    `AWQAT_ACCESS_TOKEN` env override'ını destekle. `--login` bayrağı sun.

## Node'a özel notlar
- ESM (`import`/`export`), `package.json` → `"type": "module"`.
- Yerleşik `fetch` + `AbortSignal.timeout(...)`. Ağ hatasında (reset/ölü keep-alive) isteği birkaç kez tekrar dene
  (Go'nun net/http'si otomatik yapar; fetch yapmaz).
- Token önbelleği biçimi Go ile **uyumlu** olmalı (`expiry` ISO 8601 string) — diller aynı dosyayı paylaşır.
- Test için `node:test` + `node:http` (mock) yeter; harici test kütüphanesi kullanma.

## Bitti kabul kriterleri
- [ ] `node --check` tüm dosyalarda temiz; `node --test` geçiyor.
- [ ] `.env` yokken nazik bir hata (creds gerekli mesajı).
- [ ] `.env` doluyken SPEC.md §4'teki tüm endpoint'leri çağırıp sonuçları yazdırıyor.
- [ ] İkinci çalıştırmada cache'ten token kullanıp login ATMIYOR.
- [ ] `client.test.js` mock sunucuyla tüm akışı + fallback + token-reuse'u doğruluyor (gerçek API'siz).

---

## Üst seviye reçete
Mimari ve "yeni dil ekleme" adımları: [`../CLAUDE.md`](../CLAUDE.md).
