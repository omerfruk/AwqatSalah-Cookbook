"""Mock HTTP sunucusuna karşı uçtan uca testler (gerçek API'ye gerek yok).

Çalıştırma (python/ klasöründen):  python3 -m unittest discover
"""

import threading
import unittest
from http.server import BaseHTTPRequestHandler, HTTPServer

from awqat import content, place, prayer
from awqat.client import AwqatClient, unwrap
from awqat.config import Config


def make_handler(routes):
    class Handler(BaseHTTPRequestHandler):
        def log_message(self, *args):
            pass  # sessiz

        def _respond(self):
            route = routes.get(self.path)
            if route is None:
                self.send_response(404)
                self.end_headers()
                self.wfile.write(b"not found")
                return
            status, body = route(self)
            self.send_response(status)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(body.encode("utf-8"))

        def do_GET(self):
            self._respond()

        def do_POST(self):
            length = int(self.headers.get("Content-Length", 0))
            if length:
                self.rfile.read(length)  # gövdeyi tüket
            self._respond()

    return Handler


def make_cfg(base, **over):
    d = dict(
        base_url=base, email="a", password="b", country="Türkiye", state="Isparta",
        country_id=0, state_id=0, city_id=0, access_token="", refresh_token="", token_cache_path="",
    )
    d.update(over)
    return Config(**d)


class TestClient(unittest.TestCase):
    def _server(self, routes):
        srv = HTTPServer(("127.0.0.1", 0), make_handler(routes))
        threading.Thread(target=srv.serve_forever, daemon=True).start()
        return srv, f"http://127.0.0.1:{srv.server_address[1]}"

    def test_end_to_end(self):
        routes = {
            "/api/Auth/Login": lambda h: (200, '{"success":true,"data":{"accessToken":"ACCESS","refreshToken":"R"}}'),
            "/api/DailyContent": lambda h: (200, '{"success":true,"data":{"verse":"V","hadith":"H","pray":"P"}}'),
            "/api/Place/Countries": lambda h: (200, '{"success":true,"data":[{"id":1,"name":"KUZEY KIBRIS"},{"id":2,"name":"TÜRKİYE"}]}'),
            "/api/Place/States/2": lambda h: (200, '{"success":true,"data":[{"id":538,"name":"ISPARTA"}]}'),
            "/api/Place/Cities/538": lambda h: (200, '{"success":true,"data":[{"id":9528,"name":"ISPARTA"}]}'),
            "/api/Place/CityDetail/9528": lambda h: (200, '{"success":true,"data":{"id":"9528","qiblaAngle":"151","distanceToKaaba":"2023","city":"ISPARTA","country":"TÜRKİYE"}}'),
            "/api/PrayerTime/Daily/9528": lambda h: (200, '{"success":true,"data":[{"fajr":"03:44","isha":"22:01"}]}'),
        }
        srv, base = self._server(routes)
        try:
            c = AwqatClient(make_cfg(base))
            c.ensure_auth()
            self.assertEqual(c.access_token, "ACCESS")

            self.assertEqual(content.daily_content(c)["verse"], "V")

            country = place.find_by_name(place.countries(c), "Türkiye")  # "TÜRKİYE" ile eşleşmeli
            self.assertEqual(country["id"], 2)

            state = place.find_by_name(place.states(c, country["id"]), "Isparta")
            self.assertEqual(state["id"], 538)

            ci = place.cities(c, state["id"])
            self.assertEqual(ci[0]["id"], 9528)

            detail = place.city_detail(c, ci[0]["id"])
            self.assertEqual(detail["id"], "9528")  # STRING
            self.assertEqual(detail["qiblaAngle"], "151")

            times = prayer.daily(c, ci[0]["id"])
            self.assertEqual(times[0]["fajr"], "03:44")
        finally:
            srv.shutdown()
            srv.server_close()

    def test_valid_token_skips_login(self):
        calls = {"n": 0}

        def login_route(h):
            calls["n"] += 1
            return (200, '{"success":true,"data":{"accessToken":"NEW"}}')

        def countries_route(h):
            self.assertEqual(h.headers.get("Authorization"), "Bearer SEED")
            return (200, '{"success":true,"data":[]}')

        srv, base = self._server({"/api/Auth/Login": login_route, "/api/Place/Countries": countries_route})
        try:
            c = AwqatClient(make_cfg(base, access_token="SEED"))
            c.ensure_auth()
            place.countries(c)
            self.assertEqual(calls["n"], 0)  # login atılmamalı
            self.assertEqual(c.token_source(), "env")
        finally:
            srv.shutdown()
            srv.server_close()

    def test_auth_prefix_fallback(self):
        routes = {
            "/api/Auth/Login": lambda h: (404, "nf"),
            "/Auth/Login": lambda h: (200, '{"success":true,"data":{"accessToken":"OK"}}'),
        }
        srv, base = self._server(routes)
        try:
            c = AwqatClient(make_cfg(base))
            c.ensure_auth()
            self.assertEqual(c.access_token, "OK")
        finally:
            srv.shutdown()
            srv.server_close()

    def test_unwrap_failure(self):
        with self.assertRaises(RuntimeError):
            unwrap('{"success":false,"message":"yetkisiz"}')

    def test_fold_turkish(self):
        self.assertEqual(place.fold("TÜRKİYE"), "turkiye")
        self.assertEqual(place.fold("Isparta"), "isparta")


if __name__ == "__main__":
    unittest.main()
