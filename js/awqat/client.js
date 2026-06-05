import { loadTokenCache, saveTokenCache } from './token.js';

// AwqatSalah (Diyanet Namaz Vakti) API istemcisi — sıfır bağımlılık (yerleşik fetch).
//
// Kimlik doğrulama akışı (resmi PDF'e göre):
//  1. POST /api/Auth/Login (email+password) → accessToken + refreshToken
//  2. Her istekte Authorization: Bearer <accessToken>
//  3. Süre dolmadan GET /api/Auth/RefreshToken/{refreshToken} ile yenilenir
//  4. Refresh de başarısızsa tekrar login olunur

const TOKEN_LIFETIME_MS = 30 * 60 * 1000; // access token ömrü (PDF §3: 30 dk)
const REFRESH_SAFETY_MS = 5 * 60 * 1000; // bitmeden bu kadar önce yenile
const MAX_PER_PATH_PER_DAY = 5; // Standart rol kotası: path başına / gün
const TIMEOUT_MS = 30000;

export class AwqatClient {
  constructor(cfg) {
    this.cfg = cfg;
    this.baseURL = cfg.baseURL;
    this.accessToken = '';
    this.refreshToken = '';
    this.expiry = 0; // epoch ms
    this.authPrefix = '/api'; // 404 alınırsa '' (eski örnekler) ile denenir
    this.source = ''; // 'env' | 'cache' | 'login' | 'refresh' | ''
    this.rate = new Map(); // path -> { n, date }

    // Mevcut token varsa yükle → geçerliyse login atılmaz (kota korunur).
    if (cfg.accessToken) {
      this.accessToken = cfg.accessToken;
      this.refreshToken = cfg.refreshToken || '';
      this.expiry = Date.now() + (TOKEN_LIFETIME_MS - REFRESH_SAFETY_MS);
      this.source = 'env';
    } else {
      const tc = loadTokenCache(cfg.tokenCachePath);
      if (tc) {
        this.accessToken = tc.accessToken;
        this.refreshToken = tc.refreshToken || '';
        this.expiry = Date.parse(tc.expiry) || 0;
        this.source = 'cache';
      }
    }
  }

  tokenSource() {
    return this.source;
  }

  // ---- Kimlik doğrulama -------------------------------------------------

  async ensureAuth() {
    if (this.accessToken && Date.now() < this.expiry) return; // token hâlâ geçerli
    if (this.refreshToken) {
      try {
        await this.#refresh();
        return;
      } catch {
        // refresh başarısız → login'e düş
      }
    }
    await this.#login();
  }

  async #login() {
    const payload = JSON.stringify({ email: this.cfg.email, password: this.cfg.password });
    let { status, text } = await this.#send('POST', this.authPrefix + '/Auth/Login', payload, '');
    if (status === 404 && this.authPrefix !== '') {
      this.authPrefix = ''; // eski yola geç ve hatırla
      ({ status, text } = await this.#send('POST', '/Auth/Login', payload, ''));
    }
    if (status !== 200) throw new Error(`login başarısız (HTTP ${status}): ${text}`);
    this.#storeToken(text, 'login');
  }

  async #refresh() {
    if (!this.refreshToken) throw new Error('refresh token yok');
    const path = this.authPrefix + '/Auth/RefreshToken/' + this.refreshToken;
    const { status, text } = await this.#send('GET', path, null, '');
    if (status !== 200) throw new Error(`refresh başarısız (HTTP ${status})`);
    this.#storeToken(text, 'refresh');
  }

  #storeToken(text, op) {
    const data = unwrap(text); // { accessToken, refreshToken }
    this.accessToken = data.accessToken;
    this.refreshToken = data.refreshToken;
    this.expiry = Date.now() + (TOKEN_LIFETIME_MS - REFRESH_SAFETY_MS);
    this.source = op; // 'login' veya 'refresh'
    // Token'ı kalıcı yap → sonraki çalıştırmalar geçerliyse login atmaz.
    saveTokenCache(this.cfg.tokenCachePath, {
      accessToken: this.accessToken,
      refreshToken: this.refreshToken,
      expiry: new Date(this.expiry).toISOString(),
    });
  }

  // ---- İstek katmanı (kota + auth + 401 retry) --------------------------

  // getJSON; kimlik doğrulamalı GET yapıp zarfı açarak data döndürür.
  async getJSON(path) {
    return unwrap(await this.doGet(path));
  }

  async doGet(path) {
    return this.#do('GET', path, null);
  }

  async doPost(path, body) {
    return this.#do('POST', path, JSON.stringify(body));
  }

  async #do(method, path, payload) {
    this.#checkRate(path);
    await this.ensureAuth();

    let { status, text } = await this.#authedSend(method, path, payload);

    // 401 → token geçersiz olmuş olabilir: sıfırla, yeniden doğrula, 1 kez tekrar dene.
    if (status === 401) {
      this.accessToken = '';
      this.refreshToken = '';
      await this.ensureAuth();
      ({ status, text } = await this.#authedSend(method, path, payload));
    }

    if (status !== 200) throw new Error(`${method} ${path} başarısız (HTTP ${status}): ${text}`);
    return text;
  }

  #authedSend(method, path, payload) {
    return this.#send(method, path, payload, this.accessToken);
  }

  async #send(method, path, payload, token) {
    const headers = { 'Content-Type': 'application/json' };
    if (token) headers.Authorization = 'Bearer ' + token;

    // Ağ hatasında (reset / ölü keep-alive bağlantı) kısa beklemeyle 3 kez dene.
    // (Go'nun net/http'si bunu otomatik yapar; Node fetch yapmaz.)
    let lastErr;
    for (let attempt = 0; attempt < 3; attempt++) {
      try {
        const res = await fetch(this.baseURL + path, {
          method,
          headers,
          body: payload ?? undefined,
          signal: AbortSignal.timeout(TIMEOUT_MS),
        });
        return { status: res.status, text: await res.text() };
      } catch (e) {
        lastErr = e;
        await new Promise((r) => setTimeout(r, 200 * (attempt + 1)));
      }
    }
    throw new Error(`${method} ${path}: ${lastErr?.message ?? 'ağ hatası'}`);
  }

  // path için günlük kotayı (5/gün) kontrol eder ve sayacı artırır.
  // Not: sayaç süreç-içidir; gerçek kota sunucu tarafındadır.
  #checkRate(path) {
    const today = new Date().toISOString().slice(0, 10); // YYYY-MM-DD
    const cnt = this.rate.get(path);
    if (!cnt || cnt.date !== today) {
      this.rate.set(path, { n: 1, date: today });
      return;
    }
    if (cnt.n >= MAX_PER_PATH_PER_DAY) {
      throw new Error(`kota: ${path} için bugün ${cnt.n}/${MAX_PER_PATH_PER_DAY} istek kullanıldı`);
    }
    cnt.n++;
  }
}

// unwrap; ortak zarfı açar, success kontrol eder, data döndürür.
export function unwrap(text) {
  let resp;
  try {
    resp = JSON.parse(text);
  } catch (e) {
    throw new Error(`yanıt çözümlenemedi: ${e.message}`);
  }
  if (!resp.success) throw new Error(`API başarısız: ${resp.message ?? 'bilinmeyen hata'}`);
  return resp.data;
}
