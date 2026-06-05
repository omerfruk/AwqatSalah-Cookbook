# AI.md — C# (.NET) implementasyonu (AI promptu)

> Bu dosya bir AI'a **doğrudan verilebilir**. Amacı: AwqatSalah istemcisini **C#/.NET** ile
> sıfırdan üretmek (veya bu örneği genişletmek). Tek doğruluk kaynağı: [`../SPEC.md`](../SPEC.md).
> Referans desen: kardeş [`../go`](../go), [`../js`](../js), [`../python`](../python), [`../php`](../php).

---

## Görev
AwqatSalah (Diyanet Namaz Vakti) API'si için C#'ta, **sıfır NuGet bağımlılığıyla** (yerleşik
`HttpClient` + `System.Text.Json`), tek komutla çalışan bir örnek üret. `dotnet run` çalışınca
**SPEC.md §4'teki tüm endpoint'leri** sırayla birer kez çağırmalı ve okunaklı yazdırmalı.
`dotnet run -- --login` ise sadece kimlik adımını çalıştırmalı (kotayı korumak için).

## Ön koşul: ÖNCE SPEC.md'yi oku
`../SPEC.md` API'nin tamamını içerir: base URL, auth, **tüm endpoint'ler**, DTO alanları, ortak zarf,
kota, token yeniden kullanımı (§2.5) ve config konvansiyonu (§6). **Oradan sapma.**

## Üretilecek dosyalar (namespace `Awqat`)
```
csharp/
├── AwqatSalah/AwqatSalah.csproj   # Exe, net8.0, Nullable+ImplicitUsings, InvariantGlobalization, RollForward=Major
├── AwqatSalah/Program.cs          # top-level statements: tüm endpoint turu + --login
└── AwqatSalah/Awqat/
    ├── Config.cs                  # Config.Load(): env + kök .env (yukarı doğru arama), AWQAT_* sözlüğü
    ├── TokenCache.cs              # .awqat-token.json oku/yaz (camelCase JSON, 0600)
    ├── Models.cs                  # ApiResponse<T>, Token, Place, DailyContent, CityDetail, PrayerTime, EidPrayerTime
    ├── Client.cs                  # EnsureAuthAsync, login(/api/Auth), refresh, GetJsonAsync<T>, 401 retry, kota, ağ-retry; static Unwrap<T>; enjekte edilebilir sender
    ├── Places.cs                  # CountriesAsync, AllStatesAsync, StatesAsync, AllCitiesAsync, CitiesAsync, CityDetailAsync; FindByName/Fold (Türkçe-duyarsız)
    ├── Content.cs                 # DailyContentAsync
    └── Prayer.cs                  # Daily/Weekly/Monthly/Ramadan/Eid Async
└── AwqatSalah.Tests/             # ayrı Exe proje; EnableDefaultCompileItems=false; ../AwqatSalah/Awqat/**/*.cs + TestRunner.cs (Program.cs HARİÇ)
```
> Endpoint metotları istemciyi ilk argüman alan statik metotlardır: `Prayer.DailyAsync(client, cityId)`.
> DTO sınıfı `Place`, statik endpoint sınıfı `Places` (isim çakışmasın diye çoğul).

## Kritik kurallar (SPEC.md §8 ile aynı)
1. **Tüm yollar `/api/...`** — Auth dâhil: `POST /api/Auth/Login`, `GET /api/Auth/RefreshToken/{rt}`.
   (Sağlamlık: `/api/Auth/Login` 404 dönerse `/Auth/Login`'a düş.)
2. Auth dışı her istekte `Authorization: Bearer <accessToken>`.
3. Yanıt zarfını aç (`Client.Unwrap<T>`): `Success` kontrol et, değilse `Message` ile hata fırlat, yoksa `Data` dön.
4. **401** → token'ı sıfırla, yeniden kimlik doğrula, isteği **1 kez** tekrar et.
5. Token süresini izle (~30 dk), dolmadan refresh; refresh başarısızsa login.
6. Config: **Environment.GetEnvironmentVariable → kök `.env`** (yukarı doğru aranır). Yeni env adı uydurma; `AWQAT_*` kullan.
7. Kota: aynı path'e günde 5'ten fazla isteği engelleyen süreç-içi sayaç.
8. İsim eşleştirme Türkçe + büyük/küçük harf **duyarsız** (`Fold` switch + `ToLowerInvariant`).
9. **`CityDetail.Id` STRING**, `Place.Id` long (NUMBER) — karıştırma.
10. **Token'ı yeniden kullan:** login/refresh sonrası `{accessToken,refreshToken,expiry}`'yi
    `.awqat-token.json`'a (ISO `DateTimeOffset.ToString("o")`, 0600, gitignore'lu) yaz; başlangıçta yükle; geçerliyse login ATMA (§2.5).
    `AWQAT_ACCESS_TOKEN` env override'ını destekle. `--login` bayrağı sun.

## C#'a özel notlar
- `System.Text.Json` ile `PropertyNameCaseInsensitive = true` → camelCase JSON ↔ PascalCase property otomatik.
  Token önbelleğini yazarken `PropertyNamingPolicy = JsonNamingPolicy.CamelCase` (diğer dillerle uyum).
- HTTP: `HttpClient`. Ağ hatasında (exception) isteği birkaç kez tekrar dene.
- `File.SetUnixFileMode` Windows'ta yok → `if (!OperatingSystem.IsWindows())` ile koru (CA1416).
- Test edilebilirlik: `Client` kurucusu opsiyonel `Func<...,Task<(int,string)>>? sender` alsın; testler HttpClient yerine enjekte etsin.
- `RollForward=Major` → net8 hedefi daha yeni runtime'da da çalışır. `dotnet run -- --login` (argümanlar `--`'dan sonra).

## Bitti kabul kriterleri
- [ ] `dotnet build` 0 uyarı / 0 hata; `dotnet run --project AwqatSalah.Tests` 0 başarısız.
- [ ] `.env` yokken nazik bir hata (creds gerekli mesajı).
- [ ] `.env` doluyken SPEC.md §4'teki tüm endpoint'leri çağırıp sonuçları yazdırıyor.
- [ ] İkinci çalıştırmada cache'ten token kullanıp login ATMIYOR.
- [ ] Testler uçtan uca akışı + fallback + token-reuse'u doğruluyor (gerçek API'siz).

---

## Üst seviye reçete
Mimari ve "yeni dil ekleme" adımları: [`../CLAUDE.md`](../CLAUDE.md).
