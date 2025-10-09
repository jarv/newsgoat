CREATE TABLE IF NOT EXISTS feed_folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id INTEGER NOT NULL,
    folder_name TEXT NOT NULL,
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
    UNIQUE(feed_id, folder_name)
);

CREATE INDEX IF NOT EXISTS idx_feed_folders_feed_id ON feed_folders(feed_id);
CREATE INDEX IF NOT EXISTS idx_feed_folders_folder_name ON feed_folders(folder_name);
