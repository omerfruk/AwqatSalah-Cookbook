<?php

declare(strict_types=1);

namespace Awqat;

/** DailyContent: günün Ayet / Hadis / Dua içeriği (parametresiz). */
class Content
{
    public static function dailyContent(Client $c)
    {
        return $c->getJson('/api/DailyContent');
    }
}
