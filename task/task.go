package task

import (
	"log"

	"github.com/nzai/stockrecorder/stock"
)

//	启动任务
func StartTasks() error {
	log.Print("启动任务")
	channel := make(chan int)

	//timer := time.NewTimer(time.Minute * 1)
	go func() {
		err := stock.GetToday()
		if err != nil {
			log.Fatal(err)
		}

		channel <- 0
	}()

	<-channel

	return nil
}
