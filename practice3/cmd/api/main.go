package main

import (
	"log"
	"net/http"

	pg "practice3/internal/repository/_postgres"
	"practice3/internal/usecase"
	"practice3/internal/handlers"
	"practice3/internal/middleware"
)

func main() {
	cfg := pg.DefaultPostgresConfig()

	d, err := pg.NewPGXDialect(cfg)
	if err != nil {
		log.Fatal(err)
	}

	repos := pg.NewRepositories(d)
	uc := usecase.NewUserUsecase(repos.User)
	handler := handlers.NewUserHandler(uc)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	finalHandler := middleware.Logging(middleware.Auth(mux))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", finalHandler))
}