package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-blog-diploma/internal/auth"
	"go-blog-diploma/internal/logger"
	"go-blog-diploma/internal/models"
	"go-blog-diploma/internal/storage"
)

// Handler хранит ссылку на наше хранилище, чтобы обработчики могли к нему обращаться
type Handler struct {
	Storage *storage.Storage
}

// getUserIDFromRequest - вспомогательная функция для проверки токена
func (h *Handler) getUserIDFromRequest(r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, http.ErrNoCookie
	}
	// Ожидаем формат "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, http.ErrNoCookie
	}
	return auth.ValidateToken(parts[1])
}

// Register обрабатывает POST /register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Читаем JSON из тела запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Проверяем, нет ли уже такого пользователя
	if h.Storage.GetUserByEmail(req.Email) != nil {
		http.Error(w, `{"error": "User already exists"}`, http.StatusConflict)
		return
	}

	// Хешируем пароль
	hashedPwd, _ := auth.HashPassword(req.Password)

	// Создаем пользователя и сохраняем
	user := models.User{Username: req.Username, Email: req.Email, Password: hashedPwd}
	h.Storage.CreateUser(user)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered"})
}

// Login обрабатывает POST /login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	user := h.Storage.GetUserByEmail(req.Email)
	// Если пользователя нет или пароль не совпадает
	if user == nil || !auth.CheckPasswordHash(req.Password, user.Password) {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Генерируем токен
	token, _ := auth.GenerateToken(user.ID)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// CreatePost обрабатывает POST /posts
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	// Проверяем авторизацию
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" || req.Content == "" {
		http.Error(w, `{"error": "Invalid data"}`, http.StatusBadRequest)
		return
	}

	post := models.Post{
		AuthorID:  userID,
		Title:     req.Title,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	// Сохраняем пост (метод вернет пост с присвоенным ID)
	post = h.Storage.CreatePost(post)

	// ОТПРАВЛЯЕМ СОБЫТИЕ В КАНАЛ ЛОГГЕРА (Требование диплома!)
	logger.LogAction("user " + strconv.Itoa(userID) + " created post " + strconv.Itoa(post.ID))

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// GetPosts обрабатывает GET /posts
func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	posts := h.Storage.GetAllPosts()
	// Если постов нет, вернем пустой массив, а не null
	if posts == nil {
		posts = []models.Post{}
	}
	json.NewEncoder(w).Encode(posts)
}

// GetPost обрабатывает GET /posts/{id}
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	// В Go 1.21+ параметры пути достаются через r.PathValue
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)

	post := h.Storage.GetPostByID(id)
	if post == nil {
		http.Error(w, `{"error": "Not found"}`, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(post)
}

// CreateComment обрабатывает POST /posts/{id}/comments
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	postID, _ := strconv.Atoi(r.PathValue("id"))
	if h.Storage.GetPostByID(postID) == nil {
		http.Error(w, `{"error": "Post not found"}`, http.StatusNotFound)
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" {
		http.Error(w, `{"error": "Invalid data"}`, http.StatusBadRequest)
		return
	}

	comment := models.Comment{
		PostID:    postID,
		AuthorID:  userID,
		Text:      req.Text,
		CreatedAt: time.Now(),
	}
	comment = h.Storage.CreateComment(comment)

	// ЛОГИРОВАНИЕ (Требование диплома!)
	logger.LogAction("user " + strconv.Itoa(userID) + " created comment " + strconv.Itoa(comment.ID))

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// GetComments обрабатывает GET /posts/{id}/comments
func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	postID, _ := strconv.Atoi(r.PathValue("id"))
	comments := h.Storage.GetCommentsByPostID(postID)
	if comments == nil {
		comments = []models.Comment{}
	}
	json.NewEncoder(w).Encode(comments)
}

// Health обрабатывает GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
