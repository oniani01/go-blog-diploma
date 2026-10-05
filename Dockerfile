# Используем официальный образ Go для сборки
FROM golang:1.26-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем весь код проекта
COPY . .

# Собираем приложение
RUN go build -o main ./cmd/main.go

# Финальный образ (меньше размером)
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем собранный бинарник из builder
COPY --from=builder /app/main .

# Копируем .env файл
COPY --from=builder /app/.env .

# Создаем папку для данных
RUN mkdir -p ./data

# Открываем порт 8080
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]