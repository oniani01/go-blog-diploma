package storage

import (
	"encoding/json"
	"os"
	"sync"

	"go-blog-diploma/internal/models"
)

// Storage отвечает за хранение всех данных
type Storage struct {
	mu       sync.RWMutex // Мьютекс: защищает данные, если несколько пользователей пишут одновременно
	users    []models.User
	posts    []models.Post
	comments []models.Comment
	dataDir  string // Папка, где лежат наши JSON файлы
}

// New создает хранилище и загружает старые данные из файлов
func New(dataDir string) (*Storage, error) {
	s := &Storage{dataDir: dataDir}
	s.load() // Загружаем данные при старте
	return s, nil
}

// load читает JSON-файлы с диска в память
func (s *Storage) load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.loadFile("users.json", &s.users)
	s.loadFile("posts.json", &s.posts)
	s.loadFile("comments.json", &s.comments)
}

// loadFile читает один файл
func (s *Storage) loadFile(filename string, v interface{}) {
	data, err := os.ReadFile(s.dataDir + "/" + filename)
	if err != nil {
		return // Если файла нет (первый запуск), просто ничего не делаем
	}
	json.Unmarshal(data, v) // Превращаем JSON обратно в Go-структуры
}

// saveFile записывает данные из памяти в JSON-файл на диске
func (s *Storage) saveFile(filename string, v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ") // Делаем красивый JSON с отступами
	os.WriteFile(s.dataDir+"/"+filename, data, 0644)
}

// --- Методы для работы с ПОЛЬЗОВАТЕЛЯМИ ---

func (s *Storage) GetUserByEmail(email string) *models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Email == email {
			return &user
		}
	}
	return nil
}

func (s *Storage) CreateUser(u models.User) {
	s.mu.Lock() // Блокируем для записи
	defer s.mu.Unlock()

	// Придумываем ID (последний ID + 1)
	if len(s.users) > 0 {
		u.ID = s.users[len(s.users)-1].ID + 1
	} else {
		u.ID = 1
	}

	s.users = append(s.users, u)
	s.saveFile("users.json", s.users)
}

// --- Методы для работы с ПОСТАМИ ---

func (s *Storage) CreatePost(p models.Post) models.Post {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.posts) > 0 {
		p.ID = s.posts[len(s.posts)-1].ID + 1
	} else {
		p.ID = 1
	}

	s.posts = append(s.posts, p)
	s.saveFile("posts.json", s.posts)
	return p // Возвращаем пост с уже присвоенным ID
}

func (s *Storage) GetAllPosts() []models.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.posts
}

func (s *Storage) GetPostByID(id int) *models.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.posts {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

// --- Методы для работы с КОММЕНТАРИЯМИ ---

func (s *Storage) CreateComment(c models.Comment) models.Comment {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.comments) > 0 {
		c.ID = s.comments[len(s.comments)-1].ID + 1
	} else {
		c.ID = 1
	}

	s.comments = append(s.comments, c)
	s.saveFile("comments.json", s.comments)
	return c
}

func (s *Storage) GetCommentsByPostID(postID int) []models.Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var res []models.Comment
	for _, c := range s.comments {
		if c.PostID == postID {
			res = append(res, c)
		}
	}
	return res
}
