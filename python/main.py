#!/usr/bin/env python3
"""AwqatSalah Cookbook — Python örneği.

Çalıştırma:  cd python && python3 main.py            (tüm endpoint turu)
             cd python && python3 main.py --login   (sadece token al/yenile, kotayı korur)

Resmi PDF'teki TÜM endpoint'leri sırayla çağırır (Isparta baz alınarak).
Tüm ayarlar kök .env / ortam değişkenlerinden okunur (bkz. ../.env.example).
"""

import sys

from awqat import content, place, prayer
from awqat.client import AwqatClient
from awqat.config import load_config

LOGIN_ONLY = "--login" in sys.argv or "-login" in sys.argv


def sec(n, title, path):
    print(f"\n[{n}] {title}\n     {path}")


def warn(label, err):
    print(f"   ⚠ {label} atlandı: {err}")


def short(s):
    return (s or "")[:16]


def truncate(s, n):
    s = (s or "").strip()
    return s if len(s) <= n else s[:n] + "…"


def first_date(t):
    return t[0]["gregorianDateShort"] if t else "-"


def last_date(t):
    return t[-1]["gregorianDateShort"] if t else "-"


def pick_city(items, cfg):
    """AWQAT_CITY_ID varsa o, yoksa il adıyla eşleşen (merkez), yoksa ilk ilçe."""
    if cfg.city_id:
        for c in items:
            if c["id"] == cfg.city_id:
                return c
        return {"id": cfg.city_id, "name": "(AWQAT_CITY_ID)"}
    f = place.find_by_name(items, cfg.state)
    if f:
        return f
    return items[0] if items else {"id": 0, "name": ""}


def print_prayer(city, t):
    print(f"   🕌 {city['name']} — {t.get('gregorianDateShort', '')} (Hicri: {t.get('hijriDateShort', '')})")
    rows = [
        ("İmsak  (Fajr)", t.get("fajr", "")),
        ("Güneş  (Sunrise)", t.get("sunrise", "")),
        ("Öğle   (Dhuhr)", t.get("dhuhr", "")),
        ("İkindi (Asr)", t.get("asr", "")),
        ("Akşam  (Maghrib)", t.get("maghrib", "")),
        ("Yatsı  (Isha)", t.get("isha", "")),
    ]
    for label, val in rows:
        print(f"      {label:<18} {val}")


