package workspace

import (
	"context"
	"errors"
	"fmt"
	"strings"

	workspacedb "github.com/blatessa/verbose-octo-succotash/internal/workspace/db"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNotFound         = errors.New("workspace: not found")
	ErrAlreadyMember    = errors.New("workspace: user is already a member")
	ErrInvalidWorkspace = errors.New("workspace: invalid workspace id")
	ErrInvalidUser      = errors.New("workspace: invalid user id")
)

// Workspace is the domain type for a workspace.
type Workspace struct {
	ID      string
	Name    string
	OwnerID string
}

// Member is the domain type for a workspace member.
type Member struct {
	UserID   string
	Email    string
	Role     string
	JoinedAt string
}

type Service struct {
	q *workspacedb.Queries
}

func NewService(q *workspacedb.Queries) *Service {
	return &Service{q: q}
}

// CreateWorkspace creates a new workspace owned by the given user.
func (s *Service) CreateWorkspace(ctx context.Context, ownerID, name string) (Workspace, error) {
	var ownerUUID pgtype.UUID
	if err := ownerUUID.Scan(ownerID); err != nil {
		return Workspace{}, ErrInvalidUser
	}

	row, err := s.q.CreateWorkspace(ctx, workspacedb.CreateWorkspaceParams{
		Name:    name,
		OwnerID: ownerUUID,
	})
	if err != nil {
		return Workspace{}, fmt.Errorf("workspace: create: %w", err)
	}

	return toWorkspace(row), nil
}

// GetWorkspace retrieves a workspace by its ID.
func (s *Service) GetWorkspace(ctx context.Context, workspaceID string) (Workspace, error) {
	var id pgtype.UUID
	if err := id.Scan(workspaceID); err != nil {
		return Workspace{}, ErrInvalidWorkspace
	}

	row, err := s.q.GetWorkspaceByID(ctx, id)
	if err != nil {
		return Workspace{}, ErrNotFound
	}

	return toWorkspace(row), nil
}

// AddMember adds a user to a workspace with the given role.
func (s *Service) AddMember(ctx context.Context, workspaceID, userID, role string) (Member, error) {
	var wsID pgtype.UUID
	if err := wsID.Scan(workspaceID); err != nil {
		return Member{}, ErrInvalidWorkspace
	}

	var uID pgtype.UUID
	if err := uID.Scan(userID); err != nil {
		return Member{}, ErrInvalidUser
	}

	if role == "" {
		role = "member"
	}

	row, err := s.q.AddMember(ctx, workspacedb.AddMemberParams{
		WorkspaceID: wsID,
		UserID:      uID,
		Role:        role,
	})
	if err != nil {
		// pgx surfaces unique-constraint violations as errors containing "23505"
		if isPgUniqueViolation(err) {
			return Member{}, ErrAlreadyMember
		}
		return Member{}, fmt.Errorf("workspace: add member: %w", err)
	}

	return Member{
		UserID:   row.UserID.String(),
		Role:     row.Role,
		JoinedAt: row.JoinedAt.String(),
	}, nil
}

// ListMembers returns all members of a workspace.
func (s *Service) ListMembers(ctx context.Context, workspaceID string) ([]Member, error) {
	var wsID pgtype.UUID
	if err := wsID.Scan(workspaceID); err != nil {
		return nil, ErrInvalidWorkspace
	}

	rows, err := s.q.ListMembers(ctx, wsID)
	if err != nil {
		return nil, fmt.Errorf("workspace: list members: %w", err)
	}

	members := make([]Member, len(rows))
	for i, r := range rows {
		members[i] = Member{
			UserID:   r.UserID.String(),
			Email:    r.Email,
			Role:     r.Role,
			JoinedAt: r.JoinedAt.String(),
		}
	}
	return members, nil
}

func toWorkspace(w workspacedb.Workspace) Workspace {
	return Workspace{
		ID:      w.ID.String(),
		Name:    w.Name,
		OwnerID: w.OwnerID.String(),
	}
}

func isPgUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
