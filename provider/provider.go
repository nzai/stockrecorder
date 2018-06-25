package provider

import (
	"github.com/nzai/stockrecorder/quote"
)

// DataProvider 数据提供者
type DataProvider interface {
	Exchange() *quote.Exchange
	Companies() ([]*quote.Company, error)
}
