"""Yapılandırma: ortam değişkenleri + kök .env (yukarı doğru aranır)."""

import os
from dataclasses import dataclass
from pathlib import Path

# Token önbelleği dosya adı (kök .env yanında). Tüm diller AYNI dosyayı paylaşır.
TOKEN_CACHE_FILE = ".awqat-token.json"


@dataclass
class Config:
    base_url: str
    email: str
    password: str
    country: str
    state: str
    country_id: int
    state_id: int
    city_id: int
    access_token: str
    refresh_token: str
    token_cache_path: str


def _load_dotenv_upwards():
    """CWD'den köke kadar .env arar; (değerler, kök_klasör) döndürür."""
    cwd = Path.cwd()
    for d in [cwd, *cwd.parents]:
        env = d / ".env"
        if env.is_file():
            return _parse_dotenv(env.read_text(encoding="utf-8")), d
    return {}, cwd  # .env yok → CWD kök


def _parse_dotenv(text):
    """Basit KEY=VALUE ayrıştırıcı (# yorum, boş satır, tırnak soyma)."""
    out = {}
    for line in text.splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        if line.startswith("export "):
            line = line[len("export "):]
        if "=" not in line:
            continue
        key, _, val = line.partition("=")
        key, val = key.strip(), val.strip()
        if len(val) >= 2 and val[0] == val[-1] and val[0] in "\"'":
            val = val[1:-1]
        if key:
            out[key] = val
    return out


def load_config():
    dotenv, root_dir = _load_dotenv_upwards()

    def get(key, default=""):
        v = os.environ.get(key, "").strip()
        if v:
            return v
        v = dotenv.get(key, "").strip()
        return v or default

    def num(s):
        try:
            return int(s)
        except (TypeError, ValueError):
            return 0

    cfg = Config(
        base_url=get("AWQAT_BASE_URL", "https://awqatsalah.diyanet.gov.tr").rstrip("/"),
        email=get("AWQAT_EMAIL"),
        password=get("AWQAT_PASSWORD"),
        country=get("AWQAT_COUNTRY", "Türkiye"),
        state=get("AWQAT_STATE", "Isparta"),
        country_id=num(get("AWQAT_COUNTRY_ID")),
        state_id=num(get("AWQAT_STATE_ID")),
        city_id=num(get("AWQAT_CITY_ID")),
        access_token=get("AWQAT_ACCESS_TOKEN"),
        refresh_token=get("AWQAT_REFRESH_TOKEN"),
        token_cache_path=str(Path(root_dir) / TOKEN_CACHE_FILE),
    )

    if not cfg.email or not cfg.password:
        raise RuntimeError(
            "AWQAT_EMAIL ve AWQAT_PASSWORD gerekli — kök dizinde .env oluşturup doldurun:\n"
            "    cp .env.example .env   (sonra e-posta/şifrenizi yazın)"
        )
    return cfg
