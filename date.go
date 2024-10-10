package main

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату на основе текущего времени, начальной даты и правила повторения.
func NextDate(now time.Time, date string, repeat string) (string, error) {
	// Парсим начальную дату
	startDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", errors.New("некорректный формат начальной даты")
	}

	// Если начальная дата находится в прошлом и без повторения, возвращаем ошибку
	if startDate.Before(now) && repeat == "" {
		return "", errors.New("начальная дата в прошлом и правило повторения отсутствует")
	}

	// Обрабатываем правило повторения
	switch {
	case repeat == "":
		return "", errors.New("правило повторения не указано")
	case strings.HasPrefix(repeat, "d "):
		daysStr := strings.TrimPrefix(repeat, "d ")
		days, err := strconv.Atoi(daysStr)
		if err != nil || days < 1 || days > 400 {
			return "", errors.New("некорректное значение интервала дней")
		}
		return getNextDateByDays(startDate, now, days), nil
	case repeat == "y":
		return getNextDateByYears(startDate, now), nil
	default:
		return "", errors.New("неподдерживаемый формат")
	}
}

// getNextDateByDays возвращает следующую дату с учётом интервала дней
func getNextDateByDays(startDate, now time.Time, days int) string {
	for {
		// Добавляем дни к начальной дате
		startDate = startDate.AddDate(0, 0, days)

		// Если дата после текущей, возвращаем её
		if startDate.After(now) {
			return startDate.Format("20060102")
		}
	}
}

// getNextDateByYears возвращает следующую дату через год
func getNextDateByYears(startDate, now time.Time) string {
	for {
		// Добавляем год
		startDate = startDate.AddDate(1, 0, 0)
		if startDate.After(now) {
			return startDate.Format("20060102")
		}
	}
}
