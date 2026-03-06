-- name: CreateWorkspace :one
INSERT INTO workspace.workspaces (name, owner_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetWorkspace :one
SELECT * FROM workspace.workspaces
WHERE id = $1;

-- name: ListWorkspacesByOwner :many
SELECT * FROM workspace.workspaces
WHERE owner_id = $1
ORDER BY created_at DESC;
