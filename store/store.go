package store

import (
	"time"

	"github.com/nzai/stockrecorder/market"
)

// Store 存储
type Store interface {
	// 判断是否记录过
	Exists(_market market.Market, date time.Time) (bool, error)
	// 保存
	Save(quote market.DailyQuote) error
	// 读取
	Load(_market market.Market, date time.Time) (market.DailyQuote, error)
}
