package market

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/io"
)

const (
	parseGCCount = 16
)

var (
	rawFilePath chan string = make(chan string)
)

func startProcessQueue() {
	//	启动处理协程
	go processRawFiles()

	//	搜索所有未处理的Raw文件,加入处理队列
	count, err := searchUnprocessedRawFiles()
	if err != nil {
		log.Printf("[ProcessQueue]\t搜索未处理的Raw文件时发生错误: %s", err.Error())
	}

	log.Printf("[ProcessQueue]\t一共搜索到%d个未处理的Raw文件", count)
}

//	添加到处理队列
func addProcessQueue(filePath string) {
	rawFilePath <- filePath
}

//	搜索所有未处理的Raw文件
func searchUnprocessedRawFiles() (int, error) {
	//	遍历目录
	var count int = 0

	//	遍历
	err := filepath.Walk(config.Get().DataDir, func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			return err
		}
		
		//	过滤原始数据文件
		if strings.HasSuffix(path, rawSuffix) {

			//	是否处理过(是否有error或regular文件)
			if isProcessed(path) {
				return err
			}

			//	没有处理过就加入处理队列
			addProcessQueue(path)
			count++
		}

		return err
	})

	return count, err
}

//	处理队列中的Raw文件
func processRawFiles() {

	chanSend := make(chan int, parseGCCount)
	defer close(chanSend)

	var path string
	for {
		path = <-rawFilePath

		//	并发抓取
		go func(filePath string) {

			//	处理文件
			err := processRaw(filePath)
			if err != nil {
				log.Printf("[ProcessQueue]\t处理raw数据[%s]数据失败: %s", filePath, err.Error())
			}

			<-chanSend
		}(path)

		//	流量控制
		chanSend <- 1
	}
}

//	处理一个Raw文件
func processRaw(filePath string) error {

	//	避免重复处理
	if isProcessed(filePath) {
		return nil
	}

	//	从文件名中获取信息
	market, code, date, err := retrieveRawParams(filePath)
	if err != nil {
		return err
	}

	//	读取文件
	buffer, err := io.ReadAllBytes(filePath)
	if err != nil {
		return err
	}

	return processDailyYahooJson(market, code, date, buffer)
}
