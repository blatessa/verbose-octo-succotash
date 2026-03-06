package workspace

import (
	"context"
	"errors"
	"fmt"

	workspacedb "github.com/blatessa/verbose-octo-succotash/internal/workspace/db"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrNotFound = errors.New("workspace: not found")

// Workspace is the domain type for a workspace.
type Workspace struct {
	ID      string
	Name    string
	OwnerID string
}

type Service struct {
	q *workspacedb.Queries
}

func NewService(q *workspacedb.Queries) *Service {
	return &Service{q: q}
}

// Create creates a new workspace owned by ownerID.
func (s *Service) Create(ctx context.Context, name, ownerID string) (Workspace, error) {
	var uid pgtype.UUID
	if err := uid.Scan(ownerID); err != nil {
		return Workspace{}, fmt.Errorf("workspace: invalid owner id: %w", err)
	}

	row, err := s.q.CreateWorkspace(ctx, workspacedb.CreateWorkspaceParams{
		Name:    name,
		OwnerID: uid,
	})
	if err != nil {
		return Workspace{}, fmt.Errorf("workspace: create: %w", err)
	}

	return toWorkspace(row), nil
}

// Get returns a single workspace by ID.
func (s *Service) Get(ctx context.Context, id string) (Workspace, error) {
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		return Workspace{}, ErrNotFound
	}

	row, err := s.q.GetWorkspace(ctx, uid)
	if err != nil {
		return Workspace{}, ErrNotFound
	}

	return toWorkspace(row), nil
}

// ListByOwner returns all workspaces owned by ownerID.
func (s *Service) ListByOwner(ctx context.Context, ownerID string) ([]Workspace, error) {
	var uid pgtype.UUID
	if err := uid.Scan(ownerID); err != nil {
		return nil, fmt.Errorf("workspace: invalid owner id: %w", err)
	}

	rows, err := s.q.ListWorkspacesByOwner(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("workspace: list: %w", err)
	}

	workspaces := make([]Workspace, len(rows))
	for i, r := range rows {
		workspaces[i] = toWorkspace(r)
	}
	return workspaces, nil
}

func toWorkspace(r workspacedb.Workspace) Workspace {
	return Workspace{
		ID:      r.ID.String(),
		Name:    r.Name,
		OwnerID: r.OwnerID.String(),
	}
}
