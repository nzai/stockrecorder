package market

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/io"
)

//	从雅虎财经获取上市公司分时数据
func DownloadCompanyDaily(marketName, companyCode, queryCode string, day time.Time) error {
	//	文件保存路径
	dataDir := config.GetDataDir()
	fileName := fmt.Sprintf("%s_raw.txt", day.Format("20060102"))
	filePath := filepath.Join(dataDir, marketName, companyCode, fileName)

	//	如果文件已存在就忽略
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		//	如果不存在就抓取并保存
		start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
		end := start.Add(time.Hour * 24)

		pattern := "https://finance-yql.media.yahoo.com/v7/finance/chart/%s?period2=%d&period1=%d&interval=1m&indicators=quote&includeTimestamps=true&includePrePost=true&events=div%7Csplit%7Cearn&corsDomain=finance.yahoo.com"
		url := fmt.Sprintf(pattern, queryCode, end.Unix(), start.Unix())

		html, err := io.DownloadStringRetry(url, retryTimes, retryIntervalSeconds)
		if err != nil {
			return err
		}

		//	写入文件
		return io.WriteString(filePath, html)
	}

	return nil
}
