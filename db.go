package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // Импортируем драйвер SQLite
)

func initDB() *sql.DB {
	// Определяем путь к базе данных в корневой директории проекта
	dbFile := filepath.Join(".", "scheduler.db") // Используем текущую директорию как базу

	// Открываем базу данных
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	// Проверяем существование базы данных
	_, err = os.Stat(dbFile)
	if os.IsNotExist(err) {
		// База данных не существует, создаем таблицу
		log.Println("База данных не существует. Создание таблицы...")
		createTableSQL := `
        CREATE TABLE scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date TEXT NOT NULL,
            title VARCHAR(255) NOT NULL,
            comment VARCHAR(500),
            repeat VARCHAR(128)
        );
        CREATE INDEX idx_date ON scheduler(date);
        `
		_, err = db.Exec(createTableSQL)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Таблица создана успешно.")
	} else if err != nil {
		log.Fatal(err)
	}

	return db
}
