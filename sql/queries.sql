-- name: CreateFeed :one
INSERT INTO feeds (url, title, description, last_updated, visible)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetFeed :one
SELECT * FROM feeds WHERE id = ?;

-- name: GetFeedByURL :one
SELECT * FROM feeds WHERE url = ?;

-- name: ListFeeds :many
SELECT * FROM feeds WHERE visible = TRUE ORDER BY title;

-- name: ListAllFeeds :many
SELECT * FROM feeds ORDER BY title;

-- name: UpdateFeed :exec
UPDATE feeds
SET title = ?, description = ?, last_updated = ?, etag = ?, last_modified = ?, cache_control_max_age = ?
WHERE id = ?;

-- name: UpdateFeedError :exec
UPDATE feeds
SET last_error = ?, last_error_time = ?
WHERE id = ?;

-- name: ClearFeedError :exec
UPDATE feeds
SET last_error = NULL, last_error_time = NULL
WHERE id = ?;

-- name: DeleteFeed :exec
DELETE FROM feeds WHERE id = ?;

-- name: HideFeed :exec
UPDATE feeds SET visible = FALSE WHERE id = ?;

-- name: ShowFeed :exec
UPDATE feeds SET visible = TRUE WHERE id = ?;

-- name: HideFeedByURL :exec
UPDATE feeds SET visible = FALSE WHERE url = ?;

-- name: ShowFeedByURL :exec
UPDATE feeds SET visible = TRUE WHERE url = ?;

-- name: CreateItem :one
INSERT INTO items (feed_id, guid, title, description, content, link, published)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetItem :one
SELECT * FROM items WHERE id = ?;

-- name: ListItemsByFeed :many
SELECT * FROM items
WHERE feed_id = ?
ORDER BY published DESC;

-- name: DeleteItemsByFeed :exec
DELETE FROM items WHERE feed_id = ?;

-- name: UpsertItem :one
INSERT INTO items (feed_id, guid, title, description, content, link, published)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(feed_id, guid) DO UPDATE SET
    title = excluded.title,
    description = excluded.description,
    content = excluded.content,
    link = excluded.link,
    published = excluded.published
RETURNING *;

-- name: MarkItemRead :exec
INSERT INTO read_status (item_id, read, read_at)
VALUES (?, TRUE, CURRENT_TIMESTAMP)
ON CONFLICT(item_id) DO UPDATE SET
    read = TRUE,
    read_at = CURRENT_TIMESTAMP;

-- name: MarkItemUnread :exec
INSERT INTO read_status (item_id, read)
VALUES (?, FALSE)
ON CONFLICT(item_id) DO UPDATE SET
    read = FALSE,
    read_at = NULL;

-- name: MarkAllItemsReadInFeed :exec
INSERT INTO read_status (item_id, read, read_at)
SELECT i.id, TRUE, CURRENT_TIMESTAMP
FROM items i
WHERE i.feed_id = ?
ON CONFLICT(item_id) DO UPDATE SET
    read = TRUE,
    read_at = CURRENT_TIMESTAMP;

-- name: IsItemRead :one
SELECT COALESCE(read, FALSE) as read
FROM read_status
WHERE item_id = ?;

-- name: GetFeedStats :many
SELECT
    f.id,
    f.title,
    f.url,
    f.last_error,
    f.last_error_time,
    COUNT(i.id) as total_items,
    COUNT(CASE WHEN i.id IS NOT NULL AND COALESCE(rs.read, FALSE) = FALSE THEN 1 END) as unread_items
FROM feeds f
LEFT JOIN items i ON f.id = i.feed_id
LEFT JOIN read_status rs ON i.id = rs.item_id
WHERE f.visible = TRUE
GROUP BY f.id, f.title, f.url, f.last_error, f.last_error_time
ORDER BY f.title;

-- name: GetItemsWithReadStatus :many
SELECT
    i.*,
    COALESCE(rs.read, FALSE) as read
FROM items i
LEFT JOIN read_status rs ON i.id = rs.item_id
WHERE i.feed_id = ?
ORDER BY i.published DESC;

-- name: CreateLogMessage :exec
INSERT INTO log_messages (level, message, timestamp, attributes)
VALUES (?, ?, ?, ?);

-- name: GetLogMessages :many
SELECT id, level, message, timestamp, attributes
FROM log_messages
ORDER BY timestamp DESC
LIMIT ?;

-- name: GetLogMessage :one
SELECT id, level, message, timestamp, attributes
FROM log_messages
WHERE id = ?;

-- name: DeleteAllLogMessages :exec
DELETE FROM log_messages;

-- name: GetSetting :one
SELECT key, value, updated_at FROM settings WHERE key = ?;

-- name: SetSetting :exec
INSERT INTO settings (key, value, updated_at)
VALUES (?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(key) DO UPDATE SET
    value = excluded.value,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetAllSettings :many
SELECT key, value, updated_at FROM settings ORDER BY key;

-- name: DeleteSetting :exec
DELETE FROM settings WHERE key = ?;

-- name: AddFeedFolder :exec
INSERT INTO feed_folders (feed_id, folder_name)
VALUES (?, ?)
ON CONFLICT(feed_id, folder_name) DO NOTHING;

-- name: GetFeedFolders :many
SELECT folder_name FROM feed_folders WHERE feed_id = ? ORDER BY folder_name;

-- name: DeleteFeedFolders :exec
DELETE FROM feed_folders WHERE feed_id = ?;

-- name: GetFolderStats :many
SELECT
    ff.folder_name,
    COUNT(DISTINCT i.id) as total_items,
    COUNT(DISTINCT CASE WHEN i.id IS NOT NULL AND COALESCE(rs.read, FALSE) = FALSE THEN i.id END) as unread_items
FROM feed_folders ff
INNER JOIN feeds f ON ff.feed_id = f.id
LEFT JOIN items i ON f.id = i.feed_id
LEFT JOIN read_status rs ON i.id = rs.item_id
WHERE f.visible = TRUE
GROUP BY ff.folder_name
ORDER BY ff.folder_name;