def main():
    cfg = load_config()
    c = AwqatClient(cfg)

    print("====================================================")
    print("  AwqatSalah Cookbook — Python (tüm endpoint turu)")
    print(f"  Base: {cfg.base_url}")
    print("  Not: her endpoint 1 kez çağrılır · kota: path başına 5/gün")
    print("====================================================")

    # [1] Kimlik (token)
    sec(1, "Kimlik (token)", "/api/Auth/Login")
    c.ensure_auth()
    tok = short(c.access_token)
    src = c.token_source()
    if src == "cache":
        print(f"   ✓ kayıtlı token kullanıldı, login ATILMADI (kota korundu): {tok}…")
    elif src == "env":
        print(f"   ✓ .env token'ı kullanıldı, login ATILMADI: {tok}…")
    elif src == "refresh":
        print(f"   ✓ token yenilendi (refresh): {tok}…")
    else:
        print(f"   ✓ login yapıldı, token .awqat-token.json'a kaydedildi: {tok}…")
    if LOGIN_ONLY:
        print("\n(--login) Sadece kimlik adımı çalıştırıldı; endpoint turu atlandı.")
        return

    # [2] Günün Ayet/Hadis/Dua
    sec(2, "Günlük İçerik (Ayet/Hadis/Dua)", "/api/DailyContent")
    try:
        dc = content.daily_content(c)
        print(f"   Ayet : {truncate(dc.get('verse'), 90)}")
        print(f"   Hadis: {truncate(dc.get('hadith'), 90)}")
        print(f"   Dua  : {truncate(dc.get('pray'), 90)}")
    except Exception as e:
        warn("dailyContent", e)

    # [3] Ülkeler
    sec(3, "Ülkeler", "/api/Place/Countries")
    country = {"id": cfg.country_id, "name": cfg.country}
    try:
        lst = place.countries(c)
        print(f"   {len(lst)} ülke alındı.")
        found = place.find_by_name(lst, cfg.country)
        if found:
            country = found
    except Exception as e:
        warn("countries", e)
    print(f"   → seçilen ülke: {country['name']} (id={country['id']})")
    if not country["id"]:
        raise RuntimeError(f"ülke çözümlenemedi (AWQAT_COUNTRY={cfg.country})")

    # [4] Tüm iller
    sec(4, "Tüm iller (parametresiz)", "/api/Place/States")
    try:
        allst = place.all_states(c)
        print(f"   {len(allst)} il (tüm ülkeler) alındı.")
    except Exception as e:
        warn("allStates", e)

    # [5] Ülkeye göre iller
    sec(5, "Ülkeye göre iller", "/api/Place/States/{countryId}")
    state = {"id": cfg.state_id, "name": cfg.state}
    try:
        lst = place.states(c, country["id"])
        print(f"   {country['name']} için {len(lst)} il alındı.")
        found = place.find_by_name(lst, cfg.state)
        if found:
            state = found
    except Exception as e:
        warn("states", e)
    print(f"   → seçilen il: {state['name']} (id={state['id']})")
    if not state["id"]:
        raise RuntimeError(f"il çözümlenemedi (AWQAT_STATE={cfg.state})")

    # [6] Tüm ilçeler
    sec(6, "Tüm ilçeler (parametresiz)", "/api/Place/Cities")
    try:
        allc = place.all_cities(c)
        print(f"   {len(allc)} ilçe (tüm iller) alındı.")
    except Exception as e:
        warn("allCities", e)

    # [7] İle göre ilçeler
    sec(7, "İle göre ilçeler", "/api/Place/Cities/{stateId}")
    city = {"id": cfg.city_id, "name": "(AWQAT_CITY_ID)"}
    try:
        lst = place.cities(c, state["id"])
        print(f"   {state['name']} için {len(lst)} ilçe alındı.")
        city = pick_city(lst, cfg)
    except Exception as e:
        warn("cities", e)
    print(f"   → seçilen ilçe: {city['name']} (id={city['id']})")
    if not city["id"]:
        raise RuntimeError("ilçe çözümlenemedi")

    # [8] İlçe detay (kıble)
    sec(8, "İlçe detay (kıble açısı)", "/api/Place/CityDetail/{cityId}")
    try:
        d = place.city_detail(c, city["id"])
        print(f"   {d.get('city')} / {d.get('country')} · kıble açısı: {d.get('qiblaAngle')}° · "
              f"Kâbe'ye uzaklık: {d.get('distanceToKaaba')} km")
    except Exception as e:
        warn("cityDetail", e)

    # [9] Günlük namaz vakitleri
    sec(9, "Günlük namaz vakitleri", "/api/PrayerTime/Daily/{cityId}")
    try:
        times = prayer.daily(c, city["id"])
        if times:
            print_prayer(city, times[0])
    except Exception as e:
        warn("daily", e)

    # [10] Haftalık
    sec(10, "Haftalık namaz vakitleri", "/api/PrayerTime/Weekly/{cityId}")
    try:
        t = prayer.weekly(c, city["id"])
        print(f"   {len(t)} günlük veri ({first_date(t)} … {last_date(t)})")
    except Exception as e:
        warn("weekly", e)

    # [11] Aylık
    sec(11, "Aylık namaz vakitleri", "/api/PrayerTime/Monthly/{cityId}")
    try:
        t = prayer.monthly(c, city["id"])
        print(f"   {len(t)} günlük veri ({first_date(t)} … {last_date(t)})")
    except Exception as e:
        warn("monthly", e)

    # [12] Bayram namazı
    sec(12, "Bayram namazı", "/api/PrayerTime/Eid/{cityId}")
    try:
        ev = prayer.eid(c, city["id"])
        print(f"   Ramazan B.: {ev.get('eidAlFitrDate')} {ev.get('eidAlFitrTime')} · "
              f"Kurban B.: {ev.get('eidAlAdhaDate')} {ev.get('eidAlAdhaTime')}")
    except Exception as e:
        warn("eid", e)

    # [13] Ramazan imsakiyesi
    sec(13, "Ramazan imsakiyesi", "/api/PrayerTime/Ramadan/{cityId}")
    try:
        t = prayer.ramadan(c, city["id"])
        print(f"   {len(t)} günlük imsakiye verisi")
    except Exception as e:
        warn("ramadan", e)

    print("\n====================================================")
    print("✅ Tüm endpoint'ler çağrıldı.")
    print("====================================================")


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(f"\n❌ Hata: {exc}", file=sys.stderr)
        sys.exit(1)
