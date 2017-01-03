package store

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/nzai/stockrecorder/market"
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

// objectKey 存储路径
func (s AliyunOSS) objectKey(_market market.Market, date time.Time) string {
	return fmt.Sprintf("%s%s/%s.mdq", s.config.KeyRoot, date.Format("2006/01/02"), strings.ToLower(_market.Name()))
}

// Exists 判断是否存在
func (s AliyunOSS) Exists(_market market.Market, date time.Time) (bool, error) {
	return s.bucket.IsObjectExist(s.objectKey(_market, date))
}

// Save 保存
func (s AliyunOSS) Save(quote market.DailyQuote) error {

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

	// 上传
	return s.bucket.PutObject(s.objectKey(quote.Market, quote.Date), bytes.NewReader(zipped))
}

// Load 读取
func (s AliyunOSS) Load(_market market.Market, date time.Time) (market.DailyQuote, error) {

	mdq := market.DailyQuote{Market: _market, Date: date}

	readCloser, err := s.bucket.GetObject(s.objectKey(_market, date))
	if err != nil {
		return mdq, err
	}
	defer readCloser.Close()

	reader, err := gzip.NewReader(readCloser)
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
