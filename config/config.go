package config

import (
	"os"
	"path/filepath"

	"github.com/Unknwon/goconfig"
)

const (
	dataDirName = "data"
)

type Config struct {
	filename   string
	configFile *goconfig.ConfigFile
}

var configInstance *goconfig.ConfigFile
var dataDir string

//	设置配置文件
func SetConfigFile(filePath string) error {

	configFile, err := goconfig.LoadConfigFile(filePath)
	if err != nil {
		return err
	}

	configInstance = configFile

	rootDir := filepath.Dir(filePath)
	dataDir = filepath.Join(rootDir, dataDirName)
	_, err = os.Stat(dataDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dataDir, 0x644)
		if err != nil {
			return err
		}
	}

	return nil
}

//	获取字符串
func GetString(section, key, defaultValue string) string {
	return configInstance.MustValue(section, key, defaultValue)
}

//	获取字符串数组
func GetArray(section, key string) []string {
	return configInstance.MustValueArray(section, key, ",")
}

//	获取数据保存目录
func GetDataDir() string {

	return dataDir
}
