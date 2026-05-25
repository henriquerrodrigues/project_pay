package handlers

import (
	"encoding/json"
	"net/http"

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
		http.Error(w, "body inválido", http.StatusBadRequest)
		return
	}

	// busca o usuário e compara a senha via bcrypt
	userID, ok := h.repo.VerifyCredentials(r.Context(), req.Email, req.Password)
	if !ok {
		http.Error(w, "credenciais inválidas", http.StatusUnauthorized)
		return
	}

	// gera o token JWT com o user_id nos claims
	token, err := authrepo.CreateToken(userID)
	if err != nil {
		http.Error(w, "erro ao gerar token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
