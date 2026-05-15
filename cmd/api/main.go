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

	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", userHandler.Create)
	mux.HandleFunc("GET /users/{id}", userHandler.GetByID)
	mux.HandleFunc("POST /accounts", accountHandler.CreateAccount)
	mux.HandleFunc("GET /accounts/{id}", accountHandler.GetByID)

	port := os.Getenv("API_PORT")
	log.Printf("servidor rodando na porta %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
