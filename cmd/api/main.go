package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"project_pay/cmd/internal/database"
	"project_pay/cmd/internal/handlers"
	"project_pay/cmd/internal/middleware"
	accountsrepo "project_pay/cmd/internal/repositories/accounts"
	authrepo "project_pay/cmd/internal/repositories/authentication"
	ledgerrepo "project_pay/cmd/internal/repositories/ledger_entries"
	participantsrepo "project_pay/cmd/internal/repositories/participants"
	pixkeysrepo "project_pay/cmd/internal/repositories/pix_keys"
	transfersrepo "project_pay/cmd/internal/repositories/transfers"
	usersrepo "project_pay/cmd/internal/repositories/users"
	userssvc "project_pay/cmd/internal/services/users"
)

func main() {
	godotenv.Load()

	db, err := database.Connect(context.Background())
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}
	defer db.Close()

	authRepo := authrepo.New(db)
	authHandler := handlers.NewAuthHandler(authRepo)

	userRepo := usersrepo.New(db)
	userSvc := userssvc.New(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)

	accountRepo := accountsrepo.New(db)
	accountHandler := handlers.NewAccountHandler(accountRepo)

	participantRepo := participantsrepo.New(db)
	participantHandler := handlers.NewParticipantHandler(participantRepo)

	pixKeyRepo := pixkeysrepo.New(db)
	pixKeyHandler := handlers.NewPixKeyHandler(pixKeyRepo)

	ledgerRepo := ledgerrepo.New(db)
	ledgerHandler := handlers.NewLedgerEntryHandler(ledgerRepo)

	transferRepo := transfersrepo.New(db)
	transferHandler := handlers.NewTransferHandler(transferRepo)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /users", userHandler.Create)

	mux.HandleFunc("GET /users/{id}", middleware.Auth(userHandler.GetByID))
	mux.HandleFunc("POST /accounts", middleware.Auth(accountHandler.CreateAccount))
	mux.HandleFunc("GET /accounts/{id}", middleware.Auth(accountHandler.GetByID))
	mux.HandleFunc("POST /participants", middleware.Auth(participantHandler.CreateParticipant))
	mux.HandleFunc("GET /participants/{id}", middleware.Auth(participantHandler.GetByID))
	mux.HandleFunc("POST /pix-keys", middleware.Auth(pixKeyHandler.CreatePixKey))
	mux.HandleFunc("GET /pix-keys/{key}", middleware.Auth(pixKeyHandler.GetByKey))
	mux.HandleFunc("GET /accounts/{id}/statement", middleware.Auth(ledgerHandler.GetStatement))
	mux.HandleFunc("POST /transfers", middleware.Auth(transferHandler.CreateTransfer))
	mux.HandleFunc("GET /transfers/{id}", middleware.Auth(transferHandler.GetByID))

	port := os.Getenv("API_PORT")
	log.Printf("servidor rodando na porta %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
