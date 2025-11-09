package handlers

import (
	"encoding/json"
	"goland_api/pkg/database"
	"goland_api/pkg/models"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// getOneMedia получает информацию о медиафайле по его имени
func getOneMedia(fileName string) (error, models.Media) {
	var media models.Media
	err := database.DB.QueryRow("SELECT * FROM medias WHERE name = $1", fileName).Scan(
		&media.ID,

		&media.Name,
		&media.Path,
		&media.Ext,
		&media.Size,
		&media.CreatedAt,
	)

	if err != nil {
		log.Println("Ошибка в getOneMedia", fileName, err.Error())
	}

	return err, media
}

// getMediaById получает информацию о медиафайле по его ID
func getMediaById(id int64) (error, models.Media) {
	var media models.Media
	err := database.DB.QueryRow("SELECT * FROM medias WHERE id = $1", id).Scan(
		&media.ID,
		&media.Name,
		&media.Path,
		&media.Ext,
		&media.Size,
		&media.CreatedAt,
	)

	if err != nil {
		log.Println("Ошибка в getMediaById", id, err.Error())
	}

	return err, media
}

// @Summary Загрузить медиафайл
// @Description Загрузка медиафайла
// @Tags Медиафайлы
// @Param file formData file true "Загруженный файл"
// @Success 200 {object} models.Media
// @Failure 400 {object} models.ErrorResponse
// @Failure 413 {object} models.ErrorResponse
// @Failure 415 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/media/preloader [post]
func Preloader() http.HandlerFunc {
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
			// Загрузка файла
			file, fileHeader, errFile := r.FormFile("file")
			if errFile != nil {
				log.Println("Не удалось прочитать файл")
				SendJSONError(w, http.StatusBadRequest, "Не удалось прочитать файл")
				return
			}
			defer file.Close()

			fileName := getRandomName()
			dstPath := filepath.Join("./public/upload/", fileName)

			f, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
			if err != nil {
				log.Println("Не удалось открыть файл")
				SendJSONError(w, http.StatusInternalServerError, "Не удалось открыть файл")
				return
			}
			defer f.Close()

			fileSize, err := io.Copy(f, file)
			if err != nil {
				log.Println("Не удалось скопировать файл")
				SendJSONError(w, http.StatusInternalServerError, "Не удалось скопировать файл")
				return
			}

			createdAt := time.Now()
			mimeType := getMIMEType(fileHeader.Filename)

			var media models.Media
			media.Name = fileName
			media.Path = dstPath
			media.Ext = mimeType
			media.Size = fileSize
			media.CreatedAt = createdAt

			errInsert := database.DB.QueryRow("INSERT INTO medias (name, path, ext, size) VALUES ($1, $2, $3, $4) RETURNING id", media.Name, media.Path, media.Ext, media.Size).Scan(&media.ID)
			if errInsert != nil {
				log.Println(errInsert)
			}

			json.NewEncoder(w).Encode(media)
			return
		}

		// Если метод не поддерживается
		w.Header().Set("Allow", "POST, OPTIONS")
		SendJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

func getRandomName() string {
	newUUID := uuid.New()

	return newUUID.String()
}

func getMIMEType(filename string) string {
	extFile := filepath.Ext(filename)
	extData := strings.Split(extFile, ".")
	ext := ""
	if len(extData) > 0 {
		ext = extData[1]
	} else {
		log.Println("Расширение файла не удалось получить:" + extFile)
	}

	return ext
}
