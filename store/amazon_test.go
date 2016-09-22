package store

import (
	"database/sql"
	"testing"
	// sqlite驱动
	_ "github.com/mattn/go-sqlite3"
)

// func TestSave(t *testing.T) {
//
// 	defer func() {
// 		log.Print("dsf")
// 		debug.PrintStack()
// 	}()
// 	s3 := AmazonS3{}
// 	quote := market.DailyQuote{
// 		Market: market.America{},
// 		Date:   time.Now(),
// 	}
//
// 	err := s3.Save("c:\\data", quote)
// 	if err != nil {
// 		t.Fatalf("保存失败: %v", err)
// 	}
// }

func TestGetSQLiteDB(t *testing.T) {

	db, err := sql.Open("sqlite3", "c:\\test.db")
	if err != nil {
		t.Errorf("打开数据库连接失败: %v", err)
	}
	defer db.Close()

	s3 := AmazonS3{}

	err = s3.createSQLiteTables(db)
	if err != nil {
		t.Errorf("创建表失败: %v", err)
	}
	t.Log(db)
}
