package store

import (
	"database/sql"
	"path/filepath"
	"strings"
	"time"

	"github.com/nzai/go-utility/db/sqlite"
	"github.com/nzai/go-utility/io"
	"github.com/nzai/stockrecorder/market"

	// sqlite驱动
	_ "github.com/mattn/go-sqlite3"
)

// AmazonS3 亚马逊S3存储服务
type AmazonS3 struct{}

// Exists 判断某天的数据是否存在
func (s3 AmazonS3) Exists(tempPath string, _market market.Market, date time.Time) (bool, error) {

	filePath, err := s3.sqliteFilePath(_market, date, tempPath)
	if err != nil {
		return false, err
	}

	return io.IsExists(filePath), nil
}

// Save 保存
func (s3 AmazonS3) Save(tempPath string, quote market.DailyQuote) error {

	// log.Printf("market:%v date:%v temp:%s", quote.Market, quote.Date, tempPath)
	filePath, err := s3.sqliteFilePath(quote.Market, quote.Date, tempPath)
	if err != nil {
		return err
	}

	// 创建SQLite连接
	db, err := s3.getSQLiteDB(filePath)
	if err != nil {
		return err
	}
	defer db.Close()

	// 启动事务
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// 保存到SQLite
	err = s3.saveToSQLite(tx, quote)
	if err != nil {
		// 回滚事务
		tx.Rollback()
		return err
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		// 回滚事务
		tx.Rollback()
		return err
	}

	// 压缩空间
	_, err = db.Exec("vacuum")

	return err
}

// sqliteFilePath SQLite文件路径
func (s3 AmazonS3) sqliteFilePath(_market market.Market, date time.Time, tempPath string) (string, error) {

	// 文件路径形如 ..\2016\09\17\market.db
	dir := filepath.Join(tempPath, date.Format("2006"), date.Format("01"), date.Format("02"))
	err := io.EnsureDir(dir)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, strings.ToLower(_market.Name())+".db"), nil
}

// getDB 获取数据库连接
func (s3 AmazonS3) getSQLiteDB(filePath string) (*sql.DB, error) {

	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, err
	}

	// 建表
	err = s3.createSQLiteTables(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// createSQLiteTables 确保数据表都存在
func (s3 AmazonS3) createSQLiteTables(db *sql.DB) error {

	scripts := map[string]string{
		"company": `CREATE TABLE [company] ([code] VARCHAR(20) NOT NULL, [name] VARCHAR(200) NOT NULL, CONSTRAINT [] PRIMARY KEY ([code]));`,
		"pre":     `CREATE TABLE [pre] ([code] VARCHAR(20) NOT NULL, [time] DATETIME NOT NULL, [open] INT NOT NULL, [close] INT NOT NULL, [high] INT NOT NULL, [low] INT NOT NULL, [volume] INTEGER NOT NULL);CREATE INDEX [pre_code_time] ON [pre] ([code], [time]);`,
		"regular": `CREATE TABLE [regular] ([code] VARCHAR(20) NOT NULL, [time] DATETIME NOT NULL, [open] INT NOT NULL, [close] INT NOT NULL, [high] INT NOT NULL, [low] INT NOT NULL, [volume] INTEGER NOT NULL);CREATE INDEX [regular_code_time] ON [regular] ([code], [time]);`,
		"post":    `CREATE TABLE [post] ([code] VARCHAR(20) NOT NULL, [time] DATETIME NOT NULL, [open] INT NOT NULL, [close] INT NOT NULL, [high] INT NOT NULL, [low] INT NOT NULL, [volume] INTEGER NOT NULL);CREATE INDEX [post_code_time] ON [post] ([code], [time]);`,
	}

	for name, script := range scripts {
		err := s3.createSQLiteTable(db, name, script)
		if err != nil {
			return err
		}
	}

	return nil
}

// createSQLiteTable 保存单表结构存在
func (s3 AmazonS3) createSQLiteTable(db *sql.DB, name, script string) error {

	//	判断表或索引是否存在
	found, err := sqlite.TableOrIndexExists(db, name)
	if err != nil {
		return err
	}

	if found {
		return nil
	}

	//	建表
	_, err = db.Exec(script)

	return err
}

// saveToSQLite 保存数据到SQLite
func (s3 AmazonS3) saveToSQLite(tx *sql.Tx, quote market.DailyQuote) error {

	// 保存公司
	err := s3.saveCompanyToSQLite(tx, quote.Companies)
	if err != nil {
		return err
	}

	// 保存盘前报价
	err = s3.saveQuoteToSQLite(tx, "pre", quote.Pre)
	if err != nil {
		return err
	}

	// 保存盘中报价
	err = s3.saveQuoteToSQLite(tx, "regular", quote.Regular)
	if err != nil {
		return err
	}

	// 保存盘后报价
	err = s3.saveQuoteToSQLite(tx, "post", quote.Post)
	if err != nil {
		return err
	}

	return nil
}

// saveCompanyToSQLite 保存公司到SQLite
func (s3 AmazonS3) saveCompanyToSQLite(tx *sql.Tx, companies []market.Company) error {

	if len(companies) == 0 {
		return nil
	}

	stmt, err := tx.Prepare("replace into company values(?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, company := range companies {

		//	新增
		result, err := stmt.Exec(company.Code, company.Name)
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

// saveQuoteToSQLite 保存报价数据到SQLite
func (s3 AmazonS3) saveQuoteToSQLite(tx *sql.Tx, table string, quotes []market.Quote) error {

	if len(quotes) == 0 {
		return nil
	}

	stmt, err := tx.Prepare("replace into " + table + " values(?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, quote := range quotes {

		//	新增
		result, err := stmt.Exec(quote.Code, quote.Time, int(quote.Open*100), int(quote.Close*100), int(quote.Max*100), int(quote.Min*100), quote.Volume)
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
