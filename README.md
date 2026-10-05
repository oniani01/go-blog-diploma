# REST API для блог-платформы

Дипломный проект для программы "Go-разработчик с нуля".

## Описание

Простое REST API для блога, написанное на Go 1.21+. Реализует регистрацию пользователей, авторизацию (JWT), управление постами и комментариями с хранением данных в JSON-файлах.

## Особенности

- Современный роутинг Go 1.21+ (`http.ServeMux` с поддержкой методов)
- Асинхронное логирование действий через канал и горутину с задержкой 1 сек
- Безопасное хеширование паролей через `bcrypt`
- Graceful shutdown для корректного завершения работы логгера
- Полная контейнеризация через Docker

## Структура проекта
go-blog-diploma/
├── cmd/
│ └── main.go # Точка входа
├── internal/
│ ├── models/models.go # Структуры данных (User, Post, Comment)
│ ├── storage/storage.go # Работа с JSON-файлами
│ ├── auth/auth.go # JWT и хеширование паролей
│ ├── logger/logger.go # Асинхронное логирование (горутина + канал)
│ └── handlers/handlers.go # HTTP-обработчики
├── data/ # Папка для JSON-файлов (создаётся автоматически)
├── .env # Переменные окружения
├── .gitignore # Исключения для Git
├── Dockerfile # Docker образ
├── docker-compose.yml # Docker Compose
├── go.mod # Зависимости Go
├── go.sum # Контрольные суммы зависимостей
└── README.md # Документация
## Запуск

### Локально

```bash
# Установить зависимости
go mod download

# Запустить сервер
go run cmd/main.go
Сервер запустится на http://localhost:8080
Через Docker
# Собрать и запустить
docker-compose up --build

# Остановить
docker-compose down
API Эндпоинты
Метод
Эндпоинт
Описание
Требует Auth
POST
/register
Регистрация пользователя
Нет
POST
/login
Вход (возвращает JWT токен)
Нет
POST
/posts
Создание поста
Да (Bearer)
GET
/posts
Список всех постов
Нет
GET
/posts/{id}
Получение поста по ID
Нет
POST
/posts/{id}/comments
Добавление комментария
Да (Bearer)
GET
/posts/{id}/comments
Комментарии к посту
Нет
GET
/health
Проверка состояния сервиса
Нет
Примеры запросов
Регистрация
bash
curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d '{"username":"test","email":"test@test.com","password":"123456"}'
Вход
bash
curl -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"email":"test@test.com","password":"123456"}'
Создание поста
bash
curl -X POST http://localhost:8080/posts -H "Content-Type: application/json" -H "Authorization: Bearer ВАШ_ТОКЕН" -d '{"title":"Мой пост","content":"Текст поста"}'
Получение всех постов
bash
curl http://localhost:8080/posts
Добавление комментария
bash
curl -X POST http://localhost:8080/posts/1/comments -H "Content-Type: application/json" -H "Authorization: Bearer ВАШ_ТОКЕН" -d '{"text":"Отличный пост!"}'
Технологии
Go 1.21+
JWT (github.com/golang-jwt/jwt/v5)
bcrypt (golang.org/x/crypto/bcrypt)
godotenv (github.com/joho/godotenv)
Docker & Docker Compose
