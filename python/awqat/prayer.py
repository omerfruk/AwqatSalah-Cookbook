"""PrayerTime (Namaz Vakitleri) endpoint'leri. İstemciyi ilk argüman alır."""


def daily(c, city_id):
    return c.get_json(f"/api/PrayerTime/Daily/{city_id}")


def weekly(c, city_id):
    return c.get_json(f"/api/PrayerTime/Weekly/{city_id}")


def monthly(c, city_id):
    return c.get_json(f"/api/PrayerTime/Monthly/{city_id}")


def ramadan(c, city_id):
    return c.get_json(f"/api/PrayerTime/Ramadan/{city_id}")


def eid(c, city_id):
    return c.get_json(f"/api/PrayerTime/Eid/{city_id}")
