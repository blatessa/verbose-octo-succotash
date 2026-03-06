package workspacedb

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createWorkspace = `
INSERT INTO workspace.workspaces (name, owner_id)
VALUES ($1, $2)
RETURNING id, name, owner_id, created_at, updated_at
`

type CreateWorkspaceParams struct {
	Name    string
	OwnerID pgtype.UUID
}

func (q *Queries) CreateWorkspace(ctx context.Context, arg CreateWorkspaceParams) (Workspace, error) {
	row := q.db.QueryRow(ctx, createWorkspace, arg.Name, arg.OwnerID)
	var w Workspace
	err := row.Scan(&w.ID, &w.Name, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt)
	return w, err
}

const getWorkspaceByID = `
SELECT id, name, owner_id, created_at, updated_at FROM workspace.workspaces
WHERE id = $1
`

func (q *Queries) GetWorkspaceByID(ctx context.Context, id pgtype.UUID) (Workspace, error) {
	row := q.db.QueryRow(ctx, getWorkspaceByID, id)
	var w Workspace
	err := row.Scan(&w.ID, &w.Name, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt)
	return w, err
}

const addMember = `
INSERT INTO workspace.members (workspace_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING workspace_id, user_id, role, joined_at
`

type AddMemberParams struct {
	WorkspaceID pgtype.UUID
	UserID      pgtype.UUID
	Role        string
}

func (q *Queries) AddMember(ctx context.Context, arg AddMemberParams) (Member, error) {
	row := q.db.QueryRow(ctx, addMember, arg.WorkspaceID, arg.UserID, arg.Role)
	var m Member
	err := row.Scan(&m.WorkspaceID, &m.UserID, &m.Role, &m.JoinedAt)
	return m, err
}

const listMembers = `
SELECT wm.user_id, u.email, wm.role, wm.joined_at
FROM workspace.members wm
JOIN auth.users u ON u.id = wm.user_id
WHERE wm.workspace_id = $1
ORDER BY wm.joined_at ASC
`

func (q *Queries) ListMembers(ctx context.Context, workspaceID pgtype.UUID) ([]MemberWithEmail, error) {
	rows, err := q.db.Query(ctx, listMembers, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []MemberWithEmail
	for rows.Next() {
		var m MemberWithEmail
		if err := rows.Scan(&m.UserID, &m.Email, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}
