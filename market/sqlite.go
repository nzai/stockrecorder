package market

import (
	"database/sql"
	"path/filepath"
	"strings"
	"time"

	"github.com/nzai/go-utility/db/sqlite"
	"github.com/nzai/stockrecorder/config"

	_ "github.com/mattn/go-sqlite3"
)

//	获取数据库连接
func getDB(market Market, code string) (*sql.DB, error) {

	filePath := filepath.Join(config.Get().DataDir, market.Name(), strings.ToLower(code)+".db")
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, err
	}

	//	确保数据表都存在
	err = ensureTables(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

//	保证表结构存在
func ensureTables(db *sql.DB) error {

	tables := map[string]string{
		"process": `CREATE TABLE [process] ([date] CHAR(8) NOT NULL, [success] TINYINT(1) NOT NULL, CONSTRAINT [] PRIMARY KEY ([date]));CREATE INDEX [process_success] ON [process] ([success]);`,
		"pre":     `CREATE TABLE [pre] ([time] DATETIME NOT NULL, [open] FLOAT(20, 3) NOT NULL, [close] FLOAT(20, 3) NOT NULL, [high] FLOAT(20, 3) NOT NULL, [low] FLOAT(20, 3) NOT NULL, [volume] INTEGER NOT NULL, PRIMARY KEY ([time]));`,
		"regular": `CREATE TABLE [regular] ([time] DATETIME NOT NULL, [open] FLOAT(20, 3) NOT NULL, [close] FLOAT(20, 3) NOT NULL, [high] FLOAT(20, 3) NOT NULL, [low] FLOAT(20, 3) NOT NULL, [volume] INTEGER NOT NULL, PRIMARY KEY ([time]));`,
		"post":    `CREATE TABLE [post] ([time] DATETIME NOT NULL, [open] FLOAT(20, 3) NOT NULL, [close] FLOAT(20, 3) NOT NULL, [high] FLOAT(20, 3) NOT NULL, [low] FLOAT(20, 3) NOT NULL, [volume] INTEGER NOT NULL, PRIMARY KEY ([time]));`,
		"error":   `CREATE TABLE [error] ([date] CHAR(8) NOT NULL, [message] TEXT NOT NULL, PRIMARY KEY ([date]));`}

	for name, script := range tables {
		err := ensureTable(db, name, script)
		if err != nil {
			return err
		}
	}

	return nil
}

//	保存单表结构存在
func ensureTable(db *sql.DB, tableName, createScript string) error {

	//	判断表是否存在
	found, err := sqlite.TableExists(db, tableName)
	if err != nil {
		return err
	}

	if found {
		return nil
	}

	//	建表
	_, err = db.Exec(createScript)

	return err
}

//	是否处理过
func isProcessed(tx *sql.Tx, date string) (bool, error) {

	stmt, err := tx.Prepare("select success from process where [date]=?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	result, err := stmt.Query(date)
	if err != nil {
		return false, err
	}
	defer result.Close()

	return result.Next(), nil
}

//	保存处理状态
func saveProcessStatus(tx *sql.Tx, date string, success bool) error {
	stmt, err := tx.Prepare("replace into process values(?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	//	新增
	result, err := stmt.Exec(date, success)
	if err != nil {
		return err
	}

	ra, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if ra == 0 {
		return sql.ErrNoRows
	}

	return nil
}

//	处理分时数据
func savePeroid(tx *sql.Tx, table string, peroid []Peroid60) error {

	if len(peroid) == 0 {
		return nil
	}

	stmt, err := tx.Prepare("replace into " + table + " values(?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range peroid {

		//	新增
		result, err := stmt.Exec(p.Time, p.Open, p.Close, p.High, p.Low, p.Volume)
		if err != nil {
			return err
		}

		ra, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if ra == 0 {
			return sql.ErrNoRows
		}
	}

	return nil
}

//	保存错误信息
func saveError(tx *sql.Tx, date, message string) error {

	stmt, err := tx.Prepare("replace into error values(?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	//	新增
	result, err := stmt.Exec(date, message)
	if err != nil {
		return err
	}

	ra, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if ra == 0 {
		return sql.ErrNoRows
	}

	return nil
}

//	从文件读取分时数据
func loadPeroid(market Market, code string, start, end time.Time, table string) ([]Peroid60, error) {

	filePath := filepath.Join(config.Get().DataDir, market.Name(), strings.ToLower(code)+".db")
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	stmt, err := db.Prepare("select time, open, close, high, low, volume from " + table + " where time >= ? and time <= ? order by time")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	//	查询
	row, err := stmt.Query(start, end)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var _time time.Time
	var open, _close, high, low float32
	var volume int64

	peroids := make([]Peroid60, 0)
	for row.Next() {
		err = row.Scan(&_time, &open, &_close, &high, &low, &volume)
		if err != nil {
			return nil, err
		}

		peroids = append(peroids, Peroid60{market.Name(), code, _time, open, _close, high, low, volume})
	}

	return peroids, nil
}
