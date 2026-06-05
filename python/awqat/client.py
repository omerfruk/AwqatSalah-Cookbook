"""AwqatSalah API istemcisi — sıfır bağımlılık (stdlib urllib).

Kimlik doğrulama akışı (resmi PDF'e göre):
 1. POST /api/Auth/Login (email+password) -> accessToken + refreshToken
 2. Her istekte Authorization: Bearer <accessToken>
 3. Süre dolmadan GET /api/Auth/RefreshToken/{refreshToken} ile yenilenir
 4. Refresh de başarısızsa tekrar login olunur
"""

import json
import re
import time
import urllib.error
import urllib.request
from datetime import datetime, timedelta, timezone

from .token import load_token_cache, save_token_cache

TOKEN_LIFETIME = timedelta(minutes=30)  # access token ömrü (PDF §3)
REFRESH_SAFETY = timedelta(minutes=5)   # bitmeden bu kadar önce yenile
MAX_PER_PATH_PER_DAY = 5                 # Standart rol kotası: path başına / gün
TIMEOUT = 30


class AwqatClient:
    def __init__(self, cfg):
        self.cfg = cfg
        self.base_url = cfg.base_url
        self.access_token = ""
        self.refresh_token = ""
        self.expiry = datetime.now(timezone.utc)
        self.auth_prefix = "/api"  # 404 alınırsa "" (eski örnekler) ile denenir
        self.source = ""           # "env" | "cache" | "login" | "refresh" | ""
        self.rate = {}             # path -> {"n", "date"}

        # Mevcut token varsa yükle → geçerliyse login atılmaz (kota korunur).
        if cfg.access_token:
            self.access_token = cfg.access_token
            self.refresh_token = cfg.refresh_token or ""
            self.expiry = datetime.now(timezone.utc) + (TOKEN_LIFETIME - REFRESH_SAFETY)
            self.source = "env"
        else:
            tc = load_token_cache(cfg.token_cache_path)
            if tc:
                self.access_token = tc.get("accessToken", "")
                self.refresh_token = tc.get("refreshToken", "") or ""
                self.expiry = _parse_iso(tc.get("expiry"))
                self.source = "cache"

    def token_source(self):
        return self.source

    # ---- Kimlik doğrulama --------------------------------------------------

    def ensure_auth(self):
        if self.access_token and datetime.now(timezone.utc) < self.expiry:
            return  # token hâlâ geçerli
        if self.refresh_token:
            try:
                self._refresh()
                return
            except Exception:
                pass  # refresh başarısız → login'e düş
        self._login()

    def _login(self):
        payload = json.dumps({"email": self.cfg.email, "password": self.cfg.password})
        status, text = self._send("POST", self.auth_prefix + "/Auth/Login", payload, "")
        if status == 404 and self.auth_prefix:
            self.auth_prefix = ""  # eski yola geç ve hatırla
            status, text = self._send("POST", "/Auth/Login", payload, "")
        if status != 200:
            raise RuntimeError(f"login başarısız (HTTP {status}): {text}")
        self._store_token(text, "login")

    def _refresh(self):
        if not self.refresh_token:
            raise RuntimeError("refresh token yok")
        path = self.auth_prefix + "/Auth/RefreshToken/" + self.refresh_token
        status, text = self._send("GET", path, None, "")
        if status != 200:
            raise RuntimeError(f"refresh başarısız (HTTP {status})")
        self._store_token(text, "refresh")

    def _store_token(self, text, op):
        data = unwrap(text)  # {"accessToken", "refreshToken"}
        self.access_token = data["accessToken"]
        self.refresh_token = data.get("refreshToken", "")
        self.expiry = datetime.now(timezone.utc) + (TOKEN_LIFETIME - REFRESH_SAFETY)
        self.source = op  # "login" veya "refresh"
        # Token'ı kalıcı yap → sonraki çalıştırmalar geçerliyse login atmaz.
        save_token_cache(self.cfg.token_cache_path, {
            "accessToken": self.access_token,
            "refreshToken": self.refresh_token,
            "expiry": self.expiry.isoformat(),
        })

    # ---- İstek katmanı (kota + auth + 401 retry) ---------------------------

    def get_json(self, path):
        """Kimlik doğrulamalı GET yapıp zarfı açarak data döndürür."""
        return unwrap(self.do_get(path))

    def do_get(self, path):
        return self._do("GET", path, None)

    def do_post(self, path, body):
        return self._do("POST", path, json.dumps(body))

    def _do(self, method, path, payload):
        self._check_rate(path)
        self.ensure_auth()

        status, text = self._send(method, path, payload, self.access_token)

        # 401 → token geçersiz olmuş olabilir: sıfırla, yeniden doğrula, 1 kez tekrar dene.
        if status == 401:
            self.access_token = ""
            self.refresh_token = ""
            self.ensure_auth()
            status, text = self._send(method, path, payload, self.access_token)

        if status != 200:
            raise RuntimeError(f"{method} {path} başarısız (HTTP {status}): {text}")
        return text

    def _send(self, method, path, payload, token):
        headers = {"Content-Type": "application/json"}
        if token:
            headers["Authorization"] = "Bearer " + token
        data = payload.encode("utf-8") if payload is not None else None

        # Ağ hatasında (reset / ölü bağlantı) kısa beklemeyle 3 kez dene.
        last_err = None
        for attempt in range(3):
            req = urllib.request.Request(self.base_url + path, data=data, headers=headers, method=method)
            try:
                with urllib.request.urlopen(req, timeout=TIMEOUT) as resp:
                    return resp.status, resp.read().decode("utf-8")
            except urllib.error.HTTPError as e:  # 4xx/5xx: durum + gövde (retry etme)
                return e.code, e.read().decode("utf-8", "replace")
            except OSError as e:  # URLError, timeout, connection reset ...
                last_err = e
                time.sleep(0.2 * (attempt + 1))
        raise RuntimeError(f"{method} {path}: {last_err}")

    def _check_rate(self, path):
        """Path için günlük kotayı (5/gün) kontrol eder. Sayaç süreç-içidir."""
        today = datetime.now().strftime("%Y-%m-%d")
        cnt = self.rate.get(path)
        if not cnt or cnt["date"] != today:
            self.rate[path] = {"n": 1, "date": today}
            return
        if cnt["n"] >= MAX_PER_PATH_PER_DAY:
            raise RuntimeError(f"kota: {path} için bugün {cnt['n']}/{MAX_PER_PATH_PER_DAY} istek kullanıldı")
        cnt["n"] += 1


def unwrap(text):
    """Ortak zarfı açar, success kontrol eder, data döndürür."""
    try:
        resp = json.loads(text)
    except ValueError as e:
        raise RuntimeError(f"yanıt çözümlenemedi: {e}")
    if not resp.get("success"):
        raise RuntimeError(f"API başarısız: {resp.get('message') or 'bilinmeyen hata'}")
    return resp.get("data")


def _parse_iso(s):
    """ISO 8601 (Go/JS) tarihini aware datetime'a çevirir; başarısızsa 'şimdi'."""
    if not s:
        return datetime.now(timezone.utc)
    s = s.strip().replace("Z", "+00:00")
    # Fractional saniyeyi tam 6 haneye normalize et (eski Python uyumu).
    s = re.sub(r"\.(\d+)", lambda m: "." + (m.group(1) + "000000")[:6], s)
    try:
        dt = datetime.fromisoformat(s)
        return dt if dt.tzinfo else dt.replace(tzinfo=timezone.utc)
    except ValueError:
        return datetime.now(timezone.utc)
