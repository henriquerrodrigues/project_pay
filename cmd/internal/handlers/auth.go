package handlers

import (
	"encoding/json"
	"net/http"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	authrepo "project_pay/cmd/internal/repositories/authentication"
)

type AuthHandler struct {
	repo *authrepo.Repository
}

func NewAuthHandler(repo *authrepo.Repository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.AuthJWT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("body inválido"))
		return
	}

	userID, ok := h.repo.VerifyCredentials(r.Context(), req.Email, req.Password)
	if !ok {
		apperrors.WriteHTTP(w, r, apperrors.Unauthorized("credenciais inválidas"))
		return
	}

	token, err := authrepo.CreateToken(userID)
	if err != nil {
		apperrors.WriteHTTP(w, r, apperrors.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
