package auth

import (
	"errors"
	"net/http"

	"github.com/blatessa/verbose-octo-succotash/pkg/httputil"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts the auth endpoints onto mux under the given prefix.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, prefix string) {
	mux.HandleFunc("POST "+prefix+"/register", h.register)
	mux.HandleFunc("POST "+prefix+"/login", h.login)
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		httputil.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	resp, err := h.svc.CreateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not create user")
		return
	}

	httputil.JSON(w, http.StatusCreated, resp)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httputil.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}
