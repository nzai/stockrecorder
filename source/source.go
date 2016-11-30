package source

import (
	"time"

	"github.com/nzai/stockrecorder/market"
)

// Source 数据源
type Source interface {
	// 数据能报保存多长时间(能查到的最早数据距今多长时间)
	Expiration() time.Duration
	// 获取公司每日报价
	Crawl(_market market.Market, company market.Company, date time.Time) (*market.CompanyDailyQuote, error)
	// 最大并发数
	ParallelMax() int
	// 失败重试次数
	RetryCount() int
	// 失败重试时间间隔
	RetryInterval() time.Duration
}
