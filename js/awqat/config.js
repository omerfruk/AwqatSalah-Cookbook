import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';

// Token önbelleği dosya adı (kök .env yanında). Tüm diller AYNI dosyayı paylaşır.
const TOKEN_CACHE_FILE = '.awqat-token.json';

// .env'i CWD'den YUKARI doğru arar (alt klasörden çalıştırınca da bulunsun diye).
// Bulunan .env'in klasörünü (token önbelleği için kök) ve değerlerini döndürür.
function loadDotEnvUpwards() {
  let dir = process.cwd();
  const start = dir;
  for (;;) {
    try {
      const text = readFileSync(join(dir, '.env'), 'utf8');
      return { values: parseDotEnv(text), rootDir: dir };
    } catch {
      // bu klasörde .env yok, üste çık
    }
    const parent = dirname(dir);
    if (parent === dir) return { values: {}, rootDir: start }; // kök, .env yok → CWD
    dir = parent;
  }
}

// Basit KEY=VALUE ayrıştırıcı (# yorum, boş satır, tırnak soyma).
function parseDotEnv(text) {
  const out = {};
  for (let line of text.split('\n')) {
    line = line.trim();
    if (!line || line.startsWith('#')) continue;
    if (line.startsWith('export ')) line = line.slice(7);
    const eq = line.indexOf('=');
    if (eq < 0) continue;
    const key = line.slice(0, eq).trim();
    let val = line.slice(eq + 1).trim();
    if ((val.startsWith('"') && val.endsWith('"')) || (val.startsWith("'") && val.endsWith("'"))) {
      val = val.slice(1, -1);
    }
    if (key) out[key] = val;
  }
  return out;
}

// loadConfig; ortam değişkenleri + kök .env'den yapılandırmayı üretir.
export function loadConfig() {
  const { values: dotenv, rootDir } = loadDotEnvUpwards();

  // get: önce OS ortam değişkeni, sonra .env, sonra varsayılan.
  const get = (key, def = '') => {
    const env = (process.env[key] ?? '').trim();
    if (env) return env;
    const dv = (dotenv[key] ?? '').trim();
    return dv || def;
  };
  const num = (s) => {
    const n = parseInt(s, 10);
    return Number.isNaN(n) ? 0 : n;
  };

  const cfg = {
    baseURL: get('AWQAT_BASE_URL', 'https://awqatsalah.diyanet.gov.tr').replace(/\/+$/, ''),
    email: get('AWQAT_EMAIL'),
    password: get('AWQAT_PASSWORD'),
    country: get('AWQAT_COUNTRY', 'Türkiye'),
    state: get('AWQAT_STATE', 'Isparta'),
    countryId: num(get('AWQAT_COUNTRY_ID')),
    stateId: num(get('AWQAT_STATE_ID')),
    cityId: num(get('AWQAT_CITY_ID')),
    accessToken: get('AWQAT_ACCESS_TOKEN'),
    refreshToken: get('AWQAT_REFRESH_TOKEN'),
    tokenCachePath: join(rootDir, TOKEN_CACHE_FILE),
  };

  if (!cfg.email || !cfg.password) {
    throw new Error(
      'AWQAT_EMAIL ve AWQAT_PASSWORD gerekli — kök dizinde .env oluşturup doldurun:\n' +
        '    cp .env.example .env   (sonra e-posta/şifrenizi yazın)',
    );
  }
  return cfg;
}
