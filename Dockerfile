# Стадия сборки
FROM golang:1.21 AS builder

# Установка рабочей директории внутри контейнера
WORKDIR /app

# Копируем Go-файл
COPY usePost.go .

# Сборка бинарника
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app usePost.go


# Финальный контейнер
FROM alpine:latest

# Копируем собранный бинарник из builder
COPY --from=builder /app/app /app/app

# Устанавливаем рабочую директорию
WORKDIR /app

# Указываем команду запуска
ENTRYPOINT ["./app"]
