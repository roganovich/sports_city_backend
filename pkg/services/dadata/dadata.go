package dadata

import (
	"bytes"
	"encoding/json"
	"goland_api/pkg/models"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// DadataResponse represents the actual structure of Dadata API response
type DadataResponse struct {
	Suggestions []DadataSuggestion `json:"suggestions"`
}

// DadataSuggestion represents a single suggestion from Dadata API
type DadataSuggestion struct {
	Value string         `json:"value"`
	Data  SuggestionData `json:"data"`
}

// SuggestionData contains detailed data about the suggestion
type SuggestionData struct {
	GeoLat string `json:"geo_lat"`
	GeoLon string `json:"geo_lon"`
}

func Suggest(requestBody []byte) (models.AddressResponse, error) {
	var addressResponse models.AddressResponse
	var dadataResponse DadataResponse
	apiKey := os.Getenv("DADATA_API_KEY")
	apiURL := os.Getenv("DADATA_API_URL")

	// Создаем HTTP-запрос
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("Ошибка при создании запроса:", err)
		return addressResponse, err
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Token "+apiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Ошибка при выполнении запроса:", err)
		return addressResponse, err
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Ошибка при чтении ответа:", err)
		return addressResponse, err

	}

	// Парсим ответ в промежуточную структуру
	err = json.Unmarshal(body, &dadataResponse)
	if err != nil {
		log.Println("Ошибка при парсинге ответа:", err)
		return addressResponse, err
	}

	// Преобразуем данные в нашу структуру
	var suggestions []models.AddressSuggestion
	for _, suggestion := range dadataResponse.Suggestions {
		addressSuggestion := models.AddressSuggestion{
			Value: suggestion.Value,
			Geo: models.Geo{
				Lat: suggestion.Data.GeoLat,
				Lon: suggestion.Data.GeoLon,
			},
		}
		suggestions = append(suggestions, addressSuggestion)
	}

	addressResponse.Suggestions = suggestions

	return addressResponse, nil
}
