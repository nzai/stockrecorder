package io

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

//	发送GET请求并返回结果字符串
func DownloadString(url string) (string, error) {
	return DownloadStringReferer(url, "")
}

//	发送GET请求并返回结果字符串(带Referer)
func DownloadStringReferer(url, referer string) (string, error) {
	return DownloadStringRefererRetry(url, referer, 1, 0)
}

//	访问网址并返回字符串
func DownloadStringRetry(url string, retryTimes, intervalSeconds int) (string, error) {
	return DownloadStringRefererRetry(url, "", retryTimes, intervalSeconds)
}

//	访问网址并返回字符串
func DownloadStringRefererRetry(url, referer string, retryTimes, intervalSeconds int) (string, error) {
	var err error
	for times := retryTimes - 1; times >= 0; times-- {
		//	构造请求
		request, err := http.NewRequest("GET", url, nil)
		if err == nil {
			//	引用页
			if referer != "" {
				request.Header.Set("Referer", referer)
			}

			//	发送请求
			client := &http.Client{}
			response, err := client.Do(request)
			if err == nil {
				defer response.Body.Close()

				//	读取结果
				buffer, err := ioutil.ReadAll(response.Body)
				if err == nil {
					return string(buffer), nil
				}
			}
		}

		if times > 0 {
			log.Fatalf("访问%s出错，还有%d次重试机会，%d秒后重试:%v", url, times, intervalSeconds, err)
			//	延时
			time.Sleep(time.Duration(intervalSeconds) * time.Second)
			log.Fatalf("访问%s出错，还有%d次重试机会，开始重试:%v", url, times)
		}
	}

	return "", fmt.Errorf("访问%s出错，已重试%d次，不再重试:%v", url, retryTimes, err)
}
