-- 1. order_detailsテーブルの「updated_at」を自動更新するトリガー設定を削除
DROP TRIGGER IF EXISTS update_order_details_updated_at ON order_details;

-- 2. order_detailsテーブルの削除
DROP TABLE IF EXISTS order_details;
