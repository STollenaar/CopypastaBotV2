CREATE TABLE IF NOT EXISTS speak_queue_chunks (
    id       INTEGER PRIMARY KEY,
    queue_id INTEGER NOT NULL,
    chunk_index INTEGER NOT NULL,
    audio_data BLOB NOT NULL,
    is_last_chunk BOOLEAN NOT NULL DEFAULT FALSE
);
