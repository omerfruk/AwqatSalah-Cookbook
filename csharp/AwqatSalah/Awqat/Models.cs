namespace Awqat;

/// <summary>Diyanet API'sinin tüm yanıtlarını saran ortak zarf.</summary>
public sealed class ApiResponse<T>
{
    public bool Success { get; set; }
    public string? Message { get; set; }
    public T? Data { get; set; }
}

public sealed class Token
{
    public string AccessToken { get; set; } = "";
    public string RefreshToken { get; set; } = "";
}

/// <summary>Ülke, il (state) ve ilçe (city) için ortak yapı.</summary>
public sealed class Place
{
    public long Id { get; set; }
    public string Code { get; set; } = "";
    public string Name { get; set; } = "";
}

public sealed class DailyContent
{
    public long Id { get; set; }
    public long DayOfYear { get; set; }
    public string? Verse { get; set; }
    public string? VerseSource { get; set; }
    public string? Hadith { get; set; }
    public string? HadithSource { get; set; }
    public string? Pray { get; set; }
    public string? PraySource { get; set; }
}

/// <summary>İlçe detayı. DİKKAT: Id burada STRING'tir.</summary>
public sealed class CityDetail
{
    public string Id { get; set; } = "";
    public string? Name { get; set; }
    public string? GeographicQiblaAngle { get; set; }
    public string? DistanceToKaaba { get; set; }
    public string? QiblaAngle { get; set; }
    public string? City { get; set; }
    public string? Country { get; set; }
}

public sealed class PrayerTime
{
    public string? ShapeMoonUrl { get; set; }
    public string? Fajr { get; set; }
    public string? Sunrise { get; set; }
    public string? Dhuhr { get; set; }
    public string? Asr { get; set; }
    public string? Maghrib { get; set; }
    public string? Isha { get; set; }
    public string? QiblaTime { get; set; }
    public string? HijriDateShort { get; set; }
    public string? HijriDateLong { get; set; }
    public string? GregorianDateShort { get; set; }
    public string? GregorianDateLong { get; set; }
}

public sealed class EidPrayerTime
{
    public string? EidAlFitrHijri { get; set; }
    public string? EidAlFitrDate { get; set; }
    public string? EidAlFitrTime { get; set; }
    public string? EidAlAdhaHijri { get; set; }
    public string? EidAlAdhaDate { get; set; }
    public string? EidAlAdhaTime { get; set; }
}
