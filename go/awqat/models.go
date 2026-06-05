package awqat

import "encoding/json"

// APIResponse, Diyanet AwqatSalah API'sinin TÜM yanıtlarını saran ortak zarftır.
// Her endpoint bu yapıyı döndürür: { "data": <T>, "success": bool, "message": string|null }
type APIResponse[T any] struct {
	Data    T       `json:"data"`
	Success bool    `json:"success"`
	Message *string `json:"message"`
}

// Token; login ve refresh yanıtlarındaki kimlik bilgileridir.
type Token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Place; ülke, il (state/eyalet) ve ilçe (city/şehir) için ORTAK yapıdır.
// Hiyerarşi: Country → State (Türkiye için il) → City (Türkiye için ilçe).
// Örn. ülke: {id:2, code:"TURKEY", name:"TÜRKİYE"} · il: {id:500, code:"ADANA", name:"ADANA"}
type Place struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// DailyContent; günün Ayet / Hadis / Dua içeriğidir. (/api/DailyContent)
type DailyContent struct {
	ID           int64  `json:"id"`
	DayOfYear    int64  `json:"dayOfYear"`
	Verse        string `json:"verse"`        // Ayet
	VerseSource  string `json:"verseSource"`  // Ayet kaynağı
	Hadith       string `json:"hadith"`       // Hadis
	HadithSource string `json:"hadithSource"` // Hadis kaynağı
	Pray         string `json:"pray"`         // Dua
	PraySource   string `json:"praySource"`   // Dua kaynağı (null olabilir)
}

// CityDetail; bir ilçenin kıble açısı vb. detaylarıdır. (/api/Place/CityDetail/{cityId})
// DİKKAT: id burada STRING'tir ("17885"), Place.ID'den (number) farklıdır.
type CityDetail struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Code                 string `json:"code"`
	GeographicQiblaAngle string `json:"geographicQiblaAngle"`
	DistanceToKaaba      string `json:"distanceToKaaba"`
	QiblaAngle           string `json:"qiblaAngle"`
	City                 string `json:"city"`
	CityEn               string `json:"cityEn"`
	Country              string `json:"country"`
	CountryEn            string `json:"countryEn"`
}

// PrayerTime; tek bir günün namaz vakitleridir. (Daily/Weekly/Monthly/Ramadan ortak)
type PrayerTime struct {
	ShapeMoonURL          string      `json:"shapeMoonUrl"` // ay safhası görseli
	Fajr                  string      `json:"fajr"`         // İmsak
	Sunrise               string      `json:"sunrise"`      // Güneş
	Dhuhr                 string      `json:"dhuhr"`        // Öğle
	Asr                   string      `json:"asr"`          // İkindi
	Maghrib               string      `json:"maghrib"`      // Akşam
	Isha                  string      `json:"isha"`         // Yatsı
	AstronomicalSunset    string      `json:"astronomicalSunset"`
	AstronomicalSunrise   string      `json:"astronomicalSunrise"`
	HijriDateShort        string      `json:"hijriDateShort"`
	HijriDateShortIso8601 string      `json:"hijriDateShortIso8601"`
	HijriDateLong         string      `json:"hijriDateLong"`
	HijriDateLongIso8601  string      `json:"hijriDateLongIso8601"`
	QiblaTime             string      `json:"qiblaTime"`
	GregorianDateShort    string      `json:"gregorianDateShort"` // "DD.MM.YYYY"
	GregorianDateShortIso string      `json:"gregorianDateShortIso8601"`
	GregorianDateLong     string      `json:"gregorianDateLong"`
	GregorianDateLongIso  string      `json:"gregorianDateLongIso8601"`
	GreenwichMeanTimezone json.Number `json:"greenwichMeanTimezone"` // örn. 3 = GMT+3
}

// EidPrayerTime; bayram namazı tarih/saatleridir. (/api/PrayerTime/Eid/{cityId})
type EidPrayerTime struct {
	EidAlFitrHijri string `json:"eidAlFitrHijri"` // Ramazan Bayramı (Hicri)
	EidAlFitrDate  string `json:"eidAlFitrDate"`
	EidAlFitrTime  string `json:"eidAlFitrTime"`
	EidAlAdhaHijri string `json:"eidAlAdhaHijri"` // Kurban Bayramı (Hicri)
	EidAlAdhaDate  string `json:"eidAlAdhaDate"`
	EidAlAdhaTime  string `json:"eidAlAdhaTime"`
}
