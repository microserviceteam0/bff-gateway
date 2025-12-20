-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ALTER COLUMN id TYPE BIGINT,
ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);

-- Обновляем последовательность тоже на bigint
ALTER SEQUENCE users_id_seq AS BIGINT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
ALTER COLUMN id TYPE INTEGER,
ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);

ALTER SEQUENCE users_id_seq AS INTEGER;
-- +goose StatementEnd
