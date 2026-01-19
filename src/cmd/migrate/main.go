package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/uptrace/bun/migrate"
	"github.com/urfave/cli/v2"

	"go-pg-atlas-bun/internal/infrastructure/database"
	"go-pg-atlas-bun/internal/infrastructure/migration/migrations"
)

func main() {
	// DBインスタンスを取得
	db := database.NewBunDB()
	defer db.Close()

	// マイグレーション用ツール設定
	migrator := migrate.NewMigrator(db, migrations.Migrations)

	app := &cli.App{
		Name:  "migrate",
		Usage: "database migrations tool",
		Commands: []*cli.Command{
			// 初期化用（マイグレーション管理用テーブルの作成。最初に一度だけ実行する。）
			{
				Name:   "init",
				Usage:  "create migration tables",
				Action: func(c *cli.Context) error { return migrator.Init(c.Context) },
			},
			// SQLファイル作成用（upとdown用の二つを作成する。）
			{
				Name:  "create_sql",
				Usage: "create up and down SQL migrations",
				Action: func(c *cli.Context) error {
					name := strings.Join(c.Args().Slice(), "_")
					files, err := migrator.CreateSQLMigrations(c.Context, name)
					if err != nil {
						return err
					}
					for _, f := range files {
						fmt.Printf("created %s (%s)\n", f.Name, f.Path)
					}
					return nil
				},
			},
			// マイグレーション状態の確認
			{
				Name:  "status",
				Usage: "print migrations status",
				Action: func(c *cli.Context) error {
					ms, err := migrator.MigrationsWithStatus(c.Context)
					if err != nil {
						return err
					}
					fmt.Printf("migrations: %s\n", ms)
					fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
					fmt.Printf("last migration group: %s\n", ms.LastGroup())
					return nil
				},
			},
			// マイグレーションの実行
			// （同一タイミングで実行したSQLファイルを一つのグループとして管理している。）
			{
				Name:  "migrate",
				Usage: "migrate database",
				Action: func(c *cli.Context) error {
					group, err := migrator.Migrate(c.Context)
					if err != nil {
						return err
					}
					if group.IsZero() {
						fmt.Println("no new migrations")
						return nil
					}
					fmt.Printf("migrated to %s\n", group)
					return nil
				},
			},
			// ロールバックの実行
			// （グループ単位で管理していて、それを一つ前に戻す。）
			{
				Name:  "rollback",
				Usage: "rollback the last migration group",
				Action: func(c *cli.Context) error {
					group, err := migrator.Rollback(c.Context)
					if err != nil {
						return err
					}
					if group.IsZero() {
						fmt.Println("no groups to rollback")
						return nil
					}
					fmt.Printf("rolled back %s\n", group)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
