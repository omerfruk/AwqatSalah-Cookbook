namespace Awqat;

/// <summary>DailyContent: günün Ayet / Hadis / Dua içeriği (parametresiz).</summary>
public static class Content
{
    public static Task<DailyContent> DailyContentAsync(Client c) =>
        c.GetJsonAsync<DailyContent>("/api/DailyContent");
}
