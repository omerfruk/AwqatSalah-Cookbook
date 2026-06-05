# CLAUDE.md — Mimari & AI Orkestrasyon Rehberi

Bu repo bir **AwqatSalah API Cookbook**'udur: aynı Diyanet Namaz Vakti API'sini **birden çok
programlama dilinde**, sade ve çalışır örneklerle gösterir. Kullanıcı tek bir dil adı söyler
("şimdi JS için yap") ve sen bu rehberi izleyerek o dili eklersin.

---

## 🎯 Amaç

- Kullanıcı kimlik bilgilerini **bir kez** (kök `.env`) girer, **her dildeki** örnek çalışır.
- Her dil; **login → il/ilçe çekme → namaz vakitleri** üç temel akışını tek komutla gösterir.
- Baz örnek konum: **Türkiye / Isparta**.
- Her dil hem **insan** (README.md) hem **AI** (AI.md) için belgelenir.

## 📚 Tek doğruluk kaynağı

API'nin tüm detayları **[SPEC.md](./SPEC.md)** içindedir (endpoint'ler, auth, DTO'lar, akışlar).
Herhangi bir dil eklerken önce SPEC.md'yi oku. **`go/` klasörü kanonik referans implementasyondur** —
yeni diller onu birebir taklit eder.

---

## 🗂️ Repo Yapısı

```
AwqatSalah-Cookbook/
├── README.md          # İnsan: genel bakış, dil tablosu, hızlı başlangıç
├── CLAUDE.md          # (bu dosya) Mimari + yeni dil ekleme reçetesi
├── SPEC.md            # Dilden bağımsız API sözleşmesi — TEK DOĞRULUK KAYNAĞI
├── .env.example       # Kimlik bilgisi şablonu (tüm diller için TEK dosya)
├── .env               # Kullanıcının gerçek bilgileri (gitignore'lu, COMMIT EDİLMEZ)
├── .gitignore
│
├── go/                # ✅ Referans implementasyon (CLI demo)
│   ├── README.md      #    İnsan: nasıl çalıştırılır
│   ├── AI.md          #    AI: "Go'da implemente et" promptu
│   ├── go.mod
│   ├── main.go        #    go run . → TÜM endpoint turu (A→B→C + diğerleri)
│   └── awqat/         #    Yeniden kullanılabilir istemci paketi
│       ├── config.go  #      .env + ortam değişkeni yükleme
│       ├── models.go  #      DTO'lar (zarf, token, place, dailyContent, cityDetail, prayer, eid)
│       ├── client.go  #      HTTP istemci: login(/api/Auth), refresh, auth, 401 retry, kota
│       ├── place.go   #      Countries / (All)States / (All)Cities / CityDetail
│       ├── content.go #      DailyContent (günün ayet/hadis/dua)
│       └── prayer.go  #      Daily / Weekly / Monthly / Ramadan / Eid
│
├── js/                # ✅ Node (sıfır bağımlılık, yerleşik fetch) — go/ ile aynı desen
├── python/            # ✅ Python (sıfır bağımlılık, stdlib urllib) — go/ ile aynı desen
├── php/               # ✅ PHP (sıfır bağımlılık, yerleşik curl) — go/ ile aynı desen
├── csharp/            # ✅ C# (.NET, sıfır NuGet, HttpClient+System.Text.Json) — go/ ile aynı desen
└── ...                # ⏳ sonraki diller (rust, java, kotlin, ...) aynı deseni izler
```

---

## 📐 Her Dilin Uyması Gereken Konvansiyonlar

1. **Klasör:** dil adı (`go/`, `js/`, `python/`, `php/`, `rust/` ...).
2. **Tek komutla çalışır demo:** `go run .`, `node index.js`, `python main.py` gibi —
   login → il/ilçe (Türkiye/Isparta) → günlük namaz vakitlerini çekip terminale **okunaklı** yazdırır.
   Go referansı ayrıca **SPEC.md §4'teki tüm endpoint'leri** birer kez çağıran bir tur sunar
   (DailyContent, AllStates/AllCities, CityDetail, Weekly/Monthly/Eid/Ramadan dâhil) — yeni diller de bunu hedeflemeli.
