package analyse

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/db"
	"github.com/nzai/stockrecorder/io"
	"github.com/nzai/stockrecorder/market"
)

const (
	ParseGCCount = 10
)

//	分析历史数据
func AnalyseHistory() {

	//	获取所有需要处理的文件名
	files, err := getAllRawFiles()
	if err != nil {
		log.Print("[分析]\t获取原始raw文件时发生错误:%s", err.Error())
		return
	}

	if len(files) == 0 {
		log.Print("[分析]\t没有raw文件需要处理")
		return
	}

	log.Printf("[分析]\t一共有%d个raw文件需要处理", len(files))

	chanSend := make(chan int, ParseGCCount)
	chanReceive := make(chan int)

	for _, file := range files {
		//	并发抓取
		go func(path string) {

			err := parseFile(path)
			if err != nil {
				log.Print(err)
			}

			<-chanSend
			chanReceive <- 1
		}(file)

		chanSend <- 1
	}

	//	阻塞，直到抓取所有
	for _, _ = range files {
		<-chanReceive
	}

	close(chanSend)
	close(chanReceive)
}

//	获取所有待处理的文件
func getAllRawFiles() ([]string, error) {

	files := make([]string, 0)
	//	遍历目录
	err := filepath.Walk(config.Get().DataDir, func(path string, info os.FileInfo, err error) error {

		//	过滤原始数据文件
		if strings.HasSuffix(path, "_raw.txt") {
			files = append(files, path)
		}
		return err
	})

	return files, err
}

//	解析文件
func parseFile(path string) error {

	//	如果文件不存在则忽略
	//	有可能被Market History任务处理了
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}

	marketName, code, day, err := retrieveParams(path)
	if err != nil {
		return err
	}

	//	检查数据库是否解析过
	found, err := db.DailyExists(marketName, code, day)
	if err != nil {
		return err
	}

	//	解析过就忽略
	if !found {
		//	读文件内容
		buffer, err := io.ReadAllBytes(path)
		if err != nil {
			return err
		}

		//	解析
		dar, err := market.ParseDailyYahooJson(marketName, code, day, buffer)
		if err != nil {
			return err
		}

		//	保存
		err = db.DailySave(dar)
		if err != nil {
			return err
		}
	}

	//	解析成功就删除文件
	return os.Remove(path)
}

//	从文件名中获取信息
func retrieveParams(path string) (string, string, time.Time, error) {

	other := strings.Replace(path, config.Get().DataDir, "", -1)

	//	路径处理
	if os.IsPathSeparator(other[0]) {
		other = other[1:]
	}

	parts := strings.Split(other, string(os.PathSeparator))
	if len(parts) != 3 {
		return "", "", time.Now(), fmt.Errorf("[分析]\t不规则的文件名:%s", path)
	}

	day, err := time.Parse("20060102", strings.Replace(parts[2], "_raw.txt", "", -1))
	if err != nil {
		return "", "", time.Now(), fmt.Errorf("[分析]\t不规则的文件名:%s", path)
	}

	return parts[0], parts[1], day, nil
}
