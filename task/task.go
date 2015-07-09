package task

import (
	"log"
	"time"

	"github.com/nzai/stockrecorder/stock"
)

//	启动任务
func StartTasks() error {
	log.Print("启动任务")
	channel := make(chan int)

	now := time.Now().UTC()
	tomorrow := now.AddDate(0, 0, 1)
	tomorrowZero := tomorrow.Truncate(time.Hour * 24)

	d := tomorrowZero.Sub(now)
	time.AfterFunc(d, func() {
		ticker := time.NewTicker(time.Hour * 24)
		for _ = range ticker.C {
			err := stock.GetToday()
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	<-channel

	return nil
}
