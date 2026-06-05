namespace Awqat;

/// <summary>PrayerTime (Namaz Vakitleri) endpoint'leri. İstemciyi ilk argüman alır.</summary>
public static class Prayer
{
    public static Task<List<PrayerTime>> DailyAsync(Client c, long cityId) =>
        c.GetJsonAsync<List<PrayerTime>>($"/api/PrayerTime/Daily/{cityId}");

    public static Task<List<PrayerTime>> WeeklyAsync(Client c, long cityId) =>
        c.GetJsonAsync<List<PrayerTime>>($"/api/PrayerTime/Weekly/{cityId}");

    public static Task<List<PrayerTime>> MonthlyAsync(Client c, long cityId) =>
        c.GetJsonAsync<List<PrayerTime>>($"/api/PrayerTime/Monthly/{cityId}");

    public static Task<List<PrayerTime>> RamadanAsync(Client c, long cityId) =>
        c.GetJsonAsync<List<PrayerTime>>($"/api/PrayerTime/Ramadan/{cityId}");

    public static Task<EidPrayerTime> EidAsync(Client c, long cityId) =>
        c.GetJsonAsync<EidPrayerTime>($"/api/PrayerTime/Eid/{cityId}");
}
