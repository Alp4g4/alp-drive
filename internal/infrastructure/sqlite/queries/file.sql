-- name: GetFile :one
SELECT id, name, folder_id, size, mime_type, created_at, updated_at
FROM file
WHERE id = @id;

-- name: ListFilesByFolderID :many
SELECT id, name, folder_id, size, mime_type, created_at, updated_at
FROM file
WHERE folder_id = @folder_id
ORDER BY name, id
LIMIT @limit_count;

-- name: ListFilesByFolderTree :many
WITH RECURSIVE descendant_folder(id) AS (
    SELECT folder.id FROM folder WHERE folder.id = @folder_id
    UNION ALL
    SELECT fo.id FROM folder fo JOIN descendant_folder d ON fo.parent_id = d.id
)
SELECT fi.id, fi.name, fi.folder_id, fi.size, fi.mime_type, fi.created_at, fi.updated_at
FROM file fi
WHERE fi.folder_id IN (SELECT descendant_folder.id FROM descendant_folder);

-- name: CreateFile :exec
INSERT INTO file (id, name, folder_id, size, mime_type, created_at, updated_at)
VALUES (@id, @name, @folder_id, @size, @mime_type, @created_at, @updated_at);

-- name: UpdateFile :execrows
UPDATE file
SET name = @name,
    folder_id = @folder_id,
    size = @size,
    mime_type = @mime_type,
    updated_at = @updated_at
WHERE id = @id;

-- name: DeleteFile :execrows
DELETE FROM file
WHERE id = @id;
