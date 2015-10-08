package market

import (
	"testing"
	"time"

	"github.com/nzai/stockrecorder/io"
)

func TestParse60(t *testing.T) {

	buffer, err := io.ReadAllBytes(`c:\data\America\AAPL\20151001_raw.txt`)
	if err != nil {
		t.Errorf("读取文件失败:%s", err.Error())
	}

	ar, err := ParseDailyYahooJson("America", "AAPL", time.Now(), buffer)
	if err != nil {
		t.Errorf("解析失败:%s", err.Error())
	}
	
	if ar.DailyResult.Error {
		t.Errorf("解析错误:%s", ar.DailyResult.Message)
	}
	
	t.Log(ar)

	var u1 int64 = 1443743880
	t.Logf("%d is %s", u1, time.Unix(u1, 0).Format("2006-01-02 15:04:05"))
}
