CREATE TABLE IF NOT EXISTS feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    last_updated DATETIME,
    last_error TEXT,
    last_error_time DATETIME,
    visible BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    etag TEXT,
    last_modified TEXT,
    cache_control_max_age INTEGER
);

CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id INTEGER NOT NULL,
    guid TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    link TEXT NOT NULL DEFAULT '',
    published DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
    UNIQUE(feed_id, guid)
);

CREATE TABLE IF NOT EXISTS read_status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    item_id INTEGER NOT NULL,
    read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at DATETIME,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    UNIQUE(item_id)
);

CREATE TABLE IF NOT EXISTS log_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    level TEXT NOT NULL,
    message TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    attributes TEXT -- JSON string of attributes
);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_items_feed_id ON items(feed_id);
CREATE INDEX IF NOT EXISTS idx_items_published ON items(published);
CREATE INDEX IF NOT EXISTS idx_read_status_item_id ON read_status(item_id);
CREATE INDEX IF NOT EXISTS idx_read_status_read ON read_status(read);
CREATE INDEX IF NOT EXISTS idx_log_messages_timestamp ON log_messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_log_messages_level ON log_messages(level);

CREATE TABLE IF NOT EXISTS feed_folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id INTEGER NOT NULL,
    folder_name TEXT NOT NULL,
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
    UNIQUE(feed_id, folder_name)
);

CREATE INDEX IF NOT EXISTS idx_feed_folders_feed_id ON feed_folders(feed_id);
CREATE INDEX IF NOT EXISTS idx_feed_folders_folder_name ON feed_folders(folder_name);