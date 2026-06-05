package awqat

import (
	"context"
	"fmt"
)

// DailyPrayerTimes; ilçenin bugünkü namaz vakitlerini getirir. GET /api/PrayerTime/Daily/{cityID}
func (c *Client) DailyPrayerTimes(ctx context.Context, cityID int64) ([]PrayerTime, error) {
	return getJSON[[]PrayerTime](ctx, c, fmt.Sprintf("/api/PrayerTime/Daily/%d", cityID))
}

// WeeklyPrayerTimes; haftalık vakitleri getirir. GET /api/PrayerTime/Weekly/{cityID}
func (c *Client) WeeklyPrayerTimes(ctx context.Context, cityID int64) ([]PrayerTime, error) {
	return getJSON[[]PrayerTime](ctx, c, fmt.Sprintf("/api/PrayerTime/Weekly/%d", cityID))
}

// MonthlyPrayerTimes; aylık vakitleri getirir. GET /api/PrayerTime/Monthly/{cityID}
func (c *Client) MonthlyPrayerTimes(ctx context.Context, cityID int64) ([]PrayerTime, error) {
	return getJSON[[]PrayerTime](ctx, c, fmt.Sprintf("/api/PrayerTime/Monthly/%d", cityID))
}

// MonthlyPrayerTimesFor; belirli yıl/ay için aylık vakitleri getirir.
// GET /api/PrayerTime/Monthly/{cityID}?year={year}&month={month}
func (c *Client) MonthlyPrayerTimesFor(ctx context.Context, cityID int64, year, month int) ([]PrayerTime, error) {
	return getJSON[[]PrayerTime](ctx, c, fmt.Sprintf("/api/PrayerTime/Monthly/%d?year=%d&month=%d", cityID, year, month))
}

// RamadanPrayerTimes; Ramazan ayı vakitlerini getirir. GET /api/PrayerTime/Ramadan/{cityID}
func (c *Client) RamadanPrayerTimes(ctx context.Context, cityID int64) ([]PrayerTime, error) {
	return getJSON[[]PrayerTime](ctx, c, fmt.Sprintf("/api/PrayerTime/Ramadan/%d", cityID))
}

// EidPrayerTimes; bayram namazı tarih/saatlerini getirir. GET /api/PrayerTime/Eid/{cityID}
func (c *Client) EidPrayerTimes(ctx context.Context, cityID int64) (EidPrayerTime, error) {
	return getJSON[EidPrayerTime](ctx, c, fmt.Sprintf("/api/PrayerTime/Eid/%d", cityID))
}
