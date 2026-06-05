# AwqatSalah Cookbook — C# (.NET)

Diyanet AwqatSalah (Namaz Vakti) API'sini **sıfır NuGet bağımlılığıyla** (yerleşik `HttpClient` +
`System.Text.Json`) kullanan, çalışır C# örneği. `dotnet run` ile **resmi dokümandaki tüm
endpoint'leri** sırayla çağırır.

> 🤖 Bir AI'a bu örneği oluşturtmak/genişletmek mi istiyorsunuz? → [`AI.md`](./AI.md)
> API'nin tam sözleşmesi → [`../SPEC.md`](../SPEC.md)

---

## Gereksinimler
- .NET SDK 8+ (geliştirilen: 10). Projeler `net8.0` hedefler; `RollForward=Major` ile daha yeni
  runtime'larda (net9/net10) da çalışır. **NuGet paketi yok.**
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
cd csharp/AwqatSalah
dotnet run                  # tüm endpoint turu
dotnet run -- --login       # SADECE token al/yenile (endpoint turunu atla, kotayı korur)

# Testler (sahte gönderici, gerçek API'siz):
cd ../AwqatSalah.Tests && dotnet run
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
**geçerliyse yeniden login ATMAZ**. Bu dosya **tüm dillerle ortaktır** — Go/JS/Python/PHP ile login
olunca C# da aynı token'ı kullanır. İstersen `.env`'e elle `AWQAT_ACCESS_TOKEN` da yazabilirsin.

## Dosya düzeni
```
csharp/
├── AwqatSalah/                  # demo projesi
│   ├── AwqatSalah.csproj        #   net8.0, sıfır NuGet, RollForward=Major
│   ├── Program.cs               #   tüm endpoint turu (--login ile sadece kimlik)
│   └── Awqat/
│       ├── Config.cs            #   .env + ortam değişkeni yükleme (yukarı doğru .env arar)
│       ├── TokenCache.cs        #   Token önbelleği (.awqat-token.json) oku/yaz
│       ├── Models.cs            #   DTO'lar (ApiResponse<T>, Token, Place, DailyContent, CityDetail, PrayerTime, EidPrayerTime)
│       ├── Client.cs            #   login(/api/Auth), refresh, Bearer, 401 retry, kota, ağ-retry; Unwrap<T>
│       ├── Places.cs            #   Countries/(All)States/(All)Cities/CityDetail + FindByName/Fold
│       ├── Content.cs           #   DailyContent
│       └── Prayer.cs            #   Daily/Weekly/Monthly/Ramadan/Eid
└── AwqatSalah.Tests/            # test projesi (Awqat/*.cs'i paylaşır, Program.cs hariç)
    ├── AwqatSalah.Tests.csproj
    └── TestRunner.cs            #   enjekte gönderici ile (uçtan uca + fallback + token-reuse)
```

## Kütüphane olarak kullanma
```csharp
using Awqat;

var c = new Client(Config.Load());
var times = await Prayer.DailyAsync(c, 9528); // Isparta merkez ilçe id
```

## Notlar
- **Kota:** her endpoint path'i için **günde 5 istek** (Developer rol: 100). Token önbelleği login'i korur.
- **Token:** access ~30 dk; süresi dolmadan otomatik refresh, gerekirse yeniden login.
- **Ağ:** geçici bağlantı resetlerine karşı istekler 3 kez denenir.
