# Шаг 1: Используем базовый образ Golang для компиляции программы
FROM golang:1.20 AS build

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем все файлы проекта в контейнер
COPY . .

# Загружаем зависимости
RUN go mod download

# Собираем исполняемый файл
RUN go build -o todo_server .

# Шаг 2: Используем минимальный образ для запуска приложения
FROM alpine:latest

# Устанавливаем необходимые пакеты, включая SQLite
RUN apk --no-cache add sqlite

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем исполняемый файл из стадии сборки
COPY --from=build /app/todo_server /app/todo_server

# Копируем директорию с фронтенд-файлами
COPY ./web /app/web

# Указываем порт, который будет слушать наше приложение
EXPOSE 7540

# Устанавливаем переменные окружения для конфигурации приложения
ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/scheduler.db

# Команда запуска приложения
CMD ["/app/todo_server"]
