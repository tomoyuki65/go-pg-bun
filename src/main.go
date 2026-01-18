package main

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	// echoのルーター設定
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		// レスポンス結果の設定
		res := map[string]string{
			"message": "Hello World !!",
		}

		return c.JSON(http.StatusOK, res)
	})

	// ログ出力
	slog.Info("start go-pg-atlas-bun")

	// サーバー起動
	e.Logger.Fatal(e.Start(":8080"))
}
