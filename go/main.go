// AwqatSalah Cookbook — Go örneği.
//
// Çalıştırma:  cd go && go run .
//
// Resmi PDF'teki TÜM endpoint'leri sırayla çağırır (Isparta baz alınarak):
//
//	[1]  Login              /api/Auth/Login
//	[2]  Günlük içerik       /api/DailyContent
//	[3]  Ülkeler             /api/Place/Countries
//	[4]  Tüm iller           /api/Place/States
//	[5]  Ülkeye göre iller   /api/Place/States/{countryId}
//	[6]  Tüm ilçeler         /api/Place/Cities
//	[7]  İle göre ilçeler    /api/Place/Cities/{stateId}
//	[8]  İlçe detay          /api/Place/CityDetail/{cityId}
//	[9]  Günlük vakitler     /api/PrayerTime/Daily/{cityId}
//	[10] Haftalık vakitler   /api/PrayerTime/Weekly/{cityId}
//	[11] Aylık vakitler      /api/PrayerTime/Monthly/{cityId}
//	[12] Bayram namazı       /api/PrayerTime/Eid/{cityId}
//	[13] Ramazan imsakiyesi  /api/PrayerTime/Ramadan/{cityId}
//
// (RefreshToken /api/Auth/RefreshToken/{rt} token süresi dolunca otomatik çağrılır.)
// Tüm ayarlar kök .env / ortam değişkenlerinden okunur (bkz. .env.example).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/omerfruk/AwqatSalah-Cookbook/go/awqat"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Hata: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	loginOnly := flag.Bool("login", false, "sadece token al/yenile, endpoint turunu atla (kotayı korur)")
	flag.Parse()

	cfg, err := awqat.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client := awqat.New(cfg)

	fmt.Println("====================================================")
	fmt.Println("  AwqatSalah Cookbook — Go (tüm endpoint turu)")
	fmt.Printf("  Base: %s\n", cfg.BaseURL)
	fmt.Println("  Not: her endpoint 1 kez çağrılır · kota: path başına 5/gün")
	fmt.Println("====================================================")

	// [1] Kimlik (token) ----------------------------------------------------
	sec(1, "Kimlik (token)", "/api/Auth/Login")
	if err := client.EnsureAuth(ctx); err != nil {
		return fmt.Errorf("login: %w", err) // kritik
	}
	tok := short(client.AccessToken())
	switch client.TokenSource() {
	case "cache":
		fmt.Printf("   ✓ kayıtlı token kullanıldı, login ATILMADI (kota korundu): %s…\n", tok)
	case "env":
		fmt.Printf("   ✓ .env token'ı kullanıldı, login ATILMADI: %s…\n", tok)
	case "refresh":
		fmt.Printf("   ✓ token yenilendi (refresh): %s…\n", tok)
	default:
		fmt.Printf("   ✓ login yapıldı, token .awqat-token.json'a kaydedildi: %s…\n", tok)
	}

	// -login bayrağı: sadece kimlik adımı (cache'i tohumlamak / kotayı korumak için).
	if *loginOnly {
		fmt.Println("\n(-login) Sadece kimlik adımı çalıştırıldı; endpoint turu atlandı.")
		return nil
	}

	// [2] Günün Ayet/Hadis/Dua ---------------------------------------------
	sec(2, "Günlük İçerik (Ayet/Hadis/Dua)", "/api/DailyContent")
	if dc, err := client.DailyContent(ctx); err != nil {
		warn("dailyContent", err)
	} else {
		fmt.Printf("   Ayet : %s\n", truncate(dc.Verse, 90))
		fmt.Printf("   Hadis: %s\n", truncate(dc.Hadith, 90))
		fmt.Printf("   Dua  : %s\n", truncate(dc.Pray, 90))
	}

	// [3] Ülkeler -----------------------------------------------------------
	sec(3, "Ülkeler", "/api/Place/Countries")
	country := awqat.Place{ID: cfg.CountryID, Name: cfg.Country}
	if countries, err := client.Countries(ctx); err != nil {
		warn("countries", err)
	} else {
		fmt.Printf("   %d ülke alındı.\n", len(countries))
		if found, ok := awqat.FindByName(countries, cfg.Country); ok {
			country = found
		}
	}
	fmt.Printf("   → seçilen ülke: %s (id=%d)\n", country.Name, country.ID)
	if country.ID == 0 {
		return fmt.Errorf("ülke çözümlenemedi (AWQAT_COUNTRY=%q)", cfg.Country)
	}

	// [4] Tüm iller ---------------------------------------------------------
	sec(4, "Tüm iller (parametresiz)", "/api/Place/States")
	if all, err := client.AllStates(ctx); err != nil {
		warn("allStates", err)
	} else {
		fmt.Printf("   %d il (tüm ülkeler) alındı.\n", len(all))
	}

	// [5] Ülkeye göre iller -------------------------------------------------
	sec(5, "Ülkeye göre iller", "/api/Place/States/{countryId}")
	state := awqat.Place{ID: cfg.StateID, Name: cfg.State}
	if states, err := client.States(ctx, country.ID); err != nil {
		warn("states", err)
	} else {
		fmt.Printf("   %s için %d il alındı.\n", country.Name, len(states))
		if found, ok := awqat.FindByName(states, cfg.State); ok {
			state = found
		}
	}
	fmt.Printf("   → seçilen il: %s (id=%d)\n", state.Name, state.ID)
	if state.ID == 0 {
		return fmt.Errorf("il çözümlenemedi (AWQAT_STATE=%q)", cfg.State)
	}

	// [6] Tüm ilçeler -------------------------------------------------------
	sec(6, "Tüm ilçeler (parametresiz)", "/api/Place/Cities")
	if all, err := client.AllCities(ctx); err != nil {
		warn("allCities", err)
	} else {
		fmt.Printf("   %d ilçe (tüm iller) alındı.\n", len(all))
	}

	// [7] İle göre ilçeler --------------------------------------------------
	sec(7, "İle göre ilçeler", "/api/Place/Cities/{stateId}")
	city := awqat.Place{ID: cfg.CityID, Name: "(AWQAT_CITY_ID)"}
	if cities, err := client.Cities(ctx, state.ID); err != nil {
		warn("cities", err)
	} else {
		fmt.Printf("   %s için %d ilçe alındı.\n", state.Name, len(cities))
		city = pickCity(cities, cfg)
	}
	fmt.Printf("   → seçilen ilçe: %s (id=%d)\n", city.Name, city.ID)
	if city.ID == 0 {
		return fmt.Errorf("ilçe çözümlenemedi")
	}

	// [8] İlçe detay (kıble) ------------------------------------------------
	sec(8, "İlçe detay (kıble açısı)", "/api/Place/CityDetail/{cityId}")
	if d, err := client.CityDetail(ctx, city.ID); err != nil {
		warn("cityDetail", err)
	} else {
		fmt.Printf("   %s / %s · kıble açısı: %s° · Kâbe'ye uzaklık: %s km\n",
			d.City, d.Country, d.QiblaAngle, d.DistanceToKaaba)
	}

	// [9] Günlük namaz vakitleri -------------------------------------------
	sec(9, "Günlük namaz vakitleri", "/api/PrayerTime/Daily/{cityId}")
	if times, err := client.DailyPrayerTimes(ctx, city.ID); err != nil {
		warn("daily", err)
	} else if len(times) > 0 {
		printPrayer(city, times[0])
	}

	// [10] Haftalık ---------------------------------------------------------
	sec(10, "Haftalık namaz vakitleri", "/api/PrayerTime/Weekly/{cityId}")
	if times, err := client.WeeklyPrayerTimes(ctx, city.ID); err != nil {
		warn("weekly", err)
	} else {
		fmt.Printf("   %d günlük veri (%s … %s)\n", len(times), firstDate(times), lastDate(times))
	}

	// [11] Aylık ------------------------------------------------------------
	sec(11, "Aylık namaz vakitleri", "/api/PrayerTime/Monthly/{cityId}")
	if times, err := client.MonthlyPrayerTimes(ctx, city.ID); err != nil {
		warn("monthly", err)
	} else {
		fmt.Printf("   %d günlük veri (%s … %s)\n", len(times), firstDate(times), lastDate(times))
	}

	// [12] Bayram namazı ----------------------------------------------------
	sec(12, "Bayram namazı", "/api/PrayerTime/Eid/{cityId}")
	if e, err := client.EidPrayerTimes(ctx, city.ID); err != nil {
		warn("eid", err)
	} else {
		fmt.Printf("   Ramazan B.: %s %s · Kurban B.: %s %s\n",
			e.EidAlFitrDate, e.EidAlFitrTime, e.EidAlAdhaDate, e.EidAlAdhaTime)
	}

	// [13] Ramazan imsakiyesi ----------------------------------------------
	sec(13, "Ramazan imsakiyesi", "/api/PrayerTime/Ramadan/{cityId}")
	if times, err := client.RamadanPrayerTimes(ctx, city.ID); err != nil {
		warn("ramadan", err)
	} else {
		fmt.Printf("   %d günlük imsakiye verisi\n", len(times))
	}

	fmt.Println("\n====================================================")
	fmt.Println("✅ Tüm endpoint'ler çağrıldı.")
	fmt.Println("====================================================")
	return nil
}

