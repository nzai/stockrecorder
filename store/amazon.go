package store

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"strings"
	"time"

	"github.com/nzai/stockrecorder/market"

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
		svc:    s3.New(sess, aws.NewConfig().WithRegion(s3config.Bucket)),
	}
}

// Exists 判断某天的数据是否存在
func (s AmazonS3) Exists(_market market.Market, date time.Time) (bool, error) {

	_, err := s.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(s.savePath(_market, date)),
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
func (s AmazonS3) Save(quote market.DailyQuote) error {

	// gzip 最高压缩
	buffer := bytes.NewBufferString("")
	w, err := gzip.NewWriterLevel(buffer, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.Write(quote.Marshal())
	if err != nil {
		return err
	}
	w.Flush()

	br := bytes.NewReader(buffer.Bytes())

	// 上传
	_, err = s.svc.PutObject(&s3.PutObjectInput{
		Bucket:       aws.String(s.config.Bucket),
		Key:          aws.String(s.savePath(quote.Market, quote.Date)),
		Body:         br,
		StorageClass: aws.String(s3.ObjectStorageClassReducedRedundancy),
	})

	return err
}

// savePath 保存到S3的路径
func (s AmazonS3) savePath(_market market.Market, date time.Time) string {
	return fmt.Sprintf("%s%s/%s.mdq", s.config.KeyRoot, date.Format("2006/01/02"), strings.ToLower(_market.Name()))
}
