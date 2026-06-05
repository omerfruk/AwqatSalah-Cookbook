using System.Text.Json;

namespace Awqat;

public sealed class TokenCacheData
{
    public string AccessToken { get; set; } = "";
    public string RefreshToken { get; set; } = "";
    public string Expiry { get; set; } = "";
}

/// <summary>
/// Token önbelleği: login/refresh sonrası token'ı diske kalıcı yazar.
/// Sonraki çalıştırmalarda geçerliyse yeniden login ATILMAZ (kota korunur).
/// Biçim Go/JS/Python/PHP ile UYUMLU: {"accessToken","refreshToken","expiry" (ISO 8601)}.
/// </summary>
public static class TokenCache
{
    private static readonly JsonSerializerOptions Options = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
        PropertyNameCaseInsensitive = true,
        WriteIndented = true,
    };

    public static TokenCacheData? Load(string path)
    {
        if (string.IsNullOrEmpty(path) || !File.Exists(path)) return null;
        try
        {
            var tc = JsonSerializer.Deserialize<TokenCacheData>(File.ReadAllText(path), Options);
            return tc is not null && tc.AccessToken.Length > 0 ? tc : null;
        }
        catch
        {
            return null;
        }
    }

    /// <summary>Token'ı yalnızca sahibin okuyabileceği (0600) dosyaya yazar. Hata yutulur.</summary>
    public static void Save(string path, TokenCacheData tc)
    {
        if (string.IsNullOrEmpty(path)) return;
        try
        {
            File.WriteAllText(path, JsonSerializer.Serialize(tc, Options));
            if (!OperatingSystem.IsWindows())
            {
                File.SetUnixFileMode(path, UnixFileMode.UserRead | UnixFileMode.UserWrite);
            }
        }
        catch
        {
            // önbellek kritik değil
        }
    }
}
