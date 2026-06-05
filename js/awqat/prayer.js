// PrayerTime (Namaz Vakitleri) endpoint'leri. İstemciyi ilk argüman olarak alır.

export const daily = (c, cityId) => c.getJSON(`/api/PrayerTime/Daily/${cityId}`);
export const weekly = (c, cityId) => c.getJSON(`/api/PrayerTime/Weekly/${cityId}`);
export const monthly = (c, cityId) => c.getJSON(`/api/PrayerTime/Monthly/${cityId}`);
export const ramadan = (c, cityId) => c.getJSON(`/api/PrayerTime/Ramadan/${cityId}`);
export const eid = (c, cityId) => c.getJSON(`/api/PrayerTime/Eid/${cityId}`);
