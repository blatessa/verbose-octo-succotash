package workspace

import (
	"errors"
	"net/http"

	"github.com/blatessa/verbose-octo-succotash/internal/auth"
	"github.com/blatessa/verbose-octo-succotash/pkg/httputil"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts workspace endpoints onto mux under the given prefix.
// All routes require a valid Bearer JWT (enforced via auth.Middleware).
func (h *Handler) RegisterRoutes(mux *http.ServeMux, prefix string, authCfg auth.Config) {
	protected := auth.Middleware(authCfg)

	mux.Handle("POST "+prefix, protected(http.HandlerFunc(h.createWorkspace)))
	mux.Handle("GET "+prefix+"/{id}", protected(http.HandlerFunc(h.getWorkspace)))
	mux.Handle("POST "+prefix+"/{id}/members", protected(http.HandlerFunc(h.addMember)))
	mux.Handle("GET "+prefix+"/{id}/members", protected(http.HandlerFunc(h.listMembers)))
}

// --- request / response types ---

type createWorkspaceRequest struct {
	Name string `json:"name"`
}

type workspaceJSON struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	OwnerID string `json:"owner_id"`
}

type addMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role,omitempty"`
}

type memberJSON struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}

// --- handlers ---

func (h *Handler) createWorkspace(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createWorkspaceRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "name is required")
		return
	}

	ws, err := h.svc.CreateWorkspace(r.Context(), claims.UserID, req.Name)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not create workspace")
		return
	}

	httputil.JSON(w, http.StatusCreated, workspaceJSON{
		ID:      ws.ID,
		Name:    ws.Name,
		OwnerID: ws.OwnerID,
	})
}

func (h *Handler) getWorkspace(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	ws, err := h.svc.GetWorkspace(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrInvalidWorkspace) {
			httputil.Error(w, http.StatusNotFound, "workspace not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	httputil.JSON(w, http.StatusOK, workspaceJSON{
		ID:      ws.ID,
		Name:    ws.Name,
		OwnerID: ws.OwnerID,
	})
}

func (h *Handler) addMember(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.PathValue("id")

	var req addMemberRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == "" {
		httputil.Error(w, http.StatusBadRequest, "user_id is required")
		return
	}

	member, err := h.svc.AddMember(r.Context(), workspaceID, req.UserID, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, ErrAlreadyMember):
			httputil.Error(w, http.StatusConflict, "user is already a member")
		case errors.Is(err, ErrInvalidWorkspace), errors.Is(err, ErrNotFound):
			httputil.Error(w, http.StatusNotFound, "workspace not found")
		case errors.Is(err, ErrInvalidUser):
			httputil.Error(w, http.StatusBadRequest, "invalid user_id")
		default:
			httputil.Error(w, http.StatusInternalServerError, "could not add member")
		}
		return
	}

	httputil.JSON(w, http.StatusCreated, memberJSON{
		UserID:   member.UserID,
		Role:     member.Role,
		JoinedAt: member.JoinedAt,
	})
}

func (h *Handler) listMembers(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.PathValue("id")

	members, err := h.svc.ListMembers(r.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, ErrInvalidWorkspace) {
			httputil.Error(w, http.StatusBadRequest, "invalid workspace id")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]memberJSON, len(members))
	for i, m := range members {
		out[i] = memberJSON{
			UserID:   m.UserID,
			Email:    m.Email,
			Role:     m.Role,
			JoinedAt: m.JoinedAt,
		}
	}
	httputil.JSON(w, http.StatusOK, out)
}
