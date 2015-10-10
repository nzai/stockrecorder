package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	configFile     = "project.json"
	defaultDataDir = "data"
)

type Config struct {
	RootDir       string
	DataDir       string
	MongoUrl      string
}

//	当前系统配置
var configValue *Config = nil

//	设置配置文件
func SetRootDir(root string) error {

	//	构造配置文件路径
	filePath := filepath.Join(root, configFile)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return err
	}

	//	读取文件
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	//	解析配置项
	configValue = &Config{}
	err = json.Unmarshal(content, configValue)
	if configValue == nil {
		return fmt.Errorf("配置文件错误")
	}
	if err != nil {
		return err
	}

	//	数据目录
	if strings.Trim(configValue.DataDir, " ") == "" {
		configValue.DataDir = defaultDataDir
	}

	if !filepath.IsAbs(configValue.DataDir) {
		configValue.DataDir = filepath.Join(root, configValue.DataDir)
	}

	//	数据目录不存在则创建
	_, err = os.Stat(configValue.DataDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(configValue.DataDir, 0660)
		if err != nil {
			return err
		}
	}

	return nil
}

//	获取当前系统配置
func Get() *Config {
	return configValue
}
