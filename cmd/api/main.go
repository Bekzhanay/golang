package main

import (
	"log"
	"net/http"
	"time"

	"assignment1/internal/handlers"
	"assignment1/internal/middleware"
)

func main() {
	store := handlers.NewTaskStore()
	taskHandler := handlers.NewTaskHandler(store)

	var h http.Handler = taskHandler
	h = middleware.Logging(h, "my server message")
	h = middleware.APIKeyAuth(h, "secret12345")

	http.Handle("/tasks", h)

	srv := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Server started on :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
