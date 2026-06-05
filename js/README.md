# AwqatSalah Cookbook — JavaScript (Node)

Diyanet AwqatSalah (Namaz Vakti) API'sini **sıfır bağımlılıkla** (yerleşik `fetch`, ESM) kullanan,
çalışır Node örneği. `node index.js` ile **resmi dokümandaki tüm endpoint'leri** sırayla çağırır.

> 🤖 Bir AI'a bu örneği oluşturtmak/genişletmek mi istiyorsunuz? → [`AI.md`](./AI.md)
> API'nin tam sözleşmesi → [`../SPEC.md`](../SPEC.md)

---

## Gereksinimler
- Node.js 18+ (yerleşik `fetch` için; geliştirilen sürüm: Node 22). **npm bağımlılığı yok.**
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
cd js
node index.js            # tüm endpoint turu
node index.js --login    # SADECE token al/yenile (endpoint turunu atla, kotayı korur)
npm test                 # mock sunucuya karşı testler (node --test)
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
**geçerliyse yeniden login ATMAZ**. Bu dosya **tüm dillerle ortaktır** — Go ile login olunca JS de
aynı token'ı kullanır (ve tersi). İstersen `.env`'e elle `AWQAT_ACCESS_TOKEN` da yazabilirsin.

## Dosya düzeni
```
js/
├── index.js               # Demo: tüm endpoint turu (--login ile sadece kimlik)
├── package.json           # type: module · bağımlılık yok
└── awqat/
    ├── config.js          # .env + ortam değişkeni yükleme (yukarı doğru .env arar)
    ├── token.js           # Token önbelleği (.awqat-token.json) oku/yaz
    ├── client.js          # AwqatClient: login(/api/Auth), refresh, Bearer, 401 retry, kota, ağ-retry
    ├── place.js           # countries / (all)states / (all)cities / cityDetail + findByName
    ├── content.js         # dailyContent
    ├── prayer.js          # daily / weekly / monthly / ramadan / eid
    └── client.test.js     # node:test + node:http mock (uçtan uca + fallback + token-reuse)
```

## Kütüphane olarak kullanma
```js
import { loadConfig } from './awqat/config.js';
import { AwqatClient } from './awqat/client.js';
import { daily } from './awqat/prayer.js';

const c = new AwqatClient(loadConfig());
const times = await daily(c, 9528); // Isparta merkez ilçe id
```

## Notlar
- **Kota:** her endpoint path'i için **günde 5 istek** (Developer rol: 100). Token önbelleği login'i korur.
- **Token:** access ~30 dk; süresi dolmadan otomatik refresh, gerekirse yeniden login.
- **Ağ:** geçici bağlantı resetlerine karşı istekler 3 kez denenir.
