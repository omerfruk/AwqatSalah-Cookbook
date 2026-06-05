# AwqatSalah-Cookbook

**Diyanet İşleri Başkanlığı Namaz Vakti API'sini ([AwqatSalah](https://awqatsalah.diyanet.gov.tr))
birden çok programlama dilinde** çalıştıran sade, kopyala-çalıştır örnek koleksiyonu.

Kimlik bilgilerinizi **bir kez** girersiniz → **her dildeki** örnek aynı bilgilerle çalışır.
Her örnek üç temel akışı gösterir:

1. 🔐 **Login** — token alma
2. 📍 **İl / İlçe** çekme (örnek: Türkiye → **Isparta**)
3. 🕌 **Namaz vakitleri** çekme (Isparta için günlük vakitler)

---

## 🚀 Kurulum & Çalıştırma

### 1. Adım — `.env` dosyasını oluşturun (kimlik bilgileri)

Tüm diller **repo kökündeki tek `.env` dosyasını** okur. Yani bilgilerinizi **bir kez**
girersiniz, her dildeki örnek aynı dosyayı kullanır.

**a)** `.env.example` dosyasını kopyalayıp adını **`.env`** yapın:

```bash
cp .env.example .env
```
> Windows PowerShell: `Copy-Item .env.example .env`
> Ya da dosya yöneticisinde `.env.example`'ı kopyalayıp adını `.env` olarak değiştirin.

**b)** `.env` dosyasını açın ve **yalnızca şu 2 alanı** kendi bilgilerinizle doldurun:

```dotenv
AWQAT_EMAIL=ornek@eposta.com      # ← AwqatSalah hesabınızın e-postası
AWQAT_PASSWORD=sifreniz           # ← AwqatSalah hesabınızın şifresi
```
> Diğer alanlar (base URL, Türkiye / Isparta) **varsayılan** gelir — dokunmanıza gerek yok.
> Hesabınız yoksa <https://awqatsalah.diyanet.gov.tr> üzerinden başvuruyla alınır.
> `.env` dosyası **gizlidir** (`.gitignore`) — repoya gönderilmez; yalnızca `.env.example` paylaşılır.

### 2. Adım — İstediğiniz dilin örneğini çalıştırın

Aşağıdaki [Diller](#-diller) tablosundaki komutu çalıştırın, örn:

```bash
cd go && go run .          # veya: cd python && python3 main.py
```

İlk çalıştırmada login olunur ve token gizli `.awqat-token.json` dosyasına kaydedilir;
sonraki çalıştırmalar bu token'ı **yeniden kullanır** (login günde 5 istekle sınırlıdır, kota korunur).
Sadece token tazelemek için her dilde `--login` (Go'da `-login`) bayrağı vardır.

### Ortam değişkenleri (`.env` içeriği)

| Değişken | Zorunlu | Açıklama / Varsayılan |
|----------|---------|------------------------|
| `AWQAT_EMAIL` | ✅ | AwqatSalah hesabı e-postası |
| `AWQAT_PASSWORD` | ✅ | AwqatSalah hesabı şifresi |
| `AWQAT_BASE_URL` | ✖ | `https://awqatsalah.diyanet.gov.tr` |
| `AWQAT_COUNTRY` / `AWQAT_STATE` | ✖ | `Türkiye` / `Isparta` (örnek baz konum) |
| `AWQAT_COUNTRY_ID` / `AWQAT_STATE_ID` / `AWQAT_CITY_ID` | ✖ | ID ile override (boşsa isimden bulunur) |
| `AWQAT_ACCESS_TOKEN` / `AWQAT_REFRESH_TOKEN` | ✖ | Elle token (genelde gerekmez; otomatik önbelleğe alınır) |

> Bu değişkenleri `.env`'e yazabilir veya işletim sistemi ortam değişkeni olarak da verebilirsiniz
> (ortam değişkeni `.env`'i ezer — CI/Docker için kullanışlı).

---

## 🌍 Diller

| Dil | Durum | Çalıştırma | Klasör |
|-----|-------|------------|--------|
| Go | ✅ Hazır | `cd go && go run .` | [`go/`](./go) |
| JavaScript (Node) | ✅ Hazır | `cd js && node index.js` | [`js/`](./js) |
| Python | ✅ Hazır | `cd python && python3 main.py` | [`python/`](./python) |
| PHP | ✅ Hazır | `cd php && php main.php` | [`php/`](./php) |
| Diğerleri | ⏳ | — | — |

> Yeni bir dil mi istiyorsunuz? Bir AI'a şunu söyleyin: **"`<dil>` için yap"** —
> AI, [`CLAUDE.md`](./CLAUDE.md) reçetesini ve [`SPEC.md`](./SPEC.md) sözleşmesini izleyerek ekler.

---

## 📦 Yapı

| Dosya | Ne işe yarar |
|-------|--------------|
| [`SPEC.md`](./SPEC.md) | Dilden bağımsız **API sözleşmesi** — her implementasyonun tek doğruluk kaynağı |
| [`CLAUDE.md`](./CLAUDE.md) | **Mimari** + "yeni dil nasıl eklenir" reçetesi (AI için) |
| [`.env.example`](./.env.example) | Kimlik bilgisi şablonu (tüm diller bunu okur) |
| [`go/`](./go) | **Referans** implementasyon (CLI demo) |

---

## ⚠️ Notlar

- **Kota:** Standart rolde her endpoint path'i için **günde 5 istek** sınırı vardır. Örnekler buna saygı duyar.
- **Kimlik bilgileri:** asla commit edilmez; sadece `.env.example` paylaşılır.
- **Kaynak:** Orijinal resmi C# projesi → <https://github.com/DinIsleriYuksekKurulu/AwqatSalah>
