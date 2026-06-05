// AwqatSalah Cookbook — C# örneği.
//
// Çalıştırma:  cd csharp/AwqatSalah && dotnet run            (tüm endpoint turu)
//              cd csharp/AwqatSalah && dotnet run -- --login  (sadece token al/yenile)
//
// Resmi PDF'teki TÜM endpoint'leri sırayla çağırır (Isparta baz alınarak).
// Tüm ayarlar kök .env / ortam değişkenlerinden okunur (bkz. ../../.env.example).

using Awqat;

var loginOnly = args.Contains("--login") || args.Contains("-login");

try
{
    await RunDemo(loginOnly);
}
catch (Exception e)
{
    await Console.Error.WriteLineAsync($"\n❌ Hata: {e.Message}");
    Environment.Exit(1);
}

static async Task RunDemo(bool loginOnly)
{
    var cfg = Config.Load();
    var c = new Client(cfg);

    Console.WriteLine("====================================================");
    Console.WriteLine("  AwqatSalah Cookbook — C# (tüm endpoint turu)");
    Console.WriteLine($"  Base: {cfg.BaseUrl}");
    Console.WriteLine("  Not: her endpoint 1 kez çağrılır · kota: path başına 5/gün");
    Console.WriteLine("====================================================");

    // [1] Kimlik (token)
    Sec(1, "Kimlik (token)", "/api/Auth/Login");
    await c.EnsureAuthAsync();
    var tok = Short(c.AccessToken);
    Console.WriteLine(c.TokenSource switch
    {
        "cache" => $"   ✓ kayıtlı token kullanıldı, login ATILMADI (kota korundu): {tok}…",
        "env" => $"   ✓ .env token'ı kullanıldı, login ATILMADI: {tok}…",
        "refresh" => $"   ✓ token yenilendi (refresh): {tok}…",
        _ => $"   ✓ login yapıldı, token .awqat-token.json'a kaydedildi: {tok}…",
    });
    if (loginOnly)
    {
        Console.WriteLine("\n(--login) Sadece kimlik adımı çalıştırıldı; endpoint turu atlandı.");
        return;
    }

    // [2] Günün Ayet/Hadis/Dua
    Sec(2, "Günlük İçerik (Ayet/Hadis/Dua)", "/api/DailyContent");
    try
    {
        var dc = await Content.DailyContentAsync(c);
        Console.WriteLine($"   Ayet : {Truncate(dc.Verse, 90)}");
        Console.WriteLine($"   Hadis: {Truncate(dc.Hadith, 90)}");
        Console.WriteLine($"   Dua  : {Truncate(dc.Pray, 90)}");
    }
    catch (Exception e) { Warn("dailyContent", e); }

    // [3] Ülkeler
    Sec(3, "Ülkeler", "/api/Place/Countries");
    var country = new Place { Id = cfg.CountryId, Name = cfg.Country };
    try
    {
        var list = await Places.CountriesAsync(c);
        Console.WriteLine($"   {list.Count} ülke alındı.");
        var f = Places.FindByName(list, cfg.Country);
        if (f is not null) country = f;
    }
    catch (Exception e) { Warn("countries", e); }
    Console.WriteLine($"   → seçilen ülke: {country.Name} (id={country.Id})");
    if (country.Id == 0) throw new InvalidOperationException($"ülke çözümlenemedi (AWQAT_COUNTRY={cfg.Country})");

    // [4] Tüm iller
    Sec(4, "Tüm iller (parametresiz)", "/api/Place/States");
    try { var all = await Places.AllStatesAsync(c); Console.WriteLine($"   {all.Count} il (tüm ülkeler) alındı."); }
    catch (Exception e) { Warn("allStates", e); }

    // [5] Ülkeye göre iller
    Sec(5, "Ülkeye göre iller", "/api/Place/States/{countryId}");
    var state = new Place { Id = cfg.StateId, Name = cfg.State };
    try
    {
        var list = await Places.StatesAsync(c, country.Id);
        Console.WriteLine($"   {country.Name} için {list.Count} il alındı.");
        var f = Places.FindByName(list, cfg.State);
        if (f is not null) state = f;
    }
    catch (Exception e) { Warn("states", e); }
    Console.WriteLine($"   → seçilen il: {state.Name} (id={state.Id})");
    if (state.Id == 0) throw new InvalidOperationException($"il çözümlenemedi (AWQAT_STATE={cfg.State})");

    // [6] Tüm ilçeler
    Sec(6, "Tüm ilçeler (parametresiz)", "/api/Place/Cities");
    try { var all = await Places.AllCitiesAsync(c); Console.WriteLine($"   {all.Count} ilçe (tüm iller) alındı."); }
    catch (Exception e) { Warn("allCities", e); }

    // [7] İle göre ilçeler
    Sec(7, "İle göre ilçeler", "/api/Place/Cities/{stateId}");
    var city = new Place { Id = cfg.CityId, Name = "(AWQAT_CITY_ID)" };
    try
    {
        var list = await Places.CitiesAsync(c, state.Id);
        Console.WriteLine($"   {state.Name} için {list.Count} ilçe alındı.");
        city = PickCity(list, cfg);
    }
    catch (Exception e) { Warn("cities", e); }
    Console.WriteLine($"   → seçilen ilçe: {city.Name} (id={city.Id})");
    if (city.Id == 0) throw new InvalidOperationException("ilçe çözümlenemedi");

    // [8] İlçe detay (kıble)
    Sec(8, "İlçe detay (kıble açısı)", "/api/Place/CityDetail/{cityId}");
    try
    {
        var d = await Places.CityDetailAsync(c, city.Id);
        Console.WriteLine($"   {d.City} / {d.Country} · kıble açısı: {d.QiblaAngle}° · Kâbe'ye uzaklık: {d.DistanceToKaaba} km");
    }
    catch (Exception e) { Warn("cityDetail", e); }

    // [9] Günlük namaz vakitleri
    Sec(9, "Günlük namaz vakitleri", "/api/PrayerTime/Daily/{cityId}");
    try
    {
        var times = await Prayer.DailyAsync(c, city.Id);
        if (times.Count > 0) PrintPrayer(city, times[0]);
    }
    catch (Exception e) { Warn("daily", e); }

    // [10] Haftalık
    Sec(10, "Haftalık namaz vakitleri", "/api/PrayerTime/Weekly/{cityId}");
    try { var t = await Prayer.WeeklyAsync(c, city.Id); Console.WriteLine($"   {t.Count} günlük veri ({FirstDate(t)} … {LastDate(t)})"); }
    catch (Exception e) { Warn("weekly", e); }

    // [11] Aylık
    Sec(11, "Aylık namaz vakitleri", "/api/PrayerTime/Monthly/{cityId}");
    try { var t = await Prayer.MonthlyAsync(c, city.Id); Console.WriteLine($"   {t.Count} günlük veri ({FirstDate(t)} … {LastDate(t)})"); }
    catch (Exception e) { Warn("monthly", e); }

    // [12] Bayram namazı
    Sec(12, "Bayram namazı", "/api/PrayerTime/Eid/{cityId}");
    try
    {
        var ev = await Prayer.EidAsync(c, city.Id);
        Console.WriteLine($"   Ramazan B.: {ev.EidAlFitrDate} {ev.EidAlFitrTime} · Kurban B.: {ev.EidAlAdhaDate} {ev.EidAlAdhaTime}");
    }
    catch (Exception e) { Warn("eid", e); }

    // [13] Ramazan imsakiyesi
    Sec(13, "Ramazan imsakiyesi", "/api/PrayerTime/Ramadan/{cityId}");
    try { var t = await Prayer.RamadanAsync(c, city.Id); Console.WriteLine($"   {t.Count} günlük imsakiye verisi"); }
    catch (Exception e) { Warn("ramadan", e); }

    Console.WriteLine("\n====================================================");
    Console.WriteLine("✅ Tüm endpoint'ler çağrıldı.");
    Console.WriteLine("====================================================");
}

