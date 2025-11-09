package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"goland_api/pkg/database"
	"goland_api/pkg/models"
	"goland_api/pkg/utils"
	"log"
	"net/http"
	"strings"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// GetFields возвращает функцию-обработчик, которая получает все площадки из базы данных.
// Выполняет запрос к базе данных для получения всех площадок и возвращает их в виде JSON-ответа.
//
// @Summary Получить все площадки
// @Description Получить список всех площадок
// @Tags Площадки
// @Accept application/json
// @Produces application/json
// @Success 200 {object} []models.FieldView
// @Failure 400 Bad Request
// @Failure 500 Internal Server Error
// @Router /api/fields [get]
func GetFields() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := database.DB.Query("SELECT id, name, slug, description, city, address, logo, media, status, created_at  FROM fields")
		if err != nil {
			log.Println("Ошибка в SQL запросе GetFields", err)
		}
		defer rows.Close()
		fields := []models.FieldView{}
		for rows.Next() {
			var fieldView models.FieldView
			var logo sql.NullString
			var media sql.NullString

			if err := rows.Scan(
				&fieldView.ID,
				&fieldView.Name,
				&fieldView.Slug,
				&fieldView.Description,
				&fieldView.City,
				&fieldView.Address,
				&logo,
				&media,
				&fieldView.Status,
				&fieldView.CreatedAt,
			); err != nil {
				log.Println("Ошибка в Scan", err)
			}

			if logo.Valid {
				if logo.String != "" {
					var logoFile models.Media
					errorMedia, logoFile := getOneMedia(logo.String)
					if errorMedia != nil {
						log.Println("Ошибка в Logo", logo.String, errorMedia.Error())
					} else {
						fieldView.Logo = &logoFile
					}
				}
			}
			if media.Valid && len(media.String) > 0 {
				var mediaList []models.Media
				var mediaFiles []string
				err := json.Unmarshal([]byte(media.String), &mediaFiles)
				if err != nil {
					log.Println("Ошибка при парсинге JSON:", err)
				}
				for _, mediaFile := range mediaFiles {
					if mediaFile != "" {
						errorMedia, mediaFile := getOneMedia(mediaFile)
						if errorMedia != nil {
							log.Println("Ошибка в Media", mediaFile, errorMedia.Error())
						} else {
							mediaList = append(mediaList, mediaFile)
						}
					}
				}
				fieldView.Media = &mediaList
			}
			fields = append(fields, fieldView)
		}
		if err := rows.Err(); err != nil {
			log.Println("Ошибка в Row Next", err)
		}

		json.NewEncoder(w).Encode(fields)
	}
}

// getOneFieldById получает площадку из базы данных по её ID.
// Выполняет запрос к базе данных для получения площадки с указанным ID и возвращает данные площадки вместе с любой возникшей ошибкой.

func getOneFieldById(paramId int64) (error, models.FieldView) {
	var fieldView models.FieldView
	var logo sql.NullString
	var media sql.NullString

	err := database.DB.QueryRow("SELECT id, name, slug, description, city, address, logo, media, status, created_at FROM fields WHERE id = $1", int64(paramId)).Scan(
		&fieldView.ID,
		&fieldView.Name,
		&fieldView.Slug,
		&fieldView.Description,
		&fieldView.City,
		&fieldView.Address,
		&logo,
		&media,
		&fieldView.Status,
		&fieldView.CreatedAt,
	)
	if err != nil {
		return err, fieldView
	}

	if logo.Valid {
		var logoFile models.Media
		errorMedia, logoFile := getOneMedia(logo.String)
		if errorMedia != nil {
			log.Println(errorMedia.Error())
		} else {
			fieldView.Logo = &logoFile
		}
	}
	if media.Valid && len(media.String) > 0 {
		var mediaList []models.Media
		var mediaFiles []string
		err := json.Unmarshal([]byte(media.String), &mediaFiles)
		if err != nil {
			log.Println("Ошибка при парсинге JSON:", err)
		}
		for _, mediaFile := range mediaFiles {
			errorMedia, mediaFile := getOneMedia(mediaFile)
			if errorMedia != nil {
				log.Println(errorMedia.Error())
			} else {
				mediaList = append(mediaList, mediaFile)
			}
		}
		fieldView.Media = &mediaList
	}

	return err, fieldView
}

