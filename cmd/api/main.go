package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"project_pay/cmd/internal/database"
	"project_pay/cmd/internal/handlers"
	accountsrepo "project_pay/cmd/internal/repositories/accounts"
	ledgerrepo "project_pay/cmd/internal/repositories/ledger_entries"
	participantsrepo "project_pay/cmd/internal/repositories/participants"
	pixkeysrepo "project_pay/cmd/internal/repositories/pix_keys"
	transfersrepo "project_pay/cmd/internal/repositories/transfers"
	usersrepo "project_pay/cmd/internal/repositories/users"
)

func main() {
	godotenv.Load()

	db, err := database.Connect(context.Background())
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}
	defer db.Close()

	userRepo := usersrepo.New(db)
	userHandler := handlers.NewUserHandler(userRepo)

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
	mux.HandleFunc("POST /users", userHandler.Create)
	mux.HandleFunc("GET /users/{id}", userHandler.GetByID)
	mux.HandleFunc("POST /accounts", accountHandler.CreateAccount)
	mux.HandleFunc("GET /accounts/{id}", accountHandler.GetByID)
	mux.HandleFunc("POST /participants", participantHandler.CreateParticipant)
	mux.HandleFunc("GET /participants/{id}", participantHandler.GetByID)
	mux.HandleFunc("POST /pix-keys", pixKeyHandler.CreatePixKey)
	mux.HandleFunc("GET /pix-keys/{key}", pixKeyHandler.GetByKey)
	mux.HandleFunc("GET /accounts/{id}/statement", ledgerHandler.GetStatement)
	mux.HandleFunc("POST /transfers", transferHandler.CreateTransfer)
	mux.HandleFunc("GET /transfers/{id}", transferHandler.GetByID)

	port := os.Getenv("API_PORT")
	log.Printf("servidor rodando na porta %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
