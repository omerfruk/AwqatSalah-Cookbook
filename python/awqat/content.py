"""DailyContent: günün Ayet / Hadis / Dua içeriği (parametresiz)."""


def daily_content(c):
    return c.get_json("/api/DailyContent")
