-- 1. ordersテーブルの「updated_at」を自動更新するトリガー設定を削除
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;

-- 2. ordersテーブルの削除
DROP TABLE IF EXISTS orders;
