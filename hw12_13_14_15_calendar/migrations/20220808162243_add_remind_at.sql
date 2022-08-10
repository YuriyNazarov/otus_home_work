-- +goose Up
-- +goose StatementBegin
alter table events add column if not exists remind_at timestamp;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table events drop column remind_at;
-- +goose StatementEnd
