package market

import (
	"testing"
	"time"

	"log"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/db"
	"github.com/nzai/stockrecorder/io"
)

func init() {

	//	读取配置文件
	err := config.SetRootDir(`c:\gohome\src\github.com\nzai\stockrecorder\`)
	if err != nil {
		log.Fatal("读取配置文件错误: ", err)
		return
	}
}

func TestParse60(t *testing.T) {

	buffer, err := io.ReadAllBytes(`c:\data\America\ACHN\20151001_raw.txt`)
	if err != nil {
		t.Errorf("读取文件失败:%s", err.Error())
	}

	dd, _ := time.Parse("20060102", "20151001")
	ar, err := parseDailyYahooJson("America", "AAPL", dd, buffer)
	if err != nil {
		t.Errorf("解析失败:%s", err.Error())
	}

	if ar.DailyResult.Error {
		t.Errorf("解析错误:%s", ar.DailyResult.Message)
	}

	t.Log(ar)

	err = db.DailySave(ar)
	if err != nil {
		t.Errorf("保存失败:%s", err.Error())
	}

	var u1 int64 = 14437023600
	t.Logf("%d is %s", u1, time.Unix(u1, 0).Format("2006-01-02 15:04:05"))
}
