// DailyContent; günün Ayet / Hadis / Dua içeriğini getirir (parametresiz).
// GET /api/DailyContent
export const dailyContent = (c) => c.getJSON('/api/DailyContent');