// getOneFieldBySlug получает площадку из базы данных по её slug.
// Выполняет запрос к базе данных для получения площадки с указанным slug и возвращает данные площадки вместе с любой возникшей ошибкой.
func getOneFieldBySlug(slug string) (error, models.FieldView) {
	var fieldView models.FieldView
	var logo sql.NullString
	var media sql.NullString

	err := database.DB.QueryRow("SELECT id, name, slug, description, city, address, logo, media, status, created_at FROM fields WHERE slug = $1", slug).Scan(
		&fieldView.ID,
		&fieldView.Name,
		&fieldView.Slug,
		&fieldView.Description,
		&fieldView.City,
		&fieldView.Address,
		&logo,
		&media,
		&fieldView.Status,
		&fieldView.CreatedAt,
	)
	if err != nil {
		return err, fieldView
	}

	if logo.Valid {
		var logoFile models.Media
		errorMedia, logoFile := getOneMedia(logo.String)
		if errorMedia != nil {
			log.Println(errorMedia.Error())
		} else {
			fieldView.Logo = &logoFile
		}
	}
	if media.Valid && len(media.String) > 0 {
		var mediaList []models.Media
		var mediaFiles []string
		err := json.Unmarshal([]byte(media.String), &mediaFiles)
		if err != nil {
			log.Println("Ошибка при парсинге JSON:", err)
		}
		for _, mediaFile := range mediaFiles {
			errorMedia, mediaFile := getOneMedia(mediaFile)
			if errorMedia != nil {
				log.Println(errorMedia.Error())
			} else {
				mediaList = append(mediaList, mediaFile)
			}
		}
		fieldView.Media = &mediaList
	}

	return err, fieldView
}

// GetField возвращает функцию-обработчик, которая получает конкретную площадку по её slug.
// Извлекает slug площадки из параметров URL и возвращает данные площадки в виде JSON-ответа.
//
// @Summary Получить площадку по slug
// @Description Получить информацию о площадке по её slug
// @Tags Площадки
// @Param slug path string true "Slug площадки"
// @Success 200 {object} models.FieldView
// @Failure 400 Bad Request
// @Failure 404 Not Found
// @Router /api/fields/{slug} [get]
func GetField() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		slug := vars["slug"]

		errorResponse, fieldView := getOneFieldBySlug(slug)
		if errorResponse != nil {
			SendJSONError(w, http.StatusBadRequest, errorResponse.Error())
			return
		}

		json.NewEncoder(w).Encode(fieldView)
	}
}

// validateCreateFieldRequest проверяет данные для создания новой площадки.
// Декодирует тело JSON-запроса и проверяет данные площадки с помощью структурной валидации.
// Также проверяет наличие дубликатов площадок по имени и городу.
func validateCreateFieldRequest(r *http.Request) (error, models.CreateFieldRequest) {
	var req models.CreateFieldRequest
	if validation := json.NewDecoder(r.Body).Decode(&req); validation != nil {
		return validation, req
	}
	validate := validator.New()
	if validation := validate.Struct(req); validation != nil {
		return validation, req
	}

	// Проверка на дубли по полям name и city
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM fields WHERE name = $1 AND city = $2", req.Name, req.City).Scan(&count)
	if err != nil {
		return err, req
	}
	if count > 0 {
		validation := fmt.Errorf("Площадка с именем '%s' из города '%s' уже существует", req.Name, req.City)
		return validation, req
	}

	// Generate slug if not provided
	slug := req.Slug
	if slug == "" {
		slug = utils.GenerateSlug(req.Name)
	}

	// Проверка на дубли по полю slug
	err = database.DB.QueryRow("SELECT COUNT(*) FROM fields WHERE slug = $1", slug).Scan(&count)
	if err != nil {
		return err, req
	}
	if count > 0 {
		validation := fmt.Errorf("Площадка с slug '%s' уже существует", slug)
		return validation, req
	}

	return nil, req
}

