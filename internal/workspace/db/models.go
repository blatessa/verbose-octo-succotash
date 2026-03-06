package workspacedb

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Workspace struct {
	ID        pgtype.UUID
	Name      string
	OwnerID   pgtype.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Member struct {
	WorkspaceID pgtype.UUID
	UserID      pgtype.UUID
	Role        string
	JoinedAt    time.Time
}

type MemberWithEmail struct {
	UserID   pgtype.UUID
	Email    string
	Role     string
	JoinedAt time.Time
}
