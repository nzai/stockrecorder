package market

import "time"

// Quote 报价
type Quote struct {
	Time   time.Time // 起始时间
	Open   float32   // 开盘价
	Close  float32   // 收盘价
	Max    float32   // 最高价
	Min    float32   // 最低价
	Volume int64     // 成交量
}

// CompanyDailyQuote 公司每日市场报价
type CompanyDailyQuote struct {
	Date    time.Time // 日期
	Company Company   // 公司
	Pre     []Quote   // 盘前
	Regular []Quote   // 盘中
	Post    []Quote   // 盘后
}

// DailyQuote 每日市场报价
type DailyQuote struct {
	Market
	Date      time.Time // 日期
	Companies []Company // 公司
	Pre       []Quote   // 盘前
	Regular   []Quote   // 盘中
	Post      []Quote   // 盘后
}
