package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Секретный ключ для подписи токенов.
// В реальном проекте его хранят в .env файле, но для простоты оставим здесь.
var jwtSecret = []byte("my-super-secret-diploma-key-2024")

// HashPassword превращает обычный пароль в зашифрованную строку
func HashPassword(password string) (string, error) {
	// 14 - это стоимость хеширования (чем больше, тем надежнее, но медленнее)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash сравнивает введенный пароль с зашифрованным из базы
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken создает JWT токен для пользователя
func GenerateToken(userID int) (string, error) {
	// claims - это "полезная нагрузка" токена, то, что мы в него зашиваем
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Токен живет 24 часа
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен нашим секретным ключом
	return token.SignedString(jwtSecret)
}

// ValidateToken проверяет токен и возвращает ID пользователя, если он настоящий
func ValidateToken(tokenStr string) (int, error) {
	// Разбираем токен
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что метод подписи правильный
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	// Достаем user_id из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}

	// В JWT числа хранятся как float64, поэтому приводим к int
	userID := int(claims["user_id"].(float64))
	return userID, nil
}
