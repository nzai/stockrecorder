package store

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"strings"
	"time"

	"github.com/nzai/stockrecorder/quote"
	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// AmazonS3Config 亚马逊S3存储配置
type AmazonS3Config struct {
	AccessKeyID     string `yaml:"id"`      // ID
	SecretAccessKey string `yaml:"secret"`  // Key
	Region          string `yaml:"region"`  // 区域
	Bucket          string `yaml:"bucket"`  // 存储桶
	KeyRoot         string `yaml:"keyroot"` // S3路径根目录
}

// AmazonS3 亚马逊S3存储服务
type AmazonS3 struct {
	config AmazonS3Config
	svc    *s3.S3
}

// NewAmazonS3 亚马逊S3存储服务
func NewAmazonS3(s3config AmazonS3Config) AmazonS3 {

	config := aws.Config{Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
		AccessKeyID:     s3config.AccessKeyID,
		SecretAccessKey: s3config.SecretAccessKey,
	})}

	sess, err := session.NewSession(&config)
	if err != nil {
		panic(err)
	}

	return AmazonS3{
		config: s3config,
		svc:    s3.New(sess, aws.NewConfig().WithRegion(s3config.Region).WithMaxRetries(10)),
	}
}

// savePath 保存到S3的路径
func (s AmazonS3) savePath(exchange *quote.Exchange, date time.Time) string {
	return fmt.Sprintf("%s%s/%s.mdq", s.config.KeyRoot, date.Format("2006/01/02"), strings.ToLower(exchange.Code))
}

// Exists 判断某天的数据是否存在
func (s AmazonS3) Exists(exchange *quote.Exchange, date time.Time) (bool, error) {

	_, err := s.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(s.savePath(exchange, date)),
	})

	if err == nil {
		return true, nil
	}

	ae, ok := err.(awserr.Error)
	if ok && ae.Code() == "NotFound" {
		return false, nil
	}

	return false, err
}

// Save 保存
func (s AmazonS3) Save(quote *quote.ExchangeDailyQuote) error {

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
	_, err = s.svc.PutObject(&s3.PutObjectInput{
		Bucket:       aws.String(s.config.Bucket),
		Key:          aws.String(s.savePath(quote.Exchange, quote.Date)),
		Body:         bytes.NewReader(buffer.Bytes()),
		StorageClass: aws.String(s3.ObjectStorageClassReducedRedundancy),
	})
	if err != nil {
		zap.L().Error("upload quote gzip failed", zap.Error(err), zap.Any("exchange", quote.Exchange), zap.Time("date", quote.Date), zap.Int("companies", len(quote.Companies)))
		return err
	}

	return err
}

// Load 读取
func (s AmazonS3) Load(exchange *quote.Exchange, date time.Time) (*quote.ExchangeDailyQuote, error) {

	filePath := s.savePath(exchange, date)
	output, err := s.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(s.savePath(exchange, date)),
	})
	if err != nil {
		zap.L().Error("load quote failed", zap.Error(err), zap.String("bucket", s.config.Bucket), zap.String("path", filePath))
		return nil, err
	}
	defer output.Body.Close()

	gr, err := gzip.NewReader(output.Body)
	if err != nil {
		zap.L().Error("read quote gzip failed", zap.Error(err), zap.String("bucket", s.config.Bucket), zap.String("path", filePath))
		return nil, err
	}
	defer gr.Close()

	edq := new(quote.ExchangeDailyQuote)
	err = edq.Unmarshal(gr)
	if err != nil {
		zap.L().Error("unmarshal quote failed", zap.Error(err), zap.String("bucket", s.config.Bucket), zap.String("path", filePath))
		return nil, err
	}

	return edq, nil
}