3. **Config:** ayarları SPEC.md §6'daki sırayla okur — **ortam değişkenleri → kök `.env`**.
   `.env` çalışılan klasörden **yukarı doğru aranarak** bulunur (alt klasörden de çalışsın diye).
   Asla yeni bir env değişken adı uydurma; `AWQAT_*` sözlüğünü kullan.
4. **Yeniden kullanılabilir istemci:** API mantığı, demo'dan ayrı bir modül/paket/sınıfta olur
   (`awqat` paketi gibi). Demo sadece onu çağırır.
5. **Bağımlılık minimum:** mümkünse sadece standart kütüphane (Go örneği sıfır bağımlılıktır).
   Zorunluysa minik ve yaygın bir paket kullan, README'de belirt.
6. **Dokümanlar:** her dil klasöründe **`README.md`** (insan) + **`AI.md`** (AI promptu) bulunur.
7. **Sadelik:** kullanıcının önceliği — *sade, kullanışlı, basit*. Gereksiz soyutlama yok.

---

## 🤖 Yeni Bir Dil Ekleme Reçetesi (örn. "JS için yap")

> Tetikleyici: kullanıcı bir dil adı söyler. Adımlar:

1. **Oku:** `SPEC.md` (API sözleşmesi) + `go/` referansı (özellikle `main.go` ve `awqat/`).
2. **Klasör aç:** `<dil>/` (örn. `js/`).
3. **İstemciyi yaz** (SPEC.md'yi taklit ederek, `go/awqat/`'ın eşdeğeri):
   - config yükleme (env + yukarı doğru `.env` arama)
   - HTTP istemci: login, refresh, `Authorization: Bearer`, 401 retry, kota sayacı
   - DTO'lar / parse
   - place & prayer servis fonksiyonları
4. **Demo yaz:** tek komutla A→B→C akışı, Isparta baz, okunaklı çıktı (Go demo'nun çıktısını taklit et).
5. **`<dil>/README.md`:** nasıl çalıştırılır, ne yazdırır, dosya düzeni, kota notu.
6. **`<dil>/AI.md`:** o dile özel kopyala-yapıştır AI promptu (Go'daki AI.md'yi şablon al).
7. **Kök `README.md`'deki dil tablosunu güncelle** (durumu ✅ yap).
8. **Doğrula:** mümkünse derle/çalıştır (`.env` yoksa creds hatası beklenir — bu normaldir;
   en azından derlenmeli/syntax temiz olmalı).

### Kontrol listesi (bitince hepsi ✅ olmalı)
- [ ] `<dil>/` tek komutla çalışıyor, A→B→C akışını gösteriyor.
- [ ] Kök `.env` + `AWQAT_*` ortam değişkenlerini okuyor.
- [ ] SPEC.md §8 hata-önleme listesine uyuyor (login `/Auth`, diğerleri `/api`, zarf, 401 retry...).
- [ ] `README.md` + `AI.md` yazıldı.
- [ ] Kök README dil tablosu güncellendi.

---

## 🚫 Yapılmaması Gerekenler
- `.env`'i **commit etme** (gerçek kimlik bilgileri). Sadece `.env.example` commit edilir.
- API'yi gereksiz dövme: aynı path'e **günde 5 istekten** fazla vurma (SPEC.md §1 kota).
- Env değişken adı uydurma — `AWQAT_*` sözlüğüne sadık kal.
- Aşırı mühendislik — kullanıcı sadelik istiyor.

---

## 🔗 Kaynak
Orijinal resmi C# implementasyonu: <https://github.com/DinIsleriYuksekKurulu/AwqatSalah>.
Bu cookbook o API'yi temel alır; endpoint ve davranışlar SPEC.md'de damıtılmıştır.
