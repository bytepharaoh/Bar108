-- +goose Up
ALTER TABLE users
ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
-- +goose Down
ALTER TABLE users
DROP COLUMN is_active;