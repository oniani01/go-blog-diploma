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

// respondJSON - хелпер для отправки JSON с правильным Content-Type
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError - хелпер для отправки ошибок в JSON формате
func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
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

// isValidEmail - простая, но надежная валидация формата email
func isValidEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	atIdx := strings.Index(email, "@")
	if atIdx < 1 || atIdx >= len(email)-3 {
		return false
	}
	dotIdx := strings.LastIndex(email, ".")
	if dotIdx < atIdx+2 || dotIdx >= len(email)-1 {
		return false
	}
	return true
}

// Register обрабатывает POST /register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Валидация входных данных (требование ревьюера)
	if req.Username == "" || len(req.Username) < 3 {
		respondError(w, http.StatusBadRequest, "Username must be at least 3 characters")
		return
	}
	if !isValidEmail(req.Email) {
		respondError(w, http.StatusBadRequest, "Invalid email format")
		return
	}
	if len(req.Password) < 6 {
		respondError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	if h.Storage.GetUserByEmail(req.Email) != nil {
		respondError(w, http.StatusConflict, "User already exists")
		return
	}

	hashedPwd, _ := auth.HashPassword(req.Password)

	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPwd,  // ИСПРАВЛЕНО: сохраняем хеш в правильное поле
		CreatedAt:    time.Now(), // ИСПРАВЛЕНО: добавляем время создания
	}
	h.Storage.CreateUser(user)

	respondJSON(w, http.StatusCreated, map[string]string{"message": "User registered"})
}

// Login обрабатывает POST /login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	user := h.Storage.GetUserByEmail(req.Email)
	// ИСПРАВЛЕНО: проверяем хеш в поле PasswordHash
	if user == nil || !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, _ := auth.GenerateToken(user.ID)
	respondJSON(w, http.StatusOK, map[string]string{"token": token})
}

// CreatePost обрабатывает POST /posts
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" || req.Content == "" {
		respondError(w, http.StatusBadRequest, "Invalid data")
		return
	}

	post := models.Post{
		AuthorID:  userID,
		Title:     req.Title,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}
	post = h.Storage.CreatePost(post)

	logger.LogAction("user " + strconv.Itoa(userID) + " created post " + strconv.Itoa(post.ID))

	respondJSON(w, http.StatusCreated, post)
}

// GetPosts обрабатывает GET /posts
func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	posts := h.Storage.GetAllPosts()
	if posts == nil {
		posts = []models.Post{}
	}
	respondJSON(w, http.StatusOK, posts)
}

// GetPost обрабатывает GET /posts/{id}
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)

	post := h.Storage.GetPostByID(id)
	if post == nil {
		respondError(w, http.StatusNotFound, "Not found")
		return
	}
	respondJSON(w, http.StatusOK, post)
}

// CreateComment обрабатывает POST /posts/{id}/comments
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	postID, _ := strconv.Atoi(r.PathValue("id"))
	if h.Storage.GetPostByID(postID) == nil {
		respondError(w, http.StatusNotFound, "Post not found")
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" {
		respondError(w, http.StatusBadRequest, "Invalid data")
		return
	}

	comment := models.Comment{
		PostID:    postID,
		AuthorID:  userID,
		Text:      req.Text,
		CreatedAt: time.Now(),
	}
	comment = h.Storage.CreateComment(comment)

	logger.LogAction("user " + strconv.Itoa(userID) + " created comment " + strconv.Itoa(comment.ID))

	respondJSON(w, http.StatusCreated, comment)
}

// GetComments обрабатывает GET /posts/{id}/comments
func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	postID, _ := strconv.Atoi(r.PathValue("id"))

	// ИСПРАВЛЕНО: проверяем существование поста перед выдачей комментариев
	if h.Storage.GetPostByID(postID) == nil {
		respondError(w, http.StatusNotFound, "Post not found")
		return
	}

	comments := h.Storage.GetCommentsByPostID(postID)
	if comments == nil {
		comments = []models.Comment{}
	}
	respondJSON(w, http.StatusOK, comments)
}

// Health обрабатывает GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
