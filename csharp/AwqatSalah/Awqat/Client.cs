using System.Globalization;
using System.Text;
using System.Text.Json;

namespace Awqat;

/// <summary>
/// AwqatSalah API istemcisi — sıfır NuGet bağımlılığı (yerleşik HttpClient + System.Text.Json).
///
/// Kimlik doğrulama akışı (resmi PDF'e göre):
///  1. POST /api/Auth/Login (email+password) -> accessToken + refreshToken
///  2. Her istekte Authorization: Bearer &lt;accessToken&gt;
///  3. Süre dolmadan GET /api/Auth/RefreshToken/{refreshToken} ile yenilenir
///  4. Refresh de başarısızsa tekrar login olunur
/// </summary>
public sealed class Client
{
    private static readonly TimeSpan TokenLifetime = TimeSpan.FromMinutes(30); // PDF §3
    private static readonly TimeSpan RefreshSafety = TimeSpan.FromMinutes(5);
    private const int MaxPerPathPerDay = 5; // Standart rol kotası: path başına / gün

    /// <summary>Tüm çözümlemelerde kullanılan JSON ayarları (camelCase ↔ PascalCase).</summary>
    public static readonly JsonSerializerOptions Json = new() { PropertyNameCaseInsensitive = true };

    public string BaseUrl { get; }

    private readonly Config _cfg;
    private readonly HttpClient _http;
    private readonly Func<string, string, string?, string, Task<(int, string)>>? _sender;

    private string _accessToken = "";
    private string _refreshToken = "";
    private DateTimeOffset _expiry = DateTimeOffset.MinValue;
    private string _authPrefix = "/api"; // 404 alınırsa "" (eski örnekler) ile denenir
    private string _source = "";          // "env" | "cache" | "login" | "refresh" | ""
    private readonly Dictionary<string, (int N, string Date)> _rate = new();

    /// <param name="sender">Test için enjekte edilebilir gönderici (yoksa HttpClient kullanılır).</param>
    public Client(Config cfg, Func<string, string, string?, string, Task<(int, string)>>? sender = null)
    {
        _cfg = cfg;
        BaseUrl = cfg.BaseUrl;
        _http = new HttpClient { Timeout = TimeSpan.FromSeconds(30) };
        _sender = sender;

        // Mevcut token varsa yükle → geçerliyse login atılmaz (kota korunur).
        if (!string.IsNullOrEmpty(cfg.AccessToken))
        {
            _accessToken = cfg.AccessToken;
            _refreshToken = cfg.RefreshToken;
            _expiry = DateTimeOffset.UtcNow + (TokenLifetime - RefreshSafety);
            _source = "env";
        }
        else
        {
            var tc = TokenCache.Load(cfg.TokenCachePath);
            if (tc is not null)
            {
                _accessToken = tc.AccessToken;
                _refreshToken = tc.RefreshToken;
                _expiry = DateTimeOffset.TryParse(
                    tc.Expiry, CultureInfo.InvariantCulture, DateTimeStyles.RoundtripKind, out var e)
                    ? e : DateTimeOffset.MinValue;
                _source = "cache";
            }
        }
    }

    public string AccessToken => _accessToken;
    public string TokenSource => _source;

    // ---- Kimlik doğrulama --------------------------------------------------

    public async Task EnsureAuthAsync()
    {
        if (!string.IsNullOrEmpty(_accessToken) && DateTimeOffset.UtcNow < _expiry) return;
        if (!string.IsNullOrEmpty(_refreshToken))
        {
            try { await RefreshAsync(); return; }
            catch { /* refresh başarısız → login'e düş */ }
        }
        await LoginAsync();
    }

    private async Task LoginAsync()
    {
        var payload = JsonSerializer.Serialize(new { email = _cfg.Email, password = _cfg.Password });
        var (status, text) = await SendAsync("POST", _authPrefix + "/Auth/Login", payload, "");
        if (status == 404 && _authPrefix.Length > 0)
        {
            _authPrefix = ""; // eski yola geç ve hatırla
            (status, text) = await SendAsync("POST", "/Auth/Login", payload, "");
        }
        if (status != 200) throw new InvalidOperationException($"login başarısız (HTTP {status}): {text}");
        StoreToken(text, "login");
    }

    private async Task RefreshAsync()
    {
        if (string.IsNullOrEmpty(_refreshToken)) throw new InvalidOperationException("refresh token yok");
        var (status, text) = await SendAsync("GET", _authPrefix + "/Auth/RefreshToken/" + _refreshToken, null, "");
        if (status != 200) throw new InvalidOperationException($"refresh başarısız (HTTP {status})");
        StoreToken(text, "refresh");
    }

