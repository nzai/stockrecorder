package db

import (
	"time"

	"github.com/nzai/stockrecorder/config"
	"gopkg.in/mgo.v2"
)

var startSession *mgo.Session = nil

//	获取数据库连接
func Get() (*mgo.Session, error) {
	if startSession == nil {
		session, err := mgo.DialWithTimeout(config.Get().MongoUrl, time.Minute)
		if err != nil {
			return nil, err
		}

		startSession = session
	}

	return startSession.Clone(), nil
}
