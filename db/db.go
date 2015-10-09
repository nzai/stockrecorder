package db

import (
	"gopkg.in/mgo.v2"
	"github.com/nzai/stockrecorder/config"
)

var startSession *mgo.Session = nil

//	获取数据库连接
func Get() (*mgo.Session, error) {
	if startSession == nil {
		session, err := mgo.Dial(config.Get().MongoUrl)
		if err != nil {
			return nil, err
		}
		
		startSession = session
	}
	
	return startSession.Clone(), nil
}