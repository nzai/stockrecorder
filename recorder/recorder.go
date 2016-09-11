package recorder

import (
	"log"
	"sync"
	"time"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/market"
)

// Recorder 股票记录器
type Recorder struct {
	recorders []marketRecorder // 记录器
}

// NewRecorder 新建Recorder
func NewRecorder(conf *config.Config, markets ...market.Market) *Recorder {

	recorders := make([]marketRecorder, len(markets))
	for index, m := range markets {
		recorders[index] = marketRecorder{conf, m}
	}

	return &Recorder{recorders}
}

// Run 执行
func (r Recorder) Run() {
	var wg sync.WaitGroup

	for _, m := range r.recorders {
		go func(mr marketRecorder) {
			wg.Add(1)
			// 启动记录器
			mr.RunAndWait()
			wg.Done()
		}(m)
	}

	wg.Wait()
}

// marketRecorder 市场记录器
type marketRecorder struct {
	config        *config.Config // 配置
	market.Market                // 市场
}

// RunAndWait 启动市场记录器
func (mr marketRecorder) RunAndWait() {
	go func() {
		err := mr.getHistoryData()
		if err != nil {
			log.Printf("[%s]获取历史数据时发生错误: %v", mr.Name(), err)
		}
	}()
}

// durationToNextDay 现在到明天0点的时间间隔
func (mr marketRecorder) durationToNextDay() (time.Duration, error) {
	//	本地时间
	now := time.Now()
	_, offsetLocal := now.Zone()

	//	获取市场所在时区
	location, err := time.LoadLocation(mr.Timezone())
	if err != nil {
		return 0, err
	}

	//	市场所处时区当前时间
	marketNow := now.In(location)
	_, offsetMarket := marketNow.Zone()

	// 现在到该市场所处时区0点还有多少秒
	durationSeconds := int64(offsetMarket - offsetLocal)
	if durationSeconds < 0 {
		durationSeconds += 24 * 60 * 60
	}

	return time.Second * time.Duration(durationSeconds), nil
}

// getHistoryData 获取历史数据
func (mr marketRecorder) getHistoryData() error {
	return nil
}

// startDailyTask 启动每日任务
func (mr marketRecorder) startDailyTask() error {
	return nil
}
