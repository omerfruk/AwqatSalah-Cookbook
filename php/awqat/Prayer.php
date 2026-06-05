<?php

declare(strict_types=1);

namespace Awqat;

/** PrayerTime (Namaz Vakitleri) endpoint'leri. İstemciyi ilk argüman alır. */
class Prayer
{
    public static function daily(Client $c, int $cityId)
    {
        return $c->getJson("/api/PrayerTime/Daily/$cityId");
    }

    public static function weekly(Client $c, int $cityId)
    {
        return $c->getJson("/api/PrayerTime/Weekly/$cityId");
    }

    public static function monthly(Client $c, int $cityId)
    {
        return $c->getJson("/api/PrayerTime/Monthly/$cityId");
    }

    public static function ramadan(Client $c, int $cityId)
    {
        return $c->getJson("/api/PrayerTime/Ramadan/$cityId");
    }

    public static function eid(Client $c, int $cityId)
    {
        return $c->getJson("/api/PrayerTime/Eid/$cityId");
    }
}