static void Sec(int n, string title, string path) => Console.WriteLine($"\n[{n}] {title}\n     {path}");

static void Warn(string label, Exception e) => Console.WriteLine($"   ⚠ {label} atlandı: {e.Message}");

static string Short(string s) => s.Length <= 16 ? s : s[..16];

static string Truncate(string? s, int n)
{
    s = (s ?? "").Trim();
    return s.Length <= n ? s : s[..n] + "…";
}

static string FirstDate(List<PrayerTime> t) => t.Count > 0 ? (t[0].GregorianDateShort ?? "-") : "-";

static string LastDate(List<PrayerTime> t) => t.Count > 0 ? (t[^1].GregorianDateShort ?? "-") : "-";

static Place PickCity(List<Place> items, Config cfg)
{
    if (cfg.CityId != 0)
    {
        return items.FirstOrDefault(x => x.Id == cfg.CityId) ?? new Place { Id = cfg.CityId, Name = "(AWQAT_CITY_ID)" };
    }
    var f = Places.FindByName(items, cfg.State);
    if (f is not null) return f;
    return items.Count > 0 ? items[0] : new Place();
}

static void PrintPrayer(Place city, PrayerTime t)
{
    Console.WriteLine($"   🕌 {city.Name} — {t.GregorianDateShort} (Hicri: {t.HijriDateShort})");
    var rows = new (string Label, string? Val)[]
    {
        ("İmsak  (Fajr)", t.Fajr),
        ("Güneş  (Sunrise)", t.Sunrise),
        ("Öğle   (Dhuhr)", t.Dhuhr),
        ("İkindi (Asr)", t.Asr),
        ("Akşam  (Maghrib)", t.Maghrib),
        ("Yatsı  (Isha)", t.Isha),
    };
    foreach (var (label, val) in rows)
    {
        Console.WriteLine($"      {label,-18} {val}");
    }
}
