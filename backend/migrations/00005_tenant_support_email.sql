-- +goose Up
ALTER TABLE tenants ADD COLUMN support_email citext;

-- +goose Down
ALTER TABLE tenants DROP COLUMN IF EXISTS support_email;
