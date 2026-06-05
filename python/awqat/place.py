"""Place (Ülke / İl / İlçe) endpoint'leri. İstemciyi ilk argüman alır.

Hiyerarşi: Country -> State (il) -> City (ilçe).
"""


def countries(c):
    return c.get_json("/api/Place/Countries")


def all_states(c):
    return c.get_json("/api/Place/States")


def states(c, country_id):
    return c.get_json(f"/api/Place/States/{country_id}")


def all_cities(c):
    return c.get_json("/api/Place/Cities")


def cities(c, state_id):
    return c.get_json(f"/api/Place/Cities/{state_id}")


def city_detail(c, city_id):
    return c.get_json(f"/api/Place/CityDetail/{city_id}")


# Türkçe karakterleri ASCII'ye indirgeme tablosu (arama için).
_FOLD = str.maketrans({
    "İ": "i", "I": "i", "ı": "i",
    "Ç": "c", "ç": "c", "Ğ": "g", "ğ": "g",
    "Ö": "o", "ö": "o", "Ş": "s", "ş": "s", "Ü": "u", "ü": "u",
})


def fold(s):
    return (s or "").strip().translate(_FOLD).lower()


def find_by_name(items, name):
    """Liste içinde adı (Türkçe + büyük/küçük harf duyarsız) eşleşen ilk öğeyi bulur."""
    target = fold(name)
    for p in items:
        if target in fold(p.get("name", "")):
            return p
    return None
