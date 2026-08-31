package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/yourusername/linkedin-profile-api/cache"
	"github.com/yourusername/linkedin-profile-api/handlers"
	"github.com/yourusername/linkedin-profile-api/middleware"
	"github.com/yourusername/linkedin-profile-api/scraper"
)

func main() {
	_ = godotenv.Load()

	client, err := scraper.NewClient()
	if err != nil {
		log.Fatalf("failed to init scraper client: %v", err)
	}

	profileCache := cache.NewProfileCache()
	profileHandler := handlers.NewProfileHandler(client, profileCache)
	rateLimiter := middleware.NewRateLimiter()

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(rateLimiter.Middleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	r.Get("/profile", profileHandler.GetProfile)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
