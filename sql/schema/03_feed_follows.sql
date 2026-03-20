-- +goose Up
CREATE TABLE feed_follows (
    id uuid primary key,
    created_at TIMESTAMP not null,
    updated_at TIMESTAMP not null,
    user_id uuid  not null,
    feed_id uuid  not null,

    CONSTRAINT feed_user UNIQUE (user_id,feed_id),
    CONSTRAINT user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT feed_id FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
    
);

-- +goose Down
DROP TABLE feed_follows;