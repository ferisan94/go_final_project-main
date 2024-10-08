package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Task представляет структуру задачи
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// handleNextDate обрабатывает запросы к /api/nextdate
func handleNextDate(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	// Парсим параметр now
	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "некорректный формат текущей даты", http.StatusBadRequest)
		return
	}

	// Вызываем функцию NextDate
	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Отправляем результат
	w.Write([]byte(nextDate))
}

// addTaskHandler обрабатывает HTTP-запросы для добавления новой задачи в базу данных
func addTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	// Проверяем, что заголовок задачи указан
	if task.Title == "" {
		http.Error(w, `{"error":"Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	// Получаем сегодняшнюю дату
	now := time.Now().Format("20060102")
	if task.Date == "" {
		task.Date = now // Если дата не указана, используем сегодняшнюю
	} else {
		// Проверяем формат даты
		_, err := time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, `{"error":"Неправильный формат даты. Ожидался формат 20060102"}`, http.StatusBadRequest)
			return
		}
	}

	// Если задача в прошлом и правило повторения не указано, заменяем дату на сегодняшнюю
	if task.Date < now && task.Repeat == "" {
		task.Date = now
	}

	// Если дата в будущем, не меняем её, даже если есть правило повторения
	if task.Date >= now {
		// Оставляем дату без изменений
	} else {
		// Если правило повторения указано, вычисляем следующую подходящую дату
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
			return
		}
		task.Date = nextDate
	}

	// Выполняем вставку задачи в базу данных
	res, err := db.Exec(`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, "Ошибка при добавлении задачи", http.StatusInternalServerError)
		return
	}

	// Получаем ID вставленной записи
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Ошибка получения ID новой записи", http.StatusInternalServerError)
		return
	}

	// Возвращаем результат в формате JSON
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
}

// getTasksHandler обрабатывает GET-запросы для получения списка задач
func getTasksHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	limit := 50                           // Максимальное количество возвращаемых задач
	search := r.URL.Query().Get("search") // Поиск по заголовку или комментарию

	var rows *sql.Rows
	var err error

	// Формируем SQL-запрос
	if search != "" {
		searchPattern := "%" + search + "%"
		rows, err = db.Query(`SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?`, searchPattern, searchPattern, limit)
	} else {
		rows, err = db.Query(`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?`, limit)
	}

	if err != nil {
		http.Error(w, `{"error": "Ошибка при получении задач"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := []Task{}

	// Считываем задачи из результата
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			http.Error(w, `{"error": "Ошибка при чтении задач"}`, http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	// Формируем ответ в формате JSON
	response := map[string]interface{}{"tasks": tasks}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK) // Устанавливаем статус 200
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error": "Ошибка при формировании ответа"}`, http.StatusInternalServerError)
	}
}

// getTaskHandler возвращает задачу по ID
func getTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error": "Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	var task Task
	err := db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id).
		Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error": "Ошибка при получении задачи"}`, http.StatusInternalServerError)
		}
		return
	}

	// Возвращаем задачу в формате JSON
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(task)
}

// editTaskHandler редактирует задачу
func editTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPut {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	// Проверка, что идентификатор задачи указан
	if task.ID == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	// Проверка, что заголовок задачи указан
	if task.Title == "" {
		http.Error(w, `{"error":"Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	// Проверка на корректность формата даты
	if _, err := time.Parse("20060102", task.Date); err != nil {
		http.Error(w, `{"error":"Неправильный формат даты. Ожидался формат 20060102"}`, http.StatusBadRequest)
		return
	}

	// Проверка корректности поля repeat (если оно не должно содержать произвольные значения)
	if task.Repeat != "" {
		allowedRepeats := []string{"d 1", "d 5", "d 7"}
		valid := false
		for _, v := range allowedRepeats {
			if task.Repeat == v {
				valid = true
				break
			}
		}
		if !valid {
			http.Error(w, `{"error":"Некорректное значение для repeat"}`, http.StatusBadRequest)
			return
		}
	}

	// Проверка существования задачи
	var exists int
	err = db.QueryRow("SELECT COUNT(*) FROM scheduler WHERE id = ?", task.ID).Scan(&exists)
	if err != nil {
		http.Error(w, "Ошибка при проверке существования задачи", http.StatusInternalServerError)
		return
	}

	if exists == 0 {
		http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
		return
	}

	// Выполнение обновления задачи в базе данных
	_, err = db.Exec(`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`,
		task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		http.Error(w, "Ошибка при обновлении задачи", http.StatusInternalServerError)
		return
	}

	// Возвращаем пустой JSON при успешном обновлении
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

// markTaskDoneHandler отмечает задачу выполненной
func markTaskDoneHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	var task Task
	err := db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error": "Ошибка при получении задачи"}`, http.StatusInternalServerError)
		}
		return
	}

	// Если задача одноразовая (поле repeat пустое), удаляем её
	if task.Repeat == "" {
		_, err := db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
		if err != nil {
			http.Error(w, `{"error": "Ошибка при удалении задачи"}`, http.StatusInternalServerError)
			return
		}
	} else {
		// Для повторяющейся задачи рассчитываем следующую дату
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
			return
		}

		_, err = db.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, nextDate, id)
		if err != nil {
			http.Error(w, `{"error": "Ошибка при обновлении даты задачи"}`, http.StatusInternalServerError)
			return
		}
	}

	// Возвращаем пустой JSON при успешной отметке
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

// deleteTaskHandler удаляет задачу
func deleteTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Получаем идентификатор задачи из параметров запроса
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error": "Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	// Пытаемся удалить задачу по ID
	res, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error": "Ошибка при удалении задачи"}`, http.StatusInternalServerError)
		return
	}

	// Проверяем, что задача действительно была удалена
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, `{"error": "Ошибка при проверке удаления"}`, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
		return
	}

	// Возвращаем пустой JSON при успешном удалении
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}
