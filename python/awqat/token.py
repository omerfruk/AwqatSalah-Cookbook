"""Token önbelleği: login/refresh sonrası token'ı diske kalıcı yazar.

Sonraki çalıştırmalarda geçerliyse yeniden login ATILMAZ (kota korunur).
Biçim Go/JS ile UYUMLU: {"accessToken", "refreshToken", "expiry" (ISO 8601)}.
"""

import json
import os


def load_token_cache(path):
    if not path:
        return None
    try:
        with open(path, "r", encoding="utf-8") as f:
            tc = json.load(f)
        return tc if tc.get("accessToken") else None
    except (OSError, ValueError):
        return None


def save_token_cache(path, tc):
    """Token'ı yalnızca sahibin okuyabileceği (0600) dosyaya yazar. Hata yutulur."""
    if not path:
        return
    try:
        with open(path, "w", encoding="utf-8") as f:
            json.dump(tc, f, indent=2)
        os.chmod(path, 0o600)
    except OSError:
        pass
