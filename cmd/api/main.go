package main

import (
	"context"
	"log/slog"
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
	accountssvc "project_pay/cmd/internal/services/accounts"
	ledgersvc "project_pay/cmd/internal/services/ledger_entries"
	participantssvc "project_pay/cmd/internal/services/participants"
	pixkeyssvc "project_pay/cmd/internal/services/pix_keys"
	transferssvc "project_pay/cmd/internal/services/transfers"
	userssvc "project_pay/cmd/internal/services/users"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	godotenv.Load()

	db, err := database.Connect(context.Background())
	if err != nil {
		slog.Error("erro ao conectar no banco", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	authRepo := authrepo.New(db)
	authHandler := handlers.NewAuthHandler(authRepo)

	userRepo := usersrepo.New(db)
	userSvc := userssvc.New(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)

	accountRepo := accountsrepo.New(db)
	accountSvc := accountssvc.New(accountRepo)
	accountHandler := handlers.NewAccountHandler(accountSvc)

	participantRepo := participantsrepo.New(db)
	participantSvc := participantssvc.New(participantRepo)
	participantHandler := handlers.NewParticipantHandler(participantSvc)

	pixKeyRepo := pixkeysrepo.New(db)
	pixKeySvc := pixkeyssvc.New(pixKeyRepo)
	pixKeyHandler := handlers.NewPixKeyHandler(pixKeySvc)

	ledgerRepo := ledgerrepo.New(db)
	ledgerSvc := ledgersvc.New(ledgerRepo, accountRepo)
	ledgerHandler := handlers.NewLedgerEntryHandler(ledgerSvc)

	transferRepo := transfersrepo.New(db)
	transferSvc := transferssvc.New(transferRepo, accountRepo)
	transferHandler := handlers.NewTransferHandler(transferSvc)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /users", userHandler.Create)

	mux.HandleFunc("GET /users/{id}", middleware.Auth(userHandler.GetByID))
	mux.HandleFunc("POST /accounts", middleware.Auth(accountHandler.CreateAccount))
	mux.HandleFunc("GET /accounts/{id}", middleware.Auth(accountHandler.GetByID))
	mux.HandleFunc("GET /accounts", middleware.Auth(accountHandler.List))
	mux.HandleFunc("POST /participants", middleware.Auth(participantHandler.CreateParticipant))
	mux.HandleFunc("GET /participants/{id}", middleware.Auth(participantHandler.GetByID))
	mux.HandleFunc("POST /pix-keys", middleware.Auth(pixKeyHandler.CreatePixKey))
	mux.HandleFunc("GET /pix-keys/{key}", middleware.Auth(pixKeyHandler.GetByKey))
	mux.HandleFunc("GET /accounts/{id}/statement", middleware.Auth(ledgerHandler.GetStatement))
	mux.HandleFunc("POST /transfers", middleware.Auth(transferHandler.CreateTransfer))
	mux.HandleFunc("GET /transfers/{id}", middleware.Auth(transferHandler.GetByID))

	handler := middleware.RequestLogger(mux)

	port := os.Getenv("API_PORT")
	slog.Info("servidor rodando", "port", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		slog.Error("servidor encerrou com erro", "error", err)
		os.Exit(1)
	}
}
