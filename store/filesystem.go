package store

import (
	"bytes"
	"compress/gzip"
	"io"
	"path/filepath"
	"strings"
	"time"

	gio "github.com/nzai/go-utility/io"
	"github.com/nzai/stockrecorder/quote"
	"go.uber.org/zap"
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
func (s FileSystem) storePath(exchange *quote.Exchange, date time.Time) string {
	return filepath.Join(
		s.config.StoreRoot,
		date.Format("2006"),
		date.Format("01"),
		date.Format("02"),
		strings.ToLower(exchange.Code)+".mdq",
	)
}

// Exists 判断是否存在
func (s FileSystem) Exists(exchange *quote.Exchange, date time.Time) (bool, error) {
	return gio.IsExists(s.storePath(exchange, date)), nil
}

// Save 保存
func (s FileSystem) Save(quote *quote.ExchangeDailyQuote) error {

	// gzip 最高压缩
	buffer := new(bytes.Buffer)
	gw, err := gzip.NewWriterLevel(buffer, gzip.BestCompression)
	if err != nil {
		zap.L().Error("write quote gzip failed", zap.Error(err), zap.Any("exchange", quote.Exchange), zap.Time("date", quote.Date), zap.Int("companies", len(quote.Companies)))
		return err
	}

	err = quote.Marshal(gw)
	if err != nil {
		zap.L().Error("write quote gzip failed", zap.Error(err), zap.Any("exchange", quote.Exchange), zap.Time("date", quote.Date), zap.Int("companies", len(quote.Companies)))
		return err
	}

	gw.Flush()
	gw.Close()

	// 写盘
	filePath := s.storePath(quote.Exchange, quote.Date)
	file, err := gio.OpenForWrite(filePath)
	if err != nil {
		zap.L().Error("open file for write failed", zap.Error(err), zap.String("path", filePath))
		return err
	}
	defer file.Close()

	_, err = io.Copy(buffer, file)
	if err != nil {
		zap.L().Error("save quote file failed", zap.Error(err), zap.String("path", filePath))
		return err
	}

	return nil
}

// Load 读取
func (s FileSystem) Load(exchange *quote.Exchange, date time.Time) (*quote.ExchangeDailyQuote, error) {

	filePath := s.storePath(exchange, date)
	file, err := gio.OpenForRead(filePath)
	if err != nil {
		zap.L().Error("open file for read failed", zap.Error(err), zap.String("path", filePath))
		return nil, err
	}
	defer file.Close()

	buffer := new(bytes.Buffer)
	_, err = io.Copy(file, buffer)
	if err != nil {
		zap.L().Error("read quote file failed", zap.Error(err), zap.String("path", filePath))
		return nil, err
	}

	gr, err := gzip.NewReader(buffer)
	if err != nil {
		zap.L().Error("read quote gzip failed", zap.Error(err), zap.String("path", filePath))
		return nil, err
	}
	defer gr.Close()

	edq := new(quote.ExchangeDailyQuote)
	err = edq.Unmarshal(gr)
	if err != nil {
		zap.L().Error("unmarshal quote failed", zap.Error(err), zap.String("path", filePath))
		return nil, err
	}

	return edq, nil
}
