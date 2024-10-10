package main

import (
	"log"
	"net/http"
)

func main() {
	// Инициализируем базу данных
	db := initDB()
	defer db.Close()

	// Путь к директории с фронтенд-файлами
	webDir := "./web"

	// Устанавливаем обработчики
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", handleNextDate)
	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			// Обработка добавления новой задачи
			addTaskHandler(w, r, db)
		case http.MethodGet:
			// Обработка получения задачи по ID
			getTaskHandler(w, r, db)
		case http.MethodPut:
			// Обработка редактирования задачи
			editTaskHandler(w, r, db)
		case http.MethodDelete:
			// Обработка удаления задачи
			deleteTaskHandler(w, r, db)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		getTasksHandler(w, r, db)
	})

	// Обработчик для отметки выполнения задачи
	http.HandleFunc("/api/task/done", func(w http.ResponseWriter, r *http.Request) {
		markTaskDoneHandler(w, r, db)
	})

	// Определяем порт для сервера
	port := "7540"
	log.Printf("Сервер запущен на порту %s", port)

	// Запускаем сервер
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
