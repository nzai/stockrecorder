package store

import (
	"time"

	"github.com/nzai/stockrecorder/quote"
)

// Store 存储
type Store interface {
	// 判断是否记录过
	Exists(*quote.Exchange, time.Time) (bool, error)
	// 保存
	Save(*quote.ExchangeDailyQuote) error
	// 读取
	Load(*quote.Exchange, time.Time) (*quote.ExchangeDailyQuote, error)
}
