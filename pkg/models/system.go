package models

import (
	"time"
)

type ErrorResponse struct {
	StatusCode int                       `json:"statusCode"`
	Message    string                    `json:"message"`
	Errors     []ValidationErrorResponse `json:"errors"`
}

type ValidationErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// SimpleErrorResponse представляет структуру простого JSON ответа об ошибке
type SimpleErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// PaginationResponse представляет структуру ответа с пагинацией
type PaginationResponse struct {
	Pagination Pagination    `json:"pagination"` // Информация о пагинации
	Filter     interface{}   `json:"filter"`     // Параметры фильтрации
	Data       []interface{} `json:"data"`       // Данные
}

// Pagination содержит информацию о пагинации
type Pagination struct {
	Page       int `json:"page"`     // Текущая страница
	PerPage    int `json:"per_page"` // Количество элементов на странице
	TotalPages int `json:"pages"`    // Общее количество страниц
	TotalItems int `json:"total"`    // Общее количество элементов
}

// Кастомный тип для времени
type CustomTime struct {
	time.Time
}

// Формат времени, который используется в JSON
const layout = "2006-01-02 15:04:05"

// Реализуем интерфейс json.Unmarshaler для CustomTime
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	// Убираем кавычки из JSON-строки
	str := string(b)
	str = str[1 : len(str)-1]

	// Парсим время
	t, err := time.Parse(layout, str)
	if err != nil {
		return err
	}

	ct.Time = t
	return nil
}

// Реализуем интерфейс json.Marshaler для CustomTime
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ct.Time.Format(layout) + `"`), nil
}