    private void StoreToken(string text, string op)
    {
        var token = Unwrap<Token>(text);
        _accessToken = token.AccessToken;
        _refreshToken = token.RefreshToken;
        _expiry = DateTimeOffset.UtcNow + (TokenLifetime - RefreshSafety);
        _source = op; // "login" veya "refresh"
        // Token'ı kalıcı yap → sonraki çalıştırmalar geçerliyse login atmaz.
        TokenCache.Save(_cfg.TokenCachePath, new TokenCacheData
        {
            AccessToken = _accessToken,
            RefreshToken = _refreshToken,
            Expiry = _expiry.ToString("o", CultureInfo.InvariantCulture),
        });
    }

    // ---- İstek katmanı (kota + auth + 401 retry) ---------------------------

    /// <summary>Kimlik doğrulamalı GET yapıp zarfı açarak tipli sonuç döndürür.</summary>
    public async Task<T> GetJsonAsync<T>(string path) => Unwrap<T>(await DoGetAsync(path));

    public Task<string> DoGetAsync(string path) => DoAsync("GET", path, null);

    public Task<string> DoPostAsync(string path, object body) =>
        DoAsync("POST", path, JsonSerializer.Serialize(body));

    private async Task<string> DoAsync(string method, string path, string? payload)
    {
        CheckRate(path);
        await EnsureAuthAsync();

        var (status, text) = await SendAsync(method, path, payload, _accessToken);

        // 401 → token geçersiz olmuş olabilir: sıfırla, yeniden doğrula, 1 kez tekrar dene.
        if (status == 401)
        {
            _accessToken = "";
            _refreshToken = "";
            await EnsureAuthAsync();
            (status, text) = await SendAsync(method, path, payload, _accessToken);
        }

        if (status != 200) throw new InvalidOperationException($"{method} {path} başarısız (HTTP {status}): {text}");
        return text;
    }

    private Task<(int, string)> SendAsync(string method, string path, string? payload, string token)
    {
        return _sender is not null
            ? _sender(method, path, payload, token)
            : HttpSendAsync(method, path, payload, token);
    }

    private async Task<(int, string)> HttpSendAsync(string method, string path, string? payload, string token)
    {
        var lastErr = "";
        // Ağ hatasında (reset / ölü bağlantı) kısa beklemeyle 3 kez dene.
        for (var attempt = 0; attempt < 3; attempt++)
        {
            try
            {
                using var req = new HttpRequestMessage(new HttpMethod(method), BaseUrl + path);
                if (payload is not null)
                {
                    req.Content = new StringContent(payload, Encoding.UTF8, "application/json");
                }
                if (token.Length > 0)
                {
                    req.Headers.TryAddWithoutValidation("Authorization", "Bearer " + token);
                }
                using var resp = await _http.SendAsync(req);
                var text = await resp.Content.ReadAsStringAsync();
                return ((int)resp.StatusCode, text);
            }
            catch (Exception e)
            {
                lastErr = e.Message;
                await Task.Delay(200 * (attempt + 1));
            }
        }
        throw new InvalidOperationException($"{method} {path}: {lastErr}");
    }

    // Path için günlük kotayı (5/gün) kontrol eder. Sayaç süreç-içidir.
    private void CheckRate(string path)
    {
        var today = DateTimeOffset.Now.ToString("yyyy-MM-dd", CultureInfo.InvariantCulture);
        if (!_rate.TryGetValue(path, out var cnt) || cnt.Date != today)
        {
            _rate[path] = (1, today);
            return;
        }
        if (cnt.N >= MaxPerPathPerDay)
        {
            throw new InvalidOperationException(
                $"kota: {path} için bugün {cnt.N}/{MaxPerPathPerDay} istek kullanıldı");
        }
        _rate[path] = (cnt.N + 1, today);
    }

    /// <summary>Ortak zarfı açar, success kontrol eder, data döndürür.</summary>
    public static T Unwrap<T>(string text)
    {
        ApiResponse<T>? resp;
        try
        {
            resp = JsonSerializer.Deserialize<ApiResponse<T>>(text, Json);
        }
        catch (Exception e)
        {
            throw new InvalidOperationException($"yanıt çözümlenemedi: {e.Message}");
        }
        if (resp is null) throw new InvalidOperationException("yanıt çözümlenemedi");
        if (!resp.Success) throw new InvalidOperationException($"API başarısız: {resp.Message ?? "bilinmeyen hata"}");
        return resp.Data!;
    }
}
