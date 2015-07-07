package config

import (
	"os"

	"github.com/Unknwon/goconfig"
)

const (
	configFileName     = "config.ini"
	configSection      = "path"
	configKey          = "datadir"
	configDefaultValue = "data"
)

type Config struct {
	filename   string
	configFile *goconfig.ConfigFile
}

var configInstance = New()

//	默认
func New() *goconfig.ConfigFile {
	configFile, err := goconfig.LoadConfigFile(configFileName)
	if err != nil {
		return nil
	}

	return configFile
}

//	设置配置文件
func SetConfigFile(filePath string) error {

	configFile, err := goconfig.LoadConfigFile(filePath)
	if err != nil {
		return err
	}

	configInstance = configFile

	return nil
}

//	获取配置
func GetString(section, key, defaultValue string) string {
	return configInstance.MustValue(section, key, defaultValue)
}

//	获取数据保存目录
func GetDataDir() (string, error) {

	//	数据保存目录
	dataDir := GetString(configSection, configKey, configDefaultValue)

	//	检查目录是否存在
	_, err := os.Stat(dataDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dataDir, 0x777)
		if err != nil {
			return "", err
		}
	}

	return dataDir, nil
}
