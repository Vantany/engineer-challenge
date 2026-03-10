-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS reset_tokens (
    token_hash  TEXT PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_reset_tokens_user_id ON reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_reset_tokens_expires_at ON reset_tokens(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reset_tokens_expires_at;
DROP INDEX IF EXISTS idx_reset_tokens_user_id;
DROP TABLE IF EXISTS reset_tokens CASCADE;
-- +goose StatementEnd


