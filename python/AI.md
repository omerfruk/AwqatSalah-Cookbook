# AI.md — Python implementasyonu (AI promptu)

> Bu dosya bir AI'a **doğrudan verilebilir**. Amacı: AwqatSalah istemcisini **Python** ile
> sıfırdan üretmek (veya bu örneği genişletmek). Tek doğruluk kaynağı: [`../SPEC.md`](../SPEC.md).
> Referans desen: kardeş [`../go`](../go) ve [`../js`](../js) implementasyonları.

---

## Görev
AwqatSalah (Diyanet Namaz Vakti) API'si için Python'da, **sıfır pip bağımlılığıyla** (sadece stdlib),
tek komutla çalışan bir örnek üret. `python3 main.py` çalışınca **SPEC.md §4'teki tüm endpoint'leri**
sırayla birer kez çağırmalı ve okunaklı yazdırmalı. `python3 main.py --login` ise sadece kimlik
adımını çalıştırmalı (kotayı korumak için).

## Ön koşul: ÖNCE SPEC.md'yi oku
`../SPEC.md` API'nin tamamını içerir: base URL, auth, **tüm endpoint'ler**, DTO alanları, ortak zarf,
kota, token yeniden kullanımı (§2.5) ve config konvansiyonu (§6). **Oradan sapma.**

## Üretilecek dosyalar
```
python/
├── main.py                # tüm endpoint turu + --login bayrağı, okunaklı çıktı
└── awqat/
    ├── __init__.py
    ├── config.py          # load_config(): env + kök .env (yukarı doğru arama), AWQAT_* sözlüğü; Config dataclass
    ├── token.py           # .awqat-token.json oku/yaz (login kotası koruması, 0600)
    ├── client.py          # AwqatClient: ensure_auth, _login(/api/Auth), _refresh, get_json/do_get/do_post, 401 retry, kota, ağ-retry; unwrap
    ├── place.py           # countries, all_states, states(id), all_cities, cities(id), city_detail; find_by_name/fold (Türkçe-duyarsız)
    ├── content.py         # daily_content
    ├── prayer.py          # daily/weekly/monthly/ramadan/eid
    └── test_client.py     # unittest + http.server mock (uçtan uca + auth fallback + token-reuse)
```
> Endpoint fonksiyonları istemciyi ilk argüman alan modül fonksiyonlarıdır: `prayer.daily(client, city_id)`.

## Kritik kurallar (SPEC.md §8 ile aynı)
1. **Tüm yollar `/api/...`** — Auth dâhil: `POST /api/Auth/Login`, `GET /api/Auth/RefreshToken/{rt}`.
   (Sağlamlık: `/api/Auth/Login` 404 dönerse `/Auth/Login`'a düş.)
2. Auth dışı her istekte `Authorization: Bearer <accessToken>`.
3. Yanıt zarfını aç (`unwrap`): `success` kontrol et, değilse `message` ile hata fırlat, yoksa `data` dön.
4. **401** → token'ı sıfırla, yeniden kimlik doğrula, isteği **1 kez** tekrar et.
5. Token süresini izle (~30 dk), dolmadan refresh; refresh başarısızsa login.
6. Config: **os.environ → kök `.env`** (yukarı doğru aranır). Yeni env adı uydurma; `AWQAT_*` kullan.
7. Kota: aynı path'e günde 5'ten fazla isteği engelleyen süreç-içi sayaç.
8. İsim eşleştirme Türkçe + büyük/küçük harf **duyarsız** (`str.maketrans` ile fold).
9. **`city_detail["id"]` STRING**, `place["id"]` NUMBER — karıştırma.
10. **Token'ı yeniden kullan:** login/refresh sonrası `{accessToken,refreshToken,expiry}`'yi
    `.awqat-token.json`'a (ISO tarih, 0600, gitignore'lu) yaz; başlangıçta yükle; geçerliyse login ATMA (§2.5).
    `AWQAT_ACCESS_TOKEN` env override'ını destekle. `--login` bayrağı sun.

## Python'a özel notlar
- Sadece stdlib: `urllib.request`/`urllib.error` (HTTP), `json`, `os`, `pathlib`, `datetime`, `dataclasses`.
- `urlopen` 4xx/5xx için `HTTPError` fırlatır → yakala, (code, body) döndür. Ağ hatası (`OSError`) → birkaç kez tekrar dene.
- Token önbelleği biçimi Go/JS ile **uyumlu** (`expiry` ISO 8601) — diller aynı dosyayı paylaşır.
  Eski Python uyumu için ISO ayrıştırmada `Z` ve fractional saniyeyi normalize et.
- Test için `unittest` + `http.server` (mock) yeter; harici kütüphane kullanma. `server_close()` ile soketleri kapat.

## Bitti kabul kriterleri
- [ ] `python3 -m py_compile` temiz; `python3 -m unittest discover` geçiyor (uyarısız).
- [ ] `.env` yokken nazik bir hata (creds gerekli mesajı).
- [ ] `.env` doluyken SPEC.md §4'teki tüm endpoint'leri çağırıp sonuçları yazdırıyor.
- [ ] İkinci çalıştırmada cache'ten token kullanıp login ATMIYOR.
- [ ] `test_client.py` mock sunucuyla tüm akışı + fallback + token-reuse'u doğruluyor (gerçek API'siz).

---

## Üst seviye reçete
Mimari ve "yeni dil ekleme" adımları: [`../CLAUDE.md`](../CLAUDE.md).
