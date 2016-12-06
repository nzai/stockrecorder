package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/nzai/go-utility/io"
	"github.com/nzai/go-utility/path"
	yaml "gopkg.in/yaml.v2"

	"github.com/nzai/stockrecorder/store"
)

var (
	// defaultConfigFileName 默认的配置文件名
	defaultConfigFileName = "config.yaml"
)

// Config 配置
type Config struct {
	Aliyun struct {
		OSS store.AliyunOSSConfig `yaml:"oss"`
	} `yaml:"aliyun"`
}

// parseConfig 解析配置
func parseConfig() (*Config, error) {

	configPath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	log.Printf("开始解析配置，配置文件路径: %s", configPath)

	//	读取文件
	buffer, err := io.ReadAllBytes(configPath)
	if err != nil {
		return nil, err
	}

	//	解析配置项
	config := new(Config)
	err = yaml.Unmarshal(buffer, config)

	return config, err
}

// getConfigFilePath 获取配置文件路径
func getConfigFilePath() (string, error) {

	if len(os.Args) > 1 {
		// 指定了配置文件
		return os.Args[1], nil
	}

	// 获取启动路径，默认情况配置文件和执行文件放在同一目录
	startupPath, err := path.GetStartupDir()
	if err != nil {
		return "", err
	}

	// 默认配置文件路径
	return filepath.Join(startupPath, defaultConfigFileName), nil
}
