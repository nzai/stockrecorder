package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/task"
)

const (
	configFileName           = "config.ini"
	configLogSection         = "path"
	configLogKey             = "logpath"
	configLogDefaultFileName = "main.log"
)

func main() {
	//	当前目录
	root := filepath.Dir(os.Args[0])
	filename := filepath.Join(root, configFileName)

	//	使用所有cpu
	//	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	//	读取配置文件
	err := config.SetConfigFile(filename)
	if err != nil {
		log.Fatal(err)
		return
	}

	//	日志文件路径
	logPath := config.GetString(configLogSection, configLogKey, configLogDefaultFileName)
	logDir := filepath.Dir(logPath)
	_, err = os.Stat(logDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(logDir, 0x777)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0x777)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()

	//	设置日志输出文件
	log.SetOutput(file)

	//	启动任务
	err = task.StartTasks()
	if err != nil {
		log.Fatal(err)
		return
	}
}
