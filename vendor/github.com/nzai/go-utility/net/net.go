package net

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	gio "github.com/nzai/go-utility/io"
)

// DownloadString 发送GET请求并返回字符串
func DownloadString(url string) (string, error) {
	return DownloadStringReferer(url, "")
}

// DownloadStringReferer 发送GET请求并返回字符串(带Referer)
func DownloadStringReferer(url, referer string) (string, error) {
	return DownloadStringRefererRetry(url, referer, 1, 0)
}

// DownloadStringRetry 访问网址并返回字符串
func DownloadStringRetry(url string, retryTimes int, interval time.Duration) (string, error) {
	return DownloadStringRefererRetry(url, "", retryTimes, interval)
}

// DownloadStringRefererRetry 访问网址并返回字符串
func DownloadStringRefererRetry(url, referer string, retryTimes int, interval time.Duration) (string, error) {
	buffer, err := DownloadBufferRefererRetry(url, referer, retryTimes, interval)

	return string(buffer), err
}

// DownloadBuffer 发送GET请求并返回缓冲区
func DownloadBuffer(url string) ([]byte, error) {
	return DownloadBufferReferer(url, "")
}

// DownloadBufferReferer 发送GET请求并返回缓冲区(带Referer)
func DownloadBufferReferer(url, referer string) ([]byte, error) {
	return DownloadBufferRefererRetry(url, referer, 1, 0)
}

// DownloadBufferRetry 访问网址并返回缓冲区
func DownloadBufferRetry(url string, retryTimes int, interval time.Duration) ([]byte, error) {
	return DownloadBufferRefererRetry(url, "", retryTimes, interval)
}

// DownloadBufferRefererRetry 访问网址并返回缓冲区
func DownloadBufferRefererRetry(url, referer string, retryTimes int, interval time.Duration) ([]byte, error) {
	var err error
	buffer := []byte{}
	for times := retryTimes - 1; times >= 0; times-- {

		buffer, err = DownloadBufferRefererOnce(url, referer)
		if err == nil {
			return buffer, err
		}

		if times > 0 {
			if err != nil {
				log.Printf("访问%s出错，还有%d次重试机会，%d秒后重试:%s", url, times, int64(interval.Seconds()), err.Error())
			}

			//	延时
			time.Sleep(interval)
		}
	}

	return nil, fmt.Errorf("访问%s出错，已重试%d次，不再重试:%s", url, retryTimes, err.Error())
}

// DownloadBufferRefererOnce 访问网址并返回缓冲区
func DownloadBufferRefererOnce(url, referer string) ([]byte, error) {
	//	构造请求
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	//	引用页
	if referer != "" {
		request.Header.Set("Referer", referer)
	}

	//	发送请求
	client := &http.Client{}
	//	client.Timeout = time.Second * 20
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	//	读取结果
	return ioutil.ReadAll(response.Body)
}

// DownloadFile 下载文件
func DownloadFile(url, path string) error {
	return DownloadFileReferer(url, "", path)
}

// DownloadFileReferer 下载文件(带Referer)
func DownloadFileReferer(url, referer, path string) error {
	return DownloadFileRefererRetry(url, referer, path, 1, 0)
}

// DownloadFileRetry 下载文件
func DownloadFileRetry(url, path string, retryTimes int, interval time.Duration) error {
	return DownloadFileRefererRetry(url, "", path, retryTimes, interval)
}

// DownloadFileRefererRetry 下载文件
func DownloadFileRefererRetry(url, referer, path string, retryTimes int, interval time.Duration) error {

	err := fmt.Errorf("ok")
	tempPath := path + ".downloading"
	for times := retryTimes - 1; times >= 0; times-- {

		err = downloadFileRefererOnce(url, referer, tempPath)
		if err == nil {
			return os.Rename(tempPath, path)
		}

		if times > 0 {
			if err != nil {
				log.Printf("下载%s出错，还有%d次重试机会，%d秒后重试:%s", url, times, int64(interval.Seconds()), err.Error())
			}

			//	延时
			time.Sleep(interval)
		}
	}

	//	删除临时文件
	os.Remove(tempPath)

	return fmt.Errorf("下载%s出错，已重试%d次，不再重试:%s", url, retryTimes, err.Error())
}

//	下载文件
func downloadFileRefererOnce(url, referer, path string) error {
	//	构造请求
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	//	引用页
	if referer != "" {
		request.Header.Set("Referer", referer)
	}

	//	发送请求
	client := &http.Client{}
	//	client.Timeout = time.Second * 15
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	//	tempPath := path + ".downloading"

	//	打开文件
	file, err := gio.OpenForWrite(path)
	if err != nil {
		return err
	}
	defer file.Close()

	//	写文件
	_, err = io.Copy(file, response.Body)
	return err
}
