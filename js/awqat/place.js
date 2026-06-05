// Place (Ülke / İl / İlçe) endpoint'leri. İstemciyi ilk argüman olarak alır.
// Hiyerarşi: Country → State (il) → City (ilçe).

export const countries = (c) => c.getJSON('/api/Place/Countries');
export const allStates = (c) => c.getJSON('/api/Place/States');
export const states = (c, countryId) => c.getJSON(`/api/Place/States/${countryId}`);
export const allCities = (c) => c.getJSON('/api/Place/Cities');
export const cities = (c, stateId) => c.getJSON(`/api/Place/Cities/${stateId}`);
export const cityDetail = (c, cityId) => c.getJSON(`/api/Place/CityDetail/${cityId}`);

// findByName; liste içinde adı (Türkçe + büyük/küçük harf duyarsız) eşleşen ilk öğeyi bulur.
// Örn. "Türkiye" ↔ "TÜRKİYE", "Isparta" ↔ "ISPARTA".
export function findByName(list, name) {
  const target = fold(name);
  return list.find((p) => fold(p.name).includes(target)) ?? null;
}

// fold; Türkçe karakterleri ASCII'ye indirger ve küçük harfe çevirir (arama için).
export function fold(s) {
  const map = {
    İ: 'i', I: 'i', ı: 'i',
    Ç: 'c', ç: 'c', Ğ: 'g', ğ: 'g',
    Ö: 'o', ö: 'o', Ş: 's', ş: 's', Ü: 'u', ü: 'u',
  };
  return (s ?? '')
    .trim()
    .replace(/[İIıÇçĞğÖöŞşÜü]/g, (ch) => map[ch] ?? ch)
    .toLowerCase();
}
