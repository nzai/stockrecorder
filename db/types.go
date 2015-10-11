package db

import (
	"time"
)

type DailyResult struct {
	Code    string
	Market  string
	Date    time.Time
	Error   bool
	Message string
}

type DailyAnalyzeResult struct {
	DailyResult DailyResult
	Pre         []Peroid60
	Regular     []Peroid60
	Post        []Peroid60
}

type Peroid60 struct {
	Code   string
	Market string
	Start  time.Time
	End    time.Time
	Open   float32
	Close  float32
	High   float32
	Low    float32
	Volume int64
}

type Raw60 struct {
	Code    string
	Market  string
	Date    time.Time
	Json    string
	Status  int
	Message string
}
