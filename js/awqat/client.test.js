import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createServer } from 'node:http';
import { AwqatClient, unwrap } from './client.js';
import { countries, states, cities, cityDetail, findByName, fold } from './place.js';
import { dailyContent } from './content.js';
import { daily } from './prayer.js';

// Basit mock sunucu: { '/yol': (req,res)=>... } eşlemesiyle.
function startMock(routes) {
  const server = createServer((req, res) => {
    res.setHeader('Content-Type', 'application/json');
    const handler = routes[req.url];
    if (handler) {
      req.resume(); // gövdeyi tüket
      handler(req, res);
    } else {
      res.statusCode = 404;
      res.end('not found');
    }
  });
  return new Promise((resolve) => {
    server.listen(0, () => resolve({ base: `http://127.0.0.1:${server.address().port}`, server }));
  });
}

const ok = (res, body) => res.end(body);

test('uçtan uca: login → günlük içerik → il/ilçe → detay → vakit', async () => {
  const { base, server } = await startMock({
    '/api/Auth/Login': (req, res) => ok(res, '{"success":true,"data":{"accessToken":"ACCESS","refreshToken":"REFRESH"}}'),
    '/api/DailyContent': (req, res) => ok(res, '{"success":true,"data":{"id":1,"dayOfYear":1,"verse":"V","hadith":"H","pray":"P"}}'),
    '/api/Place/Countries': (req, res) => {
      assert.equal(req.headers.authorization, 'Bearer ACCESS');
      ok(res, '{"success":true,"data":[{"id":1,"code":"NORTH CYPRUS","name":"KUZEY KIBRIS"},{"id":2,"code":"TURKEY","name":"TÜRKİYE"}]}');
    },
    '/api/Place/States/2': (req, res) => ok(res, '{"success":true,"data":[{"id":538,"code":"ISPARTA","name":"ISPARTA"}]}'),
    '/api/Place/Cities/538': (req, res) => ok(res, '{"success":true,"data":[{"id":9528,"code":"ISPARTA","name":"ISPARTA"}]}'),
    '/api/Place/CityDetail/9528': (req, res) => ok(res, '{"success":true,"data":{"id":"9528","qiblaAngle":"151","distanceToKaaba":"2023","city":"ISPARTA","country":"TÜRKİYE"}}'),
    '/api/PrayerTime/Daily/9528': (req, res) => ok(res, '{"success":true,"data":[{"fajr":"03:44","isha":"22:01","gregorianDateShort":"05.06.2026"}]}'),
  });
  try {
    const c = new AwqatClient({ baseURL: base, email: 'a', password: 'b', country: 'Türkiye', state: 'Isparta', tokenCachePath: '' });
    await c.ensureAuth();
    assert.equal(c.accessToken, 'ACCESS');

    const dc = await dailyContent(c);
    assert.equal(dc.verse, 'V');

    const country = findByName(await countries(c), 'Türkiye'); // "TÜRKİYE" ile eşleşmeli
    assert.equal(country.id, 2);

    const state = findByName(await states(c, country.id), 'Isparta');
    assert.equal(state.id, 538);

    const ci = await cities(c, state.id);
    assert.equal(ci[0].id, 9528);

    const d = await cityDetail(c, ci[0].id);
    assert.equal(d.id, '9528'); // STRING
    assert.equal(d.qiblaAngle, '151');

    const times = await daily(c, ci[0].id);
    assert.equal(times[0].fajr, '03:44');
  } finally {
    server.close();
  }
});

test('geçerli token verilince login İSTEĞİ atılmaz', async () => {
  let loginCalls = 0;
  const { base, server } = await startMock({
    '/api/Auth/Login': (req, res) => { loginCalls++; ok(res, '{"success":true,"data":{"accessToken":"NEW"}}'); },
    '/api/Place/Countries': (req, res) => {
      assert.equal(req.headers.authorization, 'Bearer SEED');
      ok(res, '{"success":true,"data":[]}');
    },
  });
  try {
    const c = new AwqatClient({ baseURL: base, email: 'a', password: 'b', accessToken: 'SEED', tokenCachePath: '' });
    await c.ensureAuth();
    await countries(c);
    assert.equal(loginCalls, 0);
    assert.equal(c.tokenSource(), 'env');
  } finally {
    server.close();
  }
});

test('/api/Auth/Login 404 → /Auth/Login fallback', async () => {
  const { base, server } = await startMock({
    '/api/Auth/Login': (req, res) => { res.statusCode = 404; res.end('nf'); },
    '/Auth/Login': (req, res) => ok(res, '{"success":true,"data":{"accessToken":"OK"}}'),
  });
  try {
    const c = new AwqatClient({ baseURL: base, email: 'a', password: 'b', tokenCachePath: '' });
    await c.ensureAuth();
    assert.equal(c.accessToken, 'OK');
  } finally {
    server.close();
  }
});

test('unwrap success=false → hata', () => {
  assert.throws(() => unwrap('{"success":false,"message":"yetkisiz","data":null}'));
});

test('fold Türkçe-duyarsız', () => {
  assert.equal(fold('TÜRKİYE'), 'turkiye');
  assert.equal(fold('Isparta'), 'isparta');
});
