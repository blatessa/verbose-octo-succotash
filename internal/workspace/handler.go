package workspace

import (
	"errors"
	"net/http"

	"github.com/blatessa/verbose-octo-succotash/pkg/httputil"
	"github.com/blatessa/verbose-octo-succotash/pkg/reqctx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts workspace endpoints onto mux under the given prefix,
// wrapping each route with the provided middleware (e.g. auth.Middleware).
func (h *Handler) RegisterRoutes(mux *http.ServeMux, prefix string, mid func(http.Handler) http.Handler) {
	mux.Handle("POST "+prefix, mid(http.HandlerFunc(h.create)))
	mux.Handle("GET "+prefix, mid(http.HandlerFunc(h.list)))
	mux.Handle("GET "+prefix+"/{id}", mid(http.HandlerFunc(h.get)))
}

type workspaceJSON struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	OwnerID string `json:"owner_id"`
}

type createRequest struct {
	Name string `json:"name"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID := reqctx.MustUserID(r.Context())

	var req createRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "name is required")
		return
	}

	ws, err := h.svc.Create(r.Context(), req.Name, userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not create workspace")
		return
	}

	httputil.JSON(w, http.StatusCreated, toJSON(ws))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID := reqctx.MustUserID(r.Context())

	workspaces, err := h.svc.ListByOwner(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not list workspaces")
		return
	}

	out := make([]workspaceJSON, len(workspaces))
	for i, ws := range workspaces {
		out[i] = toJSON(ws)
	}
	httputil.JSON(w, http.StatusOK, out)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	ws, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "workspace not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "could not get workspace")
		return
	}

	httputil.JSON(w, http.StatusOK, toJSON(ws))
}

func toJSON(ws Workspace) workspaceJSON {
	return workspaceJSON{ID: ws.ID, Name: ws.Name, OwnerID: ws.OwnerID}
}
