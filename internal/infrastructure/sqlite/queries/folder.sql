-- name: GetFolder :one
SELECT id, name, parent_id, created_at, updated_at
FROM folder
WHERE id = @id;

-- name: ExistsFolderByID :one
SELECT EXISTS (
    SELECT 1 FROM folder WHERE id = @id
);

-- name: ListFoldersByParentID :many
SELECT id, name, parent_id, created_at, updated_at
FROM folder
WHERE parent_id IS sqlc.narg(parent_id)
ORDER BY name, id
LIMIT @limit_count;

-- name: CreateFolder :exec
INSERT INTO folder (id, name, parent_id, created_at, updated_at)
VALUES (@id, @name, sqlc.narg(parent_id), @created_at, @updated_at);

-- name: UpdateFolder :execrows
UPDATE folder
SET name = @name,
    parent_id = sqlc.narg(parent_id),
    updated_at = @updated_at
WHERE id = @id;

-- name: DeleteFolder :execrows
DELETE FROM folder
WHERE id = @id;
