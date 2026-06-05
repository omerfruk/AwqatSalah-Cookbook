<?php

declare(strict_types=1);

namespace Awqat;

/**
 * Place (Ülke / İl / İlçe) endpoint'leri. İstemciyi ilk argüman alır.
 * Hiyerarşi: Country -> State (il) -> City (ilçe).
 */
class Place
{
    public static function countries(Client $c)
    {
        return $c->getJson('/api/Place/Countries');
    }

    public static function allStates(Client $c)
    {
        return $c->getJson('/api/Place/States');
    }

    public static function states(Client $c, int $countryId)
    {
        return $c->getJson("/api/Place/States/$countryId");
    }

    public static function allCities(Client $c)
    {
        return $c->getJson('/api/Place/Cities');
    }

    public static function cities(Client $c, int $stateId)
    {
        return $c->getJson("/api/Place/Cities/$stateId");
    }

    public static function cityDetail(Client $c, int $cityId)
    {
        return $c->getJson("/api/Place/CityDetail/$cityId");
    }

    /**
     * Liste içinde adı (Türkçe + büyük/küçük harf duyarsız) eşleşen ilk öğeyi bulur.
     * @param array<int,array<string,mixed>> $items
     * @return array<string,mixed>|null
     */
    public static function findByName(array $items, string $name): ?array
    {
        $target = self::fold($name);
        foreach ($items as $p) {
            if (str_contains(self::fold((string) ($p['name'] ?? '')), $target)) {
                return $p;
            }
        }
        return null;
    }

    /** Türkçe karakterleri ASCII'ye indirger ve küçük harfe çevirir (arama için). */
    public static function fold(string $s): string
    {
        $map = [
            'İ' => 'i', 'I' => 'i', 'ı' => 'i',
            'Ç' => 'c', 'ç' => 'c', 'Ğ' => 'g', 'ğ' => 'g',
            'Ö' => 'o', 'ö' => 'o', 'Ş' => 's', 'ş' => 's', 'Ü' => 'u', 'ü' => 'u',
        ];
        return strtolower(strtr(trim($s), $map));
    }
}
