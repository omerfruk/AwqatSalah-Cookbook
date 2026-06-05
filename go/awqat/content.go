package awqat

import "context"

// DailyContent; günün Ayet / Hadis / Dua içeriğini getirir (parametresiz).
// GET /api/DailyContent
func (c *Client) DailyContent(ctx context.Context) (DailyContent, error) {
	return getJSON[DailyContent](ctx, c, "/api/DailyContent")
}
