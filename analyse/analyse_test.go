package analyse

import (
	"testing"

	"log"

	"github.com/nzai/stockrecorder/config"
)

func init() {

	//	读取配置文件
	err := config.SetRootDir(`c:\gohome\src\github.com\nzai\stockrecorder\`)
	if err != nil {
		log.Fatal("读取配置文件错误: ", err)
		return
	}
}

func TestRetrieveParams(t *testing.T) {

	path := `c:\data\America\ETP\20150902_raw.txt`

	marketName, code, day, err := retrieveParams(path)
	if err != nil {
		t.Errorf("错误的路径:%s", err.Error())
	}

	t.Logf("Market:%s  Company:%s Date:%v", marketName, code, day)
}
