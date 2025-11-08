package handlers

import (
	"encoding/json"
	"goland_api/pkg/models"
	"goland_api/pkg/services/dadata"
	"log"
	"net/http"
)

// @Summary Получить подсказку по части адреса
// @Description Получить подсказку по части адреса
// @Tags Адреса
// @Accept  application/json
// @Produce  application/json
// @Param query query string true "Часть адреса для поиска"
// @Success 200 {object} models.AddressResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 413 {object} models.ErrorResponse
// @Failure 415 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/address/suggest [post]
func SuggestAddress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			// Устанавливаем заголовки для CORS
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Отправляем успешный ответ
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPost {
			// Получаем параметры из URL
			queryParams := r.URL.Query()

			// Извлекаем значение параметра "query"
			query := queryParams.Get("query")

			// Проверяем, был ли передан параметр
			if query == "" {
				log.Println("Параметр 'query' не указан")
				return
			}
			// Создаем запрос с адресом для поиска
			addressRequest := models.AddressRequest{
				Query: query,
			}

			// Преобразуем запрос в JSON
			requestBody, err := json.Marshal(addressRequest)
			if err != nil {
				log.Println("Ошибка при чтении запроса:", err)
				return
			}

			addressResponse, err := dadata.Suggest(requestBody)
			if err != nil {
				log.Println("Ошибка при обращении к сервису:", err)
				return
			}
			json.NewEncoder(w).Encode(addressResponse.Suggestions)
			return
		}

		// Если метод не поддерживается
		w.Header().Set("Allow", "POST, OPTIONS")
		SendJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}
