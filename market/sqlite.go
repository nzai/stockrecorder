package market

import (
	"database/sql"
	"path/filepath"
	"strings"
	"time"

	"github.com/nzai/go-utility/db/sqlite"

	_ "github.com/mattn/go-sqlite3"
)

//	获取数据库连接
func getDB(market Market, code string) (*sql.DB, error) {

	filePath := filepath.Join("config.Get().DataDir", market.Name(), strings.ToLower(code)+".db")
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
		"company":           `CREATE TABLE [company] ([code] VARCHAR(20) NOT NULL, [name] VARCHAR(200) NOT NULL, CONSTRAINT [] PRIMARY KEY ([code]));`,
		"pre":               `CREATE TABLE [pre] ([code] VARCHAR(20) NOT NULL, [time] DATETIME NOT NULL, [open] FLOAT(20, 2) NOT NULL, [close] FLOAT(20, 2) NOT NULL, [high] FLOAT(20, 2) NOT NULL, [low] FLOAT(20, 2) NOT NULL, [volume] INTEGER NOT NULL);`,
		"regular":           `CREATE TABLE [regular] ([code] VARCHAR(20) NOT NULL, [time] DATETIME NOT NULL, [open] FLOAT(20, 2) NOT NULL, [close] FLOAT(20, 2) NOT NULL, [high] FLOAT(20, 2) NOT NULL, [low] FLOAT(20, 2) NOT NULL, [volume] INTEGER NOT NULL);`,
		"post":              `CREATE TABLE [post] ([code] VARCHAR(20) NOT NULL, [time] DATETIME NOT NULL, [open] FLOAT(20, 2) NOT NULL, [close] FLOAT(20, 2) NOT NULL, [high] FLOAT(20, 2) NOT NULL, [low] FLOAT(20, 2) NOT NULL, [volume] INTEGER NOT NULL);`,
		"post_code_time":    `CREATE INDEX [post_code_time] ON [post] ([code], [time]);`,
		"pre_code_time":     `CREATE INDEX [pre_code_time] ON [pre] ([code], [time]);`,
		"regular_code_time": `CREATE INDEX [regular_code_time] ON [regular] ([code], [time]);`,
	}

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
	found, err := sqlite.IsExists(db, tableName)
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

	filePath := filepath.Join("config.Get().DataDir", market.Name(), strings.ToLower(code)+".db")
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