// processFieldBinaryData обрабатывает бинарные данные для логотипа и медиафайлов
// и возвращает пути к файлам
func processFieldBinaryDataCreate(req *models.CreateFieldRequest) (string, string, error) {
	var logoPath string
	var mediaPaths []string

	if req.Logo != nil && len(*req.Logo) > 0 {
		// Process logo data from Logo field (base64 string)
		var logoBytes []byte
		var err error

		// Check if it's a base64 data URL
		if strings.HasPrefix(*req.Logo, "data:") {
			// Extract base64 data from data URL
			parts := strings.SplitN(*req.Logo, ",", 2)
			if len(parts) == 2 {
				logoBytes, err = base64.StdEncoding.DecodeString(parts[1])
				if err != nil {
					return "", "", fmt.Errorf("failed to decode base64 logo data: %v", err)
				}
			} else {
				return "", "", fmt.Errorf("invalid base64 logo data format")
			}
		} else {
			// Try to decode as base64 string directly
			logoBytes, err = base64.StdEncoding.DecodeString(*req.Logo)
			if err != nil {
				return "", "", fmt.Errorf("failed to decode base64 logo data: %v", err)
			}
		}

		path, err := utils.SaveFileFromBytes(logoBytes, "logo")
		if err != nil {
			return "", "", fmt.Errorf("failed to save logo: %v", err)
		}
		logoPath = path
	}

	// Process media array data
	if req.Media != nil && len(*req.Media) > 0 {
		// Parse the JSON array of base64 strings
		var mediaArray []string
		if err := json.Unmarshal(*req.Media, &mediaArray); err != nil {
			return "", "", fmt.Errorf("failed to parse media array: %v", err)
		}

		// Process each media item
		for i, mediaData := range mediaArray {
			if mediaData != "" {
				// Process media data from Media field (base64 string)
				var mediaBytes []byte
				var err error

				// Check if it's a base64 data URL
				if strings.HasPrefix(mediaData, "data:") {
					// Extract base64 data from data URL
					parts := strings.SplitN(mediaData, ",", 2)
					if len(parts) == 2 {
						mediaBytes, err = base64.StdEncoding.DecodeString(parts[1])
						if err != nil {
							return "", "", fmt.Errorf("failed to decode base64 media data for item %d: %v", i, err)
						}
					} else {
						return "", "", fmt.Errorf("invalid base64 media data format for item %d", i)
					}
				} else {
					// Try to decode as base64 string directly
					mediaBytes, err = base64.StdEncoding.DecodeString(mediaData)
					if err != nil {
						return "", "", fmt.Errorf("failed to decode base64 media data for item %d: %v", i, err)
					}
				}

				filename := fmt.Sprintf("media_%d", i)
				path, err := utils.SaveFileFromBytes(mediaBytes, filename)
				if err != nil {
					return "", "", fmt.Errorf("failed to save media item %d: %v", i, err)
				}
				mediaPaths = append(mediaPaths, path)
			}
		}
	}

	// Convert media paths to JSON array
	var mediaJSON string
	if len(mediaPaths) > 0 {
		mediaBytes, err := json.Marshal(mediaPaths)
		if err != nil {
			return "", "", fmt.Errorf("failed to marshal media paths: %v", err)
		}
		mediaJSON = string(mediaBytes)
	}

	return logoPath, mediaJSON, nil
}

