CREATE TABLE IF NOT EXISTS speak_queue (
    id INTEGER PRIMARY KEY,
    guild_id VARCHAR NOT NULL,
    user_id VARCHAR NOT NULL,
    content TEXT NOT NULL,
    cmd_type VARCHAR NOT NULL,
    cmd_name VARCHAR NOT NULL,
    app_id VARCHAR NOT NULL,
    token VARCHAR,
    status VARCHAR DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT current_timestamp
);