package config

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/nzai/go-utility/io"
	"github.com/nzai/go-utility/path"
)

const (
	defaultConfigFile = "config.yaml"
)

var (
	// 配置文件路径
	configPath = flag.String("c", defaultConfigFile, "yaml配置文件路径")
)

// Config 配置
type Config struct {
	TempPath string `yaml:"tempPath"`
}

// getConfigFilePath 获取配置文件路径
func getConfigFilePath() (string, error) {

	flag.Parse()

	if strings.Compare(*configPath, defaultConfigFile) != 0 {
		return *configPath, nil
	}

	//	启动目录
	startupDir, err := path.GetStartupDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(startupDir, *configPath), nil
}

// Parse 解析配置
func Parse() (*Config, error) {

	// 获取配置文件路径
	filePath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	if !io.IsExists(filePath) {
		return nil, fmt.Errorf("配置文件 %s 不存在", filePath)
	}

	//	读取文件
	buffer, err := io.ReadAllBytes(filePath)
	if err != nil {
		return nil, err
	}

	//	解析配置项
	conf := &Config{}
	err = yaml.Unmarshal(buffer, conf)

	return conf, nil
}