// processFieldBinaryDataUpdate обрабатывает бинарные данные для логотипа и медиафайлов
// и возвращает пути к файлам для запросов на обновление
func processFieldBinaryDataUpdate(req *models.UpdateFieldRequest) (string, string, error) {
	var logoPath string
	var mediaPaths []string

	if req.Logo != nil && len(*req.Logo) > 0 {
		// Process logo data from Logo field (base64 string)
		var logoBytes []byte
		var err error

		// Check if it's a base64 data URL
		if strings.HasPrefix(*req.Logo, "data:") {
			// Extract base64 data from data URL
			parts := strings.SplitN(*req.Logo, ",", 2)
			if len(parts) == 2 {
				logoBytes, err = base64.StdEncoding.DecodeString(parts[1])
				if err != nil {
					return "", "", fmt.Errorf("failed to decode base64 logo data: %v", err)
				}
			} else {
				return "", "", fmt.Errorf("invalid base64 logo data format")
			}
		} else {
			// Try to decode as base64 string directly
			logoBytes, err = base64.StdEncoding.DecodeString(*req.Logo)
			if err != nil {
				return "", "", fmt.Errorf("failed to decode base64 logo data: %v", err)
			}
		}

		path, err := utils.SaveFileFromBytes(logoBytes, "logo")
		if err != nil {
			return "", "", fmt.Errorf("failed to save logo: %v", err)
		}
		logoPath = path
	}

	// Process media array data
	if req.Media != nil && len(*req.Media) > 0 {
		// Parse the JSON array of base64 strings
		var mediaArray []string
		if err := json.Unmarshal(*req.Media, &mediaArray); err != nil {
			return "", "", fmt.Errorf("failed to parse media array: %v", err)
		}

		// Process each media item
		for i, mediaData := range mediaArray {
			if mediaData != "" {
				// Process media data from Media field (base64 string)
				var mediaBytes []byte
				var err error

				// Check if it's a base64 data URL
				if strings.HasPrefix(mediaData, "data:") {
					// Extract base64 data from data URL
					parts := strings.SplitN(mediaData, ",", 2)
					if len(parts) == 2 {
						mediaBytes, err = base64.StdEncoding.DecodeString(parts[1])
						if err != nil {
							return "", "", fmt.Errorf("failed to decode base64 media data for item %d: %v", i, err)
						}
					} else {
						return "", "", fmt.Errorf("invalid base64 media data format for item %d", i)
					}
				} else {
					// Try to decode as base64 string directly
					mediaBytes, err = base64.StdEncoding.DecodeString(mediaData)
					if err != nil {
						return "", "", fmt.Errorf("failed to decode base64 media data for item %d: %v", i, err)
					}
				}

				filename := fmt.Sprintf("media_%d", i)
				path, err := utils.SaveFileFromBytes(mediaBytes, filename)
				if err != nil {
					return "", "", fmt.Errorf("failed to save media item %d: %v", i, err)
				}
				mediaPaths = append(mediaPaths, path)
			}
		}
	}

	// Convert media paths to JSON array
	var mediaJSON string
	if len(mediaPaths) > 0 {
		mediaBytes, err := json.Marshal(mediaPaths)
		if err != nil {
			return "", "", fmt.Errorf("failed to marshal media paths: %v", err)
		}
		mediaJSON = string(mediaBytes)
	}

	return logoPath, mediaJSON, nil
}

// validateUpdatedAtFieldRequest проверяет данные для обновления площадки.
// Декодирует тело JSON-запроса и проверяет данные обновления площадки с помощью структурной валидации.
func validateUpdatedAtFieldRequest(r *http.Request) (error, models.UpdateFieldRequest) {
	var req models.UpdateFieldRequest
	if validation := json.NewDecoder(r.Body).Decode(&req); validation != nil {
		return validation, req
	}
	validate := validator.New()
	if validation := validate.Struct(req); validation != nil {
		return validation, req
	}

	return nil, req
}

// CreateField возвращает функцию-обработчик, которая создает новую площадку.
// Проверяет данные запроса, вставляет новую площадку в базу данных и возвращает созданную площадку.
//
// @Summary Создать новую площадку
// @Description Создать новую спортивную площадку с предоставленными данными
// @Tags Площадки
// @Param createField body models.CreateFieldRequest true "Данные для создания новой площадки"
// @Consumes application/json
// @Produces application/json
// @Success 201 {object} models.FieldView
// @Failure 422 Unprocessable Entity
// @Router /api/fields [post]
func CreateField() http.HandlerFunc {
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
			validation, fieldRequest := validateCreateFieldRequest(r)
			if validation != nil {
				SendJSONError(w, http.StatusBadRequest, validation.Error())
				return
			}

			// Instead of processing binary data, we now directly use logo ID and media IDs
			// logo will contain an ID from the medias table
			// media will contain an array of IDs from the medias table
			var field models.Field
			field.Name = fieldRequest.Name
			// Generate slug from name if not provided
			if fieldRequest.Slug == "" {
				field.Slug = utils.GenerateSlug(fieldRequest.City, fieldRequest.Name)
			} else {
				field.Slug = fieldRequest.Slug
			}
			field.Description = fieldRequest.Description
			field.City = fieldRequest.City
			field.Address = fieldRequest.Address

			// Use logo ID directly
			field.Logo = fieldRequest.Logo

			// Use media IDs directly
			field.Media = fieldRequest.Media

			var err error
			err = database.DB.QueryRow("INSERT INTO fields (name, slug, description, city, address, logo, media) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
				field.Name,
				field.Slug,
				field.Description,
				field.City,
				field.Address,
				field.Logo,
				field.Media,
			).Scan(&field.ID)
			if err != nil {
				log.Println(err)
				SendJSONError(w, http.StatusInternalServerError, "Failed to create field")
				return
			}

			errField, fieldView := getOneFieldById(int64(field.ID))
			if errField != nil {
				SendJSONError(w, http.StatusBadRequest, errField.Error())
				return
			}
			json.NewEncoder(w).Encode(fieldView)
			return
		}

		// Если метод не поддерживается
		w.Header().Set("Allow", "POST, OPTIONS")
		SendJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

