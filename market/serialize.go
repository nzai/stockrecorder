package market

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/io"
)

type Peroid60 struct {
	Market string
	Code   string
	Time   string
	Open   float32
	Close  float32
	High   float32
	Low    float32
	Volume int64
}

const (
	rawSuffix     = "raw"
	preSuffix     = "pre"
	regularSuffix = "regular"
	postSuffix    = "post"
	errorSuffix   = "error"
	invalidSuffix = "invalid"
)

//	判断不同结尾的文件是否存在
func isExists(market Market, code string, date time.Time, suffix string) bool {

	fileName := fmt.Sprintf("%s_%s.txt", date.Format("20060102"), suffix)
	filePath := filepath.Join(config.Get().DataDir, market.Name(), code, fileName)

	return io.IsExists(filePath)
}

//	是否已下载
func isDownloaded(market Market, code string, date time.Time) bool {
	return isExists(market, code, date, rawSuffix)
}

//	是否已处理
func isProcessed(rawPath string) bool {

	//	是否有error或regular文件
	return io.IsExists(strings.Replace(rawPath, rawSuffix, errorSuffix, -1)) ||
		io.IsExists(strings.Replace(rawPath, rawSuffix, regularSuffix, -1))
}

//	是否有异常数据
func isInvalid(market Market, code string, date time.Time) bool {
	return isExists(market, code, date, invalidSuffix)
}

//	将文件标为异常
func invalidate(market Market, code string, date time.Time) error {

	fileDir := filepath.Join(config.Get().DataDir, market.Name(), code)
	dateString := date.Format("20060102")

	rawFilePath := filepath.Join(fileDir, fmt.Sprintf("%s_%s.txt", dateString, rawSuffix))
	invalidFilePath := filepath.Join(fileDir, fmt.Sprintf("%s_%s.txt", dateString, invalidSuffix))

	//	将文件改名，以便重新下载分析
	return os.Rename(rawFilePath, invalidFilePath)
}

//	记录异常
func saveError(market Market, code string, date time.Time, err error) error {

	fileName := fmt.Sprintf("%s_%s.txt", date.Format("20060102"), errorSuffix)
	filePath := filepath.Join(config.Get().DataDir, market.Name(), code, fileName)

	return io.WriteString(filePath, err.Error())
}

//	保存原始数据
func saveRaw(market Market, code string, date time.Time, buffer []byte) (string, error) {

	fileName := fmt.Sprintf("%s_%s.txt", date.Format("20060102"), rawSuffix)
	filePath := filepath.Join(config.Get().DataDir, market.Name(), code, fileName)

	return filePath, io.WriteBytes(filePath, buffer)
}

//	从文件名中获取信息
func retrieveRawParams(rawFilePath string) (Market, string, time.Time, error) {

	other := strings.Replace(rawFilePath, config.Get().DataDir, "", -1)

	//	路径处理
	if os.IsPathSeparator(other[0]) {
		other = other[1:]
	}

	parts := strings.Split(other, string(os.PathSeparator))
	if len(parts) != 3 {
		return nil, "", time.Now(), fmt.Errorf("[ProcessQueue]\t不规则的文件名:%s", rawFilePath)
	}

	market, found := markets[parts[0]]
	if !found {
		return nil, "", time.Now(), fmt.Errorf("[ProcessQueue]\t错误的市场定义:%s", parts[0])
	}

	dateString := parts[2][:8]
	day, err := time.Parse("20060102", dateString)
	if err != nil {
		return nil, "", time.Now(), fmt.Errorf("[ProcessQueue]\t不规则的文件名日期:%s", dateString)
	}

	return market, parts[1], day, nil
}

//	保存分时数据到文件
func savePeroid60(market Market, code, suffix string, date time.Time, peroids []Peroid60) error {

	fileName := fmt.Sprintf("%s_%s.txt", date.Format("20060102"), suffix)
	filePath := filepath.Join(config.Get().DataDir, market.Name(), code, fileName)

	lines := make([]string, 0)
	for _, p := range peroids {
		lines = append(lines, fmt.Sprintf("%s\t%.3f\t%.3f\t%.3f\t%.3f\t%d", p.Time, p.Open, p.Close, p.High, p.Low, p.Volume))
	}

	return io.WriteLines(filePath, lines)
}

//	从文件读取分时数据
func loadPeroid60(market Market, code string, date time.Time) ([]Peroid60, error) {

	//	检查分时数据是否存在
	if !isExists(market, code, date, regularSuffix) {
		return nil, fmt.Errorf("[%s]\t[%s]在%s的分时数据不存在", market.Name(), code, date.Format("20060102"))
	}

	fileName := fmt.Sprintf("%s_%s.txt", date.Format("20060102"), regularSuffix)
	filePath := filepath.Join(config.Get().DataDir, market.Name(), code, fileName)

	lines, err := io.ReadLines(filePath)
	if err != nil {
		return nil, err
	}

	var timeString string
	var open, _close, high, low float32
	var volume int64

	dateString := date.Format("20060102")
	peroids := make([]Peroid60, 0)
	for _, line := range lines {

		_, err = fmt.Sscanf(line, "%s\t%f\t%f\t%f\t%f\t%d", &timeString, &open, &_close, &high, &low, &volume)
		if err != nil {
			return nil, err
		}

		peroids = append(peroids, Peroid60{
			Market: market.Name(),
			Code:   strings.ToUpper(code),
			Time:   dateString + timeString,
			Open:   open,
			Close:  _close,
			High:   high,
			Low:    low,
			Volume: volume})
	}

	return peroids, nil
}
