-- 1. usersテーブルの「updated_at」を自動更新するトリガー設定を削除
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- 2. usersテーブルの削除
DROP TABLE IF EXISTS users;
