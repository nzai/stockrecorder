package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/task"
)

const (
	logFileName = "main.log"
)

func main() {

	defer func() {
		// 捕获panic异常
		log.Print("发生了致命错误")
		if err := recover(); err != nil {
			log.Print(err)
		}
	}()

	//	当前目录
	rootDir := filepath.Dir(os.Args[0])

	//	读取配置文件
	err := config.SetRootDir(rootDir)
	if err != nil {
		log.Fatal("读取配置文件错误: ", err)
		return
	}

	//	打开日志文件
	file, err := openLogFile()
	if err != nil {
		log.Fatal("打开日志文件错误: ", err)
		return
	}
	defer file.Close()

	//	设置日志输出文件
	log.SetOutput(file)

	//	启动任务
	err = task.StartTasks()
	if err != nil {
		log.Fatal("启动任务发生错误: ", err)
		
		return
	}

	//	阻塞，一直运行
	channel := make(chan int)
	<-channel
}

//	打开日志文件
func openLogFile() (*os.File, error) {
	//	日志文件路径
	logPath := filepath.Join(config.Get().DataDir, logFileName)

	return os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
}
