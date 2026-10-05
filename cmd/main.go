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
	// 1. Загрузка переменных окружения из файла .env (если он есть)
	godotenv.Load()

	// 2. Инициализация логгера (запускает фоновую горутину)
	logger.Init()

	// 3. Инициализация хранилища
	os.MkdirAll("./data", 0755) // Создаем папку data, если её еще нет
	store, err := storage.New("./data")
	if err != nil {
		log.Fatal("Не удалось инициализировать хранилище:", err)
	}

	// 4. Создаем экземпляр обработчиков и передаем туда хранилище
	h := &handlers.Handler{Storage: store}

	// 5. Настройка роутера (используем новые возможности Go 1.21)
	mux := http.NewServeMux()

	// Регистрируем маршруты. Обратите внимание на указание метода (POST, GET) прямо в пути!
	mux.HandleFunc("POST /register", h.Register)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("POST /posts", h.CreatePost)
	mux.HandleFunc("GET /posts", h.GetPosts)
	mux.HandleFunc("GET /posts/{id}", h.GetPost)
	mux.HandleFunc("POST /posts/{id}/comments", h.CreateComment)
	mux.HandleFunc("GET /posts/{id}/comments", h.GetComments)
	mux.HandleFunc("GET /health", h.Health)

	// Настраиваем сам сервер
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// 6. Запуск сервера в отдельной горутине, чтобы он не блокировал основной поток
	go func() {
		log.Println("✅ Сервер запущен на http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(" Ошибка сервера:", err)
		}
	}()

	// 7. Graceful Shutdown (Корректное завершение работы)
	// Создаем канал для получения сигналов от операционной системы (например, Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// Программа заблокируется здесь и будет ждать сигнала выключения
	<-quit
	log.Println(" Получен сигнал остановки. Завершаем работу...")

	// Даем серверу 5 секунд на завершение текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Принудительная остановка сервера:", err)
	}

	// 8. Корректное завершение работы логгера (закрывает канал и ждет записи последних логов)
	logger.Close()
	log.Println("👋 Сервер успешно остановлен")
}
