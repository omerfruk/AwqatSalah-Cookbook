// AwqatSalah Cookbook — JavaScript (Node) örneği.
//
// Çalıştırma:  cd js && node index.js          (tüm endpoint turu)
//              cd js && node index.js --login   (sadece token al/yenile, kotayı korur)
//
// Resmi PDF'teki TÜM endpoint'leri sırayla çağırır (Isparta baz alınarak).
// Tüm ayarlar kök .env / ortam değişkenlerinden okunur (bkz. ../.env.example).

import { loadConfig } from './awqat/config.js';
import { AwqatClient } from './awqat/client.js';
import { countries, allStates, states, allCities, cities, cityDetail, findByName } from './awqat/place.js';
import { dailyContent } from './awqat/content.js';
import { daily, weekly, monthly, ramadan, eid } from './awqat/prayer.js';

const loginOnly = process.argv.includes('--login') || process.argv.includes('-login');

async function main() {
  const cfg = loadConfig();
  const c = new AwqatClient(cfg);

  console.log('====================================================');
  console.log('  AwqatSalah Cookbook — JavaScript (tüm endpoint turu)');
  console.log(`  Base: ${cfg.baseURL}`);
  console.log('  Not: her endpoint 1 kez çağrılır · kota: path başına 5/gün');
  console.log('====================================================');

  // [1] Kimlik (token) ----------------------------------------------------
  sec(1, 'Kimlik (token)', '/api/Auth/Login');
  await c.ensureAuth();
  const tok = short(c.accessToken);
  switch (c.tokenSource()) {
    case 'cache': console.log(`   ✓ kayıtlı token kullanıldı, login ATILMADI (kota korundu): ${tok}…`); break;
    case 'env': console.log(`   ✓ .env token'ı kullanıldı, login ATILMADI: ${tok}…`); break;
    case 'refresh': console.log(`   ✓ token yenilendi (refresh): ${tok}…`); break;
    default: console.log(`   ✓ login yapıldı, token .awqat-token.json'a kaydedildi: ${tok}…`);
  }
  if (loginOnly) {
    console.log('\n(--login) Sadece kimlik adımı çalıştırıldı; endpoint turu atlandı.');
    return;
  }

  // [2] Günün Ayet/Hadis/Dua ---------------------------------------------
  sec(2, 'Günlük İçerik (Ayet/Hadis/Dua)', '/api/DailyContent');
  try {
    const dc = await dailyContent(c);
    console.log(`   Ayet : ${truncate(dc.verse, 90)}`);
    console.log(`   Hadis: ${truncate(dc.hadith, 90)}`);
    console.log(`   Dua  : ${truncate(dc.pray, 90)}`);
  } catch (e) { warn('dailyContent', e); }

  // [3] Ülkeler -----------------------------------------------------------
  sec(3, 'Ülkeler', '/api/Place/Countries');
  let country = { id: cfg.countryId, name: cfg.country };
  try {
    const list = await countries(c);
    console.log(`   ${list.length} ülke alındı.`);
    const f = findByName(list, cfg.country);
    if (f) country = f;
  } catch (e) { warn('countries', e); }
  console.log(`   → seçilen ülke: ${country.name} (id=${country.id})`);
  if (!country.id) throw new Error(`ülke çözümlenemedi (AWQAT_COUNTRY=${cfg.country})`);

  // [4] Tüm iller ---------------------------------------------------------
  sec(4, 'Tüm iller (parametresiz)', '/api/Place/States');
  try {
    const all = await allStates(c);
    console.log(`   ${all.length} il (tüm ülkeler) alındı.`);
  } catch (e) { warn('allStates', e); }

  // [5] Ülkeye göre iller -------------------------------------------------
  sec(5, 'Ülkeye göre iller', '/api/Place/States/{countryId}');
  let state = { id: cfg.stateId, name: cfg.state };
  try {
    const list = await states(c, country.id);
    console.log(`   ${country.name} için ${list.length} il alındı.`);
    const f = findByName(list, cfg.state);
    if (f) state = f;
  } catch (e) { warn('states', e); }
  console.log(`   → seçilen il: ${state.name} (id=${state.id})`);
  if (!state.id) throw new Error(`il çözümlenemedi (AWQAT_STATE=${cfg.state})`);

  // [6] Tüm ilçeler -------------------------------------------------------
  sec(6, 'Tüm ilçeler (parametresiz)', '/api/Place/Cities');
  try {
    const all = await allCities(c);
    console.log(`   ${all.length} ilçe (tüm iller) alındı.`);
  } catch (e) { warn('allCities', e); }

  // [7] İle göre ilçeler --------------------------------------------------
  sec(7, 'İle göre ilçeler', '/api/Place/Cities/{stateId}');
  let city = { id: cfg.cityId, name: '(AWQAT_CITY_ID)' };
  try {
    const list = await cities(c, state.id);
    console.log(`   ${state.name} için ${list.length} ilçe alındı.`);
    city = pickCity(list, cfg);
  } catch (e) { warn('cities', e); }
  console.log(`   → seçilen ilçe: ${city.name} (id=${city.id})`);
  if (!city.id) throw new Error('ilçe çözümlenemedi');

  // [8] İlçe detay (kıble) ------------------------------------------------
  sec(8, 'İlçe detay (kıble açısı)', '/api/Place/CityDetail/{cityId}');
  try {
    const d = await cityDetail(c, city.id);
    console.log(`   ${d.city} / ${d.country} · kıble açısı: ${d.qiblaAngle}° · Kâbe'ye uzaklık: ${d.distanceToKaaba} km`);
  } catch (e) { warn('cityDetail', e); }

  // [9] Günlük namaz vakitleri -------------------------------------------
  sec(9, 'Günlük namaz vakitleri', '/api/PrayerTime/Daily/{cityId}');
  try {
    const times = await daily(c, city.id);
    if (times.length) printPrayer(city, times[0]);
  } catch (e) { warn('daily', e); }

  // [10] Haftalık ---------------------------------------------------------
  sec(10, 'Haftalık namaz vakitleri', '/api/PrayerTime/Weekly/{cityId}');
  try {
    const t = await weekly(c, city.id);
    console.log(`   ${t.length} günlük veri (${firstDate(t)} … ${lastDate(t)})`);
  } catch (e) { warn('weekly', e); }

  // [11] Aylık ------------------------------------------------------------
  sec(11, 'Aylık namaz vakitleri', '/api/PrayerTime/Monthly/{cityId}');
  try {
    const t = await monthly(c, city.id);
    console.log(`   ${t.length} günlük veri (${firstDate(t)} … ${lastDate(t)})`);
  } catch (e) { warn('monthly', e); }

  // [12] Bayram namazı ----------------------------------------------------
  sec(12, 'Bayram namazı', '/api/PrayerTime/Eid/{cityId}');
  try {
    const e = await eid(c, city.id);
    console.log(`   Ramazan B.: ${e.eidAlFitrDate} ${e.eidAlFitrTime} · Kurban B.: ${e.eidAlAdhaDate} ${e.eidAlAdhaTime}`);
  } catch (e) { warn('eid', e); }

  // [13] Ramazan imsakiyesi ----------------------------------------------
  sec(13, 'Ramazan imsakiyesi', '/api/PrayerTime/Ramadan/{cityId}');
  try {
    const t = await ramadan(c, city.id);
    console.log(`   ${t.length} günlük imsakiye verisi`);
  } catch (e) { warn('ramadan', e); }

  console.log('\n====================================================');
  console.log("✅ Tüm endpoint'ler çağrıldı.");
  console.log('====================================================');
}

