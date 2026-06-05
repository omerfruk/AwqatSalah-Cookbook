package awqat

import (
	"context"
	"fmt"
	"strings"
)

// Countries; tüm ülkeleri getirir. GET /api/Place/Countries
func (c *Client) Countries(ctx context.Context) ([]Place, error) {
	return getJSON[[]Place](ctx, c, "/api/Place/Countries")
}

// AllStates; tüm eyaletleri/illeri getirir (parametresiz). GET /api/Place/States
func (c *Client) AllStates(ctx context.Context) ([]Place, error) {
	return getJSON[[]Place](ctx, c, "/api/Place/States")
}

// States; bir ülkenin illerini (eyaletlerini) getirir. GET /api/Place/States/{countryID}
func (c *Client) States(ctx context.Context, countryID int64) ([]Place, error) {
	return getJSON[[]Place](ctx, c, fmt.Sprintf("/api/Place/States/%d", countryID))
}

// AllCities; tüm şehirleri/ilçeleri getirir (parametresiz). GET /api/Place/Cities
func (c *Client) AllCities(ctx context.Context) ([]Place, error) {
	return getJSON[[]Place](ctx, c, "/api/Place/Cities")
}

// Cities; bir ilin ilçelerini getirir. GET /api/Place/Cities/{stateID}
func (c *Client) Cities(ctx context.Context, stateID int64) ([]Place, error) {
	return getJSON[[]Place](ctx, c, fmt.Sprintf("/api/Place/Cities/%d", stateID))
}

// CityDetail; bir ilçenin kıble açısı vb. detayını getirir. GET /api/Place/CityDetail/{cityID}
func (c *Client) CityDetail(ctx context.Context, cityID int64) (CityDetail, error) {
	return getJSON[CityDetail](ctx, c, fmt.Sprintf("/api/Place/CityDetail/%d", cityID))
}

// FindByName; liste içinde adı (Türkçe + büyük/küçük harf duyarsız) eşleşen ilk Place'i bulur.
// Örn. "Türkiye" ↔ "TÜRKİYE", "Isparta" ↔ "ISPARTA".
func FindByName(list []Place, name string) (Place, bool) {
	target := fold(name)
	for _, p := range list {
		if strings.Contains(fold(p.Name), target) {
			return p, true
		}
	}
	return Place{}, false
}

// fold; Türkçe karakterleri ASCII'ye indirger ve küçük harfe çevirir (arama için).
func fold(s string) string {
	r := strings.NewReplacer(
		"İ", "i", "I", "i", "ı", "i",
		"Ç", "c", "ç", "c",
		"Ğ", "g", "ğ", "g",
		"Ö", "o", "ö", "o",
		"Ş", "s", "ş", "s",
		"Ü", "u", "ü", "u",
	)
	return strings.ToLower(r.Replace(strings.TrimSpace(s)))
}
