using System.Text;

namespace Awqat;

/// <summary>
/// Place (Ülke / İl / İlçe) endpoint'leri. İstemciyi ilk argüman alır.
/// Hiyerarşi: Country -> State (il) -> City (ilçe).
/// </summary>
public static class Places
{
    public static Task<List<Place>> CountriesAsync(Client c) =>
        c.GetJsonAsync<List<Place>>("/api/Place/Countries");

    public static Task<List<Place>> AllStatesAsync(Client c) =>
        c.GetJsonAsync<List<Place>>("/api/Place/States");

    public static Task<List<Place>> StatesAsync(Client c, long countryId) =>
        c.GetJsonAsync<List<Place>>($"/api/Place/States/{countryId}");

    public static Task<List<Place>> AllCitiesAsync(Client c) =>
        c.GetJsonAsync<List<Place>>("/api/Place/Cities");

    public static Task<List<Place>> CitiesAsync(Client c, long stateId) =>
        c.GetJsonAsync<List<Place>>($"/api/Place/Cities/{stateId}");

    public static Task<CityDetail> CityDetailAsync(Client c, long cityId) =>
        c.GetJsonAsync<CityDetail>($"/api/Place/CityDetail/{cityId}");

    /// <summary>Liste içinde adı (Türkçe + büyük/küçük harf duyarsız) eşleşen ilk öğeyi bulur.</summary>
    public static Place? FindByName(IEnumerable<Place> items, string name)
    {
        var target = Fold(name);
        foreach (var p in items)
        {
            if (Fold(p.Name).Contains(target, StringComparison.Ordinal)) return p;
        }
        return null;
    }

    /// <summary>Türkçe karakterleri ASCII'ye indirger ve küçük harfe çevirir (arama için).</summary>
    public static string Fold(string s)
    {
        s = (s ?? "").Trim();
        var sb = new StringBuilder(s.Length);
        foreach (var ch in s)
        {
            sb.Append(ch switch
            {
                'İ' or 'I' or 'ı' => 'i',
                'Ç' or 'ç' => 'c',
                'Ğ' or 'ğ' => 'g',
                'Ö' or 'ö' => 'o',
                'Ş' or 'ş' => 's',
                'Ü' or 'ü' => 'u',
                _ => ch,
            });
        }
        return sb.ToString().ToLowerInvariant();
    }
}
