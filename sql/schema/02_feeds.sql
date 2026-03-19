-- +goose Up
CREATE TABLE feeds (
    name text not null,
    url text unique not null,
    user_id UUID not null,
    CONSTRAINT user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

);

-- +goose Down
DROP TABLE feeds;
