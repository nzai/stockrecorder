package quote

import (
	"fmt"
	"io"

	"github.com/nzai/go-utility/io/ioutil"
	"go.uber.org/zap"
)

// Quote 报价
type Quote struct {
	Timestamp uint64
	Open      float32
	Close     float32
	High      float32
	Low       float32
	Volume    uint64
}

// Marshal 序列化
func (q Quote) Marshal(w io.Writer) error {

	bw := ioutil.NewBinaryWriter(w)

	err := bw.UInt64(q.Timestamp)
	if err != nil {
		zap.L().Error("write quote timestamp failed", zap.Error(err), zap.Uint64("timestamp", q.Timestamp))
		return err
	}

	err = bw.Float32(q.Open)
	if err != nil {
		zap.L().Error("write quote open failed", zap.Error(err), zap.Float32("open", q.Open))
		return err
	}

	err = bw.Float32(q.Close)
	if err != nil {
		zap.L().Error("write quote close failed", zap.Error(err), zap.Float32("close", q.Close))
		return err
	}

	err = bw.Float32(q.High)
	if err != nil {
		zap.L().Error("write quote open failed", zap.Error(err), zap.Float32("high", q.High))
		return err
	}

	err = bw.Float32(q.Low)
	if err != nil {
		zap.L().Error("write quote open failed", zap.Error(err), zap.Float32("low", q.Low))
		return err
	}

	err = bw.UInt64(q.Volume)
	if err != nil {
		zap.L().Error("write quote open failed", zap.Error(err), zap.Uint64("volume", q.Volume))
		return err
	}

	return nil
}

// Unmarshal 反序列化
func (q *Quote) Unmarshal(r io.Reader) error {

	br := ioutil.NewBinaryReader(r)

	timestamp, err := br.UInt64()
	if err != nil {
		zap.L().Error("read quote timestamp failed", zap.Error(err))
		return err
	}

	open, err := br.Float32()
	if err != nil {
		zap.L().Error("read quote open failed", zap.Error(err))
		return err
	}

	_close, err := br.Float32()
	if err != nil {
		zap.L().Error("read quote close failed", zap.Error(err))
		return err
	}

	high, err := br.Float32()
	if err != nil {
		zap.L().Error("read quote high failed", zap.Error(err))
		return err
	}

	low, err := br.Float32()
	if err != nil {
		zap.L().Error("read quote low failed", zap.Error(err))
		return err
	}

	volume, err := br.UInt64()
	if err != nil {
		zap.L().Error("read quote volume failed", zap.Error(err))
		return err
	}

	q.Timestamp = timestamp
	q.Open = open
	q.Close = _close
	q.High = high
	q.Low = low
	q.Volume = volume

	return nil
}

// Equal 是否相同
func (q Quote) Equal(s Quote) error {

	if q.Timestamp != s.Timestamp {
		return fmt.Errorf("quote timestamp %d is different from %d", q.Timestamp, s.Timestamp)
	}

	if q.Open != s.Open {
		return fmt.Errorf("quote open %.2f is different from %.2f", q.Open, s.Open)
	}

	if q.Close != s.Close {
		return fmt.Errorf("quote close %.2f is different from %.2f", q.Close, s.Close)
	}

	if q.High != s.High {
		return fmt.Errorf("quote high %.2f is different from %.2f", q.High, s.High)
	}

	if q.Low != s.Low {
		return fmt.Errorf("quote low %.2f is different from %.2f", q.Low, s.Low)
	}

	if q.Volume != s.Volume {
		return fmt.Errorf("quote volume %d is different from %d", q.Volume, s.Volume)
	}

	return nil
}
