package market

import (
	"testing"
	"time"

	"log"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/io"
)

func init() {

	//	读取配置文件
	err := config.SetRootDir(`g:\gohome\src\github.com\nzai\stockrecorder\`)
	if err != nil {
		log.Fatal("读取配置文件错误: ", err)
		return
	}
}

func TestParse60(t *testing.T) {

	var u1 int64 = 1444829400
	t.Logf("%d is %s", u1, time.Unix(u1, 0).Format("2006-01-02 15:04:05"))
}

func TestProcessRaw(t *testing.T) {

	marketOffset["America"] = -43200

	path := `c:\data\America\ABEV\20151030_raw.txt`
	buffer, err := io.ReadAllBytes(path)
	if err != nil {
		t.Error(err)
	}
	market, code, date, err := retrieveParams(path)
	t.Logf("market:%s\tcode:%s\tdate:%s", market, code, date.Format("20060102"))

	err = processDailyYahooJson(market, code, date, buffer)
	if err != nil {
		t.Error(err)
	}
}
