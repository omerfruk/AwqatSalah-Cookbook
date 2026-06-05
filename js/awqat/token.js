import { readFileSync, writeFileSync } from 'node:fs';

// Token önbelleği: login/refresh sonrası token'ın diske yazılan kalıcı hâli.
// Sonraki çalıştırmalarda geçerliyse yeniden login ATILMAZ (kota korunur).
// Biçim Go ile UYUMLU: { accessToken, refreshToken, expiry (ISO 8601) }.

export function loadTokenCache(path) {
  if (!path) return null;
  try {
    const tc = JSON.parse(readFileSync(path, 'utf8'));
    return tc && tc.accessToken ? tc : null;
  } catch {
    return null;
  }
}

// Token'ı yalnızca sahibin okuyabileceği (0600) bir dosyaya yazar.
// Hata sessizce yutulur — önbellek bir kolaylıktır, kritik değildir.
export function saveTokenCache(path, tc) {
  if (!path) return;
  try {
    writeFileSync(path, JSON.stringify(tc, null, 2), { mode: 0o600 });
  } catch {
    // yok say
  }
}