// pickCity; ilçe seçer: AWQAT_CITY_ID varsa o, yoksa il adıyla eşleşen (merkez), yoksa ilk ilçe.
func pickCity(cities []awqat.Place, cfg *awqat.Config) awqat.Place {
	if cfg.CityID != 0 {
		for _, c := range cities {
			if c.ID == cfg.CityID {
				return c
			}
		}
		return awqat.Place{ID: cfg.CityID, Name: "(AWQAT_CITY_ID)"}
	}
	if c, ok := awqat.FindByName(cities, cfg.State); ok {
		return c
	}
	if len(cities) > 0 {
		return cities[0]
	}
	return awqat.Place{}
}

func printPrayer(city awqat.Place, t awqat.PrayerTime) {
	fmt.Printf("   🕌 %s — %s (Hicri: %s)\n", city.Name, t.GregorianDateShort, t.HijriDateShort)
	rows := []struct{ label, val string }{
		{"İmsak  (Fajr)", t.Fajr},
		{"Güneş  (Sunrise)", t.Sunrise},
		{"Öğle   (Dhuhr)", t.Dhuhr},
		{"İkindi (Asr)", t.Asr},
		{"Akşam  (Maghrib)", t.Maghrib},
		{"Yatsı  (Isha)", t.Isha},
	}
	for _, r := range rows {
		fmt.Printf("      %-18s %s\n", r.label, r.val)
	}
}

func sec(n int, title, path string) {
	fmt.Printf("\n[%d] %s\n     %s\n", n, title, path)
}

func warn(label string, err error) {
	fmt.Printf("   ⚠ %s atlandı: %v\n", label, err)
}

func firstDate(t []awqat.PrayerTime) string {
	if len(t) == 0 {
		return "-"
	}
	return t[0].GregorianDateShort
}

func lastDate(t []awqat.PrayerTime) string {
	if len(t) == 0 {
		return "-"
	}
	return t[len(t)-1].GregorianDateShort
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func short(s string) string {
	if len(s) <= 16 {
		return s
	}
	return s[:16]
}
