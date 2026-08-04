package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	authrepo "project_pay/cmd/internal/repositories/authentication"
)

// contextKey é um tipo próprio para evitar colisão com outras chaves no context
type contextKey string

const UserIDKey contextKey = "user_id"

// UserIDFromContext extrai o user_id injetado pelo middleware Auth.
// Services usam este helper para aplicar ownership.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

// Auth é um middleware que envolve handlers que precisam de autenticação.
// Lê o token do header, valida, extrai o user_id e injeta no context.
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			apperrors.WriteHTTP(w, r, apperrors.Unauthorized("token ausente"))
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")

		userID, err := authrepo.VerifyToken(tokenString)
		if err != nil {
			apperrors.WriteHTTP(w, r, apperrors.Unauthorized("token inválido"))
			return
		}

		// injeta o user_id no context — handlers podem ler com r.Context().Value(middleware.UserIDKey)
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}
