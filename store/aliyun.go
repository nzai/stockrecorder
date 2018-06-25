package store

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/nzai/stockrecorder/quote"
	"go.uber.org/zap"
)

// AliyunOSSConfig 阿里云对象存储服务配置
type AliyunOSSConfig struct {
	EndPoint        string `yaml:"endpoint"` // Key
	AccessKeyID     string `yaml:"id"`       // ID
	AccessKeySecret string `yaml:"secret"`   // Key
	Bucket          string `yaml:"bucket"`   // Bucket
	KeyRoot         string `yaml:"root"`     // Root
}

// AliyunOSS 阿里云对象存储服务
type AliyunOSS struct {
	config AliyunOSSConfig
	bucket *oss.Bucket
}

// NewAliyunOSS 新建阿里云对象存储服务
func NewAliyunOSS(config AliyunOSSConfig) *AliyunOSS {

	client, err := oss.New(config.EndPoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		log.Fatal(err)
	}

	bucket, err := client.Bucket(config.Bucket)
	if err != nil {
		log.Fatal(err)
	}

	return &AliyunOSS{config: config, bucket: bucket}
}

// objectKey 路径
func (s AliyunOSS) objectKey(exchange *quote.Exchange, date time.Time) string {
	return fmt.Sprintf("%s%s/%s.mdq", s.config.KeyRoot, date.Format("2006/01/02"), strings.ToLower(exchange.Code))
}

// Exists 判断是否存在
func (s AliyunOSS) Exists(exchange *quote.Exchange, date time.Time) (bool, error) {
	return s.bucket.IsObjectExist(s.objectKey(exchange, date))
}

// Save 保存
func (s AliyunOSS) Save(quote *quote.ExchangeDailyQuote) error {

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

	// 上传
	key := s.objectKey(quote.Exchange, quote.Date)
	err = s.bucket.PutObject(key, bytes.NewReader(buffer.Bytes()))
	if err != nil {
		zap.L().Error("upload quote gzip failed", zap.Error(err), zap.String("bucket", s.bucket.BucketName), zap.String("key", key))
		return err
	}

	return nil
}

// Load 读取
func (s AliyunOSS) Load(exchange *quote.Exchange, date time.Time) (*quote.ExchangeDailyQuote, error) {

	key := s.objectKey(exchange, date)
	rc, err := s.bucket.GetObject(key)
	if err != nil {
		zap.L().Error("load quote failed", zap.Error(err), zap.String("bucket", s.bucket.BucketName), zap.String("key", key))
		return nil, err
	}
	defer rc.Close()

	gr, err := gzip.NewReader(rc)
	if err != nil {
		zap.L().Error("read quote gzip failed", zap.Error(err), zap.String("bucket", s.bucket.BucketName), zap.String("key", key))
		return nil, err
	}
	defer gr.Close()

	edq := new(quote.ExchangeDailyQuote)
	err = edq.Unmarshal(gr)
	if err != nil {
		zap.L().Error("unmarshal quote failed", zap.Error(err), zap.String("bucket", s.bucket.BucketName), zap.String("key", key))
		return nil, err
	}

	return edq, nil
}
