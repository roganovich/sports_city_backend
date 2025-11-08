package handlers

import (
	"encoding/json"
	"goland_api/pkg/database"
	"goland_api/pkg/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Документация для метода GetRentals
// @Summary Возвращает список всех аренд
// @Description Получение списка всех аренд
// @Tags Аренда
// @Accept application/json
// @Produces application/json
// @Success 200 {object} []models.RentalView
// @Failure 400 Bad Request
// @Failure 500 Internal Server Error
// @Router /api/rentals [get]
func GetRentals() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем параметры из GET-запроса
		queryParams := r.URL.Query()
		// Извлекаем значение параметра "page" (с значением по умолчанию)
		page := getIntParam(queryParams, "page", 1)
		// Извлекаем значение параметра "per_page" (с значением по умолчанию)
		perPage := getIntParam(queryParams, "per_page", 10)
		// Вычисляем OFFSET
		offset := (page - 1) * perPage

		var totalCount int
		err := database.DB.QueryRow("SELECT COUNT(id) FROM rentals").Scan(&totalCount)
		if err != nil {
			log.Fatal(err)
		}

		var pages int
		pages = int(math.Ceil(float64(totalCount) / float64(perPage)))

		query := `
		SELECT id, field_id, team_id, user_id, comment, start_date, end_date, duration, status, created_at 
		FROM rentals
		ORDER BY id
		LIMIT $1 OFFSET $2
	`
		// Выполняем запрос
		rows, errQuery := database.DB.Query(query, perPage, offset)
		if errQuery != nil {
			log.Println(errQuery)
		}
		defer rows.Close()
		// Преобразуем []RentalView в []interface{}
		var rentalsInterface []interface{}
		for rows.Next() {
			var rentalView models.RentalView
			var fieldId int64
			var teamId int64
			var userId int64

			if err := rows.Scan(
				&rentalView.ID,
				&fieldId,
				&teamId,
				&userId,
				&rentalView.Comment,
				&rentalView.StartDate,
				&rentalView.EndDate,
				&rentalView.Duration,
				&rentalView.Status,
				&rentalView.CreatedAt,
			); err != nil {
				log.Println(err)
			}
			if fieldId != 0 {
				errorField, fieldView := getOneFieldById(int64(fieldId))
				if errorField != nil {
					log.Println(errorField.Error())
				} else {
					rentalView.Field = fieldView
				}
			}
			if teamId != 0 {
				errorTeam, teamView := getOneTeamById(int64(teamId))
				if errorTeam != nil {
					log.Println(errorTeam.Error())
				} else {
					rentalView.Team = teamView
				}
			}
			if userId != 0 {
				errorUser, userView := getUserViewById(int64(userId))
				if errorUser != nil {
					log.Println(errorUser.Error())
				} else {
					rentalView.User = userView
				}
			}
			rentalsInterface = append(rentalsInterface, rentalView)
			//rentals = append(rentals, rentalView)
		}
		if err := rows.Err(); err != nil {
			log.Println(err)
		}

		var fiter models.RentalSearchFilter
		validationFilterJsonError := json.NewDecoder(r.Body).Decode(&fiter)
		if validationFilterJsonError != nil {
			log.Println(validationFilterJsonError.Error())
		}
		validate := validator.New()
		validationFilterError := validate.Struct(fiter)
		if validationFilterError != nil {
			log.Println(validationFilterError.Error())
		}

		response := models.PaginationResponse{
			Pagination: models.Pagination{
				Page:       page,
				PerPage:    perPage,
				TotalPages: pages,
				TotalItems: totalCount,
			},
			Filter: fiter,
			Data:   rentalsInterface,
		}

		json.NewEncoder(w).Encode(response)
	}
}

func getOneRentalById(paramId int64) (error, models.RentalView) {
	var rentalView models.RentalView
	var fieldId int64
	var teamId int64
	var userId int64

	err := database.DB.QueryRow("SELECT id, field_id, team_id, user_id, comment, start_date, end_date, duration, status, created_at FROM rentals WHERE id = $1", int64(paramId)).Scan(
		&rentalView.ID,
		&fieldId,
		&teamId,
		&userId,
		&rentalView.Comment,
		&rentalView.StartDate,
		&rentalView.EndDate,
		&rentalView.Duration,
		&rentalView.Status,
		&rentalView.CreatedAt,
	)
	if err != nil {
		return err, rentalView
	}
	if fieldId != 0 {
		errorField, fieldView := getOneFieldById(int64(fieldId))
		if errorField != nil {
			log.Println(errorField.Error())
		} else {
			rentalView.Field = fieldView
		}
	}
	if teamId != 0 {
		errorTeam, teamView := getOneTeamById(int64(teamId))
		if errorTeam != nil {
			log.Println(errorTeam.Error())
		} else {
			rentalView.Team = teamView
		}
	}
	if userId != 0 {
		errorUser, userView := getUserViewById(int64(userId))
		if errorUser != nil {
			log.Println(errorUser.Error())
		} else {
			rentalView.User = userView
		}
	}

	return err, rentalView
}

