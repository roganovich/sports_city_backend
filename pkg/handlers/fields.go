package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"goland_api/pkg/database"
	"goland_api/pkg/models"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// generateSlug создает URL-дружественный slug из переменного количества строк
// Использование: generateSlug("строка1", "строка2", ...) или generateSlug(однаСтрока)
// Пример: generateSlug("Футбольное", "Поле") возвращает "futbolnoe-pole"
func generateSlug(texts ...string) string {
	// Concatenate all input strings with spaces
	concatenated := strings.Join(texts, " ")

	// Convert to lowercase
	slug := strings.ToLower(concatenated)

	// Transliterate common Cyrillic characters to Latin
	cyrillicToLatin := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e", 'ж': "zh", 'з': "z",
		'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r",
		'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "h", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "sch",
		'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
		'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D", 'Е': "E", 'Ё': "E", 'Ж': "Zh", 'З': "Z",
		'И': "I", 'Й': "Y", 'К': "K", 'Л': "L", 'М': "M", 'Н': "N", 'О': "O", 'П': "P", 'Р': "R",
		'С': "S", 'Т': "T", 'У': "U", 'Ф': "F", 'Х': "H", 'Ц': "Ts", 'Ч': "Ch", 'Ш': "Sh", 'Щ': "Sch",
		'Ъ': "", 'Ы': "Y", 'Ь': "", 'Э': "E", 'Ю': "Yu", 'Я': "Ya",
	}

	var transliterated strings.Builder
	for _, r := range slug {
		if replacement, exists := cyrillicToLatin[r]; exists {
			transliterated.WriteString(replacement)
		} else {
			transliterated.WriteRune(r)
		}
	}
	slug = transliterated.String()

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile("[^a-z0-9\\-]+")
	slug = reg.ReplaceAllString(slug, "_")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// Replace multiple consecutive hyphens with a single hyphen
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	return slug
}

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
		slug = generateSlug(req.Name)
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

			var field models.Field
			field.Name = fieldRequest.Name
			// Generate slug from name if not provided
			if fieldRequest.Slug == "" {
				field.Slug = generateSlug(fieldRequest.City, fieldRequest.Name)
			} else {
				field.Slug = fieldRequest.Slug
			}
			field.Description = fieldRequest.Description
			field.City = fieldRequest.City
			field.Address = fieldRequest.Address
			field.Logo = fieldRequest.Logo
			field.Media = fieldRequest.Media
			err := database.DB.QueryRow("INSERT INTO fields (name, slug, description, city, address, logo, media) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
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
			var field models.Field
			field.Name = fieldRequest.Name
			// Generate slug from name if not provided
			if fieldRequest.Slug == "" {
				field.Slug = generateSlug(fieldRequest.Name)
			} else {
				field.Slug = fieldRequest.Slug
			}
			field.Description = fieldRequest.Description
			field.City = fieldRequest.City
			field.Address = fieldRequest.Address
			field.Logo = fieldRequest.Logo
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