// UpdateField возвращает функцию-обработчик, которая обновляет существующую площадку.
// Проверяет данные запроса, обновляет площадку в базе данных и возвращает обновленную площадку.
//
// @Summary Обновить существующую площадку
// @Description Обновить существующую спортивную площадку с предоставленными данными
// @Tags Площадки
// @Param updateField body models.UpdateFieldRequest true "Данные для обновления площадки"
// @Consumes application/json
// @Produces application/json
// @Param slug path string true "Slug площадки"
// @Success 200 {object} models.FieldView
// @Failure 422 Unprocessable Entity
// @Failure 404 Not Found
// @Router /api/fields/{slug} [put]
func UpdateField() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			// Устанавливаем заголовки для CORS
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Отправляем успешный ответ
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPut {
			validation, fieldRequest := validateUpdatedAtFieldRequest(r)
			if validation != nil {
				SendJSONError(w, http.StatusBadRequest, validation.Error())
				return
			}

			// Instead of processing binary data, we now directly use logo ID and media IDs
			// logo will contain an ID from the medias table
			// media will contain an array of IDs from the medias table
			var field models.Field
			field.Name = fieldRequest.Name
			// Generate slug from name if not provided
			if fieldRequest.Slug == "" {
				field.Slug = utils.GenerateSlug(fieldRequest.Name)
			} else {
				field.Slug = fieldRequest.Slug
			}
			field.Description = fieldRequest.Description
			field.City = fieldRequest.City
			field.Address = fieldRequest.Address

			// Use logo ID directly
			field.Logo = fieldRequest.Logo

			// Use media IDs directly
			field.Media = fieldRequest.Media

			vars := mux.Vars(r)
			slug := vars["slug"]
			errorResponse, fieldView := getOneFieldBySlug(slug)

			_, errUpdate := database.DB.Exec("UPDATE fields SET name = $1, slug = $2, description = $3, city = $4, address = $5, logo = $6, media = $7 WHERE id = $8",
				field.Name,
				field.Slug,
				field.Description,
				field.City,
				field.Address,
				field.Logo,
				field.Media,
				fieldView.ID)
			if errUpdate != nil {
				log.Println(errUpdate)
				SendJSONError(w, http.StatusBadRequest, errUpdate.Error())
				return
			}

			errorResponse, fieldView = getOneFieldBySlug(slug)
			if errorResponse != nil {
				SendJSONError(w, http.StatusBadRequest, errorResponse.Error())
				return
			}
			json.NewEncoder(w).Encode(fieldView)
			return
		}

		// Если метод не поддерживается
		w.Header().Set("Allow", "PUT, OPTIONS")
		SendJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

// DeleteField возвращает функцию-обработчик, которая удаляет площадку по её ID.
// Удаляет площадку из базы данных и возвращает сообщение об успешном удалении.
//
// @Summary Удалить площадку по ID
// @Description Удалить спортивную площадку по её идентификатору
// @Tags Площадки
// @Param slug path string true "Slug площадки"
// @Success 200 {string} string "Площадка удалена"
// @Failure 404 Not Found
// @Router /api/fields/{slug} [delete]
func DeleteField() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			// Устанавливаем заголовки для CORS
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Отправляем успешный ответ
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodDelete {

			vars := mux.Vars(r)
			slug := vars["slug"]
			errorResponse, fieldView := getOneFieldBySlug(slug)

			if errorResponse != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			} else {
				_, err := database.DB.Exec("DELETE FROM fields WHERE id = $1", fieldView.ID)
				if err != nil {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				json.NewEncoder(w).Encode("Field deleted")
			}
			return
		}

		// Если метод не поддерживается
		w.Header().Set("Allow", "DELETE, OPTIONS")
		SendJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}