// Документация для метода GetRental
// @Summary Возвращает информацию об аренде по ID
// @Description Получение информации об аренде по идентификатору
// @Tags Аренда
// @Param id path int true "ID аренды"
// @Success 200 {object} models.RentalView
// @Failure 400 Bad Request
// @Failure 404 Not Found
// @Router /api/rentals/{id} [get]
func GetRental() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramId, _ := strconv.Atoi(vars["id"])

		errorResponse, rentalView := getOneRentalById(int64(paramId))
		if errorResponse != nil {
			SendJSONError(w, http.StatusBadRequest, errorResponse.Error())
			return
		}

		json.NewEncoder(w).Encode(rentalView)
	}
}

func validateCreateRentalRequest(r *http.Request) (error, models.CreateRentalRequest) {
	var req models.CreateRentalRequest
	if validation := json.NewDecoder(r.Body).Decode(&req); validation != nil {
		return validation, req
	}
	validate := validator.New()
	if validation := validate.Struct(req); validation != nil {
		return validation, req
	}

	return nil, req
}

// Документация для метода CreateRental
// @Summary Создание новой аренды
// @Description Создание новой аренды
// @Tags Аренда
// @Param createRental body models.CreateRentalRequest true "Данные для создания новой аренды"
// @Consumes application/json
// @Produces application/json
// @Success 201 {object} models.RentalView
// @Failure 422 Unprocessable Entity
// @Router /api/rentals [post]
func CreateRental() http.HandlerFunc {
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
			validation, rentalRequest := validateCreateRentalRequest(r)
			if validation != nil {
				SendJSONError(w, http.StatusBadRequest, validation.Error())
				return
			}

			var rental models.Rental
			var status int

			// Указываем формат
			layout := "2006-01-02 15:04:05"
			startDate, errTime := time.Parse(layout, rentalRequest.StartDate)
			if errTime != nil {
				SendJSONError(w, http.StatusBadRequest, errTime.Error())
				return
			}
			rental.StartDate = startDate

			endDate, errTime := time.Parse(layout, rentalRequest.EndDate)
			if errTime != nil {
				SendJSONError(w, http.StatusBadRequest, errTime.Error())
				return
			}
			rental.EndDate = endDate

			rental.FieldID = rentalRequest.FieldID
			rental.TeamID = rentalRequest.TeamID
			rental.Comment = rentalRequest.Comment

			// Вычисляем разницу между двумя временами
			duration := rental.EndDate.Sub(rental.StartDate)
			// Преобразуем разницу в секунды
			rental.Duration = int(duration.Seconds())

			err := database.DB.QueryRow("INSERT INTO rentals (field_id, team_id, user_id, comment, start_date, end_date, duration, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
				rental.FieldID,
				rental.TeamID,
				AUTH.ID,
				rental.Comment,
				rental.StartDate,
				rental.EndDate,
				rental.Duration,
				&status,
			).Scan(&rental.ID)
			if err != nil {
				log.Println(err)
			}

			errrental, rentalView := getOneRentalById(int64(rental.ID))
			if errrental != nil {
				SendJSONError(w, http.StatusBadRequest, errrental.Error())
				return
			}
			json.NewEncoder(w).Encode(rentalView)
			return
		}

		// Если метод не поддерживается
		w.Header().Set("Allow", "POST, OPTIONS")
		SendJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

// Документация для метода Deleterental
// @Summary Удаляет аренду по ID
// @Description Удаление аренды по идентификатору
// @Tags Аренда
// @Param id path int true "ID аренды"
// @Success 204 No Content
// @Failure 404 Not Found
// @Router /api/rentals/{id} [delete]
func DeleteRental() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramId, _ := strconv.Atoi(vars["id"])
		errorResponse, rentalView := getOneRentalById(int64(paramId))
		if errorResponse != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			_, err := database.DB.Exec("DELETE FROM rentals WHERE id = $1 and user_id = $2", rentalView.ID, AUTH.ID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			json.NewEncoder(w).Encode("rental deleted")
		}
	}
}
