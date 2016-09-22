package store

import (
	"time"

	"github.com/nzai/stockrecorder/market"
)

// Store 存储
type Store interface {
	// 判断是否记录过
	Exists(tempPath string, _market market.Market, date time.Time) (bool, error)
	// 保存
	Save(tempPath string, quote market.DailyQuote) error
}
