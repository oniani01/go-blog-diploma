package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"go-blog-diploma/internal/handlers"
	"go-blog-diploma/internal/logger"
	"go-blog-diploma/internal/storage"
)

func main() {
	godotenv.Load()

	logger.Init()

	os.MkdirAll("./data", 0755)

	store, err := storage.New("./data")
	if err != nil {
		log.Fatal("Failed to create storage:", err)
	}

	h := &handlers.Handler{Storage: store}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", h.Register)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("POST /posts", h.CreatePost)
	mux.HandleFunc("GET /posts", h.GetPosts)
	mux.HandleFunc("GET /posts/{id}", h.GetPost)
	mux.HandleFunc("POST /posts/{id}/comments", h.CreateComment)
	mux.HandleFunc("GET /posts/{id}/comments", h.GetComments)
	mux.HandleFunc("GET /health", h.Health)

	server := &http.Server{
		Addr:    ":" + getPort(),
		Handler: mux,
	}

	go func() {
		log.Printf("Server started on http://localhost:%s", getPort())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	logger.Close()

	log.Println("Server stopped successfully")
}

func getPort() string {
	p := os.Getenv("PORT")
	if p == "" {
		return "8080"
	}
	return p
}