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
			log.Print("致命错误:", err)
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