// pickCity; ilçe seçer: AWQAT_CITY_ID varsa o, yoksa il adıyla eşleşen (merkez), yoksa ilk ilçe.
function pickCity(list, cfg) {
  if (cfg.cityId) {
    return list.find((c) => c.id === cfg.cityId) ?? { id: cfg.cityId, name: '(AWQAT_CITY_ID)' };
  }
  const f = findByName(list, cfg.state);
  if (f) return f;
  return list[0] ?? { id: 0, name: '' };
}

function printPrayer(city, t) {
  console.log(`   🕌 ${city.name} — ${t.gregorianDateShort} (Hicri: ${t.hijriDateShort})`);
  const rows = [
    ['İmsak  (Fajr)', t.fajr],
    ['Güneş  (Sunrise)', t.sunrise],
    ['Öğle   (Dhuhr)', t.dhuhr],
    ['İkindi (Asr)', t.asr],
    ['Akşam  (Maghrib)', t.maghrib],
    ['Yatsı  (Isha)', t.isha],
  ];
  for (const [label, val] of rows) console.log(`      ${label.padEnd(18)} ${val}`);
}

const sec = (n, title, path) => console.log(`\n[${n}] ${title}\n     ${path}`);
const warn = (label, err) => console.log(`   ⚠ ${label} atlandı: ${err.message}`);
const short = (s) => (s ?? '').slice(0, 16);
const firstDate = (t) => (t.length ? t[0].gregorianDateShort : '-');
const lastDate = (t) => (t.length ? t[t.length - 1].gregorianDateShort : '-');
const truncate = (s, n) => {
  const r = [...(s ?? '').trim()];
  return r.length <= n ? r.join('') : r.slice(0, n).join('') + '…';
};

main().catch((err) => {
  console.error(`\n❌ Hata: ${err.message}`);
  process.exit(1);
});
