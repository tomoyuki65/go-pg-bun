package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"go-pg-bun/internal/infrastructure/database"
	"go-pg-bun/internal/infrastructure/database/schema"
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

	// サンプルAPI（CRUD処理）を追加
	apiV1 := e.Group("/api/v1")

	// ユーザー作成
	apiV1.POST("/users", func(c echo.Context) error {
		// リクエストボディの取得
		type CreateUsersRequestBody struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		var reqBody CreateUsersRequestBody
		if err := c.Bind(&reqBody); err != nil {
			return err
		}

		//　DBインスタンスの取得
		db := database.NewBunDB()
		defer db.Close()

		// ユーザー作成処理
		user := schema.UsersSchema{
			Name:  reqBody.Name,
			Email: reqBody.Email,
		}
		_, err := db.NewInsert().Model(&user).Exec(c.Request().Context())
		if err != nil {
			errMsg := fmt.Sprintf("failed to create user: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, errMsg)
		}

		return c.JSON(http.StatusCreated, user)
	})

	// 全てのユーザー取得
	apiV1.GET("/users", func(c echo.Context) error {
		//　DBインスタンスの取得
		db := database.NewBunDB()
		defer db.Close()

		// 全てのユーザー取得処理
		var users []schema.UsersSchema
		err := db.NewSelect().Model(&users).Scan(c.Request().Context())
		if err != nil {
			return err
		}

		// データが0件の場合、空の配列を設定
		if len(users) == 0 {
			users = []schema.UsersSchema{}
		}

		return c.JSON(http.StatusOK, users)
	})

	// 全てのユーザー取得（SQL版）
	apiV1.GET("/users/sql", func(c echo.Context) error {
		//　DBインスタンスの取得
		db := database.NewBunDB()
		defer db.Close()

		// 全てのユーザー取得処理
		var users []schema.UsersSchema
		err := db.NewRaw("SELECT * FROM users").Scan(c.Request().Context(), &users)
		if err != nil {
			return err
		}

		// データが0件の場合、空の配列を設定
		if len(users) == 0 {
			users = []schema.UsersSchema{}
		}

		return c.JSON(http.StatusOK, users)
	})

	// 対象のユーザー取得
	apiV1.GET("/users/:id", func(c echo.Context) error {
		// リクエストパラメータの取得
		id := c.Param("id")
		if id == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "id is required")
		}

		//　DBインスタンスの取得
		db := database.NewBunDB()
		defer db.Close()

		// 対象ユーザー取得処理
		var user schema.UsersSchema
		err := db.NewSelect().Model(&user).Where("id = ?", id).Scan(c.Request().Context())
		if err != nil {
			// 対象データが存在しない場合は空のオブジェクトを返す
			if errors.Is(err, sql.ErrNoRows) {
				return c.JSON(http.StatusOK, map[string]interface{}{})
			}
			return err
		}

		return c.JSON(http.StatusOK, user)
	})

	// 対象ユーザー取得（SQL版）
	apiV1.GET("/users/:id/sql", func(c echo.Context) error {
		// リクエストパラメータの取得
		id := c.Param("id")
		if id == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "id is required")
		}

		//　DBインスタンスの取得
		db := database.NewBunDB()
		defer db.Close()

		// 対象ユーザー取得処理
		var user schema.UsersSchema
		err := db.NewRaw("SELECT * FROM users WHERE id = ?", id).Scan(c.Request().Context(), &user)
		if err != nil {
			// 対象データが存在しない場合は空のオブジェクトを返す
			if errors.Is(err, sql.ErrNoRows) {
				return c.JSON(http.StatusOK, map[string]interface{}{})
			}
			return err
		}

		return c.JSON(http.StatusOK, user)
	})

	// 対象ユーザー更新
	apiV1.PUT("/users/:id", func(c echo.Context) error {
		// リクエストパラメータの取得
		id := c.Param("id")
		if id == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "id is required")
		}

		// リクエストボディの取得
		type UpdateUsersRequestBody struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		var reqBody UpdateUsersRequestBody
		if err := c.Bind(&reqBody); err != nil {
			return err
		}

		//　DBインスタンスの取得
		db := database.NewBunDB()
		defer db.Close()

		// DB操作（トランザクション有り）
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// 対象ユーザー取得
		var user schema.UsersSchema
		err = tx.NewSelect().Model(&user).Where("id = ?", id).Scan(c.Request().Context())
		if err != nil {
			// 対象データが存在しない場合
			if errors.Is(err, sql.ErrNoRows) {
				return echo.NewHTTPError(http.StatusNotFound, "user not found")
			}
			return err
		}

		// 更新値の設定
		if reqBody.Name != "" {
			user.Name = reqBody.Name
		}
		if reqBody.Email != "" {
			user.Email = reqBody.Email
		}

		// 更新処理
		_, err = tx.NewUpdate().Model(&user).Where("id = ?", id).Returning("*").Exec(c.Request().Context())
		if err != nil {
			errMsg := fmt.Sprintf("failed to update user: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, errMsg)
		}

		// コミット
		err = tx.Commit()
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, user)
	})

	// 対象ユーザー削除
	apiV1.DELETE("/users/:id", func(c echo.Context) error {
		// リクエストパラメータの取得
		id := c.Param("id")
		if id == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "id is required")
		}

		//　DBインスタンスの取得
		db := database.NewBunDB()
		defer db.Close()

		// DB操作（トランザクション有り）
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// 対象ユーザー取得
		var user schema.UsersSchema
		err = tx.NewSelect().Model(&user).Where("id = ?", id).Scan(c.Request().Context())
		if err != nil {
			// 対象データが存在しない場合
			if errors.Is(err, sql.ErrNoRows) {
				return echo.NewHTTPError(http.StatusNotFound, "user not found")
			}
			return err
		}

		// 削除処理
		_, err = tx.NewDelete().Model(&user).Where("id = ?", id).Exec(c.Request().Context())
		if err != nil {
			errMsg := fmt.Sprintf("failed to delete user: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, errMsg)
		}

		// コミット
		err = tx.Commit()
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	})

	// ログ出力
	slog.Info("start go-pg-bun")

	// サーバー起動
	e.Logger.Fatal(e.Start(":8080"))
}
