package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

// NewDB はBunのDB接続インスタンスを生成します
func NewBunDB() *bun.DB {
	// 環境変数「ENV」の値を取得
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}

	// DB接続用の環境変数の設定
	var dbHost, dbPort, dbName, dbUser, dbPassword, dbMaxOpenCons, dbMaxIdleCons, dbConnMaxLifetime string

	if dbHost = os.Getenv("DB_HOST"); dbHost == "" {
		dbHost = "localhost"
	}

	if dbPort = os.Getenv("DB_PORT"); dbPort == "" {
		dbPort = "5432"
	}

	if env == "testing" {
		dbName = "testing-db"
	} else {
		if dbName = os.Getenv("DB_NAME"); dbName == "" {
			dbName = "local-db-name"
		}
	}

	if dbUser = os.Getenv("DB_USER"); dbUser == "" {
		dbUser = "local-db-user"
	}

	if dbPassword = os.Getenv("DB_PASSWORD"); dbPassword == "" {
		dbPassword = "local-db-password"
	}

	if dbMaxOpenCons = os.Getenv("DB_MAX_OPEN_CONNS"); dbMaxOpenCons == "" {
		dbMaxOpenCons = "20"
	}

	if dbMaxIdleCons = os.Getenv("DB_MAX_IDLE_CONNS"); dbMaxIdleCons == "" {
		dbMaxIdleCons = "10"
	}

	if dbConnMaxLifetime = os.Getenv("DB_CONN_MAX_LIFETIME"); dbConnMaxLifetime == "" {
		dbConnMaxLifetime = "5"
	}

	// DSN設定
	var dsn string
	if env == "production" {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	} else {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	}

	// sql.DB の初期化（pgdriverを使用）
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	// コネクションプールの設定
	maxOpenCons, _ := strconv.Atoi(dbMaxOpenCons) // 最大接続数
	sqldb.SetMaxOpenConns(maxOpenCons)

	maxIdleCons, _ := strconv.Atoi(dbMaxIdleCons) // アイドル時の保持接続数
	sqldb.SetMaxIdleConns(maxIdleCons)

	maxLifetime, _ := strconv.Atoi(dbConnMaxLifetime) // 接続の寿命（分）
	sqldb.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Minute)

	// BunDBのインスタンス生成
	db := bun.NewDB(sqldb, pgdialect.New())

	// デバッグログ設定（SQL表示）
	if env != "production" {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithEnabled(true), // デバッグログ機能の有効化
			bundebug.WithVerbose(true), // ログ詳細表示を有効化
		))
	}

	return db
}
