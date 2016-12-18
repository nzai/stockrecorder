package store

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nzai/go-utility/io"
	"github.com/nzai/stockrecorder/market"
)

// FileSystemConfig 文件系统配置
type FileSystemConfig struct {
	StoreRoot string // 存储根目录
}

// FileSystem 文件系统存储服务
type FileSystem struct {
	config FileSystemConfig
}

// NewFileSystem 新建文件系统存储服务
func NewFileSystem(config FileSystemConfig) *FileSystem {
	return &FileSystem{config: config}
}

// storePath 存储路径
func (s FileSystem) storePath(_market market.Market, date time.Time) string {
	return filepath.Join(
		s.config.StoreRoot,
		date.Format("2006"),
		date.Format("01"),
		date.Format("02"),
		strings.ToLower(_market.Name())+".mdq",
	)
}

// Exists 判断是否存在
func (s FileSystem) Exists(_market market.Market, date time.Time) (bool, error) {
	return io.IsExists(s.storePath(_market, date)), nil
}

// Save 保存
func (s FileSystem) Save(quote market.DailyQuote) error {

	// gzip 最高压缩
	buffer := new(bytes.Buffer)
	w, err := gzip.NewWriterLevel(buffer, gzip.BestCompression)
	if err != nil {
		return err
	}
	_, err = w.Write(quote.Marshal())
	if err != nil {
		return err
	}
	w.Flush()
	w.Close()

	zipped, err := ioutil.ReadAll(buffer)
	if err != nil {
		return err
	}

	// 存盘
	return io.WriteBytes(s.storePath(quote.Market, quote.Date), zipped)
}

// Load 读取
func (s FileSystem) Load(_market market.Market, date time.Time) (market.DailyQuote, error) {

	mdq := market.DailyQuote{Market: _market, Date: date}

	file, err := os.Open(s.storePath(_market, date))
	if err != nil {
		return mdq, err
	}
	defer file.Close()

	reader, err := gzip.NewReader(file)
	if err != nil {
		return mdq, err
	}
	defer reader.Close()

	buffer, err := ioutil.ReadAll(reader)
	if err != nil {
		return mdq, err
	}

	mdq.Unmarshal(buffer)

	return mdq, nil
}
