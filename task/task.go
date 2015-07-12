package task

import (
	"log"

	"github.com/nzai/stockrecorder/analyse"
	"github.com/nzai/stockrecorder/crawl"
)

//	启动任务
func StartTasks() error {
	log.Print("启动任务")

	//	启动抓取任务
	go crawl.StartJobs()

	//	启动分析任务
	go analyse.StartJobs()

	return nil
}
