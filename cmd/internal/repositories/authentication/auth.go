package authentication

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// secretKey lê o segredo do .env em cada chamada — nunca fica hardcoded no binário
func secretKey() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

// CreateToken gera um JWT assinado com o user_id nos claims e validade de 24h
func CreateToken(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id": userID.String(),
			"exp":     time.Now().Add(time.Hour * 24).Unix(),
		})
	return token.SignedString(secretKey())
}

// VerifyToken valida a assinatura e extrai o user_id dos claims
func VerifyToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado")
		}
		return secretKey(), nil
	})
	if err != nil || !token.Valid {
		return uuid.UUID{}, fmt.Errorf("token inválido")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("claims inválidos")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("user_id ausente no token")
	}

	return uuid.Parse(userIDStr)
}

// VerifyCredentials busca o usuário por email, compara a senha com bcrypt
// retorna o user_id e true se as credenciais forem válidas
func (r *Repository) VerifyCredentials(ctx context.Context, email, password string) (uuid.UUID, bool) {
	var userID uuid.UUID
	var hashedPassword string

	err := r.db.QueryRow(ctx, sqlGetUserByEmail, email).Scan(&userID, &hashedPassword)
	if errors.Is(err, pgx.ErrNoRows) || err != nil {
		return uuid.UUID{}, false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return uuid.UUID{}, false
	}

	return userID, true
}

const sqlGetUserByEmail = `
	SELECT id, password_hash
	FROM users
	WHERE email = $1`
