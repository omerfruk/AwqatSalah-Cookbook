namespace Awqat;

/// <summary>
/// Yapılandırma: ortam değişkenleri + kök .env (yukarı doğru aranır).
/// </summary>
public sealed class Config
{
    public string BaseUrl { get; set; } = "";
    public string Email { get; set; } = "";
    public string Password { get; set; } = "";
    public string Country { get; set; } = "";
    public string State { get; set; } = "";
    public long CountryId { get; set; }
    public long StateId { get; set; }
    public long CityId { get; set; }
    public string AccessToken { get; set; } = "";
    public string RefreshToken { get; set; } = "";
    public string TokenCachePath { get; set; } = "";

    // Token önbelleği dosya adı (kök .env yanında). Tüm diller AYNI dosyayı paylaşır.
    public const string TokenCacheFile = ".awqat-token.json";

    public static Config Load()
    {
        var (dotenv, rootDir) = LoadDotEnvUpwards();

        // Get: önce OS ortam değişkeni, sonra .env, sonra varsayılan.
        string Get(string key, string def = "")
        {
            var env = Environment.GetEnvironmentVariable(key)?.Trim();
            if (!string.IsNullOrEmpty(env)) return env;
            if (dotenv.TryGetValue(key, out var v) && !string.IsNullOrWhiteSpace(v)) return v.Trim();
            return def;
        }
        long Num(string s) => long.TryParse(s, out var n) ? n : 0;

        var cfg = new Config
        {
            BaseUrl = Get("AWQAT_BASE_URL", "https://awqatsalah.diyanet.gov.tr").TrimEnd('/'),
            Email = Get("AWQAT_EMAIL"),
            Password = Get("AWQAT_PASSWORD"),
            Country = Get("AWQAT_COUNTRY", "Türkiye"),
            State = Get("AWQAT_STATE", "Isparta"),
            CountryId = Num(Get("AWQAT_COUNTRY_ID")),
            StateId = Num(Get("AWQAT_STATE_ID")),
            CityId = Num(Get("AWQAT_CITY_ID")),
            AccessToken = Get("AWQAT_ACCESS_TOKEN"),
            RefreshToken = Get("AWQAT_REFRESH_TOKEN"),
            TokenCachePath = Path.Combine(rootDir, TokenCacheFile),
        };

        if (string.IsNullOrEmpty(cfg.Email) || string.IsNullOrEmpty(cfg.Password))
        {
            throw new InvalidOperationException(
                "AWQAT_EMAIL ve AWQAT_PASSWORD gerekli — kök dizinde .env oluşturup doldurun:\n" +
                "    cp .env.example .env   (sonra e-posta/şifrenizi yazın)");
        }
        return cfg;
    }

    // CWD'den köke kadar .env arar; (değerler, kök_klasör) döndürür.
    private static (Dictionary<string, string> Values, string RootDir) LoadDotEnvUpwards()
    {
        var dir = Directory.GetCurrentDirectory();
        var start = dir;
        while (true)
        {
            var file = Path.Combine(dir, ".env");
            if (File.Exists(file))
            {
                return (ParseDotEnv(File.ReadAllText(file)), dir);
            }
            var parent = Directory.GetParent(dir)?.FullName;
            if (parent is null || parent == dir)
            {
                return (new Dictionary<string, string>(), start); // .env yok → CWD kök
            }
            dir = parent;
        }
    }

    // Basit KEY=VALUE ayrıştırıcı (# yorum, boş satır, tırnak soyma).
    private static Dictionary<string, string> ParseDotEnv(string text)
    {
        var result = new Dictionary<string, string>();
        foreach (var raw in text.Split('\n'))
        {
            var line = raw.Trim();
            if (line.Length == 0 || line[0] == '#') continue;
            if (line.StartsWith("export ")) line = line[7..];
            var eq = line.IndexOf('=');
            if (eq < 0) continue;
            var key = line[..eq].Trim();
            var val = line[(eq + 1)..].Trim();
            if (val.Length >= 2 &&
                ((val[0] == '"' && val[^1] == '"') || (val[0] == '\'' && val[^1] == '\'')))
            {
                val = val[1..^1];
            }
            if (key.Length > 0) result[key] = val;
        }
        return result;
    }
}
