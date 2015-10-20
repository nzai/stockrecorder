package market

import (
	"testing"
	"time"
)

func init() {

	//	//	读取配置文件
	//	err := config.SetRootDir(`c:\gohome\src\github.com\nzai\stockrecorder\`)
	//	if err != nil {
	//		log.Fatal("读取配置文件错误: ", err)
	//		return
	//	}
}

func TestParse60(t *testing.T) {

	var u1 int64 = 1444829400
	t.Logf("%d is %s", u1, time.Unix(u1, 0).Format("2006-01-02 15:04:05"))
}
