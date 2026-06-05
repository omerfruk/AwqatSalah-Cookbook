// Enjekte (sahte) gönderici ile testler — gerçek API'ye gerek yok, test çerçevesi gerekmez.
// Çalıştırma:  cd csharp/AwqatSalah.Tests && dotnet run

using Awqat;

int total = 0, failed = 0;

void Check(string name, bool cond)
{
    total++;
    if (cond) { Console.WriteLine($"  ok   {name}"); }
    else { Console.WriteLine($"  FAIL {name}"); failed++; }
}

Config MakeCfg(Action<Config>? over = null)
{
    var c = new Config
    {
        BaseUrl = "http://mock",
        Email = "a",
        Password = "b",
        Country = "Türkiye",
        State = "Isparta",
        TokenCachePath = "",
    };
    over?.Invoke(c);
    return c;
}

// Sahte gönderici: path -> (status, body). İstek sayısını ve son token'ı kaydeder.
Func<string, string, string?, string, Task<(int, string)>> MakeSender(
    Dictionary<string, (int, string)> routes, Dictionary<string, int> calls, Action<string>? captureToken = null)
{
    return (method, path, payload, token) =>
    {
        calls[path] = calls.GetValueOrDefault(path) + 1;
        captureToken?.Invoke(token);
        return Task.FromResult(routes.TryGetValue(path, out var r) ? r : (404, "not found"));
    };
}

// --- Test 1: uçtan uca ----------------------------------------------------
{
    var calls = new Dictionary<string, int>();
    var lastToken = "";
    var routes = new Dictionary<string, (int, string)>
    {
        ["/api/Auth/Login"] = (200, "{\"success\":true,\"data\":{\"accessToken\":\"ACCESS\",\"refreshToken\":\"R\"}}"),
        ["/api/DailyContent"] = (200, "{\"success\":true,\"data\":{\"verse\":\"V\",\"hadith\":\"H\",\"pray\":\"P\"}}"),
        ["/api/Place/Countries"] = (200, "{\"success\":true,\"data\":[{\"id\":1,\"name\":\"KUZEY KIBRIS\"},{\"id\":2,\"name\":\"TÜRKİYE\"}]}"),
        ["/api/Place/States/2"] = (200, "{\"success\":true,\"data\":[{\"id\":538,\"name\":\"ISPARTA\"}]}"),
        ["/api/Place/Cities/538"] = (200, "{\"success\":true,\"data\":[{\"id\":9528,\"name\":\"ISPARTA\"}]}"),
        ["/api/Place/CityDetail/9528"] = (200, "{\"success\":true,\"data\":{\"id\":\"9528\",\"qiblaAngle\":\"151\",\"distanceToKaaba\":\"2023\",\"city\":\"ISPARTA\",\"country\":\"TÜRKİYE\"}}"),
        ["/api/PrayerTime/Daily/9528"] = (200, "{\"success\":true,\"data\":[{\"fajr\":\"03:44\",\"isha\":\"22:01\"}]}"),
    };
    var c = new Client(MakeCfg(), MakeSender(routes, calls, t => lastToken = t));
    await c.EnsureAuthAsync();
    Check("login token", c.AccessToken == "ACCESS");

    var dc = await Content.DailyContentAsync(c);
    Check("dailyContent", dc.Verse == "V");

    var country = Places.FindByName(await Places.CountriesAsync(c), "Türkiye"); // "TÜRKİYE" ile eşleşmeli
    Check("ülke (Türkçe duyarsız)", country is { Id: 2 });
    Check("Bearer token", lastToken == "ACCESS");

    var state = Places.FindByName(await Places.StatesAsync(c, country!.Id), "Isparta");
    Check("il", state is { Id: 538 });

    var cities = await Places.CitiesAsync(c, state!.Id);
    Check("ilçe", cities[0].Id == 9528);

    var d = await Places.CityDetailAsync(c, cities[0].Id);
    Check("cityDetail id STRING", d.Id == "9528");
    Check("cityDetail qibla", d.QiblaAngle == "151");

    var times = await Prayer.DailyAsync(c, cities[0].Id);
    Check("namaz vakti", times[0].Fajr == "03:44");
}

// --- Test 2: geçerli token → login atılmaz --------------------------------
{
    var calls = new Dictionary<string, int>();
    var lastToken = "";
    var routes = new Dictionary<string, (int, string)>
    {
        ["/api/Auth/Login"] = (200, "{\"success\":true,\"data\":{\"accessToken\":\"NEW\"}}"),
        ["/api/Place/Countries"] = (200, "{\"success\":true,\"data\":[]}"),
    };
    var c = new Client(MakeCfg(cfg => cfg.AccessToken = "SEED"), MakeSender(routes, calls, t => lastToken = t));
    await c.EnsureAuthAsync();
    await Places.CountriesAsync(c);
    Check("login İSTEĞİ atılmadı", calls.GetValueOrDefault("/api/Auth/Login") == 0);
    Check("source=env", c.TokenSource == "env");
    Check("seed token kullanıldı", lastToken == "SEED");
}

// --- Test 3: /api/Auth/Login 404 → /Auth/Login fallback -------------------
{
    var calls = new Dictionary<string, int>();
    var routes = new Dictionary<string, (int, string)>
    {
        ["/api/Auth/Login"] = (404, "nf"),
        ["/Auth/Login"] = (200, "{\"success\":true,\"data\":{\"accessToken\":\"OK\"}}"),
    };
    var c = new Client(MakeCfg(), MakeSender(routes, calls));
    await c.EnsureAuthAsync();
    Check("auth prefix fallback", c.AccessToken == "OK");
}

// --- Test 4: unwrap success=false → hata ----------------------------------
{
    var threw = false;
    try { Client.Unwrap<List<Place>>("{\"success\":false,\"message\":\"yetkisiz\"}"); }
    catch { threw = true; }
    Check("unwrap success=false hata", threw);
}

// --- Test 5: fold Türkçe-duyarsız -----------------------------------------
Check("fold TÜRKİYE", Places.Fold("TÜRKİYE") == "turkiye");
Check("fold Isparta", Places.Fold("Isparta") == "isparta");

Console.WriteLine($"\n{total} test, {failed} başarısız");
Environment.Exit(failed == 0 ? 0 : 1);